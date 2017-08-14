package servicefoundation

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Prutswonder/go-servicefoundation/env"
	"github.com/Prutswonder/go-servicefoundation/logging"
	. "github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Prutswonder/go-servicefoundation/site"
)

const (
	envCORSOrigins  string = "CORS_ORIGINS"
	envHTTPpPort    string = "HTTPPORT"
	envLogMinFilter string = "LOG_MINFILTER"

	defaultHTTPPort     int    = 8080
	defaultLogMinFilter string = "Warning"

	publicSubsystem = "public"
)

type (
	serviceImpl struct {
		name            string
		port            int
		readinessPort   int
		internalPort    int
		log             Logger
		metrics         Metrics
		publicRouter    *Router
		readinessRouter *Router
		internalRouter  *Router
		handlerFactory  ServiceHandlerFactory
		versionBuilder  VersionBuilder
		exitFunc        ExitFunc
	}
)

func CreateService(name string, options ServiceOptions) Service {
	exitFunc := createExitFunc(options.Logger, options.ShutdownFunc)
	return &serviceImpl{
		name:            name,
		port:            options.Port,
		readinessPort:   options.ReadinessPort,
		internalPort:    options.InternalPort,
		log:             options.Logger,
		metrics:         options.Metrics,
		publicRouter:    options.RouterFactory.CreateRouter(),
		readinessRouter: options.RouterFactory.CreateRouter(),
		internalRouter:  options.RouterFactory.CreateRouter(),
		handlerFactory:  options.ServiceHandlerFactory,
		versionBuilder:  options.VersionBuilder,
		exitFunc:        exitFunc,
	}
}

// CreateDefaultService creates and returns a Service that uses environment variables for default configuration.
func CreateDefaultService(name string, allowedMethods []string, shutdownFunc ShutdownFunc) Service {
	corsOptions := CORSOptions{
		AllowedOrigins: env.ListOrDefault(envCORSOrigins, []string{"*"}),
		AllowedMethods: allowedMethods,
	}
	logger := logging.CreateLogger(env.OrDefault(envLogMinFilter, defaultLogMinFilter))
	metrics := logging.CreateMetrics(name, logger)
	middlewareWrapper := site.CreateMiddlewareWrapper(logger, metrics, &corsOptions)
	versionBuilder := site.CreateDefaultVersionBuilder()
	exitFunc := createExitFunc(logger, shutdownFunc)
	port := env.AsInt(envHTTPpPort, defaultHTTPPort)

	opt := ServiceOptions{
		Port:                  port,
		ReadinessPort:         port + 1,
		InternalPort:          port + 2,
		ServiceHandlerFactory: site.CreateServiceHandlerFactory(middlewareWrapper, versionBuilder, exitFunc),
		RouterFactory:         site.CreateRouterFactory(),
		Logger:                logger,
		Metrics:               metrics,
		VersionBuilder:        versionBuilder,
		ShutdownFunc:          shutdownFunc,
	}

	return CreateService(name, opt)
}

// overwrite the default os.exit() to run delayed, giving the quit handler a chance to return a status
func createExitFunc(log Logger, shutdownFunc ShutdownFunc) func(int) {
	return func(code int) {
		log.Debug("ServiceExit", "Performing service exit")
		go func(code int) {
			if shutdownFunc != nil {
				shutdownFunc(log)
			}

			if code != 0 {
				time.Sleep(500 * time.Millisecond)
			}

			log.Debug("ServiceExit", "Calling os.Exit(%v)", code)
			os.Exit(code)
		}(code)
	}
}

/* Service implementation */

func (s *serviceImpl) Run() {
	s.log.Info("Service", "%s: %s", s.name, s.versionBuilder.ToString())
	s.runReadinessServer()
	s.runInternalServer()
	s.runPublicServer() // blocks code execution
}

func (s *serviceImpl) AddRoute(name string, routes []string, methods []string, middlewares []Middleware, handler Handle) {
	s.addRoute(s.publicRouter, publicSubsystem, name, routes, methods, middlewares, handler)
}

func (s *serviceImpl) addRoute(router *Router, subsystem, name string, routes []string, methods []string, middlewares []Middleware, handler Handle) {
	for _, path := range routes {
		wrappedHandler := s.handlerFactory.WrapHandler(subsystem, name, middlewares, handler)

		for _, method := range methods {
			router.Router.Handle(method, path, wrappedHandler)
		}
	}
}

// RunReadinessServer runs the readiness service as a go-routine
func (s *serviceImpl) runReadinessServer() {
	const subsystem = "readiness"

	port := fmt.Sprintf(":%v", s.readinessPort)
	router := s.readinessRouter
	fact := s.handlerFactory

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, []Middleware{Histogram}, fact.CreateRootHandler())
	s.addRoute(router, subsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, []Middleware{Counter, NoCaching}, fact.CreateLivenessHandler())
	s.addRoute(router, subsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, []Middleware{Histogram, NoCaching}, fact.CreateReadinessHandler())

	s.log.Info("RunReadinessServer", "%s %s running on localhost%s.", s.name, subsystem, port)

	go http.ListenAndServe(port, router.Router)
}

// RunInternalServer runs the internal service as a go-routine
func (s *serviceImpl) runInternalServer() {
	const subsystem = "internal"

	port := fmt.Sprintf(":%v", s.internalPort)
	router := s.internalRouter
	fact := s.handlerFactory

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, []Middleware{Histogram}, fact.CreateRootHandler())
	s.addRoute(router, subsystem, "health_check", []string{"/health_check", "/healthz"}, MethodsForGet, []Middleware{Counter, NoCaching}, fact.CreateHealthHandler())
	s.addRoute(router, subsystem, "metrics", []string{"/metrics"}, MethodsForGet, []Middleware{Counter, NoCaching}, fact.CreateMetricsHandler())
	s.addRoute(router, subsystem, "quit", []string{"/quit"}, MethodsForGet, []Middleware{NoCaching}, fact.CreateQuitHandler())

	s.log.Info("RunInternalServer", "%s %s running on localhost%s.", s.name, subsystem, port)

	go http.ListenAndServe(port, router.Router)
}

// RunPublicServer runs the public service on the current thread.
func (s *serviceImpl) runPublicServer() {
	port := fmt.Sprintf(":%v", s.port)
	router := s.publicRouter
	fact := s.handlerFactory

	s.addRoute(router, publicSubsystem, "root", []string{"/"}, MethodsForGet, []Middleware{Histogram}, fact.CreateRootHandler())
	s.addRoute(router, publicSubsystem, "version", []string{"/service/version"}, MethodsForGet, []Middleware{Counter}, fact.CreateLivenessHandler())
	s.addRoute(router, publicSubsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, []Middleware{Counter}, fact.CreateLivenessHandler())
	s.addRoute(router, publicSubsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, []Middleware{Histogram, NoCaching}, fact.CreateReadinessHandler())

	s.log.Info("RunPublicService", "%s %s running on localhost%s.", s.name, publicSubsystem, port)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		s.log.Debug("GracefulShutdown", "Handling Sigterm/SigInt")
		s.exitFunc(0)
		done <- true
	}()

	go http.ListenAndServe(port, router.Router)

	<-done // Wait for our shutdown

	// since service.ExitFunc calls os.Exit(), we'll probably never get here
}

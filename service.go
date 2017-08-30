package servicefoundation

import (
	"context"
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
		timeout         time.Duration
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
		shutdownFunc    ShutdownFunc
		exitFunc        ExitFunc
		quitting        bool
		sendChan        chan bool
		receiveChan     chan bool
	}

	serverInstance struct {
		shutdownChan chan bool
	}
)

func CreateService(name string, options ServiceOptions) Service {
	return &serviceImpl{
		name:            name,
		timeout:         options.ServerTimeout,
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
		shutdownFunc:    options.ShutdownFunc,
		exitFunc:        options.ExitFunc,
		sendChan:        make(chan bool, 1),
		receiveChan:     make(chan bool, 1),
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
		ServerTimeout:         time.Second * 20,
		Port:                  port,
		ReadinessPort:         port + 1,
		InternalPort:          port + 2,
		ServiceHandlerFactory: site.CreateServiceHandlerFactory(middlewareWrapper, versionBuilder, exitFunc),
		RouterFactory:         site.CreateRouterFactory(),
		Logger:                logger,
		Metrics:               metrics,
		VersionBuilder:        versionBuilder,
		ShutdownFunc:          shutdownFunc,
		ExitFunc:              exitFunc,
	}

	return CreateService(name, opt)
}

// overwrite the default os.exit() to run delayed, giving the quit handler a chance to return a status
func createExitFunc(log Logger, shutdownFunc ShutdownFunc) func(int) {
	return func(code int) {
		log.Debug("ServiceExit", "Performing service exit")

		go func() {
			if shutdownFunc != nil {
				log.Debug("ShutdownFunc", "Calling shutdown func")
				shutdownFunc(log)
			}

			if code != 0 {
				time.Sleep(500 * time.Millisecond)
			}

			log.Debug("ServiceExit", "Calling os.Exit(%v)", code)
			os.Exit(code)
		}()

		// Allow the go-routine to be spawned
		time.Sleep(1 * time.Millisecond)
	}
}

/* Service implementation */

func (s *serviceImpl) Run(ctx context.Context) {
	s.log.Info("Service", "%s: %s", s.name, s.versionBuilder.ToString())

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-s.receiveChan:
			s.log.Debug("UnexpectedShutdownReceived", "Server shut down unexpectedly")
			// One of the servers has shut down unexpectedly. Because this makes the whole service unreliable, shutdown.
			break
		case <-ctx.Done():
			s.log.Debug("ServiceCancel", "Cancellation request received")

			// Shutdown any running http servers
			s.quitting = true
			s.sendChan <- true
			break
		case <-sigs:
			s.log.Debug("GracefulShutdown", "Handling Sigterm/SigInt")
			break
		}

		if !s.quitting {
			// Some other go-routine is already taking care of the shutdown
			s.quitting = true
			s.sendChan <- true
		}

		// Trigger graceful shutdown
		s.exitFunc(0)
		done <- true
	}()

	s.runReadinessServer()
	s.runInternalServer()
	s.runPublicServer()

	<-done // Wait for our shutdown

	// since service.ExitFunc calls os.Exit(), we'll never get here
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

func (s *serviceImpl) runHttpServer(port int, router *Router) {
	addr := fmt.Sprintf(":%v", port)
	svr := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Addr:         addr,
		Handler:      router.Router,
	}

	go func() {
		// Blocking until the server stops.
		svr.ListenAndServe()

		// Notify the service that the server has stopped.
		s.receiveChan <- true
	}()

	go func() {
		// Monitor sender channel and close server on signal.
		select {
		case sig := <-s.sendChan:
			// Properly close the server if possible.
			if svr != nil {
				svr.Close()
				svr = nil
			}
			// Continue sending the message
			s.sendChan <- sig
			break
		}
	}()
}

// RunReadinessServer runs the readiness service as a go-routine
func (s *serviceImpl) runReadinessServer() {
	const subsystem = "readiness"

	router := s.readinessRouter
	fact := s.handlerFactory

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, []Middleware{PanicTo500, Histogram}, fact.CreateRootHandler())
	s.addRoute(router, subsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, []Middleware{PanicTo500, Counter, NoCaching}, fact.CreateLivenessHandler())
	s.addRoute(router, subsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, []Middleware{PanicTo500, Histogram, NoCaching}, fact.CreateReadinessHandler())

	s.log.Info("RunReadinessServer", "%s %s running on localhost:%d.", s.name, subsystem, s.readinessPort)

	s.runHttpServer(s.readinessPort, router)
}

// RunInternalServer runs the internal service as a go-routine
func (s *serviceImpl) runInternalServer() {
	const subsystem = "internal"

	router := s.internalRouter
	fact := s.handlerFactory

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, []Middleware{PanicTo500, Histogram}, fact.CreateRootHandler())
	s.addRoute(router, subsystem, "health_check", []string{"/health_check", "/healthz"}, MethodsForGet, []Middleware{PanicTo500, Counter, NoCaching}, fact.CreateHealthHandler())
	s.addRoute(router, subsystem, "metrics", []string{"/metrics"}, MethodsForGet, []Middleware{PanicTo500, Counter, NoCaching}, fact.CreateMetricsHandler())
	s.addRoute(router, subsystem, "quit", []string{"/quit"}, MethodsForGet, []Middleware{PanicTo500, NoCaching}, fact.CreateQuitHandler())

	s.log.Info("RunInternalServer", "%s %s running on localhost:%d.", s.name, subsystem, s.internalPort)

	s.runHttpServer(s.internalPort, router)
}

// RunPublicServer runs the public service on the current thread.
func (s *serviceImpl) runPublicServer() {
	router := s.publicRouter
	fact := s.handlerFactory

	s.addRoute(router, publicSubsystem, "root", []string{"/"}, MethodsForGet, []Middleware{PanicTo500, Histogram}, fact.CreateRootHandler())
	s.addRoute(router, publicSubsystem, "version", []string{"/service/version"}, MethodsForGet, []Middleware{PanicTo500, Counter}, fact.CreateLivenessHandler())
	s.addRoute(router, publicSubsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, []Middleware{PanicTo500, Counter}, fact.CreateLivenessHandler())
	s.addRoute(router, publicSubsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, []Middleware{PanicTo500, Histogram, NoCaching}, fact.CreateReadinessHandler())

	s.log.Info("RunPublicService", "%s %s running on localhost:%d.", s.name, publicSubsystem, s.port)

	s.runHttpServer(s.port, router)
}

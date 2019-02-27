package servicefoundation

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Travix-International/go-servicefoundation/env"
)

const (
	envCORSOrigins       = "CORS_ORIGINS"
	envHTTPpPort         = "HTTPPORT"
	envAppName           = "APP_NAME"
	envServerName        = "SERVER_NAME"
	envDeployEnvironment = "DEPLOY_ENVIRONMENT"

	defaultHTTPPort = 8080

	publicSubsystem = "public"
)

type (
	// ShutdownFunc is a function signature for the shutdown function.
	ShutdownFunc func(log Logger)

	// ServiceGlobals contains basic service properties, like name, deployment environment and version number.
	ServiceGlobals struct {
		AppName           string
		GroupName         string
		ServerName        string
		DeployEnvironment string
		VersionNumber     string
	}

	// ServiceOptions contains value and references used by the Service implementation. The contents of ServiceOptions
	// can be used to customize or extend ServiceFoundation.
	ServiceOptions struct {
		Globals                  ServiceGlobals
		Port                     int
		ReadinessPort            int
		InternalPort             int
		LogFactory               LogFactory
		Metrics                  Metrics
		RouterFactory            RouterFactory
		MiddlewareWrapperFactory MiddlewareWrapperFactory
		CORSOptions              CORSOptions
		Handlers                 Handlers
		VersionBuilder           VersionBuilder
		ServiceStateManager      ServiceStateManager
		ShutdownFunc             ShutdownFunc
		AuthFunc                 AuthorizationFunc
		ServerTimeout            time.Duration
		IdleTimeout              time.Duration
		UsePublicRootHandler     bool
		systemLogger             Logger
	}

	// Service is the main interface for ServiceFoundation and is used to define routing and running the service.
	Service interface {
		Run(ctx context.Context)

		AddRoute(name string, routes []string, methods []string, middlewares []Middleware, metaFunc MetaFunc, handler Handle)
	}

	serviceImpl struct {
		globals              ServiceGlobals
		serverTimeout        time.Duration
		idleTimeout          time.Duration
		port                 int
		readinessPort        int
		internalPort         int
		logFactory           LogFactory
		log                  Logger
		metrics              Metrics
		publicRouter         Router
		readinessRouter      Router
		internalRouter       Router
		handlers             Handlers
		versionBuilder       VersionBuilder
		stateManager         ServiceStateManager
		shutdownFunc         ShutdownFunc
		exitFunc             ExitFunc
		quitting             bool
		sendChan             chan bool
		receiveChan          chan bool
		usePublicRootHandler bool
	}
)

// DefaultMiddlewares contains the default middleware wrappers for the predefined service endpoints.
var DefaultMiddlewares = []Middleware{PanicTo500, NoCaching}

// NewDefaultServiceOptions creates and returns ServiceOptions using a default configuration.
func NewDefaultServiceOptions(group, name string) ServiceOptions {
	opt := NewServiceOptions(group, name,
		[]string{http.MethodGet, http.MethodPost, http.MethodOptions},
		BuildVersion{},
		make(map[string]string),
	)
	opt.SetState(NewDefaultServiceStateManger())

	return opt
}

// NewServiceOptions creates and returns ServiceOptions that use environment variables for default configuration.
func NewServiceOptions(group, name string,
	allowedMethods []string,
	version BuildVersion,
	meta map[string]string,
) ServiceOptions {

	appName := env.OrDefault(envAppName, name)
	serverName := env.OrDefault(envServerName, name)
	deployEnvironment := env.OrDefault(envDeployEnvironment, "UNKNOWN")
	corsOptions := CORSOptions{
		AllowedOrigins: env.ListOrDefault(envCORSOrigins, []string{"*"}),
		AllowedMethods: allowedMethods,
	}
	globals := ServiceGlobals{
		AppName:           appName,
		GroupName:         group,
		ServerName:        serverName,
		DeployEnvironment: deployEnvironment,
		VersionNumber:     version.VersionNumber,
	}
	serviceMeta := createServiceMeta(meta, globals)
	logFactory := NewLogFactory(GetLogFilter(), serviceMeta)
	logger := logFactory.NewLogger(meta)
	metrics := NewMetrics(name, logger)
	versionBuilder := NewVersionBuilder(version)
	middlewareWrapperFactory := NewMiddlewareWrapperFactory(logger, globals)
	port := env.AsInt(envHTTPpPort, defaultHTTPPort)

	opt := ServiceOptions{
		Globals:                  globals,
		ServerTimeout:            time.Second * 30,
		IdleTimeout:              time.Second * 30,
		Port:                     port,
		ReadinessPort:            port + 1,
		InternalPort:             port + 2,
		MiddlewareWrapperFactory: middlewareWrapperFactory,
		CORSOptions:              corsOptions,
		RouterFactory:            NewRouterFactory(logFactory, metrics),
		LogFactory:               logFactory,
		Metrics:                  metrics,
		VersionBuilder:           versionBuilder,
		AuthFunc: func(WrappedResponseWriter, *http.Request, HandlerUtils) bool {
			// By default, anyone is authorized
			return true
		},
		UsePublicRootHandler: true,
		systemLogger:         logger,
	}
	return opt
}

// NewService creates and returns a Service using the provided service options.
func NewService(options ServiceOptions) Service {
	if options.systemLogger == nil {
		serviceMeta := createServiceMeta(make(map[string]string), options.Globals)
		options.systemLogger = options.LogFactory.NewLogger(serviceMeta)
	}

	service := serviceImpl{
		globals:              options.Globals,
		serverTimeout:        options.ServerTimeout,
		idleTimeout:          options.IdleTimeout,
		port:                 options.Port,
		readinessPort:        options.ReadinessPort,
		internalPort:         options.InternalPort,
		logFactory:           options.LogFactory,
		log:                  options.systemLogger,
		metrics:              options.Metrics,
		publicRouter:         options.RouterFactory.NewRouter(),
		readinessRouter:      options.RouterFactory.NewRouter(),
		internalRouter:       options.RouterFactory.NewRouter(),
		handlers:             options.Handlers,
		versionBuilder:       options.VersionBuilder,
		stateManager:         options.ServiceStateManager,
		exitFunc:             newExitFunc(options.systemLogger, options.ServiceStateManager),
		sendChan:             make(chan bool, 1),
		receiveChan:          make(chan bool, 1),
		usePublicRootHandler: options.UsePublicRootHandler,
	}

	service.stateManager.WarmUp()

	return &service
}

func createServiceMeta(baseMeta map[string]string, globals ServiceGlobals) map[string]string {
	serviceMeta := make(map[string]string)

	serviceMeta["entry.applicationgroup"] = globals.GroupName
	serviceMeta["entry.applicationname"] = globals.AppName
	serviceMeta["entry.applicationversion"] = globals.VersionNumber
	serviceMeta["entry.machinename"] = globals.ServerName

	return combineMetas(baseMeta, serviceMeta)
}

// newExitFunc returns a new exit function. It wraps the shutdownFunc and executed an os.exit after the shutdown is
// completed with a slight delay, giving the quit handler a chance to return a status.
func newExitFunc(log Logger, stateManager ServiceStateManager) func(int) {
	return func(code int) {
		log.Debug("ServiceExit", "Performing service exit")

		if stateManager != nil {
			log.Debug("ShutdownFunc", "Calling state manager's shutdown")
			stateManager.ShutDown(log)
		}

		log.Debug("ServiceExit", "Calling os.Exit(%v)", code)
		os.Exit(code)
	}
}

/* ServiceOptions implementation */

// SetState assigns a state manager and re-binds the handlers.
func (o *ServiceOptions) SetState(stateManager ServiceStateManager) {
	o.ServiceStateManager = stateManager
	o.setHandlers()
}

// setHandlers is used to update the handler references in ServiceOptions to use the correct middleware and state.
func (o *ServiceOptions) setHandlers() {
	middlewareWrapper := o.MiddlewareWrapperFactory.NewMiddlewareWrapper(&o.CORSOptions, o.AuthFunc)
	wrappedExitFunc := newExitFunc(o.systemLogger, o.ServiceStateManager)

	factory := NewServiceHandlerFactory(middlewareWrapper, o.VersionBuilder, o.ServiceStateManager, wrappedExitFunc,
		o.LogFactory, o.Metrics)

	o.Handlers = factory.NewHandlers()
}

/* Service implementation */

func (s *serviceImpl) Run(ctx context.Context) {
	s.log.Info("Service", "%s: %s", s.globals.AppName, s.versionBuilder.ToString())

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

func (s *serviceImpl) AddRoute(name string, routes []string, methods []string, middlewares []Middleware, metaFunc MetaFunc, handler Handle) {
	s.addRouteWithMetaAndPreFlight(s.publicRouter, publicSubsystem, name, routes, methods, middlewares, metaFunc, handler)
}

func (s *serviceImpl) addRoute(router Router, subsystem, name string, routes []string, methods []string, middlewares []Middleware, handler Handle) {
	defaultMetaFunc := func(_ *http.Request, _ RouteParamsFunc) Meta {
		return createServiceMeta(make(map[string]string), s.globals)
	}

	for _, path := range routes {
		for _, method := range methods {
			router.Handle(method, path, defaultMetaFunc, handler)
		}
	}
}

func (s *serviceImpl) addRouteWithMetaAndPreFlight(router Router, subsystem, name string, routes []string, methods []string,
	middlewares []Middleware, metaFunc MetaFunc, handler Handle) {

	for _, path := range routes {
		preFlightHandled := false

		for _, method := range methods {
			router.Handle(method, path, metaFunc, handler)
			preFlightHandled = preFlightHandled || method == http.MethodOptions
		}

		if preFlightHandled {
			continue
		}

		s.addPreFlightHandle(subsystem, name, middlewares, metaFunc, router, path)
	}
}

func (s *serviceImpl) addPreFlightHandle(subsystem string, name string, middlewares []Middleware, metaFunc MetaFunc,
	router Router, path string) {

	preFlightMiddlewares := []Middleware{Counter}

	for _, m := range middlewares {
		if m != CORS {
			continue
		}
		preFlightMiddlewares = append([]Middleware{CORS}, preFlightMiddlewares...)
	}

	preFlightHandler := s.handlers.PreFlightHandler.NewPreFlightHandler()
	router.Handle(http.MethodOptions, path, metaFunc, preFlightHandler)
}

func (s *serviceImpl) runHTTPServer(port int, router http.Handler) {
	addr := fmt.Sprintf(":%v", port)
	svr := &http.Server{
		ReadTimeout:  s.serverTimeout,
		WriteTimeout: s.serverTimeout,
		IdleTimeout:  s.idleTimeout,
		Addr:         addr,
		Handler:      router,
	}

	go func() {
		// Blocking until the server stops.
		_ = svr.ListenAndServe()

		// Notify the service that the server has stopped.
		s.receiveChan <- true
	}()

	go func() {
		// Monitor sender channel and close server on signal.
		select {
		case sig := <-s.sendChan:
			// Properly close the server if possible.
			if svr != nil {
				err := svr.Close()
				if err != nil {
					s.log.Error("CloseServer", err.Error())
				}
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

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, DefaultMiddlewares, s.handlers.RootHandler.NewRootHandler())
	s.addRoute(router, subsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, DefaultMiddlewares, s.handlers.LivenessHandler.NewLivenessHandler())
	s.addRoute(router, subsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, DefaultMiddlewares, s.handlers.ReadinessHandler.NewReadinessHandler())

	s.log.Info("RunReadinessServer", "%s %s running on localhost:%d.", s.globals.AppName, subsystem, s.readinessPort)

	s.runHTTPServer(s.readinessPort, router)
}

// RunInternalServer runs the internal service as a go-routine
func (s *serviceImpl) runInternalServer() {
	const subsystem = "internal"

	router := s.internalRouter

	s.addRoute(router, subsystem, "root", []string{"/"}, MethodsForGet, DefaultMiddlewares, s.handlers.RootHandler.NewRootHandler())
	s.addRoute(router, subsystem, "health_check", []string{"/health_check", "/healthz"}, MethodsForGet, DefaultMiddlewares, s.handlers.HealthHandler.NewHealthHandler())
	s.addRoute(router, subsystem, "metrics", []string{"/metrics"}, MethodsForGet, DefaultMiddlewares, s.handlers.MetricsHandler.NewMetricsHandler())
	s.addRoute(router, subsystem, "quit", []string{"/quit"}, MethodsForGet, DefaultMiddlewares, s.handlers.QuitHandler.NewQuitHandler())

	s.log.Info("RunInternalServer", "%s %s running on localhost:%d.", s.globals.AppName, subsystem, s.internalPort)

	s.runHTTPServer(s.internalPort, router)
}

// RunPublicServer runs the public service on the current thread.
func (s *serviceImpl) runPublicServer() {
	router := s.publicRouter

	if s.usePublicRootHandler {
		s.addRoute(router, publicSubsystem, "root", []string{"/"}, MethodsForGet, DefaultMiddlewares, s.handlers.RootHandler.NewRootHandler())
	}
	s.addRoute(router, publicSubsystem, "version", []string{"/service/version"}, MethodsForGet, DefaultMiddlewares, s.handlers.VersionHandler.NewVersionHandler())
	s.addRoute(router, publicSubsystem, "liveness", []string{"/service/liveness"}, MethodsForGet, DefaultMiddlewares, s.handlers.LivenessHandler.NewLivenessHandler())
	s.addRoute(router, publicSubsystem, "readiness", []string{"/service/readiness"}, MethodsForGet, DefaultMiddlewares, s.handlers.ReadinessHandler.NewReadinessHandler())

	s.log.Info("RunPublicService", "%s %s running on localhost:%d.", s.globals.AppName, publicSubsystem, s.port)

	s.runHTTPServer(s.port, router)
}

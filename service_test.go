package servicefoundation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestNewDefaultServiceOptions(t *testing.T) {
	w := &mockResponseWriter{}
	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", nil)

	// Act
	sut := sf.NewDefaultServiceOptions("some-group", "some-name")

	assert.True(t, sut.AuthFunc(w, r, sf.HandlerUtils{}))
	assert.NotNil(t, sut.LogFactory)
	assert.NotNil(t, sut.MiddlewareWrapperFactory)
	assert.NotNil(t, sut.ServiceStateManager)
	assert.NotNil(t, sut.VersionBuilder)
	assert.NotNil(t, sut.WrapHandler)
}

func TestNewService(t *testing.T) {
	opt := sf.NewDefaultServiceOptions("some-group", "some-name")

	// Act
	sut := sf.NewService(opt)

	assert.NotNil(t, sut)
}

func TestServiceImpl_AddRoute(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := httprouter.Router{}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:          logFactory,
		Metrics:             m,
		Port:                1234,
		ReadinessPort:       1235,
		InternalPort:        1236,
		ShutdownFunc:        func(log sf.Logger) {},
		VersionBuilder:      v,
		RouterFactory:       rf,
		Handlers:            handlers,
		WrapHandler:         shf,
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	var preFlightHandle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := sf.DefaultMiddlewares

	logFactory.On("NewLogger", mock.Anything).Return(log)
	shf.
		On("Wrap", "public", "do", middlewares, mock.AnythingOfType("Handle"), mock.AnythingOfType("MetaFunc")).
		Return(wrappedHandle).
		Twice() // for each route
	shf.
		On("Wrap", "public", "do-preflight", mock.Anything, mock.AnythingOfType("Handle"), mock.AnythingOfType("MetaFunc")).
		Return(wrappedHandle).
		Twice() // for each route
	rf.
		On("NewRouter").
		Return(&router).
		Times(3) // public, readiness and internal
	preFlightH.On("NewPreFlightHandler").Return(preFlightHandle)
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, metaFunc, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	preFlightH.AssertExpectations(t)
	stateManager.AssertExpectations(t)
}

func TestServiceImpl_AddRouteWithCORS(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := httprouter.Router{}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:          logFactory,
		Metrics:             m,
		Port:                1234,
		ReadinessPort:       1235,
		InternalPort:        1236,
		ShutdownFunc:        func(log sf.Logger) {},
		VersionBuilder:      v,
		RouterFactory:       rf,
		Handlers:            handlers,
		WrapHandler:         shf,
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	var preFlightHandle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := append([]sf.Middleware{sf.CORS}, sf.DefaultMiddlewares...)
	hasPreFlightCORS := false

	logFactory.On("NewLogger", mock.Anything).Return(log)
	shf.
		On("Wrap", "public", "do", middlewares, mock.AnythingOfType("Handle"), mock.AnythingOfType("MetaFunc")).
		Return(wrappedHandle).
		Once()
	shf.
		On("Wrap", "public", "do-preflight", mock.Anything, mock.AnythingOfType("Handle"), mock.AnythingOfType("MetaFunc")).
		Run(func(args mock.Arguments) {
			mw := args.Get(2).([]sf.Middleware)
			for _, m := range mw {
				hasPreFlightCORS = hasPreFlightCORS || m == sf.CORS
			}
		}).
		Return(wrappedHandle).
		Once()
	rf.
		On("NewRouter").
		Return(&router).
		Times(3) // public, readiness and internal
	preFlightH.On("NewPreFlightHandler").Return(preFlightHandle)
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do"}, []string{http.MethodGet}, middlewares, metaFunc, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	preFlightH.AssertExpectations(t)
	stateManager.AssertExpectations(t)
	assert.True(t, hasPreFlightCORS)
}

func TestServiceImpl_AddRouteWithHandledPreFlight(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := httprouter.Router{}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:          logFactory,
		Metrics:             m,
		Port:                1234,
		ReadinessPort:       1235,
		InternalPort:        1236,
		ShutdownFunc:        func(log sf.Logger) {},
		VersionBuilder:      v,
		RouterFactory:       rf,
		Handlers:            handlers,
		WrapHandler:         shf,
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := sf.DefaultMiddlewares

	logFactory.On("NewLogger", mock.Anything).Return(log)
	shf.
		On("Wrap", "public", "do", middlewares, mock.AnythingOfType("Handle"),
			mock.AnythingOfType("MetaFunc")).
		Return(wrappedHandle).
		Once()
	rf.
		On("NewRouter").
		Return(&router).
		Times(3) // public, readiness and internal
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do"}, []string{http.MethodGet, http.MethodOptions}, middlewares, metaFunc, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	stateManager.AssertExpectations(t)
}

func TestServiceImpl_Run(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	stateManager := &mockServiceStateManager{}

	publicRouter := httprouter.Router{}
	readinessRouter := httprouter.Router{}
	internalRouter := httprouter.Router{}

	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	var handle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}

	quitH := &mockQuitHandler{}
	rootH := &mockRootHandler{}
	livenessH := &mockLivenessHandler{}
	versionH := &mockVersionHandler{}
	readinessH := &mockReadinessHandler{}
	metricsH := &mockMetricsHandler{}
	healthH := &mockHealthHandler{}
	preFlightH := &mockPreFlightHandler{}

	handlers := sf.Handlers{
		QuitHandler:      quitH,
		MetricsHandler:   metricsH,
		VersionHandler:   versionH,
		HealthHandler:    healthH,
		LivenessHandler:  livenessH,
		ReadinessHandler: readinessH,
		RootHandler:      rootH,
		PreFlightHandler: preFlightH,
	}

	logFactory.On("NewLogger", mock.Anything).Return(log)
	log.On("Info", mock.Anything, mock.Anything, mock.Anything)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything)
	v.On("ToString").Return("(version)")
	rootH.On("NewRootHandler").Return(handle).Times(3)
	livenessH.On("NewLivenessHandler").Return(handle).Twice()
	readinessH.On("NewReadinessHandler").Return(handle).Twice()
	healthH.On("NewHealthHandler").Return(handle).Once()
	metricsH.On("NewMetricsHandler").Return(handle).Once()
	quitH.On("NewQuitHandler").Return(handle).Once()
	versionH.On("NewVersionHandler").Return(handle).Once()
	shf.
		On("Wrap", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(wrappedHandle)
	rf.
		On("NewRouter").
		Return(&readinessRouter).
		Once()
	rf.
		On("NewRouter").
		Return(&internalRouter).
		Once()
	rf.
		On("NewRouter").
		Return(&publicRouter).
		Once()
	stateManager.On("WarmUp").Once()
	stateManager.On("ShutDown", mock.Anything).Run(func(args mock.Arguments) {
		fmt.Println("Exit called!")
		// We need to sleep longer in order for the test to finish before this function exits
		time.Sleep(50 * time.Millisecond)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:           logFactory,
		Metrics:              m,
		Port:                 1234,
		ReadinessPort:        1235,
		InternalPort:         1236,
		ShutdownFunc:         func(log sf.Logger) {},
		VersionBuilder:       v,
		RouterFactory:        rf,
		Handlers:             handlers,
		WrapHandler:          shf,
		ServiceStateManager:  stateManager,
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: true,
	}

	sut := sf.NewService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(11 * time.Millisecond)
	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	rootH.AssertExpectations(t)
	livenessH.AssertExpectations(t)
	readinessH.AssertExpectations(t)
	healthH.AssertExpectations(t)
	metricsH.AssertExpectations(t)
	quitH.AssertExpectations(t)
	versionH.AssertExpectations(t)
}

func TestServiceImpl_Run_NoPublicRootHandler(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	stateManager := &mockServiceStateManager{}

	publicRouter := httprouter.Router{}
	readinessRouter := httprouter.Router{}
	internalRouter := httprouter.Router{}

	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	var handle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}

	quitH := &mockQuitHandler{}
	rootH := &mockRootHandler{}
	livenessH := &mockLivenessHandler{}
	versionH := &mockVersionHandler{}
	readinessH := &mockReadinessHandler{}
	metricsH := &mockMetricsHandler{}
	healthH := &mockHealthHandler{}
	preFlightH := &mockPreFlightHandler{}

	handlers := sf.Handlers{
		QuitHandler:      quitH,
		MetricsHandler:   metricsH,
		VersionHandler:   versionH,
		HealthHandler:    healthH,
		LivenessHandler:  livenessH,
		ReadinessHandler: readinessH,
		RootHandler:      rootH,
		PreFlightHandler: preFlightH,
	}

	logFactory.On("NewLogger", mock.Anything).Return(log)
	log.On("Info", mock.Anything, mock.Anything, mock.Anything)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything)
	v.On("ToString").Return("(version)")
	rootH.On("NewRootHandler").Return(handle).Twice()
	livenessH.On("NewLivenessHandler").Return(handle).Twice()
	readinessH.On("NewReadinessHandler").Return(handle).Twice()
	healthH.On("NewHealthHandler").Return(handle).Once()
	metricsH.On("NewMetricsHandler").Return(handle).Once()
	quitH.On("NewQuitHandler").Return(handle).Once()
	versionH.On("NewVersionHandler").Return(handle).Once()
	shf.
		On("Wrap", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(wrappedHandle)
	rf.
		On("NewRouter").
		Return(&readinessRouter).
		Once()
	rf.
		On("NewRouter").
		Return(&internalRouter).
		Once()
	rf.
		On("NewRouter").
		Return(&publicRouter).
		Once()
	stateManager.On("WarmUp").Once()
	stateManager.On("ShutDown", mock.Anything).Run(func(args mock.Arguments) {
		fmt.Println("Exit called!")
		// We need to sleep longer in order for the test to finish before this function exits
		time.Sleep(50 * time.Millisecond)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:           logFactory,
		Metrics:              m,
		Port:                 1234,
		ReadinessPort:        1235,
		InternalPort:         1236,
		ShutdownFunc:         func(log sf.Logger) {},
		VersionBuilder:       v,
		RouterFactory:        rf,
		Handlers:             handlers,
		WrapHandler:          shf,
		ServiceStateManager:  stateManager,
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: false,
	}

	sut := sf.NewService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(11 * time.Millisecond)
	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	rootH.AssertExpectations(t)
	livenessH.AssertExpectations(t)
	readinessH.AssertExpectations(t)
	healthH.AssertExpectations(t)
	metricsH.AssertExpectations(t)
	quitH.AssertExpectations(t)
	versionH.AssertExpectations(t)
}

func TestNewDefaultServiceStateManger(t *testing.T) {
	log := &mockLogger{}

	// Act
	sut := sf.NewDefaultServiceStateManger()

	sut.WarmUp()

	assert.True(t, sut.IsLive())
	assert.True(t, sut.IsReady())
	assert.True(t, sut.IsHealthy())

	sut.ShutDown(log)
}

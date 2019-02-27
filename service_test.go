package servicefoundation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	sf "github.com/Travix-International/go-servicefoundation"
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
	assert.NotNil(t, sut.Handlers)
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
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := &mockRouter{}
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
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	metaFunc := func(_ *http.Request, _ sf.RouteParamsFunc) sf.Meta {
		return make(map[string]string)
	}
	var preFlightHandle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := sf.DefaultMiddlewares

	logFactory.On("NewLogger", mock.Anything).Return(log)
	rf.
		On("NewRouter").
		Return(router).
		Times(3) // public, readiness and internal
	router.
		On("Handle", mock.Anything, "/do", mock.Anything, mock.Anything).
		Times(3) // OPTIONS, GET & POST
	router.
		On("Handle", mock.Anything, "/do2", mock.Anything, mock.Anything).
		Times(3) // OPTIONS, GET & POST

	preFlightH.On("NewPreFlightHandler").Return(preFlightHandle)
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, metaFunc, handle)

	rf.AssertExpectations(t)
	router.AssertExpectations(t)
	preFlightH.AssertExpectations(t)
	stateManager.AssertExpectations(t)
}

func TestServiceImpl_AddRouteWithCORS(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := &mockRouter{}
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
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	metaFunc := func(_ *http.Request, _ sf.RouteParamsFunc) sf.Meta {
		return make(map[string]string)
	}
	var preFlightHandle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := append([]sf.Middleware{sf.CORS}, sf.DefaultMiddlewares...)

	logFactory.On("NewLogger", mock.Anything).Return(log)
	rf.
		On("NewRouter").
		Return(router).
		Times(3) // public, readiness and internal
	router.
		On("Handle", mock.Anything, "/do", mock.Anything, mock.Anything).
		Times(2) // OPTIONS & GET
	preFlightH.On("NewPreFlightHandler").Return(preFlightHandle)
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do"}, []string{http.MethodGet}, middlewares, metaFunc, handle)

	rf.AssertExpectations(t)
	preFlightH.AssertExpectations(t)
	stateManager.AssertExpectations(t)
}

func TestServiceImpl_AddRouteWithHandledPreFlight(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	preFlightH := &mockPreFlightHandler{}
	stateManager := &mockServiceStateManager{}

	handlers := sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := &mockRouter{}
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
		ServiceStateManager: stateManager,
		ServerTimeout:       time.Second * 3,
		IdleTimeout:         time.Second * 3,
	}
	metaFunc := func(_ *http.Request, _ sf.RouteParamsFunc) sf.Meta {
		return make(map[string]string)
	}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {}
	middlewares := sf.DefaultMiddlewares

	logFactory.On("NewLogger", mock.Anything).Return(log)
	rf.
		On("NewRouter").
		Return(router).
		Times(3) // public, readiness and internal
	router.
		On("Handle", mock.Anything, "/do", mock.Anything, mock.Anything).
		Times(3) // OPTIONS & GET
	stateManager.On("WarmUp").Once()

	sut := sf.NewService(opt)

	// Act
	sut.AddRoute("do", []string{"/do"}, []string{http.MethodGet, http.MethodOptions}, middlewares, metaFunc, handle)

	rf.AssertExpectations(t)
	stateManager.AssertExpectations(t)
}

func TestServiceImpl_Run(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	stateManager := &mockServiceStateManager{}

	publicRouter := &mockRouter{}
	readinessRouter := &mockRouter{}
	internalRouter := &mockRouter{}

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
	rf.
		On("NewRouter").
		Return(readinessRouter).
		Once()
	rf.
		On("NewRouter").
		Return(internalRouter).
		Once()
	rf.
		On("NewRouter").
		Return(publicRouter).
		Once()
	publicRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(5)
	readinessRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(4)
	internalRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(3)
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
		ServiceStateManager:  stateManager,
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: true,
	}

	sut := sf.NewService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(11 * time.Millisecond)
	rf.AssertExpectations(t)
	publicRouter.AssertExpectations(t)
	readinessRouter.AssertExpectations(t)
	internalRouter.AssertExpectations(t)
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
	stateManager := &mockServiceStateManager{}

	publicRouter := &mockRouter{}
	readinessRouter := &mockRouter{}
	internalRouter := &mockRouter{}

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
	rf.
		On("NewRouter").
		Return(readinessRouter).
		Once()
	rf.
		On("NewRouter").
		Return(internalRouter).
		Once()
	rf.
		On("NewRouter").
		Return(publicRouter).
		Once()
	publicRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(5)
	readinessRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(4)
	internalRouter.
		On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Times(3)
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
		ServiceStateManager:  stateManager,
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: false,
	}

	sut := sf.NewService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(11 * time.Millisecond)
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

package servicefoundation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Travix-International/go-servicefoundation"
	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestCreateDefaultService(t *testing.T) {
	shutdownFn := func(log sf.Logger) {
	}

	// Act
	sut := servicefoundation.NewService("some-group", "some-name", []string{}, shutdownFn, sf.BuildVersion{}, make(map[string]string))

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

	handlers := &sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := &sf.Router{
		Router: &httprouter.Router{},
	}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:     logFactory,
		Metrics:        m,
		Port:           1234,
		ReadinessPort:  1235,
		InternalPort:   1236,
		ShutdownFunc:   func(log sf.Logger) {},
		VersionBuilder: v,
		RouterFactory:  rf,
		Handlers:       handlers,
		WrapHandler:    shf,
		ServerTimeout:  time.Second * 3,
		IdleTimeout:    time.Second * 3,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	var preFlightHandle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}
	middlewares := servicefoundation.DefaultMiddlewares

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
		Return(router).
		Times(3) // public, readiness and internal
	preFlightH.On("NewPreFlightHandler").Return(preFlightHandle)

	sut := servicefoundation.NewCustomService(opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, metaFunc, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
	preFlightH.AssertExpectations(t)
}

func TestServiceImpl_AddRouteWithHandledPreFlight(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}
	preFlightH := &mockPreFlightHandler{}

	handlers := &sf.Handlers{
		PreFlightHandler: preFlightH,
	}
	router := &sf.Router{
		Router: &httprouter.Router{},
	}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:     logFactory,
		Metrics:        m,
		Port:           1234,
		ReadinessPort:  1235,
		InternalPort:   1236,
		ShutdownFunc:   func(log sf.Logger) {},
		VersionBuilder: v,
		RouterFactory:  rf,
		Handlers:       handlers,
		WrapHandler:    shf,
		ServerTimeout:  time.Second * 3,
		IdleTimeout:    time.Second * 3,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}
	middlewares := servicefoundation.DefaultMiddlewares

	logFactory.On("NewLogger", mock.Anything).Return(log)
	shf.
		On("Wrap", "public", "do", middlewares, mock.AnythingOfType("Handle"),
			mock.AnythingOfType("MetaFunc")).
		Return(wrappedHandle).
		Once()
	rf.
		On("NewRouter").
		Return(router).
		Times(3) // public, readiness and internal

	sut := servicefoundation.NewCustomService(opt)

	// Act
	sut.AddRoute("do", []string{"/do"}, []string{http.MethodGet, http.MethodOptions}, middlewares, metaFunc, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
}

func TestServiceImpl_Run(t *testing.T) {
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}

	publicRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	readinessRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	internalRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	var handle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}

	quitH := &mockQuitHandler{}
	rootH := &mockRootHandler{}
	livenessH := &mockLivenessHandler{}
	versionH := &mockVersionHandler{}
	readinessH := &mockReadinessHandler{}
	metricsH := &mockMetricsHandler{}
	healthH := &mockHealthHandler{}
	preFlightH := &mockPreFlightHandler{}

	handlers := &sf.Handlers{
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
	log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:     logFactory,
		Metrics:        m,
		Port:           1234,
		ReadinessPort:  1235,
		InternalPort:   1236,
		ShutdownFunc:   func(log sf.Logger) {},
		VersionBuilder: v,
		RouterFactory:  rf,
		Handlers:       handlers,
		WrapHandler:    shf,
		ExitFunc: func(int) {
			fmt.Println("Exit called!")
		},
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: true,
	}

	sut := servicefoundation.NewCustomService(opt)

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

	publicRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	readinessRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	internalRouter := &sf.Router{
		Router: &httprouter.Router{},
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	var handle sf.Handle = func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}

	quitH := &mockQuitHandler{}
	rootH := &mockRootHandler{}
	livenessH := &mockLivenessHandler{}
	versionH := &mockVersionHandler{}
	readinessH := &mockReadinessHandler{}
	metricsH := &mockMetricsHandler{}
	healthH := &mockHealthHandler{}
	preFlightH := &mockPreFlightHandler{}

	handlers := &sf.Handlers{
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
	log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		LogFactory:     logFactory,
		Metrics:        m,
		Port:           1234,
		ReadinessPort:  1235,
		InternalPort:   1236,
		ShutdownFunc:   func(log sf.Logger) {},
		VersionBuilder: v,
		RouterFactory:  rf,
		Handlers:       handlers,
		WrapHandler:    shf,
		ExitFunc: func(int) {
			fmt.Println("Exit called!")
		},
		ServerTimeout:        time.Second * 3,
		IdleTimeout:          time.Second * 3,
		UsePublicRootHandler: false,
	}

	sut := servicefoundation.NewCustomService(opt)

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

func TestNewExitFunc(t *testing.T) {
	log := &mockLogger{}
	shutdownFn := func(log sf.Logger) {}

	log.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Act
	sut := sf.NewExitFunc(log, shutdownFn)

	assert.NotNil(t, sut)
	go sut(1)
}

func TestNewServiceStateReader(t *testing.T) {
	// Act
	sut := sf.NewServiceStateReader()

	assert.True(t, sut.IsLive())
	assert.True(t, sut.IsReady())
	assert.True(t, sut.IsHealthy())
}

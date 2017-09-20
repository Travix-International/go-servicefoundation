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
	sut := servicefoundation.NewService("some-name", []string{}, shutdownFn, sf.BuildVersion{})

	assert.NotNil(t, sut)
}

func TestServiceImpl_AddRoute(t *testing.T) {
	log := &mockLogger{}
	m := &mockMetrics{}
	v := &mockVersionBuilder{}
	rf := &mockRouterFactory{}
	shf := &mockServiceHandlerFactory{}

	router := &sf.Router{
		Router: &httprouter.Router{},
	}
	opt := sf.ServiceOptions{
		Globals: sf.ServiceGlobals{
			AppName: "test-service",
		},
		Logger:         log,
		Metrics:        m,
		Port:           1234,
		ReadinessPort:  1235,
		InternalPort:   1236,
		ShutdownFunc:   func(log sf.Logger) {},
		VersionBuilder: v,
		RouterFactory:  rf,
		WrapHandler:    shf,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {}
	middlewares := servicefoundation.DefaultMiddlewares

	shf.
		On("Wrap", "public", "do", middlewares, mock.AnythingOfType("Handle")).
		Return(wrappedHandle).
		Twice() // for each route
	rf.
		On("NewRouter").
		Return(router).
		Times(3) // public, readiness and internal

	sut := servicefoundation.NewCustomService(opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
}

func TestServiceImpl_Run(t *testing.T) {
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

	handlers := &sf.Handlers{
		QuitHandler:      quitH,
		MetricsHandler:   metricsH,
		VersionHandler:   versionH,
		HealthHandler:    healthH,
		LivenessHandler:  livenessH,
		ReadinessHandler: readinessH,
		RootHandler:      rootH,
	}

	log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	v.On("ToString").Return("(version)")
	rootH.On("NewRootHandler").Return(handle).Times(3)
	livenessH.On("NewLivenessHandler").Return(handle)
	readinessH.On("NewReadinessHandler").Return(handle)
	healthH.On("NewHealthHandler").Return(handle).Once()
	metricsH.On("NewMetricsHandler").Return(handle).Once()
	quitH.On("NewQuitHandler").Return(handle).Once()
	versionH.On("NewVersionHandler").Return(handle).Once()
	shf.
		On("Wrap", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
		Logger:         log,
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
	}

	sut := servicefoundation.NewCustomService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(11 * time.Millisecond)
	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
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

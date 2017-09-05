package servicefoundation_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Prutswonder/go-servicefoundation"
	"github.com/Prutswonder/go-servicefoundation/model"
	. "github.com/Prutswonder/go-servicefoundation/testing"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestCreateDefaultService(t *testing.T) {
	shutdownFn := func(log model.Logger) {
	}

	// Act
	sut := servicefoundation.CreateDefaultService("some-name", []string{}, shutdownFn)

	assert.NotNil(t, sut)
}

func TestServiceImpl_AddRoute(t *testing.T) {
	log := &MockLogger{}
	m := &MockMetrics{}
	v := &MockVersionBuilder{}
	rf := &MockRouterFactory{}
	shf := &MockServiceHandlerFactory{}

	router := &model.Router{
		Router: &httprouter.Router{},
	}
	opt := model.ServiceOptions{
		Globals: model.ServiceGlobals{
			AppName: "test-service",
		},
		Logger:                log,
		Metrics:               m,
		Port:                  1234,
		ReadinessPort:         1235,
		InternalPort:          1236,
		ShutdownFunc:          func(log model.Logger) {},
		VersionBuilder:        v,
		RouterFactory:         rf,
		ServiceHandlerFactory: shf,
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	handle := func(model.WrappedResponseWriter, *http.Request, model.RouterParams) {}
	middlewares := servicefoundation.DefaultMiddlewares

	shf.
		On("WrapHandler", "public", "do", middlewares, mock.AnythingOfType("model.Handle")).
		Return(wrappedHandle).
		Twice() // for each route
	rf.
		On("CreateRouter").
		Return(router).
		Times(3) // public, readiness and internal

	sut := servicefoundation.CreateService(opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, handle)

	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
}

func TestServiceImpl_Run(t *testing.T) {
	log := &MockLogger{}
	m := &MockMetrics{}
	v := &MockVersionBuilder{}
	rf := &MockRouterFactory{}
	shf := &MockServiceHandlerFactory{}

	publicRouter := &model.Router{
		Router: &httprouter.Router{},
	}
	readinessRouter := &model.Router{
		Router: &httprouter.Router{},
	}
	internalRouter := &model.Router{
		Router: &httprouter.Router{},
	}
	var wrappedHandle httprouter.Handle = func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	var handle model.Handle = func(model.WrappedResponseWriter, *http.Request, model.RouterParams) {}

	log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	log.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	v.On("ToString").Return("(version)")
	shf.On("CreateRootHandler").Return(handle).Times(3)
	shf.On("CreateLivenessHandler").Return(handle)
	shf.On("CreateReadinessHandler").Return(handle)
	shf.On("CreateHealthHandler").Return(handle).Once()
	shf.On("CreateMetricsHandler").Return(handle).Once()
	shf.On("CreateQuitHandler").Return(handle).Once()
	shf.On("CreateVersionHandler").Return(handle).Once()
	shf.
		On("WrapHandler", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(wrappedHandle)
	rf.
		On("CreateRouter").
		Return(readinessRouter).
		Once()
	rf.
		On("CreateRouter").
		Return(internalRouter).
		Once()
	rf.
		On("CreateRouter").
		Return(publicRouter).
		Once()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	opt := model.ServiceOptions{
		Globals: model.ServiceGlobals{
			AppName: "test-service",
		},
		Logger:                log,
		Metrics:               m,
		Port:                  1234,
		ReadinessPort:         1235,
		InternalPort:          1236,
		ShutdownFunc:          func(log model.Logger) {},
		VersionBuilder:        v,
		RouterFactory:         rf,
		ServiceHandlerFactory: shf,
		ExitFunc: func(int) {
			fmt.Println("Exit called!")
		},
	}

	sut := servicefoundation.CreateService(opt)

	// Act
	go sut.Run(ctx)

	time.Sleep(10 * time.Millisecond)
	shf.AssertExpectations(t)
	rf.AssertExpectations(t)
}

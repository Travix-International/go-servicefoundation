package servicefoundation_test

import (
	"net/http"
	"testing"

	"github.com/Prutswonder/go-servicefoundation"
	"github.com/Prutswonder/go-servicefoundation/model"
	. "github.com/Prutswonder/go-servicefoundation/testing"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	handle := func(model.WrappedResponseWriter, *http.Request, model.RouterParams) {

	}
	var wrappedHandle httprouter.Handle

	wrappedHandle = func(http.ResponseWriter, *http.Request, httprouter.Params) {

	}
	middlewares := []model.Middleware{model.NoCaching, model.CORS, model.Histogram}

	shf.
		On("WrapHandler", "public", "do", middlewares, mock.AnythingOfType("model.Handle")).
		Return(wrappedHandle).
		Twice() // for each route
	rf.
		On("CreateRouter").
		Return(router).
		Times(3) // public, readiness and internal

	sut := servicefoundation.CreateService("test-service", opt)

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{http.MethodGet, http.MethodPost}, middlewares, handle)

	shf.AssertExpectations(t)
}

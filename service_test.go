package servicefoundation_test

import (
	"net/http"
	"testing"

	"github.com/Prutswonder/go-servicefoundation"
	"github.com/Prutswonder/go-servicefoundation/model"
	. "github.com/Prutswonder/go-servicefoundation/testing"
	"github.com/stretchr/testify/assert"
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
	middlewares := []model.Middleware{model.NoCaching, model.CORS}
	sut := servicefoundation.CreateService("test-service", opt)

	shf.On("WrapHandler", "public", "do", middlewares, handle).Once()

	// Act
	sut.AddRoute("do", []string{"/do", "/do2"}, []string{"GET", "POST"}, middlewares, handle)

	shf.AssertExpectations(t)
}

package site_test

import (
	"net/http"
	"testing"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Prutswonder/go-servicefoundation/site"
	. "github.com/Prutswonder/go-servicefoundation/testing"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceHandlerFactoryImpl_CreateRootHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.CreateRootHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateReadinessHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateLivenessHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateHealthHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateVersionHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	version := make(map[string]string)
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	v.On("ToMap").Return(version).Once()
	w.On("JSON", http.StatusOK, version).Once()

	// Act
	actual := sut.CreateVersionHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
	v.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateMetricsHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	w := &MockResponseWriter{}
	rdr := &MockReader{}
	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", rdr)
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("Header").Return(http.Header{}).Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Once()

	// Act
	actual := sut.CreateMetricsHandler()
	actual(w, r, model.RouterParams{})
	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateQuitHandler(t *testing.T) {
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	called := false
	exitFn := func(int) {
		called = true
	}
	w := &MockResponseWriter{}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()
	w.On("Flush").Once()

	// Act
	actual := sut.CreateQuitHandler()
	actual(w, nil, model.RouterParams{})

	w.AssertExpectations(t)
	assert.True(t, called)
}

func TestServiceHandlerFactoryImpl_WrapHandler(t *testing.T) {
	const subSystem = "my-sub"
	const name = "my-name"
	m := &MockMiddlewareWrapper{}
	v := &MockVersionBuilder{}
	exitFn := func(int) {}
	rdr := &MockReader{}
	r, _ := http.NewRequest("GET", "https://www.site.com/some/url", rdr)
	w := &MockResponseWriter{}
	called := false
	handle := func(model.WrappedResponseWriter, *http.Request, model.RouterParams) {
		called = true
	}
	sut := site.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	m.On("Wrap", subSystem, name, model.CORS, mock.Anything).Return(handle).Once()
	m.On("Wrap", subSystem, name, model.NoCaching, mock.Anything).Return(handle).Once()

	// Act
	actual := sut.WrapHandler(subSystem, name, []model.Middleware{model.CORS, model.NoCaching}, handle)
	actual(w, r, httprouter.Params{})

	assert.True(t, called)
}

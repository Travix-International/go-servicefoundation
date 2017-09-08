package servicefoundation_test

import (
	"net/http"
	"testing"

	sf "github.com/Prutswonder/go-servicefoundation"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceHandlerFactoryImpl_CreateRootHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.CreateRootHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateReadinessHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateLivenessHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()

	// Act
	actual := sut.CreateHealthHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateVersionHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	version := make(map[string]string)
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	v.On("ToMap").Return(version).Once()
	w.On("JSON", http.StatusOK, version).Once()

	// Act
	actual := sut.CreateVersionHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
	v.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateMetricsHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	rdr := &mockReader{}
	r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("Header").Return(http.Header{}).Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Once()

	// Act
	actual := sut.CreateMetricsHandler()
	actual(w, r, sf.RouterParams{})
	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateQuitHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	called := false
	exitFn := func(int) {
		called = true
	}
	w := &mockResponseWriter{}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()
	w.On("Flush").Once()

	// Act
	actual := sut.CreateQuitHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
	assert.True(t, called)
}

func TestServiceHandlerFactoryImpl_WrapHandler(t *testing.T) {
	const subSystem = "my-sub"
	const name = "my-name"
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	rdr := &mockReader{}
	r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
	w := &mockResponseWriter{}
	called := false
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
		called = true
	}
	sut := sf.CreateServiceHandlerFactory(m, v, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	m.On("Wrap", subSystem, name, sf.CORS, mock.Anything).Return(handle).Once()
	m.On("Wrap", subSystem, name, sf.NoCaching, mock.Anything).Return(handle).Once()

	// Act
	actual := sut.WrapHandler(subSystem, name, []sf.Middleware{sf.CORS, sf.NoCaching}, handle)
	actual(w, r, httprouter.Params{})

	assert.True(t, called)
}

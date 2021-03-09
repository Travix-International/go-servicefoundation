package v8_test

import (
	"net/http"
	"testing"

	sf "github.com/Travix-International/go-servicefoundation/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceHandlerFactoryImpl_CreateRootHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.NewHandlers().RootHandler.NewRootHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsReady").Return(true)

	// Act
	actual := sut.NewHandlers().ReadinessHandler.NewReadinessHandler()
	actual(w, nil, sf.RouterParams{})
	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler_NotReady(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsReady").Return(false)

	// Act
	actual := sut.NewHandlers().ReadinessHandler.NewReadinessHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsLive").Return(true)

	// Act
	actual := sut.NewHandlers().LivenessHandler.NewLivenessHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler_NotLive(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsLive").Return(false)

	// Act
	actual := sut.NewHandlers().LivenessHandler.NewLivenessHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsHealthy").Return(true)

	// Act
	actual := sut.NewHandlers().HealthHandler.NewHealthHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler_NotHealthy(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsHealthy").Return(false)

	// Act
	actual := sut.NewHandlers().HealthHandler.NewHealthHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateVersionHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	version := make(map[string]string)
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	v.On("ToMap").Return(version).Once()
	w.On("JSON", http.StatusOK, version).Once()

	// Act
	actual := sut.NewHandlers().VersionHandler.NewVersionHandler()
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
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("Header").Return(http.Header{}).Once()
	w.On("WriteHeader", http.StatusOK).Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil).
		Once()

	// Act
	actual := sut.NewHandlers().MetricsHandler.NewMetricsHandler()
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
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()
	w.On("Flush").Once()

	// Act
	actual := sut.NewHandlers().QuitHandler.NewQuitHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
	assert.True(t, called)
}

func TestServiceHandlerFactoryImpl_CreatePreFlightHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.NewHandlers().PreFlightHandler.NewPreFlightHandler()
	actual(w, nil, sf.RouterParams{})

	w.AssertExpectations(t)
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
	handlerCalled := false
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
		handlerCalled = true
	}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	ssr := &mockServiceStateReader{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	m.On("Wrap", subSystem, name, sf.CORS, mock.Anything, mock.Anything).Return(handle).Once()
	m.On("Wrap", subSystem, name, sf.NoCaching, mock.Anything, mock.Anything).Return(handle).Once()

	// Act
	actual := sut.Wrap(subSystem, name, []sf.Middleware{sf.CORS, sf.NoCaching}, handle, metaFunc)
	actual(w, r, httprouter.Params{})

	assert.True(t, handlerCalled)
}

package servicefoundation_test

import (
	"net/http"
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceHandlerFactoryImpl_CreateRootHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.NewHandlers().RootHandler.NewRootHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsReady").Return(true)

	// Act
	actual := sut.NewHandlers().ReadinessHandler.NewReadinessHandler()
	actual(w, nil, sf.HandlerUtils{})
	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateReadinessHandler_NotReady(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsReady").Return(false)

	// Act
	actual := sut.NewHandlers().ReadinessHandler.NewReadinessHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsLive").Return(true)

	// Act
	actual := sut.NewHandlers().LivenessHandler.NewLivenessHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateLivenessHandler_NotLive(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsLive").Return(false)

	// Act
	actual := sut.NewHandlers().LivenessHandler.NewLivenessHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusOK, mock.Anything).Once()
	ssr.On("IsHealthy").Return(true)

	// Act
	actual := sut.NewHandlers().HealthHandler.NewHealthHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateHealthHandler_NotHealthy(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("JSON", http.StatusInternalServerError, mock.Anything).Once()
	ssr.On("IsHealthy").Return(false)

	// Act
	actual := sut.NewHandlers().HealthHandler.NewHealthHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

func TestServiceHandlerFactoryImpl_CreateVersionHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	version := make(map[string]string)
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	v.On("ToMap").Return(version).Once()
	w.On("JSON", http.StatusOK, version).Once()

	// Act
	actual := sut.NewHandlers().VersionHandler.NewVersionHandler()
	actual(w, nil, sf.HandlerUtils{})

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
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("WriteHeader", http.StatusOK).Once()
	w.On("Header").Return(http.Header{}).Once()
	w.
		On("Write", mock.Anything).
		Return(0, nil)

	// Act
	actual := sut.NewHandlers().MetricsHandler.NewMetricsHandler()
	actual(w, r, sf.HandlerUtils{})
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
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("WriteHeader", http.StatusOK).Once()
	w.On("Flush").Once()

	// Act
	actual := sut.NewHandlers().QuitHandler.NewQuitHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
	assert.True(t, called)
}

func TestServiceHandlerFactoryImpl_CreatePreFlightHandler(t *testing.T) {
	m := &mockMiddlewareWrapper{}
	v := &mockVersionBuilder{}
	exitFn := func(int) {}
	w := &mockResponseWriter{}
	ssr := &mockServiceStateReader{}
	logFactory := &mockLogFactory{}
	metrics := &mockMetrics{}
	sut := sf.NewServiceHandlerFactory(m, v, ssr, exitFn, logFactory, metrics)

	w.On("WriteHeader", http.StatusOK).Once()

	// Act
	actual := sut.NewHandlers().PreFlightHandler.NewPreFlightHandler()
	actual(w, nil, sf.HandlerUtils{})

	w.AssertExpectations(t)
}

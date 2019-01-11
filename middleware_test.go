package servicefoundation_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMiddlewareWrapperImpl_Wrap(t *testing.T) {
	scenarios := []sf.Middleware{
		sf.CORS,
		sf.NoCaching,
		sf.Counter,
		sf.Histogram,
		sf.RequestLogging,
		sf.RequestLogging,
		sf.RequestMetrics,
		sf.PanicTo500,
		sf.Authorize,
	}
	useTls := false

	for i, scenario := range scenarios {
		const subSystem = "my-sub"
		const name = "my-name"
		logFactory := &mockLogFactory{}
		log := &mockLogger{}
		m := &mockMetrics{}
		corsOptions := &sf.CORSOptions{}
		handleCalled := false
		handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {
			handleCalled = true
		}
		metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
			return make(map[string]string)
		}
		authFunc := func(_ sf.WrappedResponseWriter, _ *http.Request, _ sf.HandlerUtils) bool {
			return true
		}
		rdr := &mockReader{}
		r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
		if useTls {
			r.URL.Scheme = ""
			r.TLS = &tls.ConnectionState{}
		}
		useTls = !useTls
		w := &mockResponseWriter{}
		h := &mockHistogramVec{}
		s := &mockSummaryVec{}
		u := sf.HandlerUtils{LogFactory: logFactory, Metrics: m, Meta: make(map[string]string)}

		header := http.Header{}
		header.Set("foo", "bar")

		w.On("Header").Return(header)
		w.On("Status").Return(http.StatusOK)
		h.On("RecordTimeElapsed", mock.Anything)
		h.On("RecordDuration", mock.Anything, mock.Anything)
		s.On("RecordDuration", mock.Anything, mock.Anything)
		m.On("CountLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		m.On("AddHistogramVec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(h)
		m.On("AddSummaryVec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(s)
		logFactory.On("NewLogger", mock.Anything).Return(log)
		log.On("Info", mock.Anything, mock.Anything, mock.Anything)
		log.On("AddMeta", mock.Anything)

		sut := sf.NewMiddlewareWrapper(log, m, corsOptions, authFunc, sf.ServiceGlobals{})

		// Act
		actual := sut.Wrap(subSystem, name, scenario, handle, metaFunc)

		assert.NotNil(t, actual, "Scenario %x", i)
		assert.NotEqual(t, &handle, &actual, "Scenario %x", i)

		actual(w, r, u)
		assert.True(t, handleCalled, "Scenario %x", i)
	}
}

func TestMiddlewareWrapperImpl_Wrap_UnknownMiddleware_ReturnsUnwrappedHandler(t *testing.T) {
	const subSystem = "my-sub"
	const name = "my-name"
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	corsOptions := &sf.CORSOptions{}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	authFunc := func(_ sf.WrappedResponseWriter, _ *http.Request, _ sf.HandlerUtils) bool {
		return true
	}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {
	}

	logFactory.On("NewLogger", mock.Anything).Return(log)
	log.On("Warn", mock.Anything, mock.Anything, mock.Anything).Once()

	sut := sf.NewMiddlewareWrapper(log, m, corsOptions, authFunc, sf.ServiceGlobals{})

	// Act
	actual := sut.Wrap(subSystem, name, 0, handle, metaFunc)

	assert.NotNil(t, actual)
	log.AssertExpectations(t)
}

func TestMiddlewareWrapperImpl_Wrap_PanicsAreHandled(t *testing.T) {
	scenarios := []sf.Middleware{sf.PanicTo500}

	for i, scenario := range scenarios {
		const subSystem = "my-sub"
		const name = "my-name"
		logFactory := &mockLogFactory{}
		log := &mockLogger{}
		m := &mockMetrics{}
		corsOptions := &sf.CORSOptions{}
		metaCalled := false
		metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
			metaCalled = true
			return make(map[string]string)
		}
		authFunc := func(_ sf.WrappedResponseWriter, _ *http.Request, _ sf.HandlerUtils) bool {
			return true
		}
		handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {
			panic("whoa")
		}
		rdr := &mockReader{}
		r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
		w := &mockResponseWriter{}
		u := sf.HandlerUtils{LogFactory: logFactory, Metrics: m, Meta: make(map[string]string)}

		logFactory.On("NewLogger", mock.Anything).Return(log)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Once()
		w.On("WriteHeader", http.StatusInternalServerError).Once()

		sut := sf.NewMiddlewareWrapper(log, m, corsOptions, authFunc, sf.ServiceGlobals{})

		// Act
		actual := sut.Wrap(subSystem, name, scenario, handle, metaFunc)

		assert.NotNil(t, actual, "Scenario %x", i)
		assert.NotEqual(t, &handle, &actual, "Scenario %x", i)
		assert.False(t, metaCalled, "Scenario %x", i)

		actual(w, r, u)
		assert.True(t, metaCalled, "Scenario %x", i)
		log.AssertExpectations(t)
		w.AssertExpectations(t)
	}
}

func TestMiddlewareWrapperImpl_UnAuthorized_HandlerIsNotCalled(t *testing.T) {
	const subSystem = "my-sub"
	const name = "my-name"
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	corsOptions := &sf.CORSOptions{}
	handleCalled := false
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils) {
		handleCalled = true
	}
	metaFunc := func(*http.Request, sf.RouterParams) map[string]string {
		return make(map[string]string)
	}
	authFunc := func(_ sf.WrappedResponseWriter, _ *http.Request, _ sf.HandlerUtils) bool {
		return false
	}
	rdr := &mockReader{}
	r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
	w := &mockResponseWriter{}
	u := sf.HandlerUtils{}

	w.On("WriteHeader", http.StatusUnauthorized)
	logFactory.On("NewLogger", mock.Anything).Return(log)

	sut := sf.NewMiddlewareWrapper(log, m, corsOptions, authFunc, sf.ServiceGlobals{})

	// Act
	actual := sut.Wrap(subSystem, name, sf.Authorize, handle, metaFunc)

	assert.NotNil(t, actual)
	assert.NotEqual(t, &handle, &actual)

	actual(w, r, u)
	assert.False(t, handleCalled)
}

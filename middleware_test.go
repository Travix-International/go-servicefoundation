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
	}
	useTls := false

	for i, scenario := range scenarios {
		const subSystem = "my-sub"
		const name = "my-name"
		logFactory := &mockLogFactory{}
		log := &mockLogger{}
		m := &mockMetrics{}
		corsOptions := &sf.CORSOptions{}
		called := false
		handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
			called = true
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
		p := sf.RouterParams{}

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
		log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		sut := sf.NewMiddlewareWrapper(logFactory, m, corsOptions, sf.ServiceGlobals{})

		// Act
		actual := sut.Wrap(subSystem, name, scenario, handle)

		assert.NotNil(t, actual, "Scenario %n", i)
		assert.NotEqual(t, handle, actual, "Scenario %n", i)

		actual(w, r, p)
		assert.True(t, called, "Scenario %n", i)
	}
}

func TestMiddlewareWrapperImpl_Wrap_UnknownMiddleware_ReturnsUnwrappedHandler(t *testing.T) {
	const subSystem = "my-sub"
	const name = "my-name"
	logFactory := &mockLogFactory{}
	log := &mockLogger{}
	m := &mockMetrics{}
	corsOptions := &sf.CORSOptions{}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
	}

	logFactory.On("NewLogger", mock.Anything).Return(log)
	log.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	sut := sf.NewMiddlewareWrapper(logFactory, m, corsOptions, sf.ServiceGlobals{})

	// Act
	actual := sut.Wrap(subSystem, name, 0, handle)

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
		handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
			panic("whoa")
		}
		rdr := &mockReader{}
		r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
		w := &mockResponseWriter{}
		p := sf.RouterParams{}

		logFactory.On("NewLogger", mock.Anything).Return(log)
		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		w.On("WriteHeader", http.StatusInternalServerError).Once()
		w.On("Header").Return(http.Header{}).Once()
		w.On("Status").Return(http.StatusOK).Once()

		sut := sf.NewMiddlewareWrapper(logFactory, m, corsOptions, sf.ServiceGlobals{})

		// Act
		actual := sut.Wrap(subSystem, name, scenario, handle)

		assert.NotNil(t, actual, "Scenario %n", i)
		assert.NotEqual(t, handle, actual, "Scenario %n", i)

		actual(w, r, p)
		log.AssertExpectations(t)
		w.AssertExpectations(t)
	}
}

package servicefoundation_test

import (
	"net/http"
	"testing"

	sf "github.com/Prutswonder/go-servicefoundation"
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
		sf.PanicTo500,
	}

	for i, scenario := range scenarios {
		const subSystem = "my-sub"
		const name = "my-name"
		log := &mockLogger{}
		m := &mockMetrics{}
		corsOptions := &sf.CORSOptions{}
		called := false
		handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
			called = true
		}
		rdr := &mockReader{}
		r, _ := http.NewRequest("GET", "https://www.sf.com/some/url", rdr)
		w := &mockResponseWriter{}
		h := &mockMetricsHistogram{}
		p := sf.RouterParams{}
		sut := sf.NewMiddlewareWrapper(log, m, corsOptions, sf.ServiceGlobals{})

		w.On("Header").Return(http.Header{})
		w.On("Status").Return(http.StatusOK)
		h.On("RecordTimeElapsed", mock.Anything, mock.Anything)
		m.On("CountLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		m.On("AddHistogram", mock.Anything, mock.Anything, mock.Anything).Return(h)
		log.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	log := &mockLogger{}
	m := &mockMetrics{}
	corsOptions := &sf.CORSOptions{}
	handle := func(sf.WrappedResponseWriter, *http.Request, sf.RouterParams) {
	}
	sut := sf.NewMiddlewareWrapper(log, m, corsOptions, sf.ServiceGlobals{})

	log.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
		sut := sf.NewMiddlewareWrapper(log, m, corsOptions, sf.ServiceGlobals{})

		log.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		w.On("WriteHeader", http.StatusInternalServerError).Once()

		// Act
		actual := sut.Wrap(subSystem, name, scenario, handle)

		assert.NotNil(t, actual, "Scenario %n", i)
		assert.NotEqual(t, handle, actual, "Scenario %n", i)

		actual(w, r, p)
		log.AssertExpectations(t)
		w.AssertExpectations(t)
	}
}

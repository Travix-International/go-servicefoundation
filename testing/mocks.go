package testing

import (
	"io"
	"net/http"
	"time"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Travix-International/logger"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/mock"
)

/* http.ResponseWriter mock */

type MockResponseWriter struct {
	mock.Mock
	model.WrappedResponseWriter
	http.Flusher
}

func (m *MockResponseWriter) Header() http.Header {
	a := m.Called()
	return a.Get(0).(http.Header)
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	a := m.Called(b)
	return a.Get(0).(int), a.Error(1)
}

func (m *MockResponseWriter) WriteHeader(i int) {
	m.Called(i)
}

func (m *MockResponseWriter) JSON(statusCode int, content interface{}) {
	m.Called(statusCode, content)
}

func (m *MockResponseWriter) XML(statusCode int, content interface{}) {
	m.Called(statusCode, content)
}

func (m *MockResponseWriter) AcceptsXML(r *http.Request) bool {
	a := m.Called(r)
	return a.Bool(0)
}

func (m *MockResponseWriter) WriteResponse(r *http.Request, statusCode int, content interface{}) {
	m.Called(r, statusCode, content)
}

func (m *MockResponseWriter) SetCaching(maxAge int) {
	m.Called(maxAge)
}

func (m *MockResponseWriter) Status() int {
	a := m.Called()
	return a.Int(0)
}

func (m *MockResponseWriter) Flush() {
	m.Called()
}

/* io.Reader mock */

type MockReader struct {
	mock.Mock
	io.ReadCloser
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	a := m.Called(p)
	return a.Get(0).(int), a.Error(1)
}

func (m *MockReader) Close() error {
	a := m.Called()
	return a.Error(0)
}

/* model.Logger mock */

type MockLogger struct {
	mock.Mock
	model.Logger
}

func (m *MockLogger) Debug(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *MockLogger) Info(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *MockLogger) Warn(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *MockLogger) Error(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *MockLogger) GetLogger() *logger.Logger {
	i := m.Called()
	return i.Get(0).(*logger.Logger)
}

/* model.Metrics mock */

type (
	MockMetrics struct {
		mock.Mock
		model.Metrics
	}

	MockMetricsHistogram struct {
		mock.Mock
		model.MetricsHistogram
	}
)

func (m *MockMetricsHistogram) RecordTimeElapsed(start time.Time, unit time.Duration) {
	m.Called(start, unit)
}

func (m *MockMetrics) Count(subsystem, name, help string) {
	m.Called(subsystem, name, help)
}

func (m *MockMetrics) SetGauge(value float64, subsystem, name, help string) {
	m.Called(value, subsystem, name, help)
}

func (m *MockMetrics) CountLabels(subsystem, name, help string, labels, values []string) {
	m.Called(subsystem, name, help, labels, values)
}

func (m *MockMetrics) IncreaseCounter(subsystem, name, help string, increment int) {
	m.Called(subsystem, name, help, increment)
}

func (m *MockMetrics) AddHistogram(subsystem, name, help string) model.MetricsHistogram {
	a := m.Called(subsystem, name, help)
	return a.Get(0).(model.MetricsHistogram)
}

/* model.VersionBuilder mock */

type MockVersionBuilder struct {
	mock.Mock
	model.VersionBuilder
}

func (m *MockVersionBuilder) ToString() string {
	a := m.Called()
	return a.String(0)
}

func (m *MockVersionBuilder) ToMap() map[string]string {
	a := m.Called()
	return a.Get(0).(map[string]string)
}

/* model.MiddlewareWrapper mock */

type MockMiddlewareWrapper struct {
	mock.Mock
	model.MiddlewareWrapper
}

func (m *MockMiddlewareWrapper) Wrap(subsystem, name string, middleware model.Middleware, handler model.Handle) model.Handle {
	a := m.Called(subsystem, name, middleware, handler)
	return a.Get(0).(func(model.WrappedResponseWriter, *http.Request, model.RouterParams))
}

/* model.RouterFactory mock */

type MockRouterFactory struct {
	mock.Mock
	model.RouterFactory
}

func (m *MockRouterFactory) CreateRouter() *model.Router {
	a := m.Called()
	return a.Get(0).(*model.Router)
}

/* model.ServiceHandlerFactory mock */

type MockServiceHandlerFactory struct {
	mock.Mock
	model.ServiceHandlerFactory
}

func (m *MockServiceHandlerFactory) WrapHandler(subsystem, name string, middlewares []model.Middleware, handle model.Handle) httprouter.Handle {
	a := m.Called(subsystem, name, middlewares, handle)
	return a.Get(0).(httprouter.Handle)
}

func (m *MockServiceHandlerFactory) CreateRootHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateReadinessHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateLivenessHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateQuitHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateHealthHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateVersionHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

func (m *MockServiceHandlerFactory) CreateMetricsHandler() model.Handle {
	a := m.Called()
	return a.Get(0).(model.Handle)
}

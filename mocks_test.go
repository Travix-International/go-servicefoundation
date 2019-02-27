package servicefoundation_test

import (
	"io"
	"net/http"
	"time"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/mock"
)

/* http.ResponseWriter mock */

type mockResponseWriter struct {
	mock.Mock
	sf.WrappedResponseWriter
	http.Flusher
}

func (m *mockResponseWriter) Header() http.Header {
	a := m.Called()
	return a.Get(0).(http.Header)
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	a := m.Called(b)
	return a.Get(0).(int), a.Error(1)
}

func (m *mockResponseWriter) WriteHeader(i int) {
	m.Called(i)
}

func (m *mockResponseWriter) JSON(statusCode int, content interface{}) {
	m.Called(statusCode, content)
}

func (m *mockResponseWriter) XML(statusCode int, content interface{}) {
	m.Called(statusCode, content)
}

func (m *mockResponseWriter) AcceptsXML(r *http.Request) bool {
	a := m.Called(r)
	return a.Bool(0)
}

func (m *mockResponseWriter) WriteResponse(r *http.Request, statusCode int, content interface{}) {
	m.Called(r, statusCode, content)
}

func (m *mockResponseWriter) SetCaching(maxAge int) {
	m.Called(maxAge)
}

func (m *mockResponseWriter) Status() int {
	a := m.Called()
	return a.Int(0)
}

func (m *mockResponseWriter) Flush() {
	m.Called()
}

/* io.Reader mock */

type mockReader struct {
	mock.Mock
	io.ReadCloser
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	a := m.Called(p)
	return a.Get(0).(int), a.Error(1)
}

func (m *mockReader) Close() error {
	a := m.Called()
	return a.Error(0)
}

/* sf.LogFactory mock */

type mockLogFactory struct {
	mock.Mock
	sf.LogFactory
}

func (m *mockLogFactory) NewLogger(meta map[string]string) sf.Logger {
	i := m.Called(meta)
	return i.Get(0).(sf.Logger)
}

/* sf.Logger mock */

type mockLogger struct {
	mock.Mock
	sf.Logger
}

func (m *mockLogger) Debug(event, formatOrMsg string, a ...interface{}) {
	m.Called(event, formatOrMsg, a)
}

func (m *mockLogger) Info(event, formatOrMsg string, a ...interface{}) {
	m.Called(event, formatOrMsg, a)
}

func (m *mockLogger) Warn(event, formatOrMsg string, a ...interface{}) {
	m.Called(event, formatOrMsg, a)
}

func (m *mockLogger) Error(event, formatOrMsg string, a ...interface{}) {
	m.Called(event, formatOrMsg, a)
}

func (m *mockLogger) GetLogger() *logger.Logger {
	i := m.Called()
	return i.Get(0).(*logger.Logger)
}

func (m *mockLogger) AddMeta(meta map[string]string) {
	m.Called(meta)
}

/* sf.Metrics mock */

type (
	mockMetrics struct {
		mock.Mock
		sf.Metrics
	}

	mockHistogramVec struct {
		mock.Mock
		sf.HistogramVec
	}

	mockSummaryVec struct {
		mock.Mock
		sf.SummaryVec
	}
)

func (m *mockHistogramVec) RecordTimeElapsed(start time.Time) {
	m.Called(start)
}

func (m *mockHistogramVec) RecordDuration(start time.Time, unit time.Duration) {
	m.Called(start, unit)
}

func (m *mockSummaryVec) RecordTimeElapsed(start time.Time) {
	m.Called(start)
}

func (m *mockSummaryVec) RecordDuration(start time.Time, unit time.Duration) {
	m.Called(start, unit)
}

func (m *mockMetrics) Count(subsystem, name, help string) {
	m.Called(subsystem, name, help)
}

func (m *mockMetrics) SetGauge(value float64, subsystem, name, help string) {
	m.Called(value, subsystem, name, help)
}

func (m *mockMetrics) CountLabels(subsystem, name, help string, labels, values []string) {
	m.Called(subsystem, name, help, labels, values)
}

func (m *mockMetrics) IncreaseCounter(subsystem, name, help string, increment int) {
	m.Called(subsystem, name, help, increment)
}

func (m *mockMetrics) AddHistogramVec(subsystem, name, help string, labels, labelValues []string) sf.HistogramVec {
	a := m.Called(subsystem, name, help, labels, labelValues)
	return a.Get(0).(sf.HistogramVec)
}

func (m *mockMetrics) AddSummaryVec(subsystem, name, help string, labels, labelValues []string) sf.SummaryVec {
	a := m.Called(subsystem, name, help, labels, labelValues)
	return a.Get(0).(sf.SummaryVec)
}

/* sf.VersionBuilder mock */

type mockVersionBuilder struct {
	mock.Mock
	sf.VersionBuilder
}

func (m *mockVersionBuilder) ToString() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockVersionBuilder) ToMap() map[string]string {
	a := m.Called()
	return a.Get(0).(map[string]string)
}

/* sf.MiddlewareWrapperFactory mock */

type mockMiddlewareWrapperFactory struct {
	mock.Mock
	sf.MiddlewareWrapperFactory
}

func (m *mockMiddlewareWrapperFactory) NewMiddlewareWrapper(corsOptions *sf.CORSOptions, authFunc sf.AuthorizationFunc) sf.MiddlewareWrapper {
	a := m.Called(corsOptions, authFunc)
	return a.Get(0).(sf.MiddlewareWrapper)
}

/* sf.MiddlewareWrapper mock */

type mockMiddlewareWrapper struct {
	mock.Mock
	sf.MiddlewareWrapper
}

func (m *mockMiddlewareWrapper) Wrap(subsystem, name string, middleware sf.Middleware, handler sf.Handle, metaFunc sf.MetaFunc) sf.Handle {
	a := m.Called(subsystem, name, middleware, handler, metaFunc)
	return a.Get(0).(func(sf.WrappedResponseWriter, *http.Request, sf.HandlerUtils))
}

/* sf.RouterFactory mock */

type mockRouterFactory struct {
	mock.Mock
	sf.RouterFactory
}

func (m *mockRouterFactory) NewRouter() sf.Router {
	a := m.Called()
	return a.Get(0).(sf.Router)
}

/* sf.Router mock */

type mockRouter struct {
	mock.Mock
	sf.Router
}

func (m *mockRouter) Handle(method, path string, metaFunc sf.MetaFunc, handle sf.Handle) {
	m.Called(method, path, metaFunc, handle)
}

func (m *mockRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

/* sf.ServiceHandlerFactory mock */

type mockServiceHandlerFactory struct {
	mock.Mock
	sf.ServiceHandlerFactory
}

func (m *mockServiceHandlerFactory) Wrap(subsystem, name string, middlewares []sf.Middleware, handle sf.Handle,
	metaFunc sf.MetaFunc) httprouter.Handle {

	a := m.Called(subsystem, name, middlewares, handle, metaFunc)
	return a.Get(0).(httprouter.Handle)
}

func (m *mockServiceHandlerFactory) NewHandlers() sf.Handlers {
	a := m.Called()
	return a.Get(0).(sf.Handlers)
}

/* sf.RootHandler mock */

type mockRootHandler struct {
	mock.Mock
	sf.RootHandler
}

func (m *mockRootHandler) NewRootHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.ReadinessHandler mock */

type mockReadinessHandler struct {
	mock.Mock
	sf.ReadinessHandler
}

func (m *mockReadinessHandler) NewReadinessHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.LivenessHandler mock */

type mockLivenessHandler struct {
	mock.Mock
	sf.LivenessHandler
}

func (m *mockLivenessHandler) NewLivenessHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.QuitHandler mock */

type mockQuitHandler struct {
	mock.Mock
	sf.QuitHandler
}

func (m *mockQuitHandler) NewQuitHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.HealthHandler mock */

type mockHealthHandler struct {
	mock.Mock
	sf.HealthHandler
}

func (m *mockHealthHandler) NewHealthHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.RootHandler mock */

type mockVersionHandler struct {
	mock.Mock
	sf.VersionHandler
}

func (m *mockVersionHandler) NewVersionHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.MetricsHandler mock */

type mockMetricsHandler struct {
	mock.Mock
	sf.MetricsHandler
}

func (m *mockMetricsHandler) NewMetricsHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.PreFlightHandler mock */

type mockPreFlightHandler struct {
	mock.Mock
	sf.PreFlightHandler
}

func (m *mockPreFlightHandler) NewPreFlightHandler() sf.Handle {
	a := m.Called()
	return a.Get(0).(sf.Handle)
}

/* sf.ServiceStateReader mock */

type mockServiceStateReader struct {
	mock.Mock
	sf.ServiceStateReader
}

func (m *mockServiceStateReader) IsLive() bool {
	a := m.Called()
	return a.Bool(0)
}

func (m *mockServiceStateReader) IsReady() bool {
	a := m.Called()
	return a.Bool(0)
}

func (m *mockServiceStateReader) IsHealthy() bool {
	a := m.Called()
	return a.Bool(0)
}

/* sf.ServiceStateManager mock */

type mockServiceStateManager struct {
	mock.Mock
	sf.ServiceStateManager
}

func (m *mockServiceStateManager) IsLive() bool {
	a := m.Called()
	return a.Bool(0)
}

func (m *mockServiceStateManager) IsReady() bool {
	a := m.Called()
	return a.Bool(0)
}

func (m *mockServiceStateManager) IsHealthy() bool {
	a := m.Called()
	return a.Bool(0)
}

func (m *mockServiceStateManager) WarmUp() {
	m.Called()
}

func (m *mockServiceStateManager) ShutDown(logger sf.Logger) {
	m.Called(logger)
}

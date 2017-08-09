package site_test

import (
	"io"
	"net/http"
	"time"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/mock"
)

/* http.ResponseWriter mock */

type mockResponseWriter struct {
	mock.Mock
	model.WrappedResponseWriter
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

/* model.Logger mock */

type mockLogger struct {
	mock.Mock
	model.Logger
}

func (m *mockLogger) Debug(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *mockLogger) Info(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *mockLogger) Warn(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *mockLogger) Error(event, formatOrMsg string, a ...interface{}) error {
	i := m.Called(event, formatOrMsg, a)
	return i.Error(0)
}

func (m *mockLogger) GetLogger() *logger.Logger {
	i := m.Called()
	return i.Get(0).(*logger.Logger)
}

/* model.Metrics mock */

type (
	mockMetrics struct {
		mock.Mock
		model.Metrics
	}

	mockMetricsHistogram struct {
		mock.Mock
		model.MetricsHistogram
	}
)

func (m *mockMetricsHistogram) RecordTimeElapsed(start time.Time) {
	m.Called(start)
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

func (m *mockMetrics) AddHistogram(subsystem, name, help string) model.MetricsHistogram {
	a := m.Called(subsystem, name, help)
	return a.Get(0).(model.MetricsHistogram)
}

/* model.VersionBuilder mock */

type mockVersionBuilder struct {
	mock.Mock
	model.VersionBuilder
}

func (m *mockVersionBuilder) ToString() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockVersionBuilder) ToMap() map[string]string {
	a := m.Called()
	return a.Get(0).(map[string]string)
}

/* model.MiddlewareWrapper mock */

type mockMiddlewareWrapper struct {
	mock.Mock
	model.MiddlewareWrapper
}

func (m *mockMiddlewareWrapper) Wrap(subsystem, name string, middleware model.Middleware, handler model.Handle) model.Handle {
	a := m.Called(subsystem, name, middleware, handler)
	return a.Get(0).(func(model.WrappedResponseWriter, *http.Request, model.RouterParams))
}

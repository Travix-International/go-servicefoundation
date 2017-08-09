package logging_test

import (
	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/mock"
)

type mockLogger struct {
	mock.Mock
	model.Logger
}

func (l *mockLogger) Debug(event, formatOrMsg string, a ...interface{}) error {
	c := l.Called(event, formatOrMsg, a)
	return c.Error(0)
}

func (l *mockLogger) Info(event, formatOrMsg string, a ...interface{}) error {
	c := l.Called(event, formatOrMsg, a)
	return c.Error(0)
}

func (l *mockLogger) Warn(event, formatOrMsg string, a ...interface{}) error {
	c := l.Called(event, formatOrMsg, a)
	return c.Error(0)
}

func (l *mockLogger) Error(event, formatOrMsg string, a ...interface{}) error {
	c := l.Called(event, formatOrMsg, a)
	return c.Error(0)
}

func (l *mockLogger) setMinLevel(level string) {
	l.Called(level)
}

func (l *mockLogger) GetLogger() *logger.Logger {
	c := l.Called()
	return c.Get(0).(*logger.Logger)
}

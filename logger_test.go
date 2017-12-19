package servicefoundation_test

import (
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestLoggerImpl_GetLogger_DebugLevel(t *testing.T) {
	factory := sf.NewLogFactory("Debug")
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_ErrorLevel(t *testing.T) {
	factory := sf.NewLogFactory("Error")
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_UnknownLevel(t *testing.T) {
	factory := sf.NewLogFactory("Whatevah")
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_StaticMsg(t *testing.T) {
	factory := sf.NewLogFactory("Debug")
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg")
	sut.Info("event", "msg")
	sut.Warn("event", "msg")
	sut.Error("event", "msg")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

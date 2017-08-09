package logging_test

import (
	"testing"

	"github.com/Prutswonder/go-servicefoundation/logging"
	"github.com/stretchr/testify/assert"
)

func TestLoggerImpl_GetLogger_DebugLevel(t *testing.T) {
	sut := logging.CreateLogger("Debug")

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_ErrorLevel(t *testing.T) {
	sut := logging.CreateLogger("Error")

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_UnknownLevel(t *testing.T) {
	sut := logging.CreateLogger("Whatevah")

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_StaticMsg(t *testing.T) {
	sut := logging.CreateLogger("Debug")

	// Act
	sut.Debug("event", "msg")
	sut.Info("event", "msg")
	sut.Warn("event", "msg")
	sut.Error("event", "msg")
	logger := sut.GetLogger()

	assert.NotNil(t, logger)
}

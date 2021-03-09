package v8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerImpl_GetLogger_DebugLevel(t *testing.T) {
	factory := NewLogFactory("Debug", make(map[string]string))
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.(*loggerImpl).logger

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_ErrorLevel(t *testing.T) {
	factory := NewLogFactory("Error", make(map[string]string))
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.(*loggerImpl).logger

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_UnknownLevel(t *testing.T) {
	factory := NewLogFactory("Whatevah", make(map[string]string))
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.(*loggerImpl).logger

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_StaticMsg(t *testing.T) {
	factory := NewLogFactory("Debug", make(map[string]string))
	sut := factory.NewLogger(make(map[string]string))

	// Act
	sut.Debug("event", "msg")
	sut.Info("event", "msg")
	sut.Warn("event", "msg")
	sut.Error("event", "msg")
	logger := sut.(*loggerImpl).logger

	assert.NotNil(t, logger)
}

func TestLoggerImpl_GetLogger_CombineMetas(t *testing.T) {
	meta1 := make(map[string]string)
	meta1["key1"] = "value1"
	meta1["key2"] = "value2"
	meta2 := make(map[string]string)
	meta2["key3"] = "value3"
	meta1["key2"] = "value2b"
	factory := NewLogFactory("Debug", meta1)
	sut := factory.NewLogger(meta2)

	// Act
	sut.Debug("event", "msg %s %s", "arg1", "arg2")
	sut.Info("event", "msg %s %s", "arg1", "arg2")
	sut.Warn("event", "msg %s %s", "arg1", "arg2")
	sut.Error("event", "msg %s %s", "arg1", "arg2")
	logger := sut.(*loggerImpl).logger

	assert.NotNil(t, logger)
}

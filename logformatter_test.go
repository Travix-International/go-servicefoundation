package servicefoundation_test

import (
	"testing"

	"github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogFormatterImpl_Format(t *testing.T) {
	entry := &logger.Entry{
		Level:   "Debug",
		Event:   "Test",
		Message: "Test \"message\"",
		Meta:    make(map[string]string),
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.Nil(t, err)
	assert.Contains(t, actual, "\"messagetype\":\"Test\"")
	assert.Contains(t, actual, "\"level\":\"Debug\"")
	assert.Contains(t, actual, "\"message\":\"Test \\\"message\\\"\"")
	assert.NotContains(t, actual, "meta")
}

func TestLogFormatterImpl_EmptyLevel_Error(t *testing.T) {
	entry := &logger.Entry{
		Level:   "",
		Event:   "Test",
		Message: "Test \"message\"",
		Meta:    make(map[string]string),
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.NotNil(t, err)
	assert.Empty(t, actual)
}

func TestLogFormatterImpl_EmptyEvent_Error(t *testing.T) {
	entry := &logger.Entry{
		Level:   "Debug",
		Event:   "",
		Message: "Test \"message\"",
		Meta:    make(map[string]string),
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.NotNil(t, err)
	assert.Empty(t, actual)
}

func TestLogFormatterImpl_EmptyMessage_Format(t *testing.T) {
	entry := &logger.Entry{
		Level:   "Debug",
		Event:   "Test",
		Message: "",
		Meta:    make(map[string]string),
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.Nil(t, err)
	assert.Contains(t, actual, "\"messagetype\":\"Test\"")
	assert.Contains(t, actual, "\"level\":\"Debug\"")
	assert.Contains(t, actual, "\"message\":\"\"")
	assert.NotContains(t, actual, "meta")
}

func TestLogFormatterImpl_Format_NilEntry(t *testing.T) {
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(nil)

	assert.Nil(t, err)
	assert.Equal(t, "", actual)
}

func TestLogFormatterImpl_WithMetaProps_Format(t *testing.T) {
	meta := make(map[string]string)
	meta["entry.sessionId"] = "this-is-a-session"
	meta["entry.applicationGroup"] = "TestGroup"
	meta["entry.statusCode"] = "204"
	meta["something"] = "else"

	entry := &logger.Entry{
		Level:   "Debug",
		Event:   "Test",
		Message: "Test message",
		Meta:    meta,
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.Nil(t, err)
	assert.Contains(t, actual, "\"messagetype\":\"Test\"")
	assert.Contains(t, actual, "\"level\":\"Debug\"")
	assert.Contains(t, actual, "\"message\":\"Test message\"")
	assert.Contains(t, actual, "\"meta\":{\"something\":\"else\"}")
	assert.Contains(t, actual, "\"applicationGroup\":\"TestGroup\"")
	assert.Contains(t, actual, "\"statuscode\":204")
	assert.Contains(t, actual, "\"sessionId\":\"this-is-a-session\"")
	assert.NotContains(t, actual, "entry.")
}

func TestLogFormatterImpl_WithInvalidStatusCode_Format(t *testing.T) {
	meta := make(map[string]string)
	meta["entry.statusCode"] = "hmm"

	entry := &logger.Entry{
		Level:   "Debug",
		Event:   "Test",
		Message: "Test message",
		Meta:    meta,
	}
	sut := servicefoundation.NewLogFormatter()

	// Act
	actual, err := sut.Format(entry)

	assert.Nil(t, err)
	assert.Contains(t, actual, "\"messagetype\":\"Test\"")
	assert.Contains(t, actual, "\"level\":\"Debug\"")
	assert.Contains(t, actual, "\"message\":\"Test message\"")
	assert.NotContains(t, actual, "meta")
	assert.NotContains(t, actual, "statuscode")
}

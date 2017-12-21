package servicefoundation

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Travix-International/logger"
	"github.com/pkg/errors"
)

type (
	// LogFormatter formats the log entries in a LogStash-compatible JSON string
	LogFormatter interface {
		Format(entry *logger.Entry) (string, error)
	}

	logFormatterImpl struct {
	}

	flatLogEntry map[string]interface{}
)

var (
	logEntryMissingLevelError = errors.New("Missing level in log entry")
	logEntryMissingEventError = errors.New("Missing event in log entry")
)

/* LogFormatter implementation */

// NewLogFormatter instatiates a new log formatter.
func NewLogFormatter() LogFormatter {
	return &logFormatterImpl{}
}

func (f *logFormatterImpl) Format(entry *logger.Entry) (string, error) {
	if entry == nil {
		return "", nil
	}

	if entry.Level == "" {
		return "", logEntryMissingLevelError
	}
	if entry.Event == "" {
		return "", logEntryMissingEventError
	}

	var logEntry flatLogEntry = make(map[string]interface{})

	logEntry["level"] = entry.Level
	logEntry["type"] = "v2"
	logEntry["timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05.9999999Z")
	logEntry["loggername"] = "go-servicefoundation"
	logEntry["messagetype"] = entry.Event
	logEntry["message"] = entry.Message

	if statusCode := getStatusCode(entry); statusCode != nil {
		logEntry["statuscode"] = statusCode
	}

	if len(entry.Meta) > 0 {
		logEntry["meta"] = entry.Meta
	}

	appendMetaEntries(entry, logEntry)

	log, err := json.Marshal(logEntry)

	return string(log) + "\n", err
}

func appendMetaEntries(entry *logger.Entry, logEntry flatLogEntry) {
	keys := make([]string, len(entry.Meta))

	i := 0
	for k := range entry.Meta {
		keys[i] = k
		i++
	}

	for _, key := range keys {
		if !strings.HasPrefix(key, "entry.") {
			continue
		}
		logKey := key[6:]
		logEntry[logKey] = fetchMetaEntry(entry, key)
	}
}

func fetchMetaEntry(entry *logger.Entry, key string) string {
	value := entry.Meta[key]
	delete(entry.Meta, key)
	return value
}

func getStatusCode(entry *logger.Entry) *int {
	statusCode := fetchMetaEntry(entry, "entry.statusCode")

	if statusCode == "" {
		return nil
	}

	code, err := strconv.Atoi(statusCode)

	if err != nil {
		return nil
	}
	return &code
}

package servicefoundation

import (
	"fmt"
	"os"
	"strings"

	"github.com/Travix-International/logger"
)

const (
	minDebugLevel = 1
	minInfoLevel  = 2
	minWarnLevel  = 3
	defaultLevel  = 3 // Warning
)

type (
	// Logger is a wrapper around the Logger package and extending it with log level filtering and simplified formatting.
	Logger interface {
		Debug(event, formatOrMsg string, a ...interface{})
		Info(event, formatOrMsg string, a ...interface{})
		Warn(event, formatOrMsg string, a ...interface{})
		Error(event, formatOrMsg string, a ...interface{})
		GetLogger() *logger.Logger
	}

	loggerImpl struct {
		logMinLevel int
		logger      *logger.Logger
	}

	// LogFactory can be used to instantiate a new logger
	LogFactory interface {
		NewLogger(meta map[string]string) Logger
	}

	logFactoryImpl struct {
		baseMeta  map[string]string
		logFilter string
		logLevel  int
	}
)

var levels = []string{"debug", "info", "warning", "error"}

// NewLogFactory instantiates a new LogFactory implementation.
func NewLogFactory(logFilter string, baseMeta map[string]string) LogFactory {
	logLevel := 0
	levelFound := false
	lcLogFilter := strings.ToLower(logFilter)

	for i := 0; i < len(levels); i++ {
		logLevel = i + 1
		if lcLogFilter == levels[i] {
			levelFound = true
			break
		}
	}

	if !levelFound {
		logLevel = defaultLevel
	}

	return &logFactoryImpl{
		baseMeta:  baseMeta,
		logFilter: logFilter,
		logLevel:  logLevel,
	}
}

/* LogFactory implementation */

// NewLogger instantiates a new Logger implementation.
func (f *logFactoryImpl) NewLogger(meta map[string]string) Logger {
	logMeta := combineMetas(meta, f.baseMeta)
	log, _ := logger.New(logMeta)
	formatter := NewLogFormatter()
	consoleTransport := logger.NewTransport(os.Stdout, formatter)
	log.AddTransport(consoleTransport)

	return &loggerImpl{
		logger:      log,
		logMinLevel: f.logLevel,
	}
}

func combineMetas(meta1, meta2 map[string]string) map[string]string {
	meta := make(map[string]string)

	for key, value := range meta1 {
		meta[key] = value
	}

	for key, value := range meta2 {
		meta[key] = value
	}

	return meta
}

/* Logger implementation */

func (l *loggerImpl) Debug(event, formatOrMsg string, a ...interface{}) {
	if l.logMinLevel > minDebugLevel {
		return
	}

	if len(a) == 0 {
		_ = l.logger.Debug(event, formatOrMsg)
		return
	}
	_ = l.logger.Debug(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) Info(event, formatOrMsg string, a ...interface{}) {
	if l.logMinLevel > minInfoLevel {
		return
	}

	if len(a) == 0 {
		_ = l.logger.Info(event, formatOrMsg)
		return
	}
	msg := fmt.Sprintf(formatOrMsg, a...)
	_ = l.logger.Info(event, msg)
}

func (l *loggerImpl) Warn(event, formatOrMsg string, a ...interface{}) {
	if l.logMinLevel > minWarnLevel {
		return
	}

	if len(a) == 0 {
		_ = l.logger.Warn(event, formatOrMsg)
		return
	}
	_ = l.logger.Warn(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) Error(event, formatOrMsg string, a ...interface{}) {
	if len(a) == 0 {
		_ = l.logger.Error(event, formatOrMsg)
		return
	}
	_ = l.logger.Error(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) GetLogger() *logger.Logger {
	return l.logger
}

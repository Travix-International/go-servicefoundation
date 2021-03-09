package servicefoundation

import (
	"strings"

	log "github.com/Travix-International/go-log"
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
	}

	loggerImpl struct {
		logMinLevel int
		logger      log.Logger
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

	cfg := log.NewLoggerConfig()
	cfg.AppGroup = logMeta["entry.applicationgroup"]
	cfg.AppName = logMeta["entry.applicationname"]
	cfg.AppVersion = logMeta["entry.applicationversion"]
	cfg.HostName = logMeta["entry.machinename"]

	switch f.logLevel {
	case minDebugLevel:
		cfg.LogLevel = log.DebugLevel
	case minInfoLevel:
		cfg.LogLevel = log.InfoLevel
	default:
		cfg.LogLevel = log.WarnLevel
	}

	logger := log.NewLogger(cfg)

	return &loggerImpl{
		logger:      logger,
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
		l.logger.Debug(event).Log(formatOrMsg)
		return
	}
	l.logger.Debug(event).Logf(formatOrMsg, a...)
}

func (l *loggerImpl) Info(event, formatOrMsg string, a ...interface{}) {
	if l.logMinLevel > minInfoLevel {
		return
	}

	if len(a) == 0 {
		l.logger.Info(event).Log(formatOrMsg)
	}
	l.logger.Info(event).Logf(formatOrMsg, a...)
}

func (l *loggerImpl) Warn(event, formatOrMsg string, a ...interface{}) {
	if l.logMinLevel > minWarnLevel {
		return
	}

	if len(a) == 0 {
		l.logger.Warn(event).Log(formatOrMsg)
	}
	l.logger.Warn(event).Logf(formatOrMsg, a...)
}

func (l *loggerImpl) Error(event, formatOrMsg string, a ...interface{}) {
	if len(a) == 0 {
		l.logger.Error(event).Log(formatOrMsg)
	}
	l.logger.Error(event).Logf(formatOrMsg, a...)
}

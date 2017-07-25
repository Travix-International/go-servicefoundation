package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Travix-International/logger"
)

const (
	minDebugLevel = 1
	minInfoLevel  = 2
	minWarnLevel  = 3
	minErrorLevel = 4
)

type (
	loggerImpl struct {
		logMinLevel int
		logger      *logger.Logger
	}
)

var (
	loggerInstance *loggerImpl
	loggerMux      = &sync.Mutex{}
)

func CreateLogger(logMinFilter string) model.Logger {
	loggerMux.Lock()
	defer loggerMux.Unlock()

	if loggerInstance == nil {
		log := logger.New()
		consoleLogFormat := logger.NewStringFormat("[%s] ", "[%s] ", "%s\n", " (%s=", "%s)")
		consoleTransport := logger.NewTransport(os.Stdout, consoleLogFormat)
		log.AddTransport(consoleTransport)

		loggerInstance = &loggerImpl{
			logger: log,
		}
		loggerInstance.setMinLevel(logMinFilter)
	}
	return loggerInstance
}

/* Logger implementation */

func (l *loggerImpl) Debug(event, formatOrMsg string, a ...interface{}) error {
	if l.logMinLevel > minDebugLevel {
		return nil
	}

	if len(a) == 0 {
		return l.logger.Debug(event, formatOrMsg)
	}
	return l.logger.Debug(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) Info(event, formatOrMsg string, a ...interface{}) error {
	if l.logMinLevel > minInfoLevel {
		return nil
	}

	if len(a) == 0 {
		return l.logger.Info(event, formatOrMsg)
	}
	msg := fmt.Sprintf(formatOrMsg, a...)
	return l.logger.Info(event, msg)
}

func (l *loggerImpl) Warn(event, formatOrMsg string, a ...interface{}) error {
	if l.logMinLevel > minWarnLevel {
		return nil
	}

	if len(a) == 0 {
		return l.logger.Warn(event, formatOrMsg)
	}
	return l.logger.Warn(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) Error(event, formatOrMsg string, a ...interface{}) error {
	if len(a) == 0 {
		return l.logger.Error(event, formatOrMsg)
	}
	return l.logger.Error(event, fmt.Sprintf(formatOrMsg, a...))
}

func (l *loggerImpl) setMinLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		l.logMinLevel = minDebugLevel
	case "info":
		l.logMinLevel = minInfoLevel
	case "warning":
		l.logMinLevel = minWarnLevel
	case "error":
		l.logMinLevel = minErrorLevel
	default:
		l.logMinLevel = minWarnLevel
		l.Warn("LogMinLevel", "Failed parsing log level '%s', defaulting to 'Warning'", level)
	}
}

func (l *loggerImpl) GetLogger() *logger.Logger {
	return l.logger
}

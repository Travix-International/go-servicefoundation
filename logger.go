package servicefoundation

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Travix-International/logger"
)

const (
	minDebugLevel = 1
	minInfoLevel  = 2
	minWarnLevel  = 3
	defaultLevel  = "warning"
)

type (
	loggerImpl struct {
		logMinLevel int
		logger      *logger.Logger
	}
)

var (
	levels          = []string{"debug", "info", "warning", "error"}
	loggerInstances = make(map[string]*loggerImpl)
	once            sync.Once
)

func CreateLogger(logMinFilter string) Logger {
	once.Do(func() {
		for i, level := range levels {
			log := logger.New()
			consoleLogFormat := logger.NewStringFormat("[%s] ", "[%s] ", "%s\n", " (%s=", "%s)")
			consoleTransport := logger.NewTransport(os.Stdout, consoleLogFormat)
			log.AddTransport(consoleTransport)

			loggerInstances[level] = &loggerImpl{
				logger:      log,
				logMinLevel: i + 1,
			}
		}
	})
	return getLogInstance(logMinFilter)
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

func getLogInstance(level string) *loggerImpl {
	var inst *loggerImpl

	for i := 0; i < len(levels); i++ {
		if inst = loggerInstances[strings.ToLower(level)]; inst != nil {
			break
		}
	}

	if inst == nil {
		inst = loggerInstances[defaultLevel]
		inst.Warn("LogMinLevel", "Failed parsing log level '%s', defaulting to '%s'", defaultLevel)
	}
	return inst
}

func (l *loggerImpl) GetLogger() *logger.Logger {
	return l.logger
}

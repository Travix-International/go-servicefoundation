package servicefoundation

import (
	"time"

	"github.com/Travix-International/logger"
)

type (
	Logger interface {
		Debug(event, formatOrMsg string, a ...interface{}) error
		Info(event, formatOrMsg string, a ...interface{}) error
		Warn(event, formatOrMsg string, a ...interface{}) error
		Error(event, formatOrMsg string, a ...interface{}) error
		GetLogger() *logger.Logger
	}

	MetricsHistogram interface {
		RecordTimeElapsed(start time.Time, unit time.Duration)
	}

	Metrics interface {
		Count(subsystem, name, help string)
		SetGauge(value float64, subsystem, name, help string)
		CountLabels(subsystem, name, help string, labels, values []string)
		IncreaseCounter(subsystem, name, help string, increment int)
		AddHistogram(subsystem, name, help string) MetricsHistogram
	}
)

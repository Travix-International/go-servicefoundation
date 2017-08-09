package model

import (
	"time"
)

type MetricsHistogram interface {
	RecordTimeElapsed(start time.Time)
}

type Metrics interface {
	Count(subsystem, name, help string)
	SetGauge(value float64, subsystem, name, help string)
	CountLabels(subsystem, name, help string, labels, values []string)
	IncreaseCounter(subsystem, name, help string, increment int)
	AddHistogram(subsystem, name, help string) MetricsHistogram
}

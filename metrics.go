package servicefoundation

import (
	"time"

	"github.com/Travix-International/go-metrics"
)

type (
	metricsHistogramImpl struct {
		histogram *metrics.MetricsHistogram
	}

	metricsImpl struct {
		metrics *metrics.Metrics
	}
)

func CreateMetrics(namespace string, logger Logger) Metrics {
	return &metricsImpl{
		// We're not using the namespace in metrics, because we won't be able to write "basic" metrics.
		metrics: metrics.NewMetrics("", logger.GetLogger()),
	}
}

/* MetricsHistogram implementation */

func (h *metricsHistogramImpl) RecordTimeElapsed(start time.Time, unit time.Duration) {
	//TODO: Add unit as parameter to the go-metrics package
	h.histogram.RecordTimeElapsed(start)
}

/* Metrics implementation */

func (m *metricsImpl) Count(subsystem, name, help string) {
	m.metrics.Count(subsystem, name, help)
}

func (m *metricsImpl) SetGauge(value float64, subsystem, name, help string) {
	m.metrics.SetGauge(value, subsystem, name, help)
}

func (m *metricsImpl) CountLabels(subsystem, name, help string, labels, values []string) {
	m.metrics.CountLabels(subsystem, name, help, labels, values)
}

func (m *metricsImpl) IncreaseCounter(subsystem, name, help string, increment int) {
	m.metrics.IncreaseCounter(subsystem, name, help, increment)
}

func (m *metricsImpl) AddHistogram(subsystem, name, help string) MetricsHistogram {
	return &metricsHistogramImpl{m.metrics.AddHistogram(subsystem, name, help)}
}

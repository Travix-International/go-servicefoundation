package logging

import (
	"time"

	"github.com/Prutswonder/go-servicefoundation/model"
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

func CreateMetrics(namespace string, logger model.Logger) model.Metrics {
	return &metricsImpl{
		metrics: metrics.NewMetrics(namespace, logger.GetLogger()),
	}
}

/* MetricsHistogram implementation */

func (h *metricsHistogramImpl) RecordTimeElapsed(start time.Time) {
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

func (m *metricsImpl) AddHistogram(subsystem, name, help string) model.MetricsHistogram {
	return &metricsHistogramImpl{m.metrics.AddHistogram(subsystem, name, help)}
}

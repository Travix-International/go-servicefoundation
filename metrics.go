package servicefoundation

import (
	"strings"
	"time"

	"github.com/Travix-International/go-metrics"
)

type (
	// HistogramVec is a wrapper around the HistogramVec from the go-metrics package.
	HistogramVec interface {
		RecordTimeElapsed(start time.Time)
		RecordDuration(start time.Time, unit time.Duration)
	}

	// SummaryVec is a wrapper around the v from the go-metrics package.
	SummaryVec interface {
		RecordTimeElapsed(start time.Time)
		RecordDuration(start time.Time, unit time.Duration)
	}

	// Metrics is a wrapper around the Metrics from the go-metrics package.
	Metrics interface {
		Count(subsystem, name, help string)
		SetGauge(value float64, subsystem, name, help string)
		CountLabels(subsystem, name, help string, labels, values []string)
		IncreaseCounter(subsystem, name, help string, increment int)
		AddHistogramVec(subsystem, name, help string, labels, labelValues []string) HistogramVec
		AddSummaryVec(subsystem, name, help string, labels, labelValues []string) SummaryVec
	}

	histogramVecImpl struct {
		histogramVec *metrics.HistogramVec
	}

	summaryVecImpl struct {
		summaryVec *metrics.SummaryVec
	}

	metricsImpl struct {
		internalMetrics *metrics.Metrics
		externalMetrics *metrics.Metrics
	}
)

// NewMetrics instantiates a new Metrics implementation.
func NewMetrics(namespace string, logger Logger) Metrics {
	log := logger.GetLogger()

	return &metricsImpl{
		internalMetrics: metrics.NewMetrics("", log),
		externalMetrics: metrics.NewMetrics(strings.ToLower(namespace), log),
	}
}

/* HistogramVec implementation */

func (h *histogramVecImpl) RecordTimeElapsed(start time.Time) {
	h.histogramVec.RecordTimeElapsed(start)
}

func (h *histogramVecImpl) RecordDuration(start time.Time, unit time.Duration) {
	h.histogramVec.RecordDuration(start, unit)
}

/* SummaryVec implementation */

func (s *summaryVecImpl) RecordTimeElapsed(start time.Time) {
	s.summaryVec.RecordTimeElapsed(start)
}

func (s *summaryVecImpl) RecordDuration(start time.Time, unit time.Duration) {
	s.summaryVec.RecordDuration(start, unit)
}

/* Metrics implementation */

func (m *metricsImpl) Count(subsystem, name, help string) {
	m.getMetrics(subsystem).Count(subsystem, name, help)
}

func (m *metricsImpl) SetGauge(value float64, subsystem, name, help string) {
	m.getMetrics(subsystem).SetGauge(value, subsystem, name, help)
}

func (m *metricsImpl) CountLabels(subsystem, name, help string, labels, values []string) {
	m.getMetrics(subsystem).CountLabels(subsystem, name, help, labels, values)
}

func (m *metricsImpl) IncreaseCounter(subsystem, name, help string, increment int) {
	m.getMetrics(subsystem).IncreaseCounter(subsystem, name, help, increment)
}

func (m *metricsImpl) AddHistogramVec(subsystem, name, help string, labels, labelValues []string) HistogramVec {
	return &histogramVecImpl{m.getMetrics(subsystem).AddHistogramVec(subsystem, name, help, labels, labelValues)}
}

func (m *metricsImpl) AddSummaryVec(subsystem, name, help string, labels, labelValues []string) SummaryVec {
	return &summaryVecImpl{m.getMetrics(subsystem).AddSummaryVec(subsystem, name, help, labels, labelValues)}
}

func (m *metricsImpl) getMetrics(subsystem string) *metrics.Metrics {
	if subsystem == "" {
		return m.internalMetrics
	}
	return m.externalMetrics
}

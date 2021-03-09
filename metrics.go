package servicefoundation

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
		histogramVec *histogramVec
	}

	summaryVecImpl struct {
		summaryVec *summaryVec
	}

	metricsImpl struct {
		internalMetrics *metrics
		externalMetrics *metrics
	}
)

// NewMetrics instantiates a new Metrics implementation.
func NewMetrics(namespace string, logger Logger) Metrics {
	return &metricsImpl{
		internalMetrics: newMetrics("", logger),
		externalMetrics: newMetrics(strings.ToLower(namespace), logger),
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

func (m *metricsImpl) getMetrics(subsystem string) *metrics {
	if subsystem == "" {
		return m.internalMetrics
	}
	return m.externalMetrics
}

/*
	Internal metrics implementation (copied from github.com/Travix-International/go-metrics)
*/

type (
	// metrics provides a set of convenience functions that wrap Prometheus
	metrics struct {
		Namespace       string
		Counters        map[string]prometheus.Counter
		CounterVecs     map[string]*prometheus.CounterVec
		Summaries       map[string]prometheus.Summary
		SummaryVecs     map[string]*prometheus.SummaryVec
		Histograms      map[string]prometheus.Histogram
		HistogramVecs   map[string]*prometheus.HistogramVec
		Gauges          map[string]prometheus.Gauge
		Logger          Logger
		countMutex      *sync.RWMutex
		countVecMutex   *sync.RWMutex
		histMutex       *sync.RWMutex
		histVecMutex    *sync.RWMutex
		summaryVecMutex *sync.RWMutex
		gaugeMutex      *sync.RWMutex
	}

	// metricsHistogram combines a histogram and summary
	metricsHistogram struct {
		Key  string
		hist prometheus.Histogram
		sum  prometheus.Summary
	}

	// histogramVec wraps prometheus.HistogramVec
	histogramVec struct {
		Key         string
		Labels      []string
		LabelValues []string
		histVec     *prometheus.HistogramVec
	}

	// summaryVec wraps prometheus.SummaryVec
	summaryVec struct {
		Key         string
		Labels      []string
		LabelValues []string
		summaryVec  *prometheus.SummaryVec
	}
)

// newMetrics will instantiate a new Metrics wrapper object
func newMetrics(namespace string, logger Logger) *metrics {
	m := metrics{
		Namespace:       namespace,
		Logger:          logger,
		Counters:        make(map[string]prometheus.Counter),
		CounterVecs:     make(map[string]*prometheus.CounterVec),
		Histograms:      make(map[string]prometheus.Histogram),
		HistogramVecs:   make(map[string]*prometheus.HistogramVec),
		Summaries:       make(map[string]prometheus.Summary),
		SummaryVecs:     make(map[string]*prometheus.SummaryVec),
		Gauges:          make(map[string]prometheus.Gauge),
		countMutex:      &sync.RWMutex{},
		countVecMutex:   &sync.RWMutex{},
		histMutex:       &sync.RWMutex{},
		histVecMutex:    &sync.RWMutex{},
		summaryVecMutex: &sync.RWMutex{},
		gaugeMutex:      &sync.RWMutex{},
	}
	return &m
}

// defaultObjectives returns a default map of quantiles to be used in summaries.
func defaultObjectives() map[float64]float64 {
	return map[float64]float64{0.5: 0.05, 0.75: 0.025, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001, 0.999: 0.0001}
}

// Count increases the counter for the specified subsystem and name.
func (m *metrics) Count(subsystem, name, help string) {
	m.countMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	counter, exists := m.Counters[key]
	m.countMutex.RUnlock()

	if !exists {
		m.countMutex.Lock()
		if counter, exists = m.Counters[key]; !exists {
			counter = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: m.Namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      help,
			})
			m.Counters[key] = counter
			err := prometheus.Register(counter)
			if err != nil {
				m.Logger.
					Warn("MetricsCounterRegistrationFailed",
						"CounterHandler: Counter registration %v failed: %v", counter, err)
			}
		}
		m.countMutex.Unlock()
	}

	counter.Inc()
}

// SetGauge sets the gauge value for the specified subsystem and name.
func (m *metrics) SetGauge(value float64, subsystem, name, help string) {
	m.gaugeMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	gauge, exists := m.Gauges[key]
	m.gaugeMutex.RUnlock()

	if !exists {
		m.gaugeMutex.Lock()
		if gauge, exists = m.Gauges[key]; !exists {
			gauge = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: m.Namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      help,
			})
			m.Gauges[key] = gauge
			err := prometheus.Register(gauge)
			if err != nil {
				m.Logger.
					Warn("MetricsSetGaugeFailed",
						"SetGauge: Gauge registration %v failed: %v", gauge, err)
			}
		}
		m.gaugeMutex.Unlock()
	}

	gauge.Set(value)
}

// CountLabels increases the counter for the specified subsystem and name and adds the specified labels with values.
func (m *metrics) CountLabels(subsystem, name, help string, labels, values []string) {
	m.countVecMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	counter, exists := m.CounterVecs[key]
	m.countVecMutex.RUnlock()

	if !exists {
		m.countVecMutex.Lock()
		if counter, exists = m.CounterVecs[key]; !exists {
			counter = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: m.Namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      help,
			}, labels)
			m.CounterVecs[key] = counter
			err := prometheus.Register(counter)
			if err != nil {
				m.Logger.
					Warn("MetricsCounterLabelRegistrationFailed",
						"CounterLabelHandler: Counter registration %v failed: %v", counter, err)
			}
		}
		m.countVecMutex.Unlock()
	}

	counter.WithLabelValues(values...).Inc()
}

// IncreaseCounter increases the counter for the specified subsystem and name with the specified increment.
func (m *metrics) IncreaseCounter(subsystem, name, help string, increment int) {
	m.countMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	counter, exists := m.Counters[key]
	m.countMutex.RUnlock()

	if !exists {
		m.countMutex.Lock()
		if counter, exists = m.Counters[key]; !exists {
			counter = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: m.Namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      help,
			})
			m.Counters[key] = counter
			err := prometheus.Register(counter)
			if err != nil {
				m.Logger.
					Warn("MetricsIncreaseCounterRegistrationFailed",
						"CounterHandler: Counter registration failed: %v: %v", counter, err)
			}
		}
		m.countMutex.Unlock()
	}

	counter.Add(float64(increment))
}

// AddHistogram returns the MetricsHistogram for the specified subsystem and name.
func (m *metrics) AddHistogram(subsystem, name, help string) *metricsHistogram {
	return m.addHistogramWithBuckets(subsystem, name, help, prometheus.DefBuckets)
}

// AddHistogramWithCustomBuckets returns the MetricsHistogram for the specified subsystem and name with the specified buckets.
func (m *metrics) AddHistogramWithCustomBuckets(subsystem, name, help string, buckets []float64) *metricsHistogram {
	return m.addHistogramWithBuckets(subsystem, name, help, buckets)
}

func (m *metrics) addHistogramWithBuckets(subsystem, name, help string, buckets []float64) *metricsHistogram {
	m.histMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	sum, exists := m.Summaries[key]
	hist := m.Histograms[key]
	m.histMutex.RUnlock()

	if !exists {
		m.histMutex.Lock()
		if sum, exists = m.Summaries[key]; !exists {
			// todo: remove Summary creation/observation
			sum = prometheus.NewSummary(prometheus.SummaryOpts{
				Namespace:  m.Namespace,
				Subsystem:  subsystem,
				Name:       name + "_summary",
				Help:       help,
				Objectives: defaultObjectives(),
			})
			prometheus.MustRegister(sum)
			m.Summaries[key] = sum

			hist = prometheus.NewHistogram(prometheus.HistogramOpts{
				Namespace: m.Namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      help,
				Buckets:   buckets,
			})
			prometheus.MustRegister(hist)
			m.Histograms[key] = hist
		}
		m.histMutex.Unlock()
	}

	mh := metricsHistogram{
		Key:  key,
		hist: hist,
		sum:  sum,
	}
	return &mh
}

// AddHistogramVec returns the HistogramVec for the specified subsystem and name.
func (m *metrics) AddHistogramVec(subsystem, name, help string, labels, labelValues []string) *histogramVec {
	return m.addHistogramVecWithBuckets(subsystem, name, help, labels, labelValues, prometheus.DefBuckets)
}

// AddHistogramVecWithCustomBuckets returns the HistogramVec for the specified subsystem and name with the specified buckets.
func (m *metrics) AddHistogramVecWithCustomBuckets(subsystem, name, help string, labels, labelValues []string,
	buckets []float64) *histogramVec {

	return m.addHistogramVecWithBuckets(subsystem, name, help, labels, labelValues, buckets)
}

func (m *metrics) addHistogramVecWithBuckets(subsystem, name, help string, labels, labelValues []string,
	buckets []float64) *histogramVec {

	m.histVecMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	vec, exists := m.HistogramVecs[key]
	m.histVecMutex.RUnlock()

	if !exists {
		m.histVecMutex.Lock()
		vec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		}, labels)
		prometheus.MustRegister(vec)
		m.HistogramVecs[key] = vec
		m.histVecMutex.Unlock()
	}

	mh := histogramVec{
		Key:         key,
		Labels:      labels,
		LabelValues: labelValues,
		histVec:     vec,
	}
	return &mh
}

// AddSummaryVec returns the SummaryVec for the specified subsystem and name.
func (m *metrics) AddSummaryVec(subsystem, name, help string, labels, labelValues []string) *summaryVec {
	return m.addSummaryVecWithObjectives(subsystem, name, help, labels, labelValues, defaultObjectives())
}

// AddSummaryVecWithCustomObjectives returns the SummaryVec for the specified subsystem and name with the specified objectives.
func (m *metrics) AddSummaryVecWithCustomObjectives(subsystem, name, help string, labels, labelValues []string,
	objectives map[float64]float64) *summaryVec {

	return m.addSummaryVecWithObjectives(subsystem, name, help, labels, labelValues, objectives)
}

func (m *metrics) addSummaryVecWithObjectives(subsystem, name, help string, labels, labelValues []string,
	objectives map[float64]float64) *summaryVec {

	m.summaryVecMutex.RLock()
	key := fmt.Sprintf("%s/%s", subsystem, name)
	vec, exists := m.SummaryVecs[key]
	m.summaryVecMutex.RUnlock()

	if !exists {
		m.summaryVecMutex.Lock()
		vec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  m.Namespace,
			Subsystem:  subsystem,
			Name:       name,
			Help:       help,
			Objectives: objectives,
		}, labels)
		prometheus.MustRegister(vec)
		m.SummaryVecs[key] = vec
		m.summaryVecMutex.Unlock()
	}

	mh := summaryVec{
		Key:         key,
		Labels:      labels,
		LabelValues: labelValues,
		summaryVec:  vec,
	}
	return &mh
}

// RecordTimeElapsed adds the elapsed time since the specified start to the histogram in seconds and to the linked
// summary in milliseconds.
func (histogram *metricsHistogram) RecordTimeElapsed(start time.Time) {
	elapsed := float64(time.Since(start).Seconds())
	histogram.hist.Observe(elapsed)         // The default histogram buckets are recorded in seconds
	histogram.sum.Observe(elapsed * 1000.0) // While we have summaries in milliseconds
}

// RecordDuration adds the elapsed time since the specified start to the histogram in the specified unit of time
// and to the linked summary in milliseconds.
func (histogram *metricsHistogram) RecordDuration(start time.Time, unit time.Duration) {
	since := time.Since(start)
	elapsedMilliseconds := float64(since.Truncate(time.Millisecond))
	elapsedUnits := float64(since.Truncate(unit))

	histogram.hist.Observe(elapsedUnits)
	histogram.sum.Observe(elapsedMilliseconds)
}

// Observe adds the specified value to the histogram.
func (histogram *metricsHistogram) Observe(value float64) {
	histogram.hist.Observe(value)
}

// RecordTimeElapsed adds the elapsed time since the specified start to the histogram in seconds.
func (vec *histogramVec) RecordTimeElapsed(start time.Time) {
	elapsed := float64(time.Since(start).Seconds())
	vec.Observe(elapsed) // The default histogram buckets are recorded in seconds
}

// RecordDuration adds the elapsed time since the specified start to the histogram in the specified unit of time.
func (vec *histogramVec) RecordDuration(start time.Time, unit time.Duration) {
	since := time.Since(start)
	elapsedUnits := float64(since.Truncate(unit))

	vec.Observe(elapsedUnits)
}

// Observe adds the specified value to the histogram.
func (vec *histogramVec) Observe(value float64) {
	vec.histVec.WithLabelValues(vec.LabelValues...).Observe(value)
}

// RecordTimeElapsed adds the elapsed time since the specified start to the summary in milliseconds.
func (vec *summaryVec) RecordTimeElapsed(start time.Time) {
	since := time.Since(start)
	elapsed := float64(since.Truncate(time.Millisecond))
	vec.summaryVec.WithLabelValues(vec.LabelValues...).Observe(elapsed)
}

// RecordDuration adds the elapsed time since the specified start to the histogram in the specified unit of time.
func (vec *summaryVec) RecordDuration(start time.Time, unit time.Duration) {
	since := time.Since(start)
	elapsedUnits := float64(since.Truncate(unit))

	vec.summaryVec.WithLabelValues(vec.LabelValues...).Observe(elapsedUnits)
}

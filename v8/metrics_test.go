package v8_test

import (
	"testing"
	"time"

	sf "github.com/Travix-International/go-servicefoundation/v8"
	"github.com/stretchr/testify/assert"
)

func TestMetricsImpl(t *testing.T) {
	logger := mockLogger{}
	sut := sf.NewMetrics("testcount", &logger)

	// Act
	sut.Count("sub", "count", "help")
	sut.IncreaseCounter("sub", "inc", "help", 5)
	sut.CountLabels("", "lbl", "help", []string{"a", "b", "c"}, []string{"1", "2", "3"})
	sut.SetGauge(float64(55), "sub", "gauge", "help")

	h := sut.AddHistogramVec("sub", "hist", "help", []string{"a", "b", "c"}, []string{"1", "2", "3"})
	h.RecordTimeElapsed(time.Now())
	h.RecordDuration(time.Now(), time.Microsecond)

	s := sut.AddSummaryVec("sub", "sum", "help", []string{"a", "b", "c"}, []string{"1", "2", "3"})
	s.RecordTimeElapsed(time.Now())
	s.RecordDuration(time.Now(), time.Millisecond)

	assert.NotNil(t, h)
	assert.NotNil(t, s)
	logger.AssertExpectations(t)
}

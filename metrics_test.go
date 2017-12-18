package servicefoundation_test

import (
	"testing"
	"time"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/assert"
)

func TestMetricsImpl(t *testing.T) {
	log := &mockLogger{}
	log.
		On("GetLogger").
		Return(logger.New(make(map[string]string)))
	sut := sf.NewMetrics("testcount", log)

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

	assert.NotNil(t, h)
	assert.NotNil(t, s)
	log.AssertExpectations(t)
}

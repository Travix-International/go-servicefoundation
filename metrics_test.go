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
		Return(logger.New())
	sut := sf.NewMetrics("testcount", log)

	// Act
	sut.Count("sub", "count", "help")
	sut.IncreaseCounter("sub", "inc", "help", 5)
	sut.CountLabels("sub", "lbl", "help", []string{"a", "b", "c"}, []string{"1", "2", "3"})
	sut.SetGauge(float64(55), "sub", "gauge", "help")
	h := sut.AddHistogram("sub", "hist", "help")
	h.RecordTimeElapsed(time.Now())
	h.RecordDuration(time.Now(), time.Microsecond)

	assert.NotNil(t, h)
	log.AssertExpectations(t)
}

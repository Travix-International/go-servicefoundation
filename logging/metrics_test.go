package logging_test

import (
	"testing"

	"github.com/Prutswonder/go-servicefoundation/logging"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestMetricsImpl(t *testing.T) {
	log := &mockLogger{}
	log.
		On("GetLogger").
		Return(logger.New())
	sut := logging.CreateMetrics("testcount", log)

	// Act
	sut.Count("sub", "count", "help")
	sut.IncreaseCounter("sub", "inc", "help", 5)
	sut.CountLabels("sub", "lbl", "help", []string{"a", "b", "c"}, []string{"1", "2", "3"})
	sut.SetGauge(float64(55), "sub", "gauge", "help")
	h := sut.AddHistogram("sub", "hist", "help")
	h.RecordTimeElapsed(time.Now(), time.Microsecond)

	assert.NotNil(t, h)
	log.AssertExpectations(t)
}

package servicefoundation_test

import (
	"testing"

	servicefoundation "github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
	"github.com/stretchr/testify/assert"
)

var loggy, _ = logger.New(make(map[string]string))

func TestParseBuildDate(t *testing.T) {
	buildDate := "Mon.January.2.2006.15:04:05.-0700.MST"

	// act
	parsedBuildDate := servicefoundation.ParseBuildDate(buildDate, loggy)

	assert.True(t, parsedBuildDate.Year() == 2006)
}

func TestParseBuildDateOtherDate(t *testing.T) {
	buildDate := "Wed.February.9.1977.15:04:05.-0700.MST"

	// act
	parsedBuildDate := servicefoundation.ParseBuildDate(buildDate, loggy)

	assert.True(t, parsedBuildDate.Day() == 9)
}

func TestParseBuildDateErr(t *testing.T) {
	buildDate := "Thu.February.31.1977.15:04:05.-0700.MST"

	// act
	parsedBuildDate := servicefoundation.ParseBuildDate(buildDate, loggy)

	assert.False(t, parsedBuildDate.Year() == 1977)
}

func BenchmarkParseBuildDate(b *testing.B) {
	buildDate := "Wed.February.9.1977.15:04:05.-0700.MST"

	//benchmark
	for i := 0; i < b.N; i++ {
		result = servicefoundation.ParseBuildDate(buildDate, loggy)
	}
}

func BenchmarkParseBuildDateErr(b *testing.B) {
	loggy, _ = logger.New(make(map[string]string))

	buildDate := "Thu.February.31.1977.15:04:05.-0700.MST"

	//benchmark
	for i := 0; i < b.N; i++ {
		result = servicefoundation.ParseBuildDate(buildDate, loggy)
	}
}

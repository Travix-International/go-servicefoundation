package servicefoundation_test

import (
	"testing"

	sf "github.com/Prutswonder/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestCreateDefaultVersionBuilder(t *testing.T) {
	sut := sf.NewVersionBuilder()

	actual := sut.ToString()
	actualMap := sut.ToMap()

	assert.Equal(t, "version: ? - buildDate: ? - git hash: ?", actual)
	assert.EqualValues(t, map[string]string{
		"version":   "?",
		"buildDate": "?",
		"gitHash":   "?",
	}, actualMap)
}

func TestCreateVersionBuilder(t *testing.T) {
	version := sf.BuildVersion{
		BuildDate:     "date",
		VersionNumber: "nmbr",
		GitHash:       "hash",
	}

	sut := sf.NewCustomVersionBuilder(version)

	actual := sut.ToString()
	actualMap := sut.ToMap()

	assert.Equal(t, "version: nmbr - buildDate: date - git hash: hash", actual)
	assert.EqualValues(t, map[string]string{
		"version":   "nmbr",
		"buildDate": "date",
		"gitHash":   "hash",
	}, actualMap)
}

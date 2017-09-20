package servicefoundation_test

import (
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestCreateVersionBuilder(t *testing.T) {
	version := sf.BuildVersion{
		BuildDate:     "date",
		VersionNumber: "nmbr",
		GitHash:       "hash",
	}

	sut := sf.NewVersionBuilder(version)

	actual := sut.ToString()
	actualMap := sut.ToMap()

	assert.Equal(t, "version: nmbr - buildDate: date - git hash: hash", actual)
	assert.EqualValues(t, map[string]string{
		"version":   "nmbr",
		"buildDate": "date",
		"gitHash":   "hash",
	}, actualMap)
}

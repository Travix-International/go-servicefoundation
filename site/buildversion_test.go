package site_test

import (
	"testing"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/Prutswonder/go-servicefoundation/site"
	"github.com/stretchr/testify/assert"
)

func TestCreateDefaultVersionBuilder(t *testing.T) {
	sut := site.CreateDefaultVersionBuilder()

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
	version := model.BuildVersion{
		BuildDate:     "date",
		VersionNumber: "nmbr",
		GitHash:       "hash",
	}

	sut := site.CreateVersionBuilder(version)

	actual := sut.ToString()
	actualMap := sut.ToMap()

	assert.Equal(t, "version: nmbr - buildDate: date - git hash: hash", actual)
	assert.EqualValues(t, map[string]string{
		"version":   "nmbr",
		"buildDate": "date",
		"gitHash":   "hash",
	}, actualMap)
}

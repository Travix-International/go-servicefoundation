package env_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/Travix-International/go-servicefoundation/env"
	"github.com/stretchr/testify/assert"
)

func TestOrDefault(t *testing.T) {
	const (
		name     = "Test1"
		expected = "Value1"
	)
	os.Setenv(name, expected)

	// Act
	actual := env.OrDefault(name, "SomethingElse")

	assert.Equal(t, expected, actual)
}

func TestOrDefault_UseDefault(t *testing.T) {
	const (
		name     = "Test2"
		expected = "Value2"
	)

	// Act
	actual := env.OrDefault(name, expected)

	assert.Equal(t, expected, actual)
}

func TestList(t *testing.T) {
	const name = "Test3"

	os.Setenv(name, "A,B,C")

	// Act
	actual := env.List(name)

	assert.Equal(t, 3, len(actual))
	assert.Equal(t, "A", actual[0])
	assert.Equal(t, "B", actual[1])
	assert.Equal(t, "C", actual[2])
}

func TestList_NoList(t *testing.T) {
	const name = "Test4"

	os.Setenv(name, "ABC")

	// Act
	actual := env.List(name)

	assert.Equal(t, 1, len(actual))
	assert.Equal(t, "ABC", actual[0])
}

func TestListOrDefault(t *testing.T) {
	const name = "Test5"

	os.Setenv(name, "A,B,C")

	// Act
	actual := env.ListOrDefault(name, []string{"D", "E", "F"})

	assert.Equal(t, 3, len(actual))
	assert.Equal(t, "A", actual[0])
	assert.Equal(t, "B", actual[1])
	assert.Equal(t, "C", actual[2])
}

func TestListOrDefault_UseDefault(t *testing.T) {
	const name = "Test6"

	// Act
	actual := env.ListOrDefault(name, []string{"D", "E", "F"})

	assert.Equal(t, 3, len(actual))
	assert.Equal(t, "D", actual[0])
	assert.Equal(t, "E", actual[1])
	assert.Equal(t, "F", actual[2])
}

func TestAsInt(t *testing.T) {
	const (
		name     = "Test7"
		expected = 6
	)

	os.Setenv(name, strconv.Itoa(expected))

	// Act
	actual := env.AsInt(name, 7)

	assert.Equal(t, expected, actual)
}

func TestAsInt_UseDefault(t *testing.T) {
	const (
		name     = "Test8"
		expected = 4
	)

	// Act
	actual := env.AsInt(name, expected)

	assert.Equal(t, expected, actual)
}

package v8_test

import (
	"testing"

	sf "github.com/Travix-International/go-servicefoundation/v8"
	"github.com/stretchr/testify/assert"
)

func TestNewRouterFactory(t *testing.T) {
	sut := sf.NewRouterFactory()

	assert.NotNil(t, sut)

	actual := sut.NewRouter()

	assert.NotNil(t, actual)
}

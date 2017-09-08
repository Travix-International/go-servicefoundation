package servicefoundation_test

import (
	"testing"

	sf "github.com/Prutswonder/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestCreateRouterFactory(t *testing.T) {
	sut := sf.CreateRouterFactory()

	assert.NotNil(t, sut)

	actual := sut.CreateRouter()

	assert.NotNil(t, actual)
}

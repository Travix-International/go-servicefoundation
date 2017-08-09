package site_test

import (
	"testing"

	"github.com/Prutswonder/go-servicefoundation/site"
	"github.com/stretchr/testify/assert"
)

func TestCreateRouterFactory(t *testing.T) {
	sut := site.CreateRouterFactory()

	assert.NotNil(t, sut)

	actual := sut.CreateRouter()

	assert.NotNil(t, actual)
}

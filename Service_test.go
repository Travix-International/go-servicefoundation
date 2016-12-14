package servicefoundation_test

import (
	"testing"

	"net/http/httptest"
	"strconv"

	servicefoundation "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestSetCaching(t *testing.T) {
	const expectedAge int = 4352
	w := httptest.NewRecorder()

	// Act
	servicefoundation.SetCaching(w, expectedAge)

	assert.NotEmpty(t, w.Header().Get("Vary"))
	assert.Contains(t, w.Header().Get("Cache-Control"), strconv.Itoa(expectedAge))

}

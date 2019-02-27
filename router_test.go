package servicefoundation_test

import (
	"net/http"
	"testing"

	sf "github.com/Travix-International/go-servicefoundation"
	"github.com/stretchr/testify/assert"
)

func TestNewRouterFactory(t *testing.T) {
	logFactory := &mockLogFactory{}
	m := &mockMetrics{}

	sut := sf.NewRouterFactory(logFactory, m)

	assert.NotNil(t, sut)

	router := sut.NewRouter()

	assert.NotNil(t, router)

	metaFunc := func(_ *http.Request, _ sf.RouteParamsFunc) sf.Meta {
		return make(map[string]string)
	}
	handleCalled := false
	var rp sf.RouteParams
	handle := func(_ sf.WrappedResponseWriter, _ *http.Request, u sf.HandlerUtils) {
		rp = u.ParamsFunc()
		handleCalled = true
	}
	w := &mockResponseWriter{}
	r, _ := http.NewRequest("GET", "https://www.go.to/something/val1", nil)
	router.Handle("GET", "/something/:p1", metaFunc, handle)

	router.ServeHTTP(w, r)

	assert.True(t, handleCalled)
	assert.Equal(t, "val1", rp["p1"])
}

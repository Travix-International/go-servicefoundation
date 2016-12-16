package servicefoundation_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	servicefoundation "github.com/Travix-International/go-servicefoundation"
	"github.com/Travix-International/logger"
	"github.com/rs/cors"
	"github.com/stretchr/testify/assert"
)

func TestVersionHandler(t *testing.T) {
	expectedResponseMessage := `{"buildDate":"2006-01-02 15:04:05 -0700 MST","gitHash":"gitHash","version":"mainVersionminVersion"}
`
	buildDate := "Mon.January.2.2006.15:04:05.-0700.MST"
	parsedBuildDate, _ := time.Parse(buildDate, buildDate)

	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo",
		servicefoundation.AppVersion{
			MainVersion: "mainVersion",
			MinVersion:  "minVersion",
			BuildDate:   parsedBuildDate,
			GitHash:     "gitHash",
		}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	//act
	fn := servicefoundation.VersionHandler(ctx)
	fn(w, r)

	data, _ := ioutil.ReadAll(w.Body)
	actual := string(data)

	assert.Equal(t, expectedResponseMessage, actual)
}

func TestHealthHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	//act
	servicefoundation.HealthHandler(w, r)

	assert.True(t, w.Code == http.StatusOK)
}

func TestLivenessHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	//act
	servicefoundation.LivenessHandler(w, r)

	assert.True(t, w.Code == http.StatusOK)
}

func TestRootHandler(t *testing.T) {
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	//act
	handler := servicefoundation.RootHandler("root", ctx)
	handler(w, r)

	assert.True(t, w.Code == http.StatusOK)
	assert.Zero(t, w.Body.Len())
}

func TestQuitHandler(t *testing.T) {
	exitIsCalled := false
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	// we need to overwrite the function so it will not exit the testing routine
	servicefoundation.ExitFunc = func(int) {
		exitIsCalled = true
		return
	}

	//act
	servicefoundation.QuitHandler(w, r)

	assert.True(t, w.Code == http.StatusOK)
	assert.True(t, exitIsCalled)
}

func TestMetricsHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	//act
	servicefoundation.MetricsHandler(w, r)

	assert.True(t, w.Code == http.StatusOK)
}

func TestCounterHandler(t *testing.T) {
	const expectedSubsystem = "testSubSystem"
	const expectedName = "counterTest"
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	expectedHandlerCalled := false
	expectedHandler := func(http.ResponseWriter, *http.Request) {
		expectedHandlerCalled = true
	}

	//act
	handler := servicefoundation.CounterHandler(expectedSubsystem, expectedName, ctx, expectedHandler)
	handler(w, r)

	_, exist := ctx.Metrics().Counters["testsubsystem/get_countertest_requests_total"]
	assert.True(t, exist)
	assert.True(t, expectedHandlerCalled)
}

func TestDontCacheHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	expectedHandler := func(http.ResponseWriter, *http.Request) {}

	//act
	testHandle := servicefoundation.DontCacheHandler(expectedHandler)
	testHandle(w, r)

	assert.True(t, w.HeaderMap["Cache-Control"][0] == "max-age: 0, private")
}

func TestDontCacheContextHandler(t *testing.T) {
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	contextHandlerCalled := false
	contextHandler := func(servicefoundation.AppContext) http.HandlerFunc {
		return func(http.ResponseWriter, *http.Request) {
			contextHandlerCalled = true
		}
	}

	//act
	handler := servicefoundation.DontCacheContextHandler(ctx, contextHandler)
	handler(w, r)

	assert.True(t, contextHandlerCalled)
	assert.Equal(t, "max-age: 0, private", w.Header().Get("Cache-Control"))
}

func TestRootHandlerWithUnknownRoute(t *testing.T) {
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "/whoops", nil)
	w := httptest.NewRecorder()

	//act
	handler := servicefoundation.RootHandler("foo", ctx)
	handler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCorsHandler(t *testing.T) {
	const expectedOrigins = "my-origins"
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	corsOptions := &cors.Options{AllowedOrigins: []string{expectedOrigins}}
	innerHandlerCalled := false
	innerHandler := func(http.ResponseWriter, *http.Request) {
		innerHandlerCalled = true
	}

	//act
	handler := servicefoundation.CorsHandler(corsOptions, innerHandler)
	handler(w, r)

	assert.True(t, innerHandlerCalled)
}

func TestCorsHandlerFallbackOrigin(t *testing.T) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	corsOptions := &cors.Options{AllowedOrigins: []string{}}
	innerHandlerCalled := false
	innerHandler := func(http.ResponseWriter, *http.Request) {
		innerHandlerCalled = true
	}

	//act
	handler := servicefoundation.CorsHandler(corsOptions, innerHandler)
	handler(w, r)

	assert.True(t, innerHandlerCalled)
}

func TestHistogramHandler(t *testing.T) {
	const subsystem = "testSubSystem"
	const histogramName = "histogramTest"
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	innerHandlerCalled := false
	contextHandler := func(http.ResponseWriter, *http.Request) {
		innerHandlerCalled = true
	}

	//act
	handler := servicefoundation.HistogramHandler(subsystem, histogramName, ctx, contextHandler)
	handler(w, r)

	_, exist := ctx.Metrics().Histograms["testsubsystem/get_histogramtest_request_duration_milliseconds"]
	assert.True(t, exist)
	assert.True(t, innerHandlerCalled)
}

func TestHistogramContextHandler(t *testing.T) {
	const subsystem = "testSubSystem"
	const histogramName = "histogramContextTest"
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	innerHandlerCalled := false
	contextHandler := func(servicefoundation.AppContext) http.HandlerFunc {
		return func(http.ResponseWriter, *http.Request) {
			innerHandlerCalled = true
		}
	}

	//act
	handler := servicefoundation.HistogramContextHandler(subsystem, histogramName, ctx, contextHandler)
	handler(w, r)

	_, exist := ctx.Metrics().Histograms["testsubsystem/get_histogramcontexttest_request_duration_milliseconds"]
	assert.True(t, exist)
	assert.True(t, innerHandlerCalled)
}

// BENCHMARKS

// RESULT is a container for the result of the function called,
// this is to make sure the compiler is not going to optimize the function if the result is not used.
// see: http://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go
var result interface{}

func BenchmarkHistogramHandler(b *testing.B) {
	const subsystem = "testSubSystem"
	const histogramName = "histogramTest"
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	innerHandlerCalled := false
	contextHandler := func(http.ResponseWriter, *http.Request) {
		innerHandlerCalled = true
	}

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.HistogramHandler(subsystem, histogramName, ctx, contextHandler)
		handler(w, r)
	}
}

func BenchmarkCounterHandler(b *testing.B) {
	const expectedSubsystem = "testSubSystem"
	const expectedName = "counterTest"
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	expectedHandlerCalled := false
	expectedHandler := func(http.ResponseWriter, *http.Request) {
		expectedHandlerCalled = true
	}

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.CounterHandler(expectedSubsystem, expectedName, ctx, expectedHandler)
		handler(w, r)
	}
}

func BenchmarkRootHandler(b *testing.B) {
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "/whoops", nil)
	w := httptest.NewRecorder()

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.RootHandler("foo", ctx)
		handler(w, r)
	}
}

func BenchmarkHistogramContextHandler(b *testing.B) {
	const subsystem = "testSubSystem"
	const histogramName = "histogramContextTest"
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	innerHandlerCalled := false
	contextHandler := func(servicefoundation.AppContext) http.HandlerFunc {
		return func(http.ResponseWriter, *http.Request) {
			innerHandlerCalled = true
		}
	}

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.HistogramContextHandler(subsystem, histogramName, ctx, contextHandler)
		handler(w, r)
	}
}

func BenchmarkCorsHandler(b *testing.B) {
	const expectedOrigins = "my-origins"
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	corsOptions := &cors.Options{AllowedOrigins: []string{expectedOrigins}}
	innerHandlerCalled := false
	innerHandler := func(http.ResponseWriter, *http.Request) {
		innerHandlerCalled = true
	}

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.CorsHandler(corsOptions, innerHandler)
		handler(w, r)
	}
}

func BenchmarkDontCacheContextHandler(b *testing.B) {
	loggy, _ := logger.New(make(map[string]string))
	ctx := NewAppContext("foo", servicefoundation.AppVersion{}, loggy)
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	contextHandlerCalled := false
	contextHandler := func(servicefoundation.AppContext) http.HandlerFunc {
		return func(http.ResponseWriter, *http.Request) {
			contextHandlerCalled = true
		}
	}

	// benchmark
	for i := 0; i < b.N; i++ {
		handler := servicefoundation.DontCacheContextHandler(ctx, contextHandler)
		handler(w, r)
	}
}

func BenchmarkLivenessHandler(b *testing.B) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	// benchmark
	for i := 0; i < b.N; i++ {
		servicefoundation.LivenessHandler(w, r)
	}
}

func BenchmarkDontCacheHandler(b *testing.B) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()
	expectedHandler := func(http.ResponseWriter, *http.Request) {}

	// benchmark
	for i := 0; i < b.N; i++ {
		testHandle := servicefoundation.DontCacheHandler(expectedHandler)
		testHandle(w, r)
	}
}

func BenchmarkHealthHandler(b *testing.B) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	// benchmark
	for i := 0; i < b.N; i++ {
		servicefoundation.HealthHandler(w, r)
	}
}

func BenchmarkJSON(b *testing.B) {
	w := httptest.NewRecorder()

	//benchmark
	for i := 0; i < b.N; i++ {
		servicefoundation.JSON(w, 200, nil)
	}
}

func BenchmarkMetricsHandler(b *testing.B) {
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	//benchmark
	for i := 0; i < b.N; i++ {
		servicefoundation.MetricsHandler(w, r)
	}
}

func BenchmarkQuitHandler(b *testing.B) {
	exitIsCalled := false
	r, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	// we need to overwrite the function so it will not exit the testing routine
	servicefoundation.ExitFunc = func(int) {
		exitIsCalled = true
		return
	}

	//benchmark
	for i := 0; i < b.N; i++ {
		servicefoundation.QuitHandler(w, r)
	}
}

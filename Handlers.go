package servicefoundation

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Travix-International/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
)

const (
	counterNameTemplate    string = "%v_%v_requests_total"
	counterHelpTemplate    string = "Total number of %v requests to %v."
	histogramNameTemplate  string = "%v_%v_request_duration_milliseconds"
	histogramHelpTemplate  string = "Response times for %v requests to %v in milliseconds."
	statusCodeNameTemplate string = "%v_%v_response_statuscode_count"
	statusCodeHelpTemplate string = "Status code counts for %v responses to %v of %v."
	versionFormatString    string = "%s%s"
	dateFormatString       string = "%v"
)

type (
	// ContextHandler is an http handler that passes down the application context as an argument
	ContextHandler func(AppContext) http.HandlerFunc

	version_struct struct {
		version   string
		buildDate string
		gitHash   string
	}

	probe struct {
		Requests requests
	}
	requests struct {
		Counters    map[string]uint64
		Percentiles map[string]int64
	}
)

var (
	// We need this to overwrite for unittesting
	ExitFunc = os.Exit

	VersionHandler = func(ctx AppContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			version := ctx.Version()
			output := map[string]string{
				"version":   fmt.Sprintf(versionFormatString, version.MainVersion, version.MinVersion),
				"buildDate": fmt.Sprintf(dateFormatString, version.BuildDate),
				"gitHash":   version.GitHash,
			}
			JSON(w, http.StatusOK, output)
		}
	}
)

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, "ok")
}

func LivenessHandler(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, "ok")
}

func RootHandler(subsystem string, ctx AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			ctx.Metrics().Loggy.Warn("RootHandlerPathNotFound", fmt.Sprintf("Not found: %v %v %v", subsystem, r.Method, r.URL.Path))
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func QuitHandler(w http.ResponseWriter, _ *http.Request) {
	defer ExitFunc(0)

	w.WriteHeader(http.StatusOK)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	prometheus.Handler().ServeHTTP(w, r)
}

func CounterHandler(subsystem, name string, ctx AppContext, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverFunc(ctx.Metrics().Loggy)
		counterName := fmt.Sprintf(counterNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
		counterHelp := fmt.Sprintf(counterHelpTemplate, r.Method, name)

		ctx.Metrics().Count(strings.ToLower(subsystem), counterName, counterHelp)
		ww := NewWrappedResponseWriter(w)

		fn(ww, r)

		countStatus(ctx, ww, r, subsystem, name)
	}
}

func HistogramHandler(subsystem, name string, ctx AppContext, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverFunc(ctx.Metrics().Loggy)
		histogramName := fmt.Sprintf(histogramNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
		histogramHelp := fmt.Sprintf(histogramHelpTemplate, r.Method, name)
		ww := NewWrappedResponseWriter(w)

		func() {
			histogram := ctx.Metrics().AddHistogram(strings.ToLower(subsystem), histogramName, histogramHelp)
			start := time.Now()
			defer func() {
				histogram.RecordTimeElapsed(start)
			}()

			fn(ww, r)
		}()

		countStatus(ctx, ww, r, subsystem, name)
	}
}

func countStatus(ctx AppContext, ww *WrappedResponseWriter, r *http.Request, subsystem, name string) {
	statusName := fmt.Sprintf(statusCodeNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
	statusHelp := fmt.Sprintf(statusCodeHelpTemplate, r.Method, name, subsystem)
	ctx.Metrics().CountLabels(strings.ToLower(subsystem), statusName, statusHelp,
		[]string{"status", "method"}, []string{strconv.Itoa(ww.status), r.Method})
}

func HistogramContextHandler(subsystem, name string, ctx AppContext, fn ContextHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverFunc(ctx.Metrics().Loggy)
		histogramName := fmt.Sprintf(histogramNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
		histogramHelp := fmt.Sprintf(histogramHelpTemplate, r.Method, name)
		ww := NewWrappedResponseWriter(w)

		func() {
			histogram := ctx.Metrics().AddHistogram(strings.ToLower(subsystem), histogramName, histogramHelp)
			start := time.Now()
			defer func() {
				histogram.RecordTimeElapsed(start)
			}()

			execute := fn(ctx)
			execute(ww, r)
		}()

		countStatus(ctx, ww, r, subsystem, name)
	}
}

func DontCacheHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age: 0, private")
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
		w.Header().Set("Expires", time.Now().AddDate(-1, 0, 0).Format(http.TimeFormat))
		fn(w, r)
	}
}

func DontCacheContextHandler(ctx AppContext, fn ContextHandler) http.HandlerFunc {
	execute := fn(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age: 0, private")
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
		w.Header().Set("Expires", time.Now().AddDate(-1, 0, 0).Format(http.TimeFormat))
		execute(w, r)
	}
}

func CorsHandler(corsOptions *cors.Options, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(corsOptions.AllowedOrigins) == 0 {
			corsOptions.AllowedOrigins = append(corsOptions.AllowedOrigins, "*")
		}

		c := cors.New(*corsOptions)

		corsHandler := c.Handler(fn)
		corsHandler.ServeHTTP(w, r)
	}
}

var recoverFunc = func(loggy *logger.Logger) {
	if rec := recover(); rec != nil {
		loggy.Error("HandlerRecovery", fmt.Sprintf("PANIC recovered: %v", rec))
	}
}

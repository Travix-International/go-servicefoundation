package site

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/Prutswonder/go-servicefoundation/model"
	"github.com/rs/cors"
)

const (
	counterNameTemplate    string = "%v_%v_requests_total"
	counterHelpTemplate    string = "Total number of %v requests to %v."
	histogramNameTemplate  string = "%v_%v_request_duration_milliseconds"
	histogramHelpTemplate  string = "Response times for %v requests to %v in milliseconds."
	statusCodeNameTemplate string = "%v_%v_response_statuscode_count"
	statusCodeHelpTemplate string = "Status code counts for %v responses to %v of %v."
)

type middlewareWrapperImpl struct {
	logger      Logger
	metrics     Metrics
	corsOptions *cors.Options
}

func CreateMiddlewareWrapper(logger Logger, metrics Metrics, corsOptions *CORSOptions) MiddlewareWrapper {
	m := &middlewareWrapperImpl{
		logger:  logger,
		metrics: metrics,
	}
	m.corsOptions = m.mergeCORSOptions(corsOptions)
	return m
}

/* MiddlewareWrapper implementation */

func (m *middlewareWrapperImpl) Wrap(subsystem, name string, middleware Middleware, handler Handle) Handle {
	switch middleware {
	case CORS:
		return m.wrapWithCORS(subsystem, name, handler)
	case NoCaching:
		return m.wrapWithNoCache(subsystem, name, handler)
	case Counter:
		return m.wrapWithCounter(subsystem, name, handler)
	case Histogram:
		return m.wrapWithHistogram(subsystem, name, handler)
	case PanicTo500:
		return m.wrapWithPanicHandler(subsystem, name, handler)
	default:
		m.logger.Warn("UnhandledMiddleware", "Unhandled middleware: %v", middleware)
	}
	return handler
}

func (m *middlewareWrapperImpl) wrapWithCounter(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		counterName := fmt.Sprintf(counterNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
		counterHelp := fmt.Sprintf(counterHelpTemplate, r.Method, name)

		m.metrics.Count(subsystem, counterName, counterHelp)

		handler(w, r, p)

		m.countStatus(w, r, subsystem, name)
	}
}

func (m *middlewareWrapperImpl) wrapWithHistogram(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		histogramName := fmt.Sprintf(histogramNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
		histogramHelp := fmt.Sprintf(histogramHelpTemplate, r.Method, name)

		hist := m.metrics.AddHistogram(strings.ToLower(subsystem), histogramName, histogramHelp)
		start := time.Now()

		handler(w, r, p)

		hist.RecordTimeElapsed(start)
		m.countStatus(w, r, subsystem, name)
	}
}

func (m *middlewareWrapperImpl) wrapWithNoCache(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		w.Header().Set("Cache-Control", "max-age: 0, private")
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
		w.Header().Set("Expires", time.Now().AddDate(-1, 0, 0).Format(http.TimeFormat))

		handler(w, r, p)
	}
}

func (m *middlewareWrapperImpl) wrapWithCORS(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		c := cors.New(*m.corsOptions)

		h := func(ww http.ResponseWriter, r *http.Request) {
			w := CreateWrappedResponseWriter(ww)
			handler(w, r, p)
		}
		c.ServeHTTP(w, r, h)
	}
}
func (m *middlewareWrapperImpl) countStatus(w WrappedResponseWriter, r *http.Request, subsystem, name string) {
	statusName := fmt.Sprintf(statusCodeNameTemplate, strings.ToLower(r.Method), strings.ToLower(name))
	statusHelp := fmt.Sprintf(statusCodeHelpTemplate, r.Method, name, subsystem)
	m.metrics.CountLabels(strings.ToLower(subsystem), statusName, statusHelp,
		[]string{"status", "method"}, []string{strconv.Itoa(w.Status()), r.Method})
}

func (m *middlewareWrapperImpl) mergeCORSOptions(options *CORSOptions) *cors.Options {
	// TODO: de-duplicate elements in slices
	corsOptions := cors.Options{
		AllowedOrigins: options.AllowedOrigins,
		AllowedMethods: append(options.AllowedMethods, "HEAD", "OPTIONS"),
		AllowedHeaders: append(options.AllowedHeaders,
			"Origin", "Accept", "Content-Type", "X-Requested-With", "X-CSRF-Token"),
		AllowCredentials: true,
		ExposedHeaders: append(options.ExposedHeaders,
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Methods",
			"Access-Control-Max-Age",
			"Access-Control-Allow-Credentials",
			"Access-Controll-Allow-Origin"),
		MaxAge: options.MaxAge,
	}
	return &corsOptions
}

func (m *middlewareWrapperImpl) wrapWithPanicHandler(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		defer func() {
			if rec := recover(); rec != nil {
				m.logger.Error("PanicAutorecover", "PANIC recovered: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler(w, r, p)
	}
}

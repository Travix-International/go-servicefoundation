package servicefoundation

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
)

const (
	// CORS is a Middleware enumeration for validating cross-domain requests.
	CORS Middleware = 1
	// NoCaching is a middleware enumeration to adding no-caching headers to the response.
	NoCaching Middleware = 2
	// Counter is a middleware enumeration to add counter metrics to the current request/response.
	Counter Middleware = 3
	// Histogram is a middleware enumeration to add histogram metrics to the current request/response.
	Histogram Middleware = 4
	// PanicTo500 is a middleware enumeration to log panics as errors and respond with http status-code 500.
	PanicTo500 Middleware = 5
	// RequestLogging is a middleware enumeration to log the incoming request and response times.
	RequestLogging Middleware = 6
)

type (
	// Middleware is an enumeration to indicare the available middleware wrappers.
	Middleware int

	// MiddlewareWrapper is an interface to wrap an existing handler with the specified middleware.
	MiddlewareWrapper interface {
		Wrap(subsystem, name string, middleware Middleware, handler Handle) Handle
	}
)

type middlewareWrapperImpl struct {
	logger      Logger
	metrics     Metrics
	globals     ServiceGlobals
	corsOptions *cors.Options
}

// NewMiddlewareWrapper instantiates a new MiddelwareWrapper implementation.
func NewMiddlewareWrapper(logger Logger, metrics Metrics, corsOptions *CORSOptions, globals ServiceGlobals) MiddlewareWrapper {
	m := &middlewareWrapperImpl{
		logger:  logger,
		metrics: metrics,
		globals: globals,
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
		return m.wrapWithCounter("", name, handler)
	case Histogram:
		return m.wrapWithHistogram(subsystem, name, handler)
	case PanicTo500:
		return m.wrapWithPanicHandler(subsystem, name, handler)
	case RequestLogging:
		return m.wrapWithRequestLogging(subsystem, name, handler)
	default:
		m.logger.Warn("UnhandledMiddleware", "Unhandled middleware: %v", middleware)
	}
	return handler
}

func (m *middlewareWrapperImpl) wrapWithCounter(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		lcName := strings.ToLower(name)
		counterName := fmt.Sprintf("%v_total", lcName)
		counterHelp := fmt.Sprintf("Totals for %v.", name)

		m.metrics.CountLabels("", counterName, counterHelp,
			[]string{"app", "server", "env", "code", "method", "handler", "version", "subsystem"},
			[]string{
				m.globals.AppName,
				m.globals.ServerName,
				m.globals.DeployEnvironment,
				strconv.Itoa(w.Status()),
				strings.ToLower(r.Method),
				lcName,
				m.globals.VersionNumber,
				subsystem,
			},
		)

		handler(w, r, p)
	}
}

func (m *middlewareWrapperImpl) wrapWithHistogram(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		histogramName := fmt.Sprintf("%v_duration_milliseconds", strings.ToLower(name))
		histogramHelp := fmt.Sprintf("Response times for %v in milliseconds.", name)

		hist := m.metrics.AddHistogram(subsystem, histogramName, histogramHelp)
		start := time.Now()

		handler(w, r, p)

		hist.RecordTimeElapsed(start)
	}
}

func (m *middlewareWrapperImpl) wrapWithRequestLogging(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		lcName := strings.ToLower(name)
		log := m.logger
		start := time.Now()

		//TODO: Log message for requests
		//log.Info(fmt.Sprintf("Request-%s", name), "TODO")
		m.metrics.CountLabels("", "http_requests_total", "Total requests.",
			[]string{"app", "server", "env", "code", "method", "handler", "version", "subsystem"},
			[]string{
				m.globals.AppName,
				m.globals.ServerName,
				m.globals.DeployEnvironment,
				strconv.Itoa(w.Status()),
				strings.ToLower(r.Method),
				lcName,
				m.globals.VersionNumber,
				subsystem,
			},
		)
		histSeconds := m.metrics.AddHistogram("", "http_request_duration_seconds",
			"Response times for requests in seconds.")
		histMicroSeconds := m.metrics.AddHistogram("", "http_request_duration_microseconds",
			"Response times for requests in microseconds.")

		handler(w, r, p)

		elapsedMicroSeconds := time.Since(start).Nanoseconds() / int64(time.Microsecond)

		histMicroSeconds.RecordDuration(start, time.Microsecond)
		histSeconds.RecordTimeElapsed(start)

		//TODO: Log message for responses
		log.Info(fmt.Sprintf("Response-%s", name), "Elapsed (microsec): %d", elapsedMicroSeconds)
		m.metrics.CountLabels("", "http_responses_total", "Total responses.",
			[]string{"app", "server", "env", "code", "method", "handler", "version", "subsystem"},
			[]string{
				m.globals.AppName,
				m.globals.ServerName,
				m.globals.DeployEnvironment,
				strconv.Itoa(w.Status()),
				strings.ToLower(r.Method),
				lcName,
				m.globals.VersionNumber,
				subsystem,
			},
		)
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
			w := NewWrappedResponseWriter(ww)
			handler(w, r, p)
		}
		c.ServeHTTP(w, r, h)
	}
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
			"Access-Control-Allow-Origin"),
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

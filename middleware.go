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
	// RequestMetrics is a middleware enumeration to measure the incoming request and response times.
	RequestMetrics Middleware = 7
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
	log         Logger
	logFactory  LogFactory
	metrics     Metrics
	globals     ServiceGlobals
	corsOptions *cors.Options
}

// NewMiddlewareWrapper instantiates a new MiddelwareWrapper implementation.
func NewMiddlewareWrapper(logFactory LogFactory, metrics Metrics, corsOptions *CORSOptions, globals ServiceGlobals) MiddlewareWrapper {
	m := &middlewareWrapperImpl{
		log:        logFactory.NewLogger(make(map[string]string)),
		logFactory: logFactory,
		metrics:    metrics,
		globals:    globals,
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
	case RequestMetrics:
		return m.wrapWithRequestMetrics(subsystem, name, handler)
	default:
		m.log.Warn("UnhandledMiddleware", "Unhandled middleware: %v", middleware)
	}
	return handler
}

func (m *middlewareWrapperImpl) wrapWithCounter(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		counterName := fmt.Sprintf("%v_total", strings.ToLower(name))
		counterHelp := fmt.Sprintf("Totals for %v.", name)
		labels, values := m.getLabelsAndValues(subsystem, name, w, r)

		m.metrics.CountLabels("", counterName, counterHelp, labels, values)

		handler(w, r, p)
	}
}

func (m *middlewareWrapperImpl) wrapWithHistogram(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		histogramName := fmt.Sprintf("%v_duration_milliseconds", strings.ToLower(name))
		histogramHelp := fmt.Sprintf("Response times for %v in milliseconds.", name)

		labels, values := m.getLabelsAndValues(subsystem, name, w, r)
		hist := m.metrics.AddHistogramVec(subsystem, histogramName, histogramHelp, labels, values)
		start := time.Now()

		handler(w, r, p)

		hist.RecordTimeElapsed(start)
	}
}

func (m *middlewareWrapperImpl) wrapWithRequestLogging(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		meta := make(map[string]string)
		start := time.Now()

		log := m.getMetaLog(subsystem, name, nil, r, p, meta)
		log.Info("ApiRequest", m.getRequestStartMessage(r, p, meta))

		handler(w, r, p)

		elapsedMs := float64(time.Since(start).Nanoseconds()/int64(time.Microsecond)) / 1000.0
		durationMs := strconv.FormatFloat(elapsedMs, 'f', 3, 64)

		meta["entry.duration"] = durationMs
		log = m.getMetaLog(subsystem, name, w, r, p, meta)
		log.Info("ApiResponse", m.getRequestEndMessage(w, r, p, meta, durationMs))
	}
}

func (m *middlewareWrapperImpl) getRequestStartMessage(r *http.Request, p RouterParams, meta map[string]string) string {
	return fmt.Sprintf("%s %s", r.Method, meta["entry.http.url"])
}

func (m *middlewareWrapperImpl) getRequestEndMessage(w WrappedResponseWriter, r *http.Request, p RouterParams, meta map[string]string, durationMs string) string {
	status := strconv.Itoa(w.Status())
	contentType := w.Header().Get("content-type")

	return fmt.Sprintf("%s %s finished. Duration: %sms. Status: %s, ContentType: %s",
		r.Method,
		meta["entry.http.url"],
		durationMs,
		status,
		contentType,
	)
}

func (m *middlewareWrapperImpl) getMetaLog(subsystem, name string, w WrappedResponseWriter, r *http.Request, p RouterParams, meta map[string]string) Logger {
	m.addMetaEntry(meta, "http.method", r.Method)
	m.addMetaEntry(meta, "http.host", r.Host)

	url := r.RequestURI

	if r.URL != nil {
		scheme := "http"
		if r.URL.Scheme != "" {
			scheme = r.URL.Scheme
		} else if r.TLS != nil {
			scheme = "https"
		}
		url = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
		m.addMetaEntry(meta, "http.url", url)
		m.addMetaEntry(meta, "http.query", r.URL.RawQuery)
		m.addMetaEntry(meta, "http.route", r.URL.RawPath)
		m.addMetaEntry(meta, "http.scheme", scheme)
	}

	m.addMetaEntry(meta, "request", fmt.Sprintf("%s %s", r.Method, url))

	if w != nil {
		m.addMetaEntry(meta, "statuscode", strconv.Itoa(w.Status()))

		for key := range w.Header() {
			m.addMetaEntry(meta, "http.header."+strings.ToLower(key), w.Header().Get(key))
		}
	}

	return m.logFactory.NewLogger(meta)
}

func (m *middlewareWrapperImpl) addMetaEntry(meta map[string]string, key, value string) {
	if value == "" {
		return
	}
	meta["entry."+key] = value
}

func (m *middlewareWrapperImpl) getLabelsAndValues(subsystem, name string, w WrappedResponseWriter,
	r *http.Request) ([]string, []string) {
	return []string{"app", "server", "env", "code", "method", "handler", "version", "subsystem"},
		[]string{
			m.globals.AppName,
			m.globals.ServerName,
			m.globals.DeployEnvironment,
			strconv.Itoa(w.Status()),
			strings.ToLower(r.Method),
			strings.ToLower(name),
			m.globals.VersionNumber,
			subsystem,
		}
}

func (m *middlewareWrapperImpl) wrapWithRequestMetrics(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, p RouterParams) {
		labels, values := m.getLabelsAndValues(subsystem, name, w, r)
		start := time.Now()

		histSeconds := m.metrics.AddHistogramVec("", "http_request_duration_seconds",
			"Response times for requests in seconds.", labels, values)
		sumMicroSeconds := m.metrics.AddSummaryVec("", "http_request_duration_microseconds",
			"Response times for requests in microseconds.", labels, values)

		m.metrics.CountLabels("", "http_requests_total", "Total requests.", labels, values)

		handler(w, r, p)

		sumMicroSeconds.RecordDuration(start, time.Microsecond)
		histSeconds.RecordDuration(start, time.Second)

		labels, values = m.getLabelsAndValues(subsystem, name, w, r)
		m.metrics.CountLabels("", "http_responses_total", "Total responses.", labels, values)
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
				log := m.getMetaLog(subsystem, name, w, r, p, make(map[string]string))
				log.Error("PanicAutorecover", "PANIC recovered: %v", rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler(w, r, p)
	}
}

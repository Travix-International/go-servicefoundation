package servicefoundation

import (
	"fmt"
	"net/http"
	"runtime/debug"
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
	// Authorize is a middle enumeration to authorize the current user before continuing.
	Authorize Middleware = 8
)

type (
	// Middleware is an enumeration to indicate the available middleware wrappers.
	Middleware int

	//MiddlewareWrapperFactory is a factory for MiddlewareWrapper.
	MiddlewareWrapperFactory interface {
		NewMiddlewareWrapper(corsOptions *CORSOptions, authFunc AuthorizationFunc) MiddlewareWrapper
	}

	// MiddlewareWrapper is an interface to wrap an existing handler with the specified middleware.
	MiddlewareWrapper interface {
		Wrap(subsystem, name string, middleware Middleware, handler Handle, metaFunc MetaFunc) Handle
	}

	// AuthorizationFunc is a middleware function that should return true if authorization is successful.
	AuthorizationFunc func(WrappedResponseWriter, *http.Request, HandlerUtils) bool

	middlewareWrapperFactoryImpl struct {
		log     Logger
		globals ServiceGlobals
	}

	middlewareWrapperImpl struct {
		log         Logger
		globals     ServiceGlobals
		corsOptions *cors.Options
		authFunc    AuthorizationFunc
	}
)

// NewMiddlewareWrapperFactory instantiates and returns a new NewMiddlewareWrapperFactory implementation.
func NewMiddlewareWrapperFactory(logger Logger, globals ServiceGlobals) MiddlewareWrapperFactory {
	return &middlewareWrapperFactoryImpl{
		log:     logger,
		globals: globals,
	}
}

// NewMiddlewareWrapper instantiates a new MiddlewareWrapper implementation.
func (f *middlewareWrapperFactoryImpl) NewMiddlewareWrapper(corsOptions *CORSOptions,
	authFunc AuthorizationFunc) MiddlewareWrapper {

	m := &middlewareWrapperImpl{
		log:      f.log,
		globals:  f.globals,
		authFunc: authFunc,
	}
	m.corsOptions = m.mergeCORSOptions(corsOptions)
	return m
}

/* MiddlewareWrapper implementation */

func (m *middlewareWrapperImpl) Wrap(subsystem, name string, middleware Middleware, handler Handle, metaFunc MetaFunc) Handle {
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
		return m.wrapWithPanicHandler(subsystem, name, handler, metaFunc)
	case RequestLogging:
		return m.wrapWithRequestLogging(subsystem, name, handler, metaFunc)
	case RequestMetrics:
		return m.wrapWithRequestMetrics(subsystem, name, handler)
	case Authorize:
		return m.wrapWithAuthorization(subsystem, name, handler)
	default:
		m.log.Warn("UnhandledMiddleware", "Unhandled middleware: %v", middleware)
	}
	return handler
}

func (m *middlewareWrapperImpl) wrapWithCounter(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		counterName := fmt.Sprintf("%v_total", strings.ToLower(name))
		counterHelp := fmt.Sprintf("Totals for %v.", name)
		labels, values := m.getLabelsAndValues(subsystem, name, w, r)

		u.Metrics.CountLabels("", counterName, counterHelp, labels, values)

		handler(w, r, u)
	}
}

func (m *middlewareWrapperImpl) wrapWithHistogram(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		histogramName := fmt.Sprintf("%v_duration_milliseconds", strings.ToLower(name))
		histogramHelp := fmt.Sprintf("Response times for %v in milliseconds.", name)

		labels, values := m.getLabelsAndValues(subsystem, name, w, r)
		hist := u.Metrics.AddHistogramVec(subsystem, histogramName, histogramHelp, labels, values)
		start := time.Now()

		handler(w, r, u)

		hist.RecordTimeElapsed(start)
	}
}

func (m *middlewareWrapperImpl) wrapWithRequestLogging(subsystem, name string, handler Handle, metaFunc MetaFunc) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		meta := metaFunc(r, u.ParamsFunc)
		m.addRequestResponseToMeta(w, r, meta)
		log := u.NewLoggerWithMeta(meta, r)
		log.Info("ApiRequest", m.getRequestStartMessage(r, u.ParamsFunc, meta))

		start := time.Now()

		handler(w, r, u)

		elapsedMs := float64(time.Since(start).Nanoseconds()/int64(time.Microsecond)) / 1000.0
		durationMs := strconv.FormatFloat(elapsedMs, 'f', 3, 64)

		meta["entry.duration"] = durationMs
		m.addRequestResponseToMeta(w, r, meta)
		log = u.NewLoggerWithMeta(meta, r)
		log.Info("ApiResponse", m.getRequestEndMessage(w, r, u.ParamsFunc, meta, durationMs))
	}
}

func (m *middlewareWrapperImpl) getRequestStartMessage(r *http.Request, p RouteParamsFunc, meta map[string]string) string {
	return fmt.Sprintf("%s %s", r.Method, meta["entry.http.url"])
}

func (m *middlewareWrapperImpl) getRequestEndMessage(w WrappedResponseWriter, r *http.Request, p RouteParamsFunc, meta map[string]string, durationMs string) string {
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

func (m *middlewareWrapperImpl) addRequestResponseToMeta(w WrappedResponseWriter, r *http.Request, meta map[string]string) {
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
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		labels, values := m.getLabelsAndValues(subsystem, name, w, r)
		start := time.Now()

		histSeconds := u.Metrics.AddHistogramVec("", "http_request_duration_seconds",
			"Response times for requests in seconds.", labels, values)
		sumMicroSeconds := u.Metrics.AddSummaryVec("", "http_request_duration_microseconds",
			"Response times for requests in microseconds.", labels, values)

		u.Metrics.CountLabels("", "http_requests_total", "Total requests.", labels, values)

		handler(w, r, u)

		sumMicroSeconds.RecordDuration(start, time.Microsecond)
		histSeconds.RecordDuration(start, time.Second)

		labels, values = m.getLabelsAndValues(subsystem, name, w, r)
		u.Metrics.CountLabels("", "http_responses_total", "Total responses.", labels, values)
	}
}

func (m *middlewareWrapperImpl) wrapWithAuthorization(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		if !m.authFunc(w, r, u) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(w, r, u)
	}
}

func (m *middlewareWrapperImpl) wrapWithNoCache(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		w.Header().Set("Cache-Control", "max-age: 0, private")
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
		w.Header().Set("Expires", time.Now().AddDate(-1, 0, 0).Format(http.TimeFormat))

		handler(w, r, u)
	}
}

func (m *middlewareWrapperImpl) wrapWithCORS(subsystem, name string, handler Handle) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		c := cors.New(*m.corsOptions)

		h := func(ww http.ResponseWriter, r *http.Request) {
			w := NewWrappedResponseWriter(ww)
			handler(w, r, u)
		}
		c.ServeHTTP(w, r, h)
	}
}

func (m *middlewareWrapperImpl) mergeCORSOptions(options *CORSOptions) *cors.Options {
	corsOptions := cors.Options{
		AllowedOrigins: options.AllowedOrigins,
		AllowedMethods: appendAndDedupe(options.AllowedMethods, "HEAD", "OPTIONS"),
		AllowedHeaders: appendAndDedupe(options.AllowedHeaders,
			"Origin", "Accept", "Content-Type", "X-Requested-With", "X-CSRF-Token"),
		AllowCredentials: true,
		ExposedHeaders: appendAndDedupe(options.ExposedHeaders,
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Methods",
			"Access-Control-Max-Age",
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Origin"),
		MaxAge: options.MaxAge,
	}
	return &corsOptions
}

func appendAndDedupe(slice []string, elements ...string) []string {
	var result []string

	temp := append(slice, elements...)

	for i := 0; i < len(temp); i++ {
		for j := 0; j < len(result); j++ {
			if temp[i] == result[j] {
				break
			}
		}
		result = append(result, temp[i])
	}
	return result
}

func (m *middlewareWrapperImpl) wrapWithPanicHandler(subsystem, name string, handler Handle, metaFunc MetaFunc) Handle {
	return func(w WrappedResponseWriter, r *http.Request, u HandlerUtils) {
		defer func() {
			if rec := recover(); rec != nil {
				meta := metaFunc(r, u.ParamsFunc)
				log := u.NewLoggerWithMeta(meta, r)
				log.Error("PanicAutorecover", "PANIC recovered: %v\n%s", rec, string(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler(w, r, u)
	}
}

package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	// ExitFunc is the function signature for the exit function used by Service.
	ExitFunc func(int)

	// WrapHandler is an interface for wrapping a Handle with middleware.
	WrapHandler interface {
		Wrap(string, string, []Middleware, Handle) httprouter.Handle
	}

	// RootHandler is an interface to instantiate a new root handler.
	RootHandler interface {
		NewRootHandler() Handle
	}

	// ReadinessHandler is an interface to instantiate a new readiness handler.
	ReadinessHandler interface {
		NewReadinessHandler() Handle
	}

	// LivenessHandler is an interface to instantiate a new liveness handler.
	LivenessHandler interface {
		NewLivenessHandler() Handle
	}

	// HealthHandler is an interface to instantiate a new health handler.
	HealthHandler interface {
		NewHealthHandler() Handle
	}

	// VersionHandler is an interface to instantiate a new version handler.
	VersionHandler interface {
		NewVersionHandler() Handle
	}

	// MetricsHandler is an interface to instantiate a new metrics handler.
	MetricsHandler interface {
		NewMetricsHandler() Handle
	}

	// QuitHandler is an interface to instantiate a new quit handler.
	QuitHandler interface {
		NewQuitHandler() Handle
	}

	// ServiceHandlerFactory is an interface to get access to implemented handlers.
	ServiceHandlerFactory interface {
		NewHandlers() *Handlers
		WrapHandler
	}

	// Handlers is a struct containing references to handler implementations.
	Handlers struct {
		RootHandler      RootHandler
		ReadinessHandler ReadinessHandler
		LivenessHandler  LivenessHandler
		HealthHandler    HealthHandler
		VersionHandler   VersionHandler
		MetricsHandler   MetricsHandler
		QuitHandler      QuitHandler
	}

	serviceHandlerFactoryImpl struct {
		versionBuilder    VersionBuilder
		exitFunc          ExitFunc
		middlewareWrapper MiddlewareWrapper
		stateReader       ServiceStateReader
	}
)

// NewServiceHandlerFactory creates a new factory with handler implementations.
func NewServiceHandlerFactory(middlewareWrapper MiddlewareWrapper, versionBuilder VersionBuilder,
	stateReader ServiceStateReader, exitFunc ExitFunc) ServiceHandlerFactory {

	return &serviceHandlerFactoryImpl{
		versionBuilder:    versionBuilder,
		exitFunc:          exitFunc,
		middlewareWrapper: middlewareWrapper,
		stateReader:       stateReader,
	}
}

/* ServiceHandlerFactory implementation */

// Wrap wraps the specified Handle with the specified middleware wrappers.
func (f *serviceHandlerFactoryImpl) Wrap(subsystem, name string, middlewares []Middleware, handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h := handle

		for _, middleware := range middlewares {
			h = f.middlewareWrapper.Wrap(subsystem, name, middleware, h)
		}
		h(NewWrappedResponseWriter(w), r, RouterParams{Params: p})
	}
}

// NewHandlers instantiates a new Handlers struct containing implemented handlers.
func (f *serviceHandlerFactoryImpl) NewHandlers() *Handlers {
	return &Handlers{
		RootHandler:      f,
		QuitHandler:      f,
		MetricsHandler:   f,
		VersionHandler:   f,
		HealthHandler:    f,
		LivenessHandler:  f,
		ReadinessHandler: f,
	}
}

func (f *serviceHandlerFactoryImpl) NewRootHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		w.WriteHeader(http.StatusOK)
	}
}

func (f *serviceHandlerFactoryImpl) NewReadinessHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		if f.stateReader.IsReady() {
			w.JSON(http.StatusOK, "ok")
		} else {
			w.JSON(http.StatusInternalServerError, "not ready")
		}
	}
}

func (f *serviceHandlerFactoryImpl) NewLivenessHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		if f.stateReader.IsLive() {
			w.JSON(http.StatusOK, "ok")
		} else {
			w.JSON(http.StatusInternalServerError, "not ready")
		}
	}
}

func (f *serviceHandlerFactoryImpl) NewQuitHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		defer f.exitFunc(0)

		w.WriteHeader(http.StatusOK)

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (f *serviceHandlerFactoryImpl) NewHealthHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		if f.stateReader.IsHealthy() {
			w.JSON(http.StatusOK, "ok")
		} else {
			w.JSON(http.StatusInternalServerError, "not healthy")
		}
	}
}

func (f *serviceHandlerFactoryImpl) NewVersionHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		version := f.versionBuilder.ToMap()
		w.JSON(http.StatusOK, version)
	}
}

func (f *serviceHandlerFactoryImpl) NewMetricsHandler() Handle {
	return func(w WrappedResponseWriter, r *http.Request, _ RouterParams) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

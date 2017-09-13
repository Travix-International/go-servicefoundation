package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	ExitFunc func(int)

	WrapHandler interface {
		Wrap(string, string, []Middleware, Handle) httprouter.Handle
	}

	RootHandler interface {
		NewRootHandler() Handle
	}

	ReadinessHandler interface {
		NewReadinessHandler() Handle
	}

	LivenessHandler interface {
		NewLivenessHandler() Handle
	}

	HealthHandler interface {
		NewHealthHandler() Handle
	}

	VersionHandler interface {
		NewVersionHandler() Handle
	}

	MetricsHandler interface {
		NewMetricsHandler() Handle
	}

	QuitHandler interface {
		NewQuitHandler() Handle
	}

	ServiceHandlerFactory interface {
		NewHandlers() *Handlers
		WrapHandler
	}

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

func (f *serviceHandlerFactoryImpl) Wrap(subsystem, name string, middlewares []Middleware, handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h := handle

		for _, middleware := range middlewares {
			h = f.middlewareWrapper.Wrap(subsystem, name, middleware, h)
		}
		h(NewWrappedResponseWriter(w), r, RouterParams{Params: p})
	}
}

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

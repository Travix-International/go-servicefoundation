package site

import (
	"net/http"

	. "github.com/Prutswonder/go-servicefoundation/model"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type serviceHandlerFactoryImpl struct {
	versionBuilder    VersionBuilder
	exitFunc          ExitFunc
	middlewareWrapper MiddlewareWrapper
}

func CreateServiceHandlerFactory(middlewareWrapper MiddlewareWrapper, versionBuilder VersionBuilder, exitFunc ExitFunc) ServiceHandlerFactory {
	return &serviceHandlerFactoryImpl{
		versionBuilder:    versionBuilder,
		exitFunc:          exitFunc,
		middlewareWrapper: middlewareWrapper,
	}
}

/* ServiceHandlerFactory implementation */

func (f *serviceHandlerFactoryImpl) WrapHandler(subsystem, name string, middlewares []Middleware, handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h := handle

		for _, middleware := range middlewares {
			h = f.middlewareWrapper.Wrap(subsystem, name, middleware, h)
		}
		h(CreateWrappedResponseWriter(w), r, RouterParams{p})
	}
}

func (f *serviceHandlerFactoryImpl) CreateRootHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		w.WriteHeader(http.StatusOK)
	}
}

func (f *serviceHandlerFactoryImpl) CreateReadinessHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateLivenessHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateQuitHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		defer f.exitFunc(0)

		w.WriteHeader(http.StatusOK)

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (f *serviceHandlerFactoryImpl) CreateHealthHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateVersionHandler() Handle {
	return func(w WrappedResponseWriter, _ *http.Request, _ RouterParams) {
		version := f.versionBuilder.ToMap()
		w.JSON(http.StatusOK, version)
	}
}

func (f *serviceHandlerFactoryImpl) CreateMetricsHandler() Handle {
	return func(w WrappedResponseWriter, r *http.Request, _ RouterParams) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

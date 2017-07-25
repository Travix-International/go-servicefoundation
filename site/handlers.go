package site

import (
	"net/http"

	"github.com/Prutswonder/go-servicefoundation/model"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type serviceHandlerFactoryImpl struct {
	exitFunc          model.ExitFunc
	middlewareWrapper model.MiddlewareWrapper
}

func CreateServiceHandlerFactory(middlewareWrapper model.MiddlewareWrapper, exitFunc model.ExitFunc) model.ServiceHandlerFactory {
	return &serviceHandlerFactoryImpl{
		exitFunc:          exitFunc,
		middlewareWrapper: middlewareWrapper,
	}
}

/* ServiceHandlerFactory implementation */

func (f *serviceHandlerFactoryImpl) WrapHandler(subsystem, name string, middlewares []model.Middleware, handle model.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h := handle

		for _, middleware := range middlewares {
			h = f.middlewareWrapper.Wrap(subsystem, name, middleware, h)
		}
		h(CreateWrappedResponseWriter(w), r, p)
	}
}

func (f *serviceHandlerFactoryImpl) CreateRootHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}

func (f *serviceHandlerFactoryImpl) CreateReadinessHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateLivenessHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateQuitHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		defer f.exitFunc(0)

		w.WriteHeader(http.StatusOK)

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (f *serviceHandlerFactoryImpl) CreateHealthHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.JSON(http.StatusOK, "ok")
	}
}

func (f *serviceHandlerFactoryImpl) CreateVersionHandler() model.Handle {
	return func(w model.WrappedResponseWriter, _ *http.Request, _ httprouter.Params) {
		//TODO: Construct application version
		version := "ok"
		w.JSON(http.StatusOK, version)
	}
}

func (f *serviceHandlerFactoryImpl) CreateMetricsHandler() model.Handle {
	return func(w model.WrappedResponseWriter, r *http.Request, _ httprouter.Params) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

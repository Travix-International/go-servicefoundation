package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	// Handle is a function signature for the ServiceFoundation handlers
	Handle func(WrappedResponseWriter, *http.Request, HandlerUtils)

	// MetaFunc is a function that returns a map containing meta data used to enrich log messages.
	MetaFunc func(*http.Request, RouterParams) map[string]string

	// HandlerUtils contains utilities that can be used by the current handler.
	HandlerUtils struct {
		Meta       map[string]string
		Params     RouterParams
		LogFactory LogFactory
		Metrics    Metrics
	}

	// RouterParams is a struct that wraps httprouter.Params
	RouterParams struct {
		Params httprouter.Params
	}

	// Router is an interface containing the used methods from httprouter.Router
	Router interface {
		Handle(method, path string, handle httprouter.Handle)
		ServeHTTP(http.ResponseWriter, *http.Request)
	}

	// RouterFactory is an interface to create a new Router.
	RouterFactory interface {
		NewRouter() Router
	}

	routerFactoryImpl struct {
	}
)

var (
	// MethodsForGet contains a slice with the supported http methods for GET.
	MethodsForGet = []string{http.MethodGet}
	// MethodsForPost contains a slice with the supported http methods for POST.
	MethodsForPost = []string{http.MethodPost}
)

// NewRouterFactory instantiates a new RouterFactory implementation.
func NewRouterFactory() RouterFactory {
	return &routerFactoryImpl{}
}

/* RouterFactory implementation */

func (r *routerFactoryImpl) NewRouter() Router {
	return httprouter.New()
}

/* HandlerUtils methods */

// NewLoggerWithMeta instantiates and returns a new logger containing the provided meta.
func (u HandlerUtils) NewLoggerWithMeta(meta map[string]string) Logger {
	combinedMeta := combineMetas(u.Meta, meta)
	return u.LogFactory.NewLogger(combinedMeta)
}

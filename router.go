package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	// Handle is a function signature for the ServiceFoundation handlers
	Handle func(WrappedResponseWriter, *http.Request, RouterParams)

	// MetaFunc is a function that returns a map containing meta data used to enrich log messages.
	MetaFunc func(*http.Request, RouterParams) map[string]string

	// RouterParams is a struct that wraps httprouter.Params
	RouterParams struct {
		Params httprouter.Params
	}

	// Router is a struct that wraps httprouter.Router
	Router struct {
		Router *httprouter.Router
	}

	// RouterFactory is an interface to create a new Router.
	RouterFactory interface {
		NewRouter() *Router
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

func (r *routerFactoryImpl) NewRouter() *Router {
	return &Router{Router: httprouter.New()}
}

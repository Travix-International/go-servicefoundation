package servicefoundation

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	Handle func(WrappedResponseWriter, *http.Request, RouterParams)

	RouterParams struct {
		Params httprouter.Params
	}

	Router struct {
		Router *httprouter.Router
	}

	RouterFactory interface {
		CreateRouter() *Router
	}

	routerFactoryImpl struct {
	}
)

var (
	MethodsForGet  = []string{http.MethodGet}
	MethodsForPost = []string{http.MethodPost}
)

func NewRouterFactory() RouterFactory {
	return &routerFactoryImpl{}
}

/* RouterFactory implementation */

func (r *routerFactoryImpl) CreateRouter() *Router {
	return &Router{Router: httprouter.New()}
}

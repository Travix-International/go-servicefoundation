package model

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
)

var (
	MethodsForGet  = []string{http.MethodGet}
	MethodsForPost = []string{http.MethodPost}
)

package model

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	Handle func(WrappedResponseWriter, *http.Request, httprouter.Params)

	RouterFactory interface {
		CreateRouter() *httprouter.Router
	}
)

var (
	MethodsForGet  = []string{http.MethodGet}
	MethodsForPost = []string{http.MethodPost}
)

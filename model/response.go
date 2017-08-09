package model

import "net/http"

type WrappedResponseWriter interface {
	http.ResponseWriter
	JSON(statusCode int, content interface{})
	XML(statusCode int, content interface{})
	AcceptsXML(r *http.Request) bool
	WriteResponse(r *http.Request, statusCode int, content interface{})
	SetCaching(maxAge int)
	Status() int
}

package servicefoundation

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

type (
	// WrappedResponseWriter is a wrapper around the http.ResponseWriter and extending it with commonly used writing
	// methods.
	WrappedResponseWriter interface {
		http.ResponseWriter
		JSON(statusCode int, content interface{})
		XML(statusCode int, content interface{})
		AcceptsXML(r *http.Request) bool
		WriteResponse(r *http.Request, statusCode int, content interface{})
		SetCaching(maxAge int)
		Status() int
	}

	wrappedResponseWriterImpl struct {
		http.ResponseWriter
		status      int
		wroteHeader bool
	}
)

const (
	// AcceptHeader is the name of the http Accept header.
	AcceptHeader = "Accept"
	// ContentTypeHeader is the name of the http content type header.
	ContentTypeHeader = "Content-Type"
	// ContentTypeJSON is the value of the http content type header for JSON documents.
	ContentTypeJSON = "application/json"
	// ContentTypeXML is the value of the http content type header for XML documents.
	ContentTypeXML = "application/xml"
)

// NewWrappedResponseWriter instantiates a new WrappedResponseWriter implementation.
func NewWrappedResponseWriter(w http.ResponseWriter) WrappedResponseWriter {
	return &wrappedResponseWriterImpl{ResponseWriter: w, status: http.StatusOK}
}

/* WrappedResponseWriter implementation */

func (w *wrappedResponseWriterImpl) Status() int {
	return w.status
}

func (w *wrappedResponseWriterImpl) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *wrappedResponseWriterImpl) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)

	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

func (w *wrappedResponseWriterImpl) JSON(statusCode int, content interface{}) {
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(content)
}

func (w *wrappedResponseWriterImpl) XML(statusCode int, content interface{}) {
	w.Header().Set(ContentTypeHeader, ContentTypeXML)
	w.WriteHeader(statusCode)

	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(content)
}

func (w *wrappedResponseWriterImpl) AcceptsXML(r *http.Request) bool {
	return !strings.Contains(r.Header.Get(AcceptHeader), ContentTypeJSON) &&
		strings.Contains(r.Header.Get(AcceptHeader), ContentTypeXML)
}

func (w *wrappedResponseWriterImpl) WriteResponse(r *http.Request, statusCode int, content interface{}) {
	if w.AcceptsXML(r) {
		w.XML(statusCode, content)
		return
	}
	w.JSON(statusCode, content)
}

func (w *wrappedResponseWriterImpl) SetCaching(maxAge int) {
	w.Header().Set("Vary", "Accept, Origin") // Because we don't want to mix XML and JSON in the cache!
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", maxAge))
}

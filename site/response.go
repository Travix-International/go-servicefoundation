package site

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/Prutswonder/go-servicefoundation/model"
)

const (
	ACCEPT_HEADER = "Accept"
	ACCEPT_JSON   = "application/json"
	ACCEPT_XML    = "application/xml"
)

type wrappedResponseWriterImpl struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func CreateWrappedResponseWriter(w http.ResponseWriter) model.WrappedResponseWriter {
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(content)
}

func (w *wrappedResponseWriterImpl) XML(statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(content)
}

func (w *wrappedResponseWriterImpl) AcceptsXML(r *http.Request) bool {
	return !strings.Contains(r.Header.Get(ACCEPT_HEADER), ACCEPT_JSON) && strings.Contains(r.Header.Get(ACCEPT_HEADER), ACCEPT_XML)
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

func (w *wrappedResponseWriterImpl) GetStatus() int {
	return w.status
}

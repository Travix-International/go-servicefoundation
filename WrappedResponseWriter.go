//Copied from: https://gist.github.com/ciaranarcher/abccf50cb37645ca27fa
package servicefoundation

import (
	"net/http"
)

type WrappedResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func NewWrappedResponseWriter(w http.ResponseWriter) *WrappedResponseWriter {
	return &WrappedResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *WrappedResponseWriter) Status() int {
	return w.status
}

func (w *WrappedResponseWriter) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *WrappedResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)

	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

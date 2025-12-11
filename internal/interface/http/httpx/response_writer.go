// Package httpx provides HTTP utilities and wrappers.
package httpx

import "net/http"

// ResponseWriter wraps http.ResponseWriter to capture status code.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// NewResponseWriter creates a new ResponseWriter wrapper.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200
	}
}

// WriteHeader captures the status code.
func (rw *ResponseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes data and ensures status is set.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// StatusCode returns the captured status code.
func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}

// Unwrap returns the original ResponseWriter.
func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

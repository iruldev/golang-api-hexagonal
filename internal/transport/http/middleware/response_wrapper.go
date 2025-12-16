// Package middleware provides HTTP middleware components.
package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// ResponseWrapper wraps http.ResponseWriter to capture response metrics.
// It tracks the HTTP status code and total bytes written.
type ResponseWrapper struct {
	http.ResponseWriter
	status int
	bytes  int
}

// NewResponseWrapper creates a new ResponseWrapper with default status 200.
func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
		status:         http.StatusOK, // Default status if WriteHeader is never called
	}
}

// WriteHeader captures the status code and passes it to the underlying writer.
func (w *ResponseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write captures the bytes written and passes data to the underlying writer.
func (w *ResponseWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Status returns the captured HTTP status code.
func (w *ResponseWrapper) Status() int {
	return w.status
}

// BytesWritten returns the total number of bytes written to the response.
func (w *ResponseWrapper) BytesWritten() int {
	return w.bytes
}

// Unwrap returns the underlying ResponseWriter.
// This allows access to optional interfaces like http.Flusher.
func (w *ResponseWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush delegates to the underlying http.Flusher if supported.
func (w *ResponseWrapper) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack delegates to the underlying http.Hijacker if supported.
func (w *ResponseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not support hijacking")
}

// CloseNotify delegates to the underlying http.CloseNotifier if supported.
// Deprecated in stdlib but still used by some handlers.
func (w *ResponseWrapper) CloseNotify() <-chan bool {
	if notifier, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	return nil
}

// Push delegates to the underlying http.Pusher if supported (HTTP/2 server push).
func (w *ResponseWrapper) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

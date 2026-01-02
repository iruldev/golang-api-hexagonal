package handler

import "net/http"

// benchResponseWriter is a lightweight response writer for benchmarking (zero allocations)
// It avoids the overhead of httptest.ResponseRecorder which captures the body buffers.
type benchResponseWriter struct {
	code int
}

func newBenchResponseWriter() *benchResponseWriter {
	return &benchResponseWriter{code: http.StatusOK}
}

func (w *benchResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *benchResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *benchResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

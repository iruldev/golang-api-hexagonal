package middleware

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logging middleware logs HTTP requests with structured fields.
// Logs: method, path, status, latency, request_id, trace_id
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			latency := time.Since(start)

			// Extract trace_id from OTEL span context (Story 5.8)
			traceID := ""
			spanCtx := trace.SpanContextFromContext(r.Context())
			if spanCtx.HasTraceID() {
				traceID = spanCtx.TraceID().String()
			}

			logger.Info("request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.statusCode),
				zap.Duration("latency", latency),
				zap.String("request_id", GetRequestID(r.Context())),
				zap.String("trace_id", traceID),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and delegates to the wrapped writer.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

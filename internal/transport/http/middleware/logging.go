package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	logKeyRequestID = "requestId"
	logKeyMethod    = "method"
	logKeyRoute     = "route"
	logKeyStatus    = "status"
	logKeyDuration  = "duration_ms"
	logKeyBytes     = "bytes"
	logKeyTraceID   = "traceId"
)

// RequestLogger returns a middleware that logs HTTP request completion.
// It captures method, route, status, duration, and response size.
// The requestId field is populated from the context (set by RequestID middleware).
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status and bytes
			ww := NewResponseWrapper(w)

			// Process the request
			next.ServeHTTP(ww, r)

			// Capture request ID from context (set by RequestID middleware)
			requestID := GetRequestID(r.Context())
			if requestID == "" {
				// Fallback to prevent empty requestId in logs when RequestID middleware is missing/misordered.
				requestID = generateRequestID()
			}

			// Get route pattern from chi router context
			routeCtx := chi.RouteContext(r.Context())
			routePattern := ""
			if routeCtx != nil {
				routePattern = routeCtx.RoutePattern()
			}
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			// Calculate duration
			duration := time.Since(start)

			// Log request completion with structured fields
			logger.Info("request completed",
				logKeyMethod, r.Method,
				logKeyRoute, routePattern,
				logKeyStatus, ww.Status(),
				logKeyDuration, duration.Milliseconds(),
				logKeyBytes, ww.BytesWritten(),
				logKeyRequestID, requestID,
				logKeyTraceID, GetTraceID(r.Context()),
			)
		})
	}
}

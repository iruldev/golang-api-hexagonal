package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

const (
	logKeyRequestID = "requestId"
	logKeyMethod    = "method"
	logKeyRoute     = "route"
	logKeyStatus    = "status"
	logKeyDuration  = "duration_ms"
	logKeyBytes     = "bytes"
)

// RequestLogger returns a middleware that logs HTTP request completion.
// It captures method, route, status, duration, and response size.
// The requestId field will be populated when Request ID middleware is added (Story 2.2).
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status and bytes
			ww := NewResponseWrapper(w)

			// Capture request ID injected by chi middleware.RequestID
			requestID := chiMiddleware.GetReqID(r.Context())

			// Process the request
			next.ServeHTTP(ww, r)

			// Get route pattern from chi router context
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
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
			)
		})
	}
}

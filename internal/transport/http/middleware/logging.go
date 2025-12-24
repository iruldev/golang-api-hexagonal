package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
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
// The requestId field is populated from the context (set by RequestID middleware).
// The traceId and spanId fields are conditionally added when tracing is enabled.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status and bytes
			ww := NewResponseWrapper(w)

			// Process the request
			next.ServeHTTP(ww, r)

			// Capture request ID from context (set by RequestID middleware)
			requestID := ctxutil.GetRequestID(r.Context())
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

			// Build log args, conditionally adding trace fields only if present (AC#1, AC#2)
			args := []any{
				logKeyMethod, r.Method,
				logKeyRoute, routePattern,
				logKeyStatus, ww.Status(),
				logKeyDuration, duration.Milliseconds(),
				logKeyBytes, ww.BytesWritten(),
				logKeyRequestID, requestID,
			}

			// Add traceId only if present (absent when tracing disabled)
			if traceID := GetTraceID(r.Context()); traceID != "" {
				args = append(args, observability.LogKeyTraceID, traceID)
			}

			// Add spanId only if present (absent when tracing disabled)
			if spanID := GetSpanID(r.Context()); spanID != "" {
				args = append(args, observability.LogKeySpanID, spanID)
			}

			// Log request completion with structured fields using request context (for context-aware handlers).
			logger.InfoContext(r.Context(), "request completed", args...)
		})
	}
}

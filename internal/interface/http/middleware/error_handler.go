// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// ErrorHandler is a middleware that recovers from panics and logs errors.
// It wraps panic errors in a consistent Envelope response format.
//
// The middleware:
//   - Recovers from panics in downstream handlers
//   - Logs panic with trace_id correlation
//   - Returns HTTP 500 with Envelope error format
//   - Returns HTTP 500 with Envelope error format
//   - Includes meta.trace_id in response
//
// Dependencies:
//   - RequestID middleware must run BEFORE this middleware to ensure trace_id is available.
//
// Usage:
//
//	router.Use(middleware.ErrorHandler)
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				traceID := ctxutil.RequestIDFromContext(r.Context())
				slog.Error("panic recovered",
					"trace_id", traceID,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				// Use Envelope format for consistent error responses
				response.InternalServerErrorCtx(w, r.Context(), "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

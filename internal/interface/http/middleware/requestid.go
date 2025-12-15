// Package middleware contains HTTP middleware for the API.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
)

// RequestIDHeader is the header key for request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID middleware generates or uses existing request ID.
// If X-Request-ID header is present, it uses the provided ID.
// Otherwise, it generates a new UUID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in response header
		w.Header().Set(RequestIDHeader, requestID)

		// Store in context for downstream handlers using ctxutil
		ctx := ctxutil.NewRequestIDContext(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context.
// Returns empty string if request ID is not present.
// Deprecated: Use ctxutil.RequestIDFromContext directly.
func GetRequestID(ctx context.Context) string {
	return ctxutil.RequestIDFromContext(ctx)
}

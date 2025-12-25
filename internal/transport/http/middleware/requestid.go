package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// headerXRequestID is the HTTP header name for request ID.
const headerXRequestID = "X-Request-ID"

// RequestID returns a middleware that generates or passes through a request ID.
// If the incoming request has an X-Request-ID header, it uses that value (passthrough).
// Otherwise, it generates a new random ID (16 bytes hex = 32 characters).
// The request ID is injected into the request context and set in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(headerXRequestID)

		// Validate request ID (prevent DoS/DB truncation/Injection)
		if !isValidRequestID(requestID) {
			requestID = generateRequestID()
		}

		// Set response header
		w.Header().Set(headerXRequestID, requestID)

		// Inject into context
		ctx := ctxutil.SetRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a new UUID v7 request ID.
func generateRequestID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to v4 if v7 fails (rare)
		return uuid.NewString()
	}
	return id.String()
}

// isValidRequestID checks if the ID is safe and valid (length <= 64, printable ASCII safe chars).
func isValidRequestID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	// Allow alphanumeric, hyphen, underscore, colon, dot
	for _, c := range id {
		if !isSafeChar(c) {
			return false
		}
	}
	return true
}

func isSafeChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == ':' || c == '.'
}

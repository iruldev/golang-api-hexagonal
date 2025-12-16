package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"
)

// contextKey is a custom type to avoid context key collisions.
type contextKey string

// requestIDKey is the context key for storing request ID.
const requestIDKey contextKey = "requestId"

// headerXRequestID is the HTTP header name for request ID.
const headerXRequestID = "X-Request-ID"

// RequestID returns a middleware that generates or passes through a request ID.
// If the incoming request has an X-Request-ID header, it uses that value (passthrough).
// Otherwise, it generates a new random ID (16 bytes hex = 32 characters).
// The request ID is injected into the request context and set in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(headerXRequestID)

		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set response header
		w.Header().Set(headerXRequestID, requestID)

		// Inject into context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// generateRequestID creates a new random request ID.
// It generates 16 random bytes and encodes them as hex (32 characters).
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		// Fallback to time-based hash to avoid empty/partial IDs if rand fails
		fallback := sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
		copy(b, fallback[:])
	}
	return hex.EncodeToString(b)
}

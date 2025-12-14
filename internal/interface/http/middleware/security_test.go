package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	// Create a mock handler that always returns 200 OK
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the handler with the SecurityHeaders middleware
	handler := SecurityHeaders(nextHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check response body
	assert.Equal(t, "OK", rec.Body.String())

	// Check security headers
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
	}

	for key, value := range expectedHeaders {
		assert.Equal(t, value, rec.Header().Get(key), "Header %s should be %s", key, value)
	}
}

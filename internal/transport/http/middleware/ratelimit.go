// Package middleware provides HTTP middleware for the transport layer.
// This file implements rate limiting middleware with per-user and per-IP support.
package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httprate"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// RateLimitWindow is the time window for rate limiting.
const RateLimitWindow = time.Second

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per second.
	RequestsPerSecond int
	// TrustProxy enables trusting X-Forwarded-For/X-Real-IP headers for client IP.
	TrustProxy bool
}

// RateLimiter returns middleware that limits requests per key (user ID or IP).
// For authenticated requests, limits are per-user (claims.Subject).
// For unauthenticated requests, limits are per-IP.
//
// AC #1: Unauthenticated requests use resolved client IP as key.
// AC #2: Authenticated requests use claims.Subject (userId) as key.
// AC #3,4: IP resolution relies on global RealIP middleware (respects TRUST_PROXY).
// AC #5: Returns 429 with RFC 7807 and Retry-After header.
// AC #6: Uses go-chi/httprate.
// AC #7: Rate limit is configurable via RATE_LIMIT_RPS.
func RateLimiter(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return httprate.Limit(
		cfg.RequestsPerSecond,
		RateLimitWindow,
		httprate.WithKeyFuncs(keyFunc()),
		httprate.WithLimitHandler(rateLimitExceededHandler),
	)
}

// keyFunc returns the rate limit key based on JWT claims or IP.
// It checks for authenticated user first (per-user limiting),
// then falls back to IP-based limiting for unauthenticated requests.
func keyFunc() httprate.KeyFunc {
	return func(r *http.Request) (string, error) {
		// Try to get user ID from JWT claims first (AC #2)
		if claims := ctxutil.GetClaims(r.Context()); claims != nil {
			// Use Subject (sub claim) as the user identifier
			if strings.TrimSpace(claims.Subject) != "" {
				return "user:" + claims.Subject, nil
			}
		}

		// Fallback to IP-based limiting (AC #1)
		ip := resolveClientIP(r)
		return "ip:" + ip, nil
	}
}

// resolveClientIP extracts the client IP address from the request.
// It relies on r.RemoteAddr, which is normalized by the global RealIP middleware
// if TRUST_PROXY is enabled.
func resolveClientIP(r *http.Request) string {
	// AC #4: Use RemoteAddr (already normalized by RealIP middleware if configured)
	// RemoteAddr format is "ip:port" (default) or "ip" (if RealIP ran)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails (e.g. valid IP with no port), use as is
		return r.RemoteAddr
	}
	return ip
}

// rateLimitExceededHandler handles 429 responses with RFC 7807 format.
// It sets the Retry-After header and writes the error response (AC #5).
func rateLimitExceededHandler(w http.ResponseWriter, r *http.Request) {
	// Set Retry-After header (based on window)
	w.Header().Set("Retry-After", strconv.Itoa(int(RateLimitWindow.Seconds())))

	// Write RFC 7807 error response
	contract.WriteProblemJSON(w, r, &app.AppError{
		Op:      "RateLimiter",
		Code:    app.CodeRateLimitExceeded,
		Message: "Rate limit exceeded",
	})
}

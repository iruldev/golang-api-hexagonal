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

// DefaultRateLimitWindow is the default time window for rate limiting.
const DefaultRateLimitWindow = time.Second

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per window.
	// Required (> 0).
	RequestsPerSecond int

	// Window is the time window for rate limiting.
	// Optional (default: 1 second).
	Window time.Duration
}

// RateLimiter returns middleware that limits requests per key (user ID or IP).
// For authenticated requests, limits are per-user (claims.Subject).
// For unauthenticated requests, limits are per-IP.
//
// The middleware sets the following headers on ALL responses:
//   - X-RateLimit-Limit: Maximum requests allowed per window
//   - X-RateLimit-Remaining: Remaining requests in current window
//   - X-RateLimit-Reset: Unix timestamp when limit resets
//
// When rate limit is exceeded (429), additional headers are set:
//   - Retry-After: Seconds until retry is allowed
//
// AC #1: Unauthenticated requests use resolved client IP as key.
// AC #2: Authenticated requests use claims.Subject (userId) as key.
// AC #3,4: IP resolution relies on global RealIP middleware (respects TRUST_PROXY).
// AC #5: Returns 429 with RFC 7807 and Retry-After header.
// AC #6: Uses go-chi/httprate.
// AC #7: Rate limit is configurable via RATE_LIMIT_RPS.
// AC #8: Rate limit headers on all responses (Story 2.6).
func RateLimiter(cfg RateLimitConfig) func(http.Handler) http.Handler {
	// Validate config
	if cfg.RequestsPerSecond <= 0 {
		// Log warning or set sensible default? For middleware, usually panic on invalid startup config
		// or log. Here we'll default to 10 to ensure safety if not provided.
		// A panic might be better to signal misconfiguration, but let's be safe.
		cfg.RequestsPerSecond = 10
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultRateLimitWindow
	}

	return httprate.Limit(
		cfg.RequestsPerSecond,
		cfg.Window,
		httprate.WithKeyFuncs(keyFunc()),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			// AC #2: Calculate reset time
			// Try to get precise reset time from header set by httprate
			var resetTimeStr string
			if resetTimestampStr := w.Header().Get("X-RateLimit-Reset"); resetTimestampStr != "" {
				if ts, err := time.Parse(time.RFC3339, resetTimestampStr); err == nil {
					resetTimeStr = ts.Format(time.RFC3339)
				} else if ts, err := strconv.ParseInt(resetTimestampStr, 10, 64); err == nil {
					resetTimeStr = time.Unix(ts, 0).Format(time.RFC3339)
				}
			}

			// Fallback if header missing or unparseable
			if resetTimeStr == "" {
				resetTimeStr = time.Now().Add(cfg.Window).Format(time.RFC3339)
			}

			// AC #1: Write RFC 7807 error response with RATE-001 code and dynamic detail
			contract.WriteProblemJSON(w, r, &app.AppError{
				Op:      "RateLimiter",
				Code:    app.CodeRateLimitExceeded,
				Message: "Rate limit exceeded. Try again after " + resetTimeStr,
			})
		}),
		httprate.WithResponseHeaders(httprate.ResponseHeaders{
			Limit:      "X-RateLimit-Limit",
			Remaining:  "X-RateLimit-Remaining",
			Reset:      "X-RateLimit-Reset",
			RetryAfter: "Retry-After",
		}),
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

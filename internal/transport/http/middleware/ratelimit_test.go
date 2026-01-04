// Package middleware provides HTTP middleware for the transport layer.
// This file contains unit tests for rate limiting middleware.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// TestResolveClientIP tests the IP resolution logic.
func TestResolveClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		wantIP     string
	}{
		{
			name:       "RemoteAddr with port",
			remoteAddr: "203.0.113.50:12345",
			wantIP:     "203.0.113.50",
		},
		{
			name:       "RemoteAddr without port (normalized by RealIP)",
			remoteAddr: "192.168.1.1",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "IPv6 with port",
			remoteAddr: "[::1]:12345",
			wantIP:     "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			got := resolveClientIP(req)
			assert.Equal(t, tt.wantIP, got)
		})
	}
}

// TestKeyFunc tests the rate limit key function.
func TestKeyFunc(t *testing.T) {
	t.Run("unauthenticated request uses IP key", func(t *testing.T) {
		// AC #1: Unauthenticated requests use IP-based key
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		kf := keyFunc()
		key, err := kf(req)

		require.NoError(t, err)
		assert.Equal(t, "ip:192.168.1.1", key)
	})

	t.Run("authenticated request uses user ID key", func(t *testing.T) {
		// AC #2: Authenticated requests use claims.Subject (userId) as key
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		// Set claims in context
		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-12345",
			},
		}
		ctx := ctxutil.SetClaims(req.Context(), claims)
		req = req.WithContext(ctx)

		kf := keyFunc()
		key, err := kf(req)

		require.NoError(t, err)
		assert.Equal(t, "user:user-12345", key)
	})

	t.Run("authenticated request with empty subject falls back to IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:9999"

		// Set claims with empty subject
		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "   ", // whitespace only
			},
		}
		ctx := ctxutil.SetClaims(req.Context(), claims)
		req = req.WithContext(ctx)

		kf := keyFunc()
		key, err := kf(req)

		require.NoError(t, err)
		assert.Equal(t, "ip:10.0.0.1", key)
	})
}

// TestRateLimitExceededHandler removed as logic is now internal to RateLimiter closure.
// Verification is done via TestRateLimiterMiddleware.

// TestRateLimiterMiddleware tests the complete middleware integration.
func TestRateLimiterMiddleware(t *testing.T) {
	t.Run("requests under limit pass through", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 10,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := RateLimiter(cfg)(handler)

		// Make 5 requests (under the limit)
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("requests exceeding limit return 429", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 3,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		var successCount, rateLimitedCount int

		// Make more requests than the limit
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			// Use unique source to avoid race conditions
			req.RemoteAddr = "10.0.0.1:12345"
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			} else if rec.Code == http.StatusTooManyRequests {
				rateLimitedCount++
				// Verify Retry-After header
				assert.Equal(t, "1", rec.Header().Get("Retry-After"))
				// Verify Content-Type
				assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

				// AC #1, #2: Verify RFC 7807 body
				var problem testProblemDetail
				err := json.Unmarshal(rec.Body.Bytes(), &problem)
				require.NoError(t, err)

				assert.Equal(t, http.StatusTooManyRequests, problem.Status)
				assert.Equal(t, contract.CodeRateLimitExceeded, problem.Code)
				assert.Equal(t, "Rate Limit Exceeded", problem.Title)
				assert.True(t, strings.HasSuffix(problem.Type, "rate-limit-exceeded"))

				// AC #2: Verify detail includes reset time information
				assert.Contains(t, problem.Detail, "Try again after", "Detail should mention retry time")
				assert.Contains(t, problem.Detail, "20", "Detail should contain year in timestamp")
			}
		}

		// Some requests should succeed, some should be rate limited
		assert.LessOrEqual(t, successCount, cfg.RequestsPerSecond+1) // httprate may allow burst
		assert.Greater(t, rateLimitedCount, 0, "Expected some requests to be rate limited")
	})

	t.Run("different users have separate rate limits", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// User A makes requests
		userAClaims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-a",
			},
		}

		// User B makes requests
		userBClaims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-b",
			},
		}

		// Both users should be able to make requests independently
		for i := 0; i < 2; i++ {
			// User A request
			reqA := httptest.NewRequest(http.MethodGet, "/test", nil)
			reqA.RemoteAddr = "192.168.1.1:12345"
			ctxA := ctxutil.SetClaims(reqA.Context(), userAClaims)
			reqA = reqA.WithContext(ctxA)
			recA := httptest.NewRecorder()
			middleware.ServeHTTP(recA, reqA)
			assert.Equal(t, http.StatusOK, recA.Code, "User A request %d should succeed", i+1)

			// User B request
			reqB := httptest.NewRequest(http.MethodGet, "/test", nil)
			reqB.RemoteAddr = "192.168.1.2:12345"
			ctxB := ctxutil.SetClaims(reqB.Context(), userBClaims)
			reqB = reqB.WithContext(ctxB)
			recB := httptest.NewRecorder()
			middleware.ServeHTTP(recB, reqB)
			assert.Equal(t, http.StatusOK, recB.Code, "User B request %d should succeed", i+1)
		}
	})

	t.Run("custom window configuration", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
			Window:            2 * time.Second,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Exhaust the limit
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.3.1:12345"
			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Next request should be blocked
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.3.1:12345"
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)

		// Verify Retry-After is likely 1 or 2 (depending on window consumed)
		retryAfter := rec.Header().Get("Retry-After")
		assert.NotEmpty(t, retryAfter)
	})
}

// TestRateLimitConfig tests configuration struct.
func TestRateLimitConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		cfg := RateLimitConfig{}
		assert.Equal(t, 0, cfg.RequestsPerSecond)
	})

	t.Run("custom config values", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 100,
		}
		assert.Equal(t, 100, cfg.RequestsPerSecond)
	})
}

// BenchmarkKeyFunc benchmarks the key function performance.
func BenchmarkKeyFunc(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	kf := keyFunc()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kf(req)
	}
}

// BenchmarkKeyFuncWithClaims benchmarks key function with JWT claims.
func BenchmarkKeyFuncWithClaims(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-12345",
		},
	}
	ctx := ctxutil.SetClaims(context.Background(), claims)
	req = req.WithContext(ctx)

	kf := keyFunc()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kf(req)
	}
}

// BenchmarkResolveClientIP benchmarks IP resolution.
func BenchmarkResolveClientIP(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolveClientIP(req)
	}
}

// TestRateLimitHeaders tests that X-RateLimit-* headers are set on all responses.
// Story 2.6: Rate Limit Headers Enhancement.
func TestRateLimitHeaders(t *testing.T) {
	t.Run("headers present on successful requests", func(t *testing.T) {
		// AC #1: Response includes X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
		cfg := RateLimitConfig{
			RequestsPerSecond: 10,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := RateLimiter(cfg)(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.100.1:12345"
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		// Verify successful response
		assert.Equal(t, http.StatusOK, rec.Code)

		// AC #1: X-RateLimit-Limit header present
		limitHeader := rec.Header().Get("X-RateLimit-Limit")
		assert.NotEmpty(t, limitHeader, "X-RateLimit-Limit header should be present")
		limit, err := strconv.Atoi(limitHeader)
		require.NoError(t, err, "X-RateLimit-Limit should be a valid integer")
		assert.Equal(t, cfg.RequestsPerSecond, limit, "X-RateLimit-Limit should match configured RequestsPerSecond")

		// AC #1: X-RateLimit-Remaining header present
		remainingHeader := rec.Header().Get("X-RateLimit-Remaining")
		assert.NotEmpty(t, remainingHeader, "X-RateLimit-Remaining header should be present")
		remaining, err := strconv.Atoi(remainingHeader)
		require.NoError(t, err, "X-RateLimit-Remaining should be a valid integer")
		assert.GreaterOrEqual(t, remaining, 0, "X-RateLimit-Remaining should be >= 0")
		assert.Less(t, remaining, cfg.RequestsPerSecond, "X-RateLimit-Remaining should be less than limit after request")

		// AC #1: X-RateLimit-Reset header present
		resetHeader := rec.Header().Get("X-RateLimit-Reset")
		assert.NotEmpty(t, resetHeader, "X-RateLimit-Reset header should be present")
		resetTimestamp, err := strconv.ParseInt(resetHeader, 10, 64)
		require.NoError(t, err, "X-RateLimit-Reset should be a valid Unix timestamp")
		assert.Greater(t, resetTimestamp, time.Now().Unix()-10, "X-RateLimit-Reset should be a recent timestamp")
	})

	t.Run("remaining decrements with each request", func(t *testing.T) {
		// AC #1: X-RateLimit-Remaining decrements correctly
		cfg := RateLimitConfig{
			RequestsPerSecond: 5,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		var previousRemaining int = -1

		// Make 3 requests and verify remaining decrements
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.200.1:12345" // Unique IP
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			remainingHeader := rec.Header().Get("X-RateLimit-Remaining")
			require.NotEmpty(t, remainingHeader)
			remaining, err := strconv.Atoi(remainingHeader)
			require.NoError(t, err)

			if previousRemaining != -1 {
				assert.Equal(t, previousRemaining-1, remaining, "Remaining should decrement by 1")
			}
			previousRemaining = remaining
		}
	})

	t.Run("headers present on 429 responses", func(t *testing.T) {
		// AC #2: Rate limit headers still show current limit state on 429
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Exhaust the rate limit
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.201.1:12345" // Unique IP
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code == http.StatusTooManyRequests {
				// Verify rate limit headers on 429
				assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit should be present on 429")
				assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining should be present on 429")
				assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"), "X-RateLimit-Reset should be present on 429")
				assert.NotEmpty(t, rec.Header().Get("Retry-After"), "Retry-After should be present on 429")

				// Remaining should be 0 when rate limited
				remainingHeader := rec.Header().Get("X-RateLimit-Remaining")
				remaining, _ := strconv.Atoi(remainingHeader)
				assert.Equal(t, 0, remaining, "Remaining should be 0 when rate limited")
				return
			}
		}
		t.Fatal("Expected to hit rate limit")
	})

	t.Run("headers present for authenticated requests", func(t *testing.T) {
		// AC #1: Headers present for authenticated requests
		cfg := RateLimitConfig{
			RequestsPerSecond: 10,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Set up authenticated request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.202.1:12345"
		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-test-headers",
			},
		}
		ctx := ctxutil.SetClaims(req.Context(), claims)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		// Verify headers are present
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit should be present for authenticated requests")
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining should be present for authenticated requests")
		assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"), "X-RateLimit-Reset should be present for authenticated requests")
	})
}

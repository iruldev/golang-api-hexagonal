// Package middleware provides HTTP middleware for the transport layer.
// This file contains unit tests for rate limiting middleware.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// TestResolveClientIP tests the IP resolution logic.
func TestResolveClientIP(t *testing.T) {
	tests := []struct {
		name       string
		trustProxy bool
		xff        string
		xri        string
		remoteAddr string
		wantIP     string
	}{
		{
			name:       "TrustProxy=false uses RemoteAddr",
			trustProxy: false,
			xff:        "192.168.1.1",
			xri:        "10.0.0.1",
			remoteAddr: "203.0.113.50:12345",
			wantIP:     "203.0.113.50",
		},
		{
			name:       "TrustProxy=true with X-Forwarded-For single IP",
			trustProxy: true,
			xff:        "192.168.1.1",
			remoteAddr: "10.0.0.1:12345",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "TrustProxy=true with X-Forwarded-For multiple IPs uses first",
			trustProxy: true,
			xff:        "192.168.1.1, 10.0.0.1, 172.16.0.1",
			remoteAddr: "127.0.0.1:12345",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "TrustProxy=true with X-Forwarded-For with spaces",
			trustProxy: true,
			xff:        "  192.168.1.100  , 10.0.0.2",
			remoteAddr: "127.0.0.1:12345",
			wantIP:     "192.168.1.100",
		},
		{
			name:       "TrustProxy=true with X-Real-IP no X-Forwarded-For",
			trustProxy: true,
			xff:        "",
			xri:        "10.10.10.10",
			remoteAddr: "127.0.0.1:12345",
			wantIP:     "10.10.10.10",
		},
		{
			name:       "TrustProxy=true falls back to RemoteAddr if no headers",
			trustProxy: true,
			xff:        "",
			xri:        "",
			remoteAddr: "172.17.0.5:8080",
			wantIP:     "172.17.0.5",
		},
		{
			name:       "RemoteAddr without port",
			trustProxy: false,
			remoteAddr: "192.168.1.1",
			wantIP:     "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := resolveClientIP(req, tt.trustProxy)
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

		kf := keyFunc(false)
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

		kf := keyFunc(false)
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

		kf := keyFunc(false)
		key, err := kf(req)

		require.NoError(t, err)
		assert.Equal(t, "ip:10.0.0.1", key)
	})

	t.Run("TrustProxy=true with authenticated uses user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.100")

		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "admin-user",
			},
		}
		ctx := ctxutil.SetClaims(req.Context(), claims)
		req = req.WithContext(ctx)

		kf := keyFunc(true)
		key, err := kf(req)

		require.NoError(t, err)
		// Should use user ID, not IP
		assert.Equal(t, "user:admin-user", key)
	})

	t.Run("TrustProxy=true with unauthenticated uses X-Forwarded-For", func(t *testing.T) {
		// AC #3: With TRUST_PROXY=true, IP is extracted from X-Forwarded-For
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.100")

		kf := keyFunc(true)
		key, err := kf(req)

		require.NoError(t, err)
		assert.Equal(t, "ip:203.0.113.100", key)
	})
}

// TestRateLimitExceededHandler tests the 429 response handler.
func TestRateLimitExceededHandler(t *testing.T) {
	// AC #5: Returns 429 with RFC 7807 and Retry-After header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	rateLimitExceededHandler(rec, req)

	// Check status code
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Check Retry-After header
	assert.Equal(t, "1", rec.Header().Get("Retry-After"))

	// Check Content-Type
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	// Check RFC 7807 body
	var problem contract.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &problem)
	require.NoError(t, err)

	assert.Equal(t, http.StatusTooManyRequests, problem.Status)
	assert.Equal(t, app.CodeRateLimitExceeded, problem.Code)
	assert.Equal(t, "Too Many Requests", problem.Title)
	assert.True(t, strings.HasSuffix(problem.Type, "rate-limit-exceeded"))
	assert.Equal(t, "/api/v1/users", problem.Instance)
}

// TestRateLimiterMiddleware tests the complete middleware integration.
func TestRateLimiterMiddleware(t *testing.T) {
	t.Run("requests under limit pass through", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 10,
			TrustProxy:        false,
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
			TrustProxy:        false,
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
			}
		}

		// Some requests should succeed, some should be rate limited
		assert.LessOrEqual(t, successCount, cfg.RequestsPerSecond+1) // httprate may allow burst
		assert.Greater(t, rateLimitedCount, 0, "Expected some requests to be rate limited")
	})

	t.Run("different users have separate rate limits", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
			TrustProxy:        false,
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

	t.Run("context without claims uses IP-based limiting", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
			TrustProxy:        false,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Make requests without claims - should use IP-based limiting
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "172.16.0.100:9999"
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
}

// TestRateLimiterWithProxyTrust tests proxy header handling.
func TestRateLimiterWithProxyTrust(t *testing.T) {
	t.Run("TrustProxy=true uses X-Forwarded-For for rate limit key", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
			TrustProxy:        true,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Two different real IPs behind the same proxy should have separate limits
		for i := 0; i < 2; i++ {
			// Client 1
			req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
			req1.RemoteAddr = "127.0.0.1:12345"
			req1.Header.Set("X-Forwarded-For", "203.0.113.1")
			rec1 := httptest.NewRecorder()
			middleware.ServeHTTP(rec1, req1)
			assert.Equal(t, http.StatusOK, rec1.Code)

			// Client 2
			req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
			req2.RemoteAddr = "127.0.0.1:12345"
			req2.Header.Set("X-Forwarded-For", "203.0.113.2")
			rec2 := httptest.NewRecorder()
			middleware.ServeHTTP(rec2, req2)
			assert.Equal(t, http.StatusOK, rec2.Code)
		}
	})

	t.Run("TrustProxy=false ignores X-Forwarded-For", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 2,
			TrustProxy:        false,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RateLimiter(cfg)(handler)

		// Different X-Forwarded-For but same RemoteAddr should hit same limit
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "10.0.0.99:12345"
			req.Header.Set("X-Forwarded-For", "192.168."+string(rune('0'+i))+".1") // Different XFF each time
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			} else if rec.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		// Should be rate limited based on RemoteAddr, not X-Forwarded-For
		assert.LessOrEqual(t, successCount, cfg.RequestsPerSecond+1)
		assert.Greater(t, rateLimitedCount, 0, "Expected rate limiting to kick in")
	})
}

// TestRateLimitConfig tests configuration struct.
func TestRateLimitConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		cfg := RateLimitConfig{}
		assert.Equal(t, 0, cfg.RequestsPerSecond)
		assert.False(t, cfg.TrustProxy)
	})

	t.Run("custom config values", func(t *testing.T) {
		cfg := RateLimitConfig{
			RequestsPerSecond: 100,
			TrustProxy:        true,
		}
		assert.Equal(t, 100, cfg.RequestsPerSecond)
		assert.True(t, cfg.TrustProxy)
	})
}

// BenchmarkKeyFunc benchmarks the key function performance.
func BenchmarkKeyFunc(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	kf := keyFunc(false)

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

	kf := keyFunc(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kf(req)
	}
}

// BenchmarkResolveClientIP benchmarks IP resolution.
func BenchmarkResolveClientIP(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 172.16.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = resolveClientIP(req, true)
	}
}

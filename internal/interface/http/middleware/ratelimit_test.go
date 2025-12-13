package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

func TestTokenBucket_Allow(t *testing.T) {
	tests := []struct {
		name        string
		rate        float64 // tokens per second
		capacity    float64
		requests    int
		expectAllow []bool
	}{
		{
			name:        "all requests within limit",
			rate:        10,
			capacity:    5,
			requests:    5,
			expectAllow: []bool{true, true, true, true, true},
		},
		{
			name:        "requests exceed limit",
			rate:        2,
			capacity:    2,
			requests:    4,
			expectAllow: []bool{true, true, false, false},
		},
		{
			name:        "single request allowed",
			rate:        1,
			capacity:    1,
			requests:    2,
			expectAllow: []bool{true, false},
		},
		{
			name:        "high capacity burst",
			rate:        1,
			capacity:    10,
			requests:    10,
			expectAllow: []bool{true, true, true, true, true, true, true, true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			bucket := NewTokenBucket(tt.rate, tt.capacity)

			// Act & Assert
			for i, want := range tt.expectAllow {
				got := bucket.Allow()
				if got != want {
					t.Errorf("request %d: Allow() = %v, want %v", i+1, got, want)
				}
			}
		})
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	// Arrange: bucket with 2 tokens/sec, capacity 2
	bucket := NewTokenBucket(2.0, 2.0)

	// Act: exhaust all tokens
	bucket.Allow() // 1st token
	bucket.Allow() // 2nd token

	// Assert: should be denied now
	if bucket.Allow() {
		t.Error("expected denial after exhausting tokens")
	}

	// Act: wait for refill (1 token takes 0.5 seconds at 2 tokens/sec)
	time.Sleep(600 * time.Millisecond)

	// Assert: should be allowed after refill
	if !bucket.Allow() {
		t.Error("expected Allow after refill time")
	}
}

func TestTokenBucket_RetryAfter(t *testing.T) {
	// Arrange: bucket with 1 token/sec, capacity 1
	bucket := NewTokenBucket(1.0, 1.0)

	// Act: exhaust token
	bucket.Allow()

	// Assert: retry after should be about 1 second
	retryAfter := bucket.RetryAfter()
	if retryAfter != 1 {
		t.Errorf("RetryAfter() = %d, want 1", retryAfter)
	}
}

func TestTokenBucket_RetryAfter_WhenAvailable(t *testing.T) {
	// Arrange: full bucket
	bucket := NewTokenBucket(1.0, 1.0)

	// Assert: retry after should be 0 when tokens available
	retryAfter := bucket.RetryAfter()
	if retryAfter != 0 {
		t.Errorf("RetryAfter() = %d, want 0", retryAfter)
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	// Arrange: bucket with 1000 tokens capacity
	bucket := NewTokenBucket(1000.0, 1000.0)

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	// Act: 1000 concurrent requests
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if bucket.Allow() {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Assert: exactly 1000 should be allowed
	if allowedCount != 1000 {
		t.Errorf("concurrent requests allowed = %d, want 1000", allowedCount)
	}
}

func TestInMemoryRateLimiter_Allow(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(2, time.Second)),
	)
	defer limiter.Stop()

	ctx := context.Background()
	key := "test-user"

	tests := []struct {
		name        string
		expectAllow bool
	}{
		{"first request", true},
		{"second request", true},
		{"third request denied", false},
	}

	// Act & Assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := limiter.Allow(ctx, key)
			if err != nil {
				t.Fatalf("Allow() error = %v", err)
			}
			if allowed != tt.expectAllow {
				t.Errorf("Allow() = %v, want %v", allowed, tt.expectAllow)
			}
		})
	}
}

func TestInMemoryRateLimiter_DifferentKeys(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Second)),
	)
	defer limiter.Stop()

	ctx := context.Background()

	// Act & Assert: different keys have separate buckets
	allowed1, _ := limiter.Allow(ctx, "user-1")
	allowed2, _ := limiter.Allow(ctx, "user-2")

	if !allowed1 {
		t.Error("user-1 first request should be allowed")
	}
	if !allowed2 {
		t.Error("user-2 first request should be allowed")
	}

	// Both should be denied on second attempt
	denied1, _ := limiter.Allow(ctx, "user-1")
	denied2, _ := limiter.Allow(ctx, "user-2")

	if denied1 {
		t.Error("user-1 second request should be denied")
	}
	if denied2 {
		t.Error("user-2 second request should be denied")
	}
}

func TestInMemoryRateLimiter_Limit(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Minute)), // very restrictive default
	)
	defer limiter.Stop()

	ctx := context.Background()
	key := "premium-user"

	// Act: set custom rate (high limit)
	err := limiter.Limit(ctx, key, runtimeutil.NewRate(100, time.Second))
	if err != nil {
		t.Fatalf("Limit() error = %v", err)
	}

	// Assert: should allow many requests
	allowedCount := 0
	for i := 0; i < 100; i++ {
		allowed, _ := limiter.Allow(ctx, key)
		if allowed {
			allowedCount++
		}
	}

	if allowedCount != 100 {
		t.Errorf("allowed count = %d, want 100", allowedCount)
	}
}

func TestInMemoryRateLimiter_RetryAfter(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Second)),
	)
	defer limiter.Stop()

	ctx := context.Background()
	key := "test-user"

	// Act: exhaust limit
	limiter.Allow(ctx, key)

	// Assert: retry after should be ~1 second
	retryAfter := limiter.RetryAfter(key)
	if retryAfter != 1 {
		t.Errorf("RetryAfter() = %d, want 1", retryAfter)
	}
}

func TestInMemoryRateLimiter_Concurrent(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
	)
	defer limiter.Stop()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Act: concurrent requests for same key
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Allow(ctx, "shared-key")
		}()
	}
	wg.Wait()

	// Assert: should not panic (thread-safe)
	// If we got here without panic, test passes
}

func TestRateLimitMiddleware_AllowsRequest(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(10, time.Second)),
	)
	defer limiter.Stop()

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRateLimitMiddleware_BlocksExcessiveRequests(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Minute)),
	)
	defer limiter.Stop()

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Second request should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitMiddleware_RetryAfterHeader(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Second)),
	)
	defer limiter.Stop()

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Exhaust limit
	handler.ServeHTTP(rec, req)

	// Act: second request
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	// Assert: Retry-After header should be set
	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Retry-After header not set")
	}
}

func TestRateLimitMiddleware_ErrorResponseFormat(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Minute)),
	)
	defer limiter.Stop()

	handler := RateLimitMiddleware(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Exhaust limit
	handler.ServeHTTP(rec, req)

	// Act: second request
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	// Assert: response should be JSON with correct format
	var errResp response.ErrorResponse
	if err := json.Unmarshal(rec2.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if errResp.Success {
		t.Error("response.success = true, want false")
	}
	if errResp.Error.Code != ErrRateLimited {
		t.Errorf("error.code = %s, want %s", errResp.Error.Code, ErrRateLimited)
	}
}

func TestRateLimitMiddleware_CustomKeyExtractor(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Minute)),
	)
	defer limiter.Stop()

	// Custom key extractor using a header
	customExtractor := func(r *http.Request) string {
		return r.Header.Get("X-Custom-Key")
	}

	handler := RateLimitMiddleware(limiter, WithKeyExtractor(customExtractor))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// First request with key "user-A"
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Custom-Key", "user-A")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request with different key "user-B" - should be allowed
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Custom-Key", "user-B")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// Assert
	if rec1.Code != http.StatusOK {
		t.Errorf("user-A request: status = %d, want %d", rec1.Code, http.StatusOK)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("user-B request: status = %d, want %d", rec2.Code, http.StatusOK)
	}
}

func TestRateLimitMiddleware_FixedRetryAfter(t *testing.T) {
	// Arrange
	limiter := NewInMemoryRateLimiter(
		WithDefaultRate(runtimeutil.NewRate(1, time.Minute)),
	)
	defer limiter.Stop()

	handler := RateLimitMiddleware(limiter, WithRetryAfterSeconds(120))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Exhaust limit
	handler.ServeHTTP(rec, req)

	// Act: second request
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	// Assert: Retry-After should be 120
	if rec2.Header().Get("Retry-After") != "120" {
		t.Errorf("Retry-After = %s, want 120", rec2.Header().Get("Retry-After"))
	}
}

func TestIPKeyExtractor(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			want:       "192.168.1.1",
		},
		{
			name:       "behind proxy with X-Forwarded-For",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			want:       "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For with multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			want:       "203.0.113.50",
		},
		{
			name:       "with X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.60"},
			want:       "203.0.113.60",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"X-Real-IP":       "203.0.113.60",
			},
			want: "203.0.113.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := IPKeyExtractor(req)
			if got != tt.want {
				t.Errorf("IPKeyExtractor() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUserIDKeyExtractor(t *testing.T) {
	tests := []struct {
		name       string
		claims     *Claims
		remoteAddr string
		want       string
	}{
		{
			name:       "with authenticated user",
			claims:     &Claims{UserID: "user-123"},
			remoteAddr: "192.168.1.1:12345",
			want:       "user:user-123",
		},
		{
			name:       "without claims falls back to IP",
			claims:     nil,
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
		{
			name:       "empty UserID falls back to IP",
			claims:     &Claims{UserID: ""},
			remoteAddr: "192.168.1.1:12345",
			want:       "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.claims != nil {
				ctx := NewContext(req.Context(), *tt.claims)
				req = req.WithContext(ctx)
			}

			got := UserIDKeyExtractor(req)
			if got != tt.want {
				t.Errorf("UserIDKeyExtractor() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRateLimitMiddleware_FailOpen(t *testing.T) {
	// Arrange: Create a limiter that always errors
	errorLimiter := &erroringRateLimiter{}

	handler := RateLimitMiddleware(errorLimiter)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert: Should allow request through (fail-open)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (fail-open)", rec.Code, http.StatusOK)
	}
}

// erroringRateLimiter is a test limiter that always returns an error.
type erroringRateLimiter struct{}

func (e *erroringRateLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return false, context.DeadlineExceeded
}

func (e *erroringRateLimiter) Limit(_ context.Context, _ string, _ runtimeutil.Rate) error {
	return context.DeadlineExceeded
}

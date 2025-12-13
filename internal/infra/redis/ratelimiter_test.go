package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// setupMiniredis creates a miniredis instance for testing.
func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return mr, client
}

func TestRedisRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		period      time.Duration
		requests    int
		expectAllow []bool
	}{
		{
			name:        "within limit - all allowed",
			limit:       5,
			period:      time.Minute,
			requests:    3,
			expectAllow: []bool{true, true, true},
		},
		{
			name:        "at limit - all allowed",
			limit:       3,
			period:      time.Minute,
			requests:    3,
			expectAllow: []bool{true, true, true},
		},
		{
			name:        "exceeds limit - last rejected",
			limit:       3,
			period:      time.Minute,
			requests:    5,
			expectAllow: []bool{true, true, true, false, false},
		},
		{
			name:        "single request - allowed",
			limit:       100,
			period:      time.Minute,
			requests:    1,
			expectAllow: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			_, client := setupMiniredis(t)
			limiter := NewRedisRateLimiter(client,
				WithRedisDefaultRate(runtimeutil.NewRate(tt.limit, tt.period)),
			)
			ctx := context.Background()

			// Act & Assert
			for i := 0; i < tt.requests; i++ {
				allowed, err := limiter.Allow(ctx, "test-key")
				require.NoError(t, err)
				assert.Equal(t, tt.expectAllow[i], allowed, "request %d", i+1)
			}
		})
	}
}

func TestRedisRateLimiter_Allow_DifferentKeys(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(2, time.Minute)),
	)
	ctx := context.Background()

	// Act & Assert - each key should have independent limit
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "user-1")
		require.NoError(t, err)
		if i < 2 {
			assert.True(t, allowed, "user-1 request %d should be allowed", i+1)
		} else {
			assert.False(t, allowed, "user-1 request %d should be rate limited", i+1)
		}
	}

	// user-2 should have fresh limit
	allowed, err := limiter.Allow(ctx, "user-2")
	require.NoError(t, err)
	assert.True(t, allowed, "user-2 first request should be allowed")
}

func TestRedisRateLimiter_Limit(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
	)
	ctx := context.Background()

	// Set custom rate for specific key
	err := limiter.Limit(ctx, "vip-user", runtimeutil.NewRate(1000, time.Minute))
	require.NoError(t, err)

	// Act - VIP user should have higher limit
	for i := 0; i < 500; i++ {
		allowed, err := limiter.Allow(ctx, "vip-user")
		require.NoError(t, err)
		assert.True(t, allowed, "vip-user request %d should be allowed", i+1)
	}
}

func TestRedisRateLimiter_KeyPrefix(t *testing.T) {
	// Arrange
	mr, client := setupMiniredis(t)
	prefix := "api:v1:"
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(10, time.Minute)),
		WithKeyPrefix(prefix),
	)
	ctx := context.Background()

	// Act
	_, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)

	// Assert - key should have prefix
	keys := mr.Keys()
	assert.Len(t, keys, 1)
	assert.Equal(t, prefix+"test-key", keys[0])
}

func TestRedisRateLimiter_Timeout(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisTimeout(50*time.Millisecond),
	)

	// Assert - timeout should be set
	assert.Equal(t, 50*time.Millisecond, limiter.timeout)
}

func TestRedisRateLimiter_RetryAfter(t *testing.T) {
	// Arrange
	mr, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(1, time.Minute)),
	)
	ctx := context.Background()

	// Exhaust the limit
	_, _ = limiter.Allow(ctx, "test-key")
	_, _ = limiter.Allow(ctx, "test-key")

	// Fast forward miniredis time slightly to ensure TTL is set
	mr.FastForward(1 * time.Second)

	// Act
	retryAfter := limiter.RetryAfter("test-key")

	// Assert - should return remaining TTL (approximately 59 seconds)
	assert.Greater(t, retryAfter, 0)
	assert.LessOrEqual(t, retryAfter, 60)
}

func TestRedisRateLimiter_RetryAfter_NoKey(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client)

	// Act
	retryAfter := limiter.RetryAfter("nonexistent-key")

	// Assert
	assert.Equal(t, 0, retryAfter)
}

func TestRedisRateLimiter_FallbackOnRedisFailure(t *testing.T) {
	// Arrange - use a mock fallback limiter
	fallback := &mockRateLimiter{allowResult: true}
	mr, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithFallbackLimiter(fallback),
		WithCircuitBreakerConfig(1, 30*time.Second), // Open after 1 failure
	)
	ctx := context.Background()

	// Close miniredis to simulate Redis failure
	mr.Close()

	// Act
	allowed, err := limiter.Allow(ctx, "test-key")

	// Assert - should use fallback
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.True(t, fallback.allowCalled)
}

func TestRedisRateLimiter_CircuitBreaker_Opens(t *testing.T) {
	// Arrange
	fallback := &mockRateLimiter{allowResult: true}
	mr, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithFallbackLimiter(fallback),
		WithCircuitBreakerConfig(2, 30*time.Second), // Open after 2 failures
	)
	ctx := context.Background()

	// First request should work
	allowed, err := limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Close Redis to cause failures
	mr.Close()

	// These should cause circuit breaker to open (2 failures)
	_, _ = limiter.Allow(ctx, "test-key")
	_, _ = limiter.Allow(ctx, "test-key")

	// Reset call count
	fallback.allowCalled = false

	// Next request should skip Redis entirely (circuit open)
	allowed, err = limiter.Allow(ctx, "test-key")
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.True(t, fallback.allowCalled, "should use fallback when circuit is open")
}

func TestRedisRateLimiter_CircuitBreaker_Recovery(t *testing.T) {
	// Arrange
	cb := newCircuitBreaker(2, 100*time.Millisecond)

	// Record failures to open circuit
	cb.recordFailure()
	cb.recordFailure()

	// Assert circuit is open
	assert.True(t, cb.isOpen())

	// Wait for recovery time
	time.Sleep(150 * time.Millisecond)

	// Assert circuit is closed (recovery attempt)
	assert.False(t, cb.isOpen())
}

func TestRedisRateLimiter_FailOpen_NoFallback(t *testing.T) {
	// Arrange - no fallback configured
	mr, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client) // No fallback
	ctx := context.Background()

	// Close Redis to simulate failure
	mr.Close()

	// Act
	allowed, err := limiter.Allow(ctx, "test-key")

	// Assert - fail-open, allow request
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestRedisRateLimiter_LuaScriptAtomic(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(10, time.Minute)),
	)
	ctx := context.Background()

	// Act - make concurrent requests
	results := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			allowed, _ := limiter.Allow(ctx, "concurrent-key")
			results <- allowed
		}()
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < 20; i++ {
		if <-results {
			allowedCount++
		}
	}

	// Assert - exactly 10 should be allowed (atomic operations)
	assert.Equal(t, 10, allowedCount, "exactly 10 requests should be allowed")
}

func TestRedisRateLimiter_WindowReset(t *testing.T) {
	// Arrange
	mr, client := setupMiniredis(t)
	limiter := NewRedisRateLimiter(client,
		WithRedisDefaultRate(runtimeutil.NewRate(2, time.Second)), // 2 per second
	)
	ctx := context.Background()

	// Exhaust limit
	_, _ = limiter.Allow(ctx, "test-key")
	_, _ = limiter.Allow(ctx, "test-key")
	allowed, _ := limiter.Allow(ctx, "test-key")
	assert.False(t, allowed, "should be rate limited")

	// Fast forward past window
	mr.FastForward(2 * time.Second)

	// Act - should be allowed again
	allowed, err := limiter.Allow(ctx, "test-key")

	// Assert
	require.NoError(t, err)
	assert.True(t, allowed, "should be allowed after window reset")
}

func TestNewRedisRateLimiter_Defaults(t *testing.T) {
	// Arrange
	_, client := setupMiniredis(t)

	// Act
	limiter := NewRedisRateLimiter(client)

	// Assert defaults
	assert.Equal(t, 100, limiter.defaultRate.Limit)
	assert.Equal(t, time.Minute, limiter.defaultRate.Period)
	assert.Equal(t, "rl:", limiter.keyPrefix)
	assert.Equal(t, 100*time.Millisecond, limiter.timeout)
	assert.NotNil(t, limiter.circuit)
}

func TestIsNoScriptError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "redis.Nil error",
			err:      redis.Nil,
			expected: false,
		},
		{
			name:     "NOSCRIPT full message",
			err:      fmt.Errorf("NOSCRIPT No matching script. Please use EVAL."),
			expected: true,
		},
		{
			name:     "NOSCRIPT prefix only",
			err:      fmt.Errorf("NOSCRIPT some other message"),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNoScriptError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockRateLimiter is a mock implementation of runtimeutil.RateLimiter for testing.
type mockRateLimiter struct {
	allowResult bool
	allowCalled bool
}

func (m *mockRateLimiter) Allow(_ context.Context, _ string) (bool, error) {
	m.allowCalled = true
	return m.allowResult, nil
}

func (m *mockRateLimiter) Limit(_ context.Context, _ string, _ runtimeutil.Rate) error {
	return nil
}

func (m *mockRateLimiter) RetryAfter(_ string) int {
	return 60
}

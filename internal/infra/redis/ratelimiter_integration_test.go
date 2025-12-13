//go:build integration

package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	infraredis "github.com/iruldev/golang-api-hexagonal/internal/infra/redis"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// setupRedisContainer starts a Redis container for integration testing.
func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	// Verify connection
	err = client.Ping(ctx).Err()
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestIntegration_RedisRateLimiter_MultiInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Arrange - create two limiter instances pointing to same Redis
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter1 := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(5, time.Minute)),
		infraredis.WithKeyPrefix("test:integration:"),
	)
	limiter2 := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(5, time.Minute)),
		infraredis.WithKeyPrefix("test:integration:"),
	)

	ctx := context.Background()
	key := "shared-user"

	// Act - make requests from both instances
	var allowedCount int
	for i := 0; i < 3; i++ {
		if allowed, _ := limiter1.Allow(ctx, key); allowed {
			allowedCount++
		}
		if allowed, _ := limiter2.Allow(ctx, key); allowed {
			allowedCount++
		}
	}

	// Assert - both instances should share the same limit
	// Total requests: 6, limit: 5, so 5 should be allowed
	assert.Equal(t, 5, allowedCount, "exactly 5 requests should be allowed across instances")
}

func TestIntegration_RedisRateLimiter_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Arrange
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
		infraredis.WithKeyPrefix("test:concurrent:"),
	)

	ctx := context.Background()
	key := "concurrent-test"

	// Act - make concurrent requests
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowedCount := 0

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if allowed, _ := limiter.Allow(ctx, key); allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Assert - exactly 100 should be allowed (atomic operations)
	assert.Equal(t, 100, allowedCount, "exactly 100 concurrent requests should be allowed")
}

func TestIntegration_RedisRateLimiter_WindowExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Arrange
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Use short window for faster test
	limiter := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(2, 2*time.Second)),
		infraredis.WithKeyPrefix("test:expiry:"),
	)

	ctx := context.Background()
	key := "expiry-test"

	// Exhaust limit
	_, _ = limiter.Allow(ctx, key)
	_, _ = limiter.Allow(ctx, key)
	allowed, _ := limiter.Allow(ctx, key)
	assert.False(t, allowed, "should be rate limited")

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// Act - should be allowed again
	allowed, err := limiter.Allow(ctx, key)

	// Assert
	require.NoError(t, err)
	assert.True(t, allowed, "should be allowed after window expiry")
}

func TestIntegration_RedisRateLimiter_RetryAfter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Arrange
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(1, time.Minute)),
		infraredis.WithKeyPrefix("test:retry:"),
	)

	ctx := context.Background()
	key := "retry-test"

	// Exhaust limit
	_, _ = limiter.Allow(ctx, key)
	_, _ = limiter.Allow(ctx, key)

	// Act
	retryAfter := limiter.RetryAfter(key)

	// Assert - should return approximate TTL (around 60 seconds)
	assert.Greater(t, retryAfter, 0)
	assert.LessOrEqual(t, retryAfter, 60)
}

func TestIntegration_RedisRateLimiter_DifferentKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Arrange
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter := infraredis.NewRedisRateLimiter(client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(2, time.Minute)),
		infraredis.WithKeyPrefix("test:keys:"),
	)

	ctx := context.Background()

	// Act - exhaust limit for key1
	_, _ = limiter.Allow(ctx, "key1")
	_, _ = limiter.Allow(ctx, "key1")
	limited, _ := limiter.Allow(ctx, "key1")

	// key2 should have its own limit
	allowed, _ := limiter.Allow(ctx, "key2")

	// Assert
	assert.False(t, limited, "key1 should be rate limited")
	assert.True(t, allowed, "key2 should not be affected by key1's limit")
}

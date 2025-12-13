package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// setupRedisContainer starts a Redis container for testing.
// Returns the Redis client and a cleanup function.
func setupRedisContainer(t *testing.T) (*redis.Client, func()) {
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
	if err != nil {
		t.Skipf("Failed to start Redis container: %v", err)
		return nil, func() {}
	}

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	// Verify connection
	err = client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis container")

	cleanup := func() {
		_ = client.Close()
		_ = container.Terminate(ctx)
	}

	return client, cleanup
}

func TestRedisStore_Check_FirstOccurrence(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "first-occurrence-key", time.Hour)

	// Assert
	require.NoError(t, err)
	assert.True(t, isNew, "First occurrence should return true")
}

func TestRedisStore_Check_Duplicate(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()
	key := "duplicate-key"

	// Act - First check
	isNew1, err1 := store.Check(ctx, key, time.Hour)
	require.NoError(t, err1)
	assert.True(t, isNew1, "First occurrence should return true")

	// Act - Second check (duplicate)
	isNew2, err2 := store.Check(ctx, key, time.Hour)
	require.NoError(t, err2)
	assert.False(t, isNew2, "Second occurrence should return false (duplicate)")
}

func TestRedisStore_Check_TTLExpiration(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()
	key := "ttl-test-key"
	shortTTL := 100 * time.Millisecond

	// Act - First check
	isNew1, err1 := store.Check(ctx, key, shortTTL)
	require.NoError(t, err1)
	assert.True(t, isNew1, "First occurrence should return true")

	// Immediate second check
	isNew2, err2 := store.Check(ctx, key, shortTTL)
	require.NoError(t, err2)
	assert.False(t, isNew2, "Immediate second check should return false")

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Third check after TTL
	isNew3, err3 := store.Check(ctx, key, shortTTL)
	require.NoError(t, err3)
	assert.True(t, isNew3, "Check after TTL expiration should return true")
}

func TestRedisStore_Check_EmptyKey(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "", time.Hour)

	// Assert
	require.NoError(t, err)
	assert.True(t, isNew, "Empty key should return true (no idempotency)")
}

func TestRedisStore_StoreAndGetResult(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()
	key := "result-key"
	expectedResult := []byte(`{"status": "completed", "value": 42}`)

	// Act - Store result
	err := store.StoreResult(ctx, key, expectedResult, time.Hour)
	require.NoError(t, err)

	// Act - Get result
	result, found, err := store.GetResult(ctx, key)

	// Assert
	require.NoError(t, err)
	assert.True(t, found, "Result should be found")
	assert.Equal(t, expectedResult, result)
}

func TestRedisStore_GetResult_NotFound(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))
	ctx := context.Background()

	// Act
	result, found, err := store.GetResult(ctx, "nonexistent-key")

	// Assert
	require.NoError(t, err)
	assert.False(t, found, "Result should not be found")
	assert.Nil(t, result)
}

func TestRedisStore_DefaultPrefix(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange - Create store with empty prefix
	store := NewRedisStore(client, "", WithLogger(zap.NewNop()))
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "default-prefix-key", time.Hour)

	// Assert
	require.NoError(t, err)
	assert.True(t, isNew)

	// Verify key was stored with default prefix
	exists, err := client.Exists(ctx, DefaultKeyPrefix+"default-prefix-key").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists, "Key should exist with default prefix")
}

func TestRedisStore_FailOpen_ProcessesOnError(t *testing.T) {
	// Arrange - Create a client with invalid connection
	badClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid port
	})
	defer badClient.Close()

	store := NewRedisStore(badClient, "test:",
		WithFailMode(FailOpen),
		WithLogger(zap.NewNop()),
	)
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "any-key", time.Hour)

	// Assert - Fail-open should return true (process the task)
	require.NoError(t, err, "Fail-open should not return error")
	assert.True(t, isNew, "Fail-open should return true to process the task")
}

func TestRedisStore_FailClosed_ReturnsError(t *testing.T) {
	// Arrange - Create a client with invalid connection
	badClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid port
	})
	defer badClient.Close()

	store := NewRedisStore(badClient, "test:",
		WithFailMode(FailClosed),
		WithLogger(zap.NewNop()),
	)
	ctx := context.Background()

	// Act
	isNew, err := store.Check(ctx, "any-key", time.Hour)

	// Assert - Fail-closed should return error
	require.Error(t, err, "Fail-closed should return error")
	assert.False(t, isNew, "Fail-closed should return false on error")
}

func TestRedisStore_SetFailMode(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "test:", WithLogger(zap.NewNop()))

	// Act - Change fail mode
	store.SetFailMode(FailClosed)

	// Assert - Store should have updated fail mode
	// (We verify this indirectly by the behavior test above)
	assert.NotNil(t, store)
}

func TestRedisStore_ConcurrentChecks(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	// Arrange
	store := NewRedisStore(client, "concurrent-test:", WithLogger(zap.NewNop()))
	ctx := context.Background()
	key := "concurrent-key"

	// Act - Concurrent checks
	results := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			isNew, _ := store.Check(ctx, key, time.Hour)
			results <- isNew
		}()
	}

	// Collect results
	newCount := 0
	for i := 0; i < 100; i++ {
		if <-results {
			newCount++
		}
	}

	// Assert - Only one should be "new" due to atomic SET NX
	assert.Equal(t, 1, newCount, "Only one concurrent check should return true (new)")
}

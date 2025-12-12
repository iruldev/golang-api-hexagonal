//go:build integration
// +build integration

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// containerTimeout is the default timeout for container operations.
// May need to be increased in slow CI environments.
const containerTimeout = 3 * time.Minute

func TestPostgresContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), containerTimeout)
	defer cancel()

	// Start PostgreSQL container
	container, err := NewPostgresContainer(ctx)
	require.NoError(t, err, "should start PostgreSQL container")
	defer func() {
		// Use background context for cleanup to avoid cancellation issues
		cleanupCtx := context.Background()
		if termErr := container.Terminate(cleanupCtx); termErr != nil {
			t.Logf("warning: failed to terminate postgres container: %v", termErr)
		}
	}()

	// Verify DSN is not empty
	assert.NotEmpty(t, container.DSN, "DSN should not be empty")

	// Setup test database
	pool, cleanup, err := SetupTestDatabase(ctx, container.DSN)
	require.NoError(t, err, "should setup test database")
	defer cleanup()

	// Verify we can query the database
	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err, "should execute query")
	assert.Equal(t, 1, result)

	// Verify notes table exists
	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'notes'
		)
	`).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "notes table should exist")
}

func TestRedisContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), containerTimeout)
	defer cancel()

	// Start Redis container
	container, err := NewRedisContainer(ctx)
	require.NoError(t, err, "should start Redis container")
	defer func() {
		// Use background context for cleanup to avoid cancellation issues
		cleanupCtx := context.Background()
		if termErr := container.Terminate(cleanupCtx); termErr != nil {
			t.Logf("warning: failed to terminate redis container: %v", termErr)
		}
	}()

	// Verify address is not empty
	assert.NotEmpty(t, container.Addr, "address should not be empty")
	assert.Contains(t, container.Addr, ":", "address should contain port")

	// Verify Redis actually works by performing a PING
	client := redis.NewClient(&redis.Options{
		Addr: container.Addr,
	})
	defer client.Close()

	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err, "should ping Redis")
	assert.Equal(t, "PONG", pong, "should receive PONG response")

	// Test SET/GET operation
	err = client.Set(ctx, "test-key", "test-value", 0).Err()
	require.NoError(t, err, "should SET value")

	val, err := client.Get(ctx, "test-key").Result()
	require.NoError(t, err, "should GET value")
	assert.Equal(t, "test-value", val, "should receive correct value")
}

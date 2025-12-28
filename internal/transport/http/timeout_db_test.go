//go:build integration

package http

import (
	"context"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// TestDBQueryTimeout verifies that DB queries respect context timeout (AC2).
func TestDBQueryTimeout(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*DockerContainer).startLogProducer.func1"),
		)
	})

	pool := containers.NewPostgres(t)
	containers.MigrateWithPath(t, pool, "../../../migrations")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute slow query - should timeout
	_, err := pool.Exec(ctx, "SELECT pg_sleep(5)")

	// Should timeout with context error
	assert.Error(t, err, "query should timeout")
	assert.ErrorIs(t, err, context.DeadlineExceeded,
		"error should be deadline exceeded")
}

// TestDBQueryWithinTimeout verifies successful queries complete within timeout (AC2).
func TestDBQueryWithinTimeout(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*DockerContainer).startLogProducer.func1"),
		)
	})

	pool := containers.NewPostgres(t)
	containers.MigrateWithPath(t, pool, "../../../migrations")

	// Create context with generous timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute fast query - should complete
	var result int
	err := pool.QueryRow(ctx, "SELECT 1").Scan(&result)

	assert.NoError(t, err, "fast query should complete within timeout")
	assert.Equal(t, 1, result)
}

// TestMultipleQueriesTimeout verifies multiple queries all respect timeout (AC2).
func TestMultipleQueriesTimeout(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*DockerContainer).startLogProducer.func1"),
		)
	})

	pool := containers.NewPostgres(t)
	containers.MigrateWithPath(t, pool, "../../../migrations")

	// Create context with short timeout shared across queries
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	const numQueries = 3
	errCh := make(chan error, numQueries)

	// Start multiple slow queries with shared context
	for i := 0; i < numQueries; i++ {
		go func() {
			_, err := pool.Exec(ctx, "SELECT pg_sleep(5)")
			errCh <- err
		}()
	}

	// All should timeout
	for i := 0; i < numQueries; i++ {
		select {
		case err := <-errCh:
			assert.Error(t, err, "query %d should timeout", i)
		case <-time.After(3 * time.Second):
			t.Fatalf("query %d did not return in time", i)
		}
	}
}

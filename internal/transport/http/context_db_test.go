//go:build integration

package http

import (
	"context"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/testutil/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// TestContextPropagationToDB verifies that context cancellation propagates
// to downstream database operations (AC3).
func TestContextPropagationToDB(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			// Ignore testcontainers cleanup goroutines
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*DockerContainer).startLogProducer.func1"),
		)
	})

	// Create PostgreSQL container
	pool := containers.NewPostgres(t)
	containers.MigrateWithPath(t, pool, "../../../migrations")

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Channel to receive query result
	errCh := make(chan error, 1)

	// Start a slow query that respects context
	go func() {
		// pg_sleep respects context cancellation
		_, err := pool.Exec(ctx, "SELECT pg_sleep(10)")
		errCh <- err
	}()

	// Wait for query to be active in Postgres
	waitForActiveQuery(t, pool, "SELECT pg_sleep(10)")

	// Cancel context mid-query
	cancel()

	// Verify query was cancelled
	select {
	case err := <-errCh:
		assert.Error(t, err, "query should be cancelled")
		assert.Contains(t, err.Error(), "cancel",
			"error should indicate cancellation")
	case <-time.After(3 * time.Second):
		t.Fatal("query did not cancel in time - context propagation failed")
	}
}

// TestMultipleDBOperationsCancellation verifies that multiple concurrent
// DB operations are all cancelled when context is cancelled.
func TestMultipleDBOperationsCancellation(t *testing.T) {
	t.Cleanup(func() {
		goleak.VerifyNone(t,
			goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*DockerContainer).startLogProducer.func1"),
		)
	})

	pool := containers.NewPostgres(t)
	containers.MigrateWithPath(t, pool, "../../../migrations")

	ctx, cancel := context.WithCancel(context.Background())

	const numQueries = 3
	errCh := make(chan error, numQueries)

	// Start multiple slow queries
	for i := 0; i < numQueries; i++ {
		go func() {
			_, err := pool.Exec(ctx, "SELECT pg_sleep(10)")
			errCh <- err
		}()
	}

	// Wait for queries to be active
	// We just check if at least one is active, or we could wait for count
	waitForActiveQuery(t, pool, "SELECT pg_sleep(10)")

	// Cancel all
	cancel()

	// Collect results
	var errors []error
	for i := 0; i < numQueries; i++ {
		select {
		case err := <-errCh:
			errors = append(errors, err)
		case <-time.After(3 * time.Second):
			t.Fatalf("query %d did not cancel in time", i)
		}
	}

	// All should have errored
	assert.Len(t, errors, numQueries, "all queries should return")
	for i, err := range errors {
		assert.Error(t, err, "query %d should be cancelled", i)
	}
}

// waitForActiveQuery polls pg_stat_activity until the expected query appears
func waitForActiveQuery(t *testing.T, pool *pgxpool.Pool, querySnippet string) {
	deadline := time.Now().Add(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		var count int
		// Check for active query, excluding this check itself
		err := pool.QueryRow(context.Background(),
			"SELECT count(*) FROM pg_stat_activity WHERE query LIKE $1 AND query NOT LIKE '%pg_stat_activity%' AND state = 'active'",
			"%"+querySnippet+"%").Scan(&count)

		if err == nil && count > 0 {
			return
		}

		<-ticker.C
	}
	t.Fatalf("timeout waiting for active query containing %q", querySnippet)
}

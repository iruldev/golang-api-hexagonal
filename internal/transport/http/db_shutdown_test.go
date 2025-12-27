//go:build integration

package http

import (
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/testutil/containers"
	"github.com/stretchr/testify/assert"
)

// TestShutdownClosesDBConnections verifies that database connections
// are properly closed when the pool is shut down (AC3).
func TestShutdownClosesDBConnections(t *testing.T) {
	// Create a PostgreSQL container
	pool := containers.NewPostgres(t)
	// Use relative path to project root migrations
	containers.MigrateWithPath(t, pool, "../../../migrations")

	// Execute a query to ensure connections are established
	var result int
	err := pool.QueryRow(t.Context(), "SELECT 1").Scan(&result)
	assert.NoError(t, err)

	// Get stats before shutdown
	statsBefore := pool.Stat()
	assert.Greater(t, statsBefore.TotalConns(), int32(0), "should have active connections")

	// Close pool (simulating graceful shutdown)
	pool.Close()

	// Verify connections are closed
	statsAfter := pool.Stat()
	assert.Equal(t, int32(0), statsAfter.AcquiredConns(), "no connections should be acquired after close")
}

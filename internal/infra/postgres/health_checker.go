// Package postgres provides PostgreSQL database adapters.
package postgres

import (
	"context"
	"time"
)

// pingable is an interface for types that can be pinged.
type pingable interface {
	Ping(ctx context.Context) error
}

// DatabaseHealthChecker checks PostgreSQL connection health for readiness probes.
// It implements handler.DependencyChecker interface.
type DatabaseHealthChecker struct {
	pool pingable
}

// NewDatabaseHealthChecker creates a new database health checker.
// Accepts any type that implements Ping(ctx) method, such as postgres.Pooler.
func NewDatabaseHealthChecker(pool pingable) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{pool: pool}
}

// Name returns "database" as the dependency identifier.
func (c *DatabaseHealthChecker) Name() string {
	return "database"
}

// CheckHealth pings the database and measures latency.
// Returns "healthy" if ping succeeds, "unhealthy" with error if it fails.
// The context should include a timeout for the check.
func (c *DatabaseHealthChecker) CheckHealth(ctx context.Context) (string, time.Duration, error) {
	start := time.Now()
	err := c.pool.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return "unhealthy", latency, err
	}
	return "healthy", latency, nil
}

// =============================================================================
// Story 3.4: Health Check Library Integration
// =============================================================================

// NewDatabaseCheck returns a healthcheck.Check function for database connectivity.
// This is compatible with heptiolabs/healthcheck library.
// The timeout parameter controls how long to wait for the ping before timing out.
func NewDatabaseCheck(pool pingable, timeout time.Duration) func() error {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return pool.Ping(ctx)
	}
}

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

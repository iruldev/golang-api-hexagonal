// Package testing provides test helpers for integration tests.
package testing

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a testcontainers PostgreSQL container.
type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

// NewPostgresContainer starts a PostgreSQL container for testing.
// Returns the container wrapper with DSN for connecting.
// Caller must call Terminate() when done.
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	return &PostgresContainer{
		Container: container,
		DSN:       dsn,
	}, nil
}

// Terminate stops and removes the PostgreSQL container.
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// RedisContainer wraps a testcontainers Redis container.
type RedisContainer struct {
	Container testcontainers.Container
	Addr      string
}

// NewRedisContainer starts a Redis container for testing.
// Returns the container wrapper with address for connecting.
// Caller must call Terminate() when done.
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	container, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start redis container: %w", err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get endpoint: %w", err)
	}

	return &RedisContainer{
		Container: container,
		Addr:      endpoint,
	}, nil
}

// Terminate stops and removes the Redis container.
func (c *RedisContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

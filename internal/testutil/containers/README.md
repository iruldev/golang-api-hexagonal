# Container Helpers Package

This package provides helpers for integration testing with Docker containers.

## Quick Start

```go
func TestUserRepo(t *testing.T) {
    // Start PostgreSQL container
    pool := containers.NewPostgres(t)
    
    // Apply migrations
    containers.Migrate(t, pool)
    
    // Test with transaction isolation (rollback after test)
    containers.WithTx(t, pool, func(tx pgx.Tx) {
        repo := postgres.NewUserRepo(tx)
        // ... test code
    })
}
```

## Helpers

### NewPostgres(t) *pgxpool.Pool

Creates a PostgreSQL container and returns a connected pool.
Container is automatically cleaned up when test ends.

### Migrate(t, pool)

Applies goose migrations from `migrations/` directory.

### WithTx(t, pool, fn)

Runs function in a transaction that is rolled back after.
Provides test isolation without truncate overhead.

### Truncate(t, pool, tables...)

Truncates specified tables with CASCADE.
Use when transaction isolation isn't suitable.

## Patterns

### Transaction Isolation (Recommended)

Use `WithTx` for most tests - it's faster than truncate:

```go
containers.WithTx(t, pool, func(tx pgx.Tx) {
    // Changes are automatically rolled back
})
```

### Truncate Cleanup

Use `Truncate` when you need committed data across tests:

```go
t.Cleanup(func() {
    containers.Truncate(t, pool, "users", "audit_events")
})
```

## Requirements

- Docker must be running
- Use `-short` flag to skip container tests: `go test -short ./...`

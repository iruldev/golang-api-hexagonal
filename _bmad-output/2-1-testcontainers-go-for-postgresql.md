# Story 2.1: testcontainers-go for PostgreSQL

Status: done

## Story

As a **developer**,
I want containerized database tests,
so that tests are reproducible without external setup.

## Acceptance Criteria

1. **AC1:** `github.com/testcontainers/testcontainers-go` added to dependencies
2. **AC2:** `containers.NewPostgres(t)` returns working pool
3. **AC3:** Container auto-cleanup via `t.Cleanup()`
4. **AC4:** Container starts in ≤30 seconds

## Tasks / Subtasks

- [x] Task 1: Add testcontainers-go dependency (AC: #1)
  - [x] Run `go get github.com/testcontainers/testcontainers-go`
  - [x] Run `go get github.com/testcontainers/testcontainers-go/modules/postgres`
  - [x] Run `go mod tidy`
- [x] Task 2: Implement NewPostgres helper (AC: #2, #3, #4)
  - [x] Create `internal/testutil/containers/postgres.go`
  - [x] Implement `NewPostgres(t)` function
  - [x] Return `*pgxpool.Pool` ready to use
  - [x] Register cleanup with `t.Cleanup()`
- [x] Task 3: Add test for container helper (AC: #2, #4)
  - [x] Create `containers/postgres_test.go`
  - [x] Verify container starts in ≤30 seconds
  - [x] Verify pool is usable (ping)
- [x] Task 4: Update containers package doc (AC: #2)
  - [x] Update `containers.go` with usage examples

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-004:** testcontainers-go for integration testing
- **Pattern 4:** Container lifecycle management

### NewPostgres Implementation

```go
// internal/testutil/containers/postgres.go
package containers

import (
    "context"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

// NewPostgres creates a PostgreSQL container for testing.
// Returns a pool connected to the container.
// Container is automatically cleaned up when test ends.
func NewPostgres(t testing.TB) *pgxpool.Pool {
    t.Helper()
    ctx := context.Background()

    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    if err != nil {
        t.Fatalf("failed to start postgres container: %v", err)
    }

    t.Cleanup(func() {
        if err := container.Terminate(ctx); err != nil {
            t.Errorf("failed to terminate container: %v", err)
        }
    })

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("failed to get connection string: %v", err)
    }

    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatalf("failed to create pool: %v", err)
    }

    t.Cleanup(func() {
        pool.Close()
    })

    return pool
}
```

### Testing Standards

- Run `go test ./internal/testutil/containers/...` to verify
- Container startup should complete in ≤30 seconds
- Verify pool can execute simple query

### Docker Requirement

- Tests require Docker daemon running
- CI must have Docker available

### Previous Epic Learnings (Epic 1)

- Use `t.Helper()` in test helpers
- Register cleanup with `t.Cleanup()` for auto-teardown
- Use consistent error handling patterns

### References

- [Source: _bmad-output/architecture.md#AD-004 testcontainers-go]
- [Source: _bmad-output/epics.md#Story 2.1]
- [Source: _bmad-output/prd.md#FR16, NFR4]
- [testcontainers-go/modules/postgres](https://golang.testcontainers.org/modules/postgres/)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

- [x] Code Review (AI): Verified implementation, fixed task status sync, ensured git tracking, and ran `go mod tidy`. All tests passed.

### File List

_Files created/modified during implementation:_
- [x] `go.mod` (add testcontainers-go)
- [x] `internal/testutil/containers/postgres.go` (new)
- [x] `internal/testutil/containers/postgres_test.go` (new)
- [x] `internal/testutil/containers/containers.go` (update doc)

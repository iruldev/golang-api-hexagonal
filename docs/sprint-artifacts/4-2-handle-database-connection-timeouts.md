# Story 4.2: Handle Database Connection Timeouts

Status: done

## Story

As a SRE,
I want database connections to timeout gracefully,
So that slow database doesn't hang the application.

## Acceptance Criteria

### AC1: Connection timeout on startup
**Given** database is unreachable
**When** connection attempt times out
**Then** error is logged with context
**And** application fails startup gracefully

### AC2: Query timeout with context cancellation
**Given** query takes longer than timeout
**When** context is cancelled
**Then** query is cancelled
**And** appropriate error is returned

---

## Tasks / Subtasks

- [x] **Task 1: Add connection timeout config** (AC: #1)
  - [x] Add `DB_CONN_TIMEOUT` to DatabaseConfig (default: 10s)
  - [x] Update .env.example with new variable

- [x] **Task 2: Implement connection timeout** (AC: #1)
  - [x] Create context with timeout for NewPool
  - [x] Wrap connection errors with timeout context
  - [x] Log connection failure with host/port info

- [x] **Task 3: Add query timeout helper** (AC: #2)
  - [x] Add `DB_QUERY_TIMEOUT` to DatabaseConfig (default: 30s)
  - [x] Create `QueryContext(ctx, timeout)` helper
  - [x] Handle context.DeadlineExceeded error

- [x] **Task 4: Create domain error for timeout** (AC: #1, #2)
  - [x] Add `ErrTimeout` to domain/errors.go
  - [x] Map to HTTP 504 Gateway Timeout

- [x] **Task 5: Create timeout tests** (AC: #1, #2)
  - [x] Test: Connection timeout returns error
  - [x] Test: Query timeout cancels operation
  - [x] Test: Timeout errors have proper messages

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Config Extension

```go
// internal/config/config.go
type DatabaseConfig struct {
    // ... existing fields
    ConnTimeout  time.Duration `koanf:"conn_timeout"`  // 10s default
    QueryTimeout time.Duration `koanf:"query_timeout"` // 30s default
}
```

### Connection Timeout

```go
// internal/infra/postgres/postgres.go
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
    timeout := cfg.Database.ConnTimeout
    if timeout == 0 {
        timeout = 10 * time.Second
    }

    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("connection timeout after %s: %w", timeout, err)
        }
        return nil, fmt.Errorf("create pool: %w", err)
    }
    // ...
}
```

### Query Timeout Helper

```go
// internal/infra/postgres/timeout.go
package postgres

import (
    "context"
    "time"
)

// QueryContext returns a context with query timeout.
func QueryContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    return context.WithTimeout(ctx, timeout)
}
```

### Domain Error

```go
// internal/domain/errors.go
// ErrTimeout indicates an operation timed out.
// Maps to HTTP 504 Gateway Timeout.
var ErrTimeout = errors.New("operation timed out")
```

### Environment Variables

```env
# Connection timeout (how long to wait for initial connection)
DB_CONN_TIMEOUT=10s

# Query timeout (how long to wait for queries)
DB_QUERY_TIMEOUT=30s
```

### Architecture Compliance

**Layer:** `internal/infra/postgres`, `internal/domain`
**Pattern:** Timeout handling at infrastructure layer
**Benefit:** Prevents hanging operations, improves SRE visibility

### References

- [Source: docs/epics.md#Story-4.2]
- [Story 4.1 - PostgreSQL Connection](file:///docs/sprint-artifacts/4-1-setup-postgresql-connection-with-pgx.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Second story in Epic 4: Database & Persistence.
Adds timeout handling for database operations.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/infra/postgres/timeout.go` - Query timeout helper

Files to modify:
- `internal/config/config.go` - Add timeout fields
- `internal/infra/postgres/postgres.go` - Add connection timeout
- `internal/domain/errors.go` - Add ErrTimeout
- `internal/interface/http/response/mapper.go` - Map ErrTimeout
- `.env.example` - Add timeout variables
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

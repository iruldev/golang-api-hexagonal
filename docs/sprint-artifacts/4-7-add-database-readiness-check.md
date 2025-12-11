# Story 4.7: Add Database Readiness Check

Status: done

## Story

As a SRE,
I want database connectivity checked in readiness probe,
So that unhealthy instances are removed from load balancer.

## Acceptance Criteria

### AC1: Database connected returns 200
**Given** database is connected
**When** `/readyz` is requested
**Then** response is 200

### AC2: Database disconnected returns 503
**Given** database connection is lost
**When** `/readyz` is requested
**Then** response is 503
**And** response body indicates database unavailable

---

## Tasks / Subtasks

- [x] **Task 1: Create DB health checker interface** (AC: #1, #2)
  - [x] Create `DBHealthChecker` interface with `Ping()` method
  - [x] Implement for pgxpool

- [x] **Task 2: Update readiness handler** (AC: #1, #2)
  - [x] Add DB checker dependency to handler
  - [x] Call Ping() and return appropriate status

- [x] **Task 3: Wire up database pool to handler** (AC: #1, #2)
  - [x] Pass pool to readiness handler in main.go
  - [x] Handle nil pool case gracefully

- [x] **Task 4: Add tests** (AC: #1, #2)
  - [x] Test with mock healthy DB
  - [x] Test with mock unhealthy DB
  - [x] Test nil pool scenario

- [x] **Task 5: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### DBHealthChecker Interface

```go
// internal/infra/postgres/health.go
package postgres

import "context"

// DBHealthChecker checks database health
type DBHealthChecker interface {
    Ping(ctx context.Context) error
}

// PoolHealthChecker implements DBHealthChecker for pgxpool
type PoolHealthChecker struct {
    pool *pgxpool.Pool
}

func NewPoolHealthChecker(pool *pgxpool.Pool) *PoolHealthChecker {
    return &PoolHealthChecker{pool: pool}
}

func (c *PoolHealthChecker) Ping(ctx context.Context) error {
    return c.pool.Ping(ctx)
}
```

### Updated Readiness Handler

```go
// internal/interface/http/handlers/health.go
func (h *ReadyzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Check database if available
    if h.dbChecker != nil {
        if err := h.dbChecker.Ping(ctx); err != nil {
            response.Error(w, http.StatusServiceUnavailable, response.ErrServiceUnavailable, "database unavailable")
            return
        }
    }
    
    response.Success(w, HealthData{Status: "ready"})
}
```

### Response Body for 503

```json
{
  "success": false,
  "error": {
    "code": "ERR_SERVICE_UNAVAILABLE",
    "message": "database unavailable"
  }
}
```

### Architecture Compliance

**Layer:** `internal/interface/http/handlers/` + `internal/infra/postgres/`
**Pattern:** Dependency injection for testability
**Benefit:** Readiness reflects actual database connectivity

### References

- [Source: docs/epics.md#Story-4.7]
- [Story 4.1 - PostgreSQL Connection](file:///docs/sprint-artifacts/4-1-setup-postgresql-connection-with-pgx.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Seventh and final story in Epic 4: Database & Persistence.
Completes the epic by integrating DB into health checks.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/infra/postgres/health.go` - DB health checker

Files to modify:
- `internal/interface/http/handlers/health.go` - Add DB check to readyz
- `internal/interface/http/response/errors.go` - Add ErrDatabaseUnavailable
- `cmd/server/main.go` - Wire up pool to handler
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

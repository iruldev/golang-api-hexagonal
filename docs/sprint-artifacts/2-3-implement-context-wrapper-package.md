# Story 2.3: Implement Context Wrapper Package

Status: done

## Story

As a developer,
I want wrapper functions for DB and HTTP with mandatory context,
So that context propagation is consistent.

## Acceptance Criteria

1. **Given** `internal/infra/wrapper/` package
   **When** I call `wrapper.Query(ctx, pool, query, args...)`
   **Then** context is propagated correctly
   **And** default timeout is applied if context has no deadline

2. **Given** a wrapper function for HTTP client
   **When** I call `wrapper.DoHTTP(ctx, client, req)`
   **Then** context is propagated to the request
   **And** default timeout is applied if context has no deadline

3. **Given** a wrapper function for Redis operations
   **When** I call `wrapper.DoRedis(ctx, func())`
   **Then** context is checked before execution
   **And** returns `context.Canceled` or `context.DeadlineExceeded` if context is done

4. **Given** any wrapper function
   **When** context deadline is already set
   **Then** the existing deadline is preserved (not overwritten)
   **And** the wrapper does not add a new timeout

5. **Given** the wrapper package
   **When** imported by `internal/infra/` packages
   **Then** linter passes (respects layer boundaries: infra → domain only)
   **And** all tests pass with `make verify`

## Tasks / Subtasks

- [x] Task 1: Create wrapper package structure (AC: #1, #2, #3, #5)
  - [x] 1.1 Create `internal/infra/wrapper/` directory
  - [x] 1.2 Create `doc.go` with package documentation
  - [x] 1.3 Define default timeout constants (30s for query, 30s for HTTP)

- [x] Task 2: Implement DB wrapper functions (AC: #1, #4)
  - [x] 2.1 Create `db.go` with `Query/QueryRow/Exec` wrapper functions
  - [x] 2.2 Port/refactor existing `postgres/timeout.go` logic into wrapper
  - [x] 2.3 Implement deadline detection: only add timeout if no deadline set
  - [x] 2.4 Return context error immediately if context is already done
  - [x] 2.5 Create `db_test.go` with comprehensive tests

- [x] Task 3: Implement HTTP client wrapper (AC: #2, #4)
  - [x] 3.1 Create `http.go` with `DoRequest` wrapper function
  - [x] 3.2 Implement `req.WithContext(ctx)` propagation
  - [x] 3.3 Implement deadline detection for timeout
  - [x] 3.4 Create `http_test.go` with comprehensive tests

- [x] Task 4: Implement Redis wrapper (AC: #3, #4)
  - [x] 4.1 Create `redis.go` with context-aware wrapper function
  - [x] 4.2 Implement context check before execution
  - [x] 4.3 Create `redis_test.go` with comprehensive tests

- [x] Task 5: Integration and cleanup (AC: #5)
  - [x] 5.1 Update `internal/infra/postgres/timeout.go` to deprecate with migration guide
  - [x] 5.2 Run `make lint` to verify layer boundaries
  - [x] 5.3 Run `make verify` to ensure all tests pass
  - [x] 5.4 Update documentation if needed

## Dev Notes

### Architecture Decision Reference

**Decision 3: Context Propagation Enforcement (Both)** from `docs/architecture-decisions.md`:
- Linter `contextcheck` enforces context as first parameter
- Wrapper pattern for DB, Redis, HTTP provides defense in depth
- Default timeout applied when context has no deadline

**Target Implementation Pattern:**
```go
// internal/infra/wrapper/db.go
package wrapper

import (
    "context"
    "time"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

const (
    DefaultQueryTimeout = 30 * time.Second
    DefaultHTTPTimeout  = 30 * time.Second
)

// Query wraps pool.Query with context timeout.
// If ctx has no deadline, DefaultQueryTimeout is applied.
func Query(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) (pgx.Rows, error) {
    // Return early if context is already cancelled
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    
    // Add timeout only if no deadline is set
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, DefaultQueryTimeout)
        defer cancel()
    }
    
    return pool.Query(ctx, query, args...)
}

// internal/infra/wrapper/http.go
func DoRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    
    req = req.WithContext(ctx)
    
    // Add timeout only if no deadline is set
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, DefaultHTTPTimeout)
        defer cancel()
        req = req.WithContext(ctx)
    }
    
    return client.Do(req)
}
```

### Existing State Analysis

**Current `internal/infra/postgres/timeout.go`:**
```go
func QueryContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    if timeout == 0 {
        timeout = DefaultQueryTimeout
    }
    return context.WithTimeout(ctx, timeout)
}
```
- Already has partial implementation of timeout logic
- Needs to be refactored into wrapper package for reuse
- Current approach always adds timeout (should check deadline first)

**Current `internal/infra/postgres/note/repository.go`:**
- All methods accept `ctx context.Context` as first parameter
- Directly calls `r.queries.XYZ(ctx, ...)` without timeout wrapper
- Will benefit from wrapper functions for consistent timeout

**Current `internal/infra/redis/redis.go`:**
- Has Ping method with `ctx context.Context` parameter
- Exposes raw `*redis.Client` for direct access
- Wrapper can add context checks for operations

### Key Implementation Points

1. **Deadline Detection Pattern:**
   ```go
   if _, ok := ctx.Deadline(); !ok {
       ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
       defer cancel()
   }
   ```

2. **Early Return on Cancelled Context:**
   ```go
   if err := ctx.Err(); err != nil {
       return nil, err
   }
   ```

3. **Layer Boundary Compliance:**
   - `internal/infra/wrapper/` is in infra layer
   - Can only import: `domain` (if needed)
   - Cannot import: `usecase`, `interface`
   - Can be used by other `infra` packages

### File Structure

```
internal/infra/wrapper/              # [NEW] Context wrapper package
├── doc.go                           # [NEW] Package documentation
├── db.go                            # [NEW] Database wrapper functions
├── db_test.go                       # [NEW] Database wrapper tests
├── http.go                          # [NEW] HTTP client wrapper
├── http_test.go                     # [NEW] HTTP wrapper tests
├── redis.go                         # [NEW] Redis operation wrapper
└── redis_test.go                    # [NEW] Redis wrapper tests

internal/infra/postgres/
├── timeout.go                       # [REFACTOR or DEPRECATE]
```

### Critical Points from Previous Stories

From Story 2.1 & 2.2 learnings:
- All tests must pass with `make verify`
- UPPER_SNAKE error codes if custom errors needed
- Use `ctxutil.RequestIDFromContext(ctx)` for trace correlation
- Test coverage is critical - CI enforces lint+test
- Follow existing code patterns in infra layer

### Testing Strategy

1. **Unit Tests:**
   - Wrapper functions correctly propagate context
   - Default timeout applied when no deadline
   - Existing deadline preserved when set
   - Early return on cancelled context
   - `context.DeadlineExceeded` handling

2. **Mock Tests:**
   - Use mock DB pool for database tests
   - Use httptest for HTTP client tests
   - Use mock Redis for Redis tests

3. **Table-Driven Tests:**
   ```go
   tests := []struct {
       name           string
       ctxFactory     func() context.Context
       expectTimeout  bool
       expectError    bool
   }{
       {"no deadline", context.Background, true, false},
       {"with deadline", withDeadline, false, false},
       {"cancelled", cancelled, false, true},
   }
   ```

### NFR Targets

| NFR | Requirement | Verification |
|-----|-------------|--------------|
| FR26 | All IO functions receive context first | Wrapper enforces pattern |
| FR28 | Wrapper pattern for DB, Redis, HTTP | Package implementation |
| FR29 | Default timeout in wrappers | Test verification |
| NFR-M1 | Coverage ≥80% for wrapper | Unit tests |

### Dependencies

- **Story 2.2 (Done):** Central error registry (may use for wrapper errors)
- **Story 2.1 (Done):** Response envelope (context utilities)
- **Story 2.4 (Future):** HTTP error mapping will use wrapper for outgoing calls

### Critical Points

1. **No breaking changes:** Existing code continues to work
2. **Optional adoption:** Wrappers are opt-in helpers, not mandatory refactor
3. **Layer boundaries:** Wrapper stays in infra layer
4. **Deadline preservation:** Never overwrite existing context deadline
5. **Testing:** Comprehensive tests for all scenarios

### References

- [Source: docs/epics.md#Story 2.3](file:///docs/epics.md) - FR26, FR28, FR29
- [Source: docs/architecture-decisions.md#Decision 3](file:///docs/architecture-decisions.md) - Context propagation pattern
- [Source: project_context.md](file:///project_context.md) - Context propagation rules
- [Source: internal/infra/postgres/timeout.go](file:///internal/infra/postgres/timeout.go) - Existing timeout logic
- [Source: internal/infra/postgres/note/repository.go](file:///internal/infra/postgres/note/repository.go) - Current usage pattern
- [Source: internal/infra/redis/redis.go](file:///internal/infra/redis/redis.go) - Redis client pattern
- [Source: docs/sprint-artifacts/2-2-create-central-error-code-registry.md](file:///docs/sprint-artifacts/2-2-create-central-error-code-registry.md) - Previous story learnings

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 2: API Standards & Response Contract (MVP) - in-progress
- Previous story: Story 2.2 (Create Central Error Code Registry) - done

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

None required.

### Completion Notes List

- ✅ Created `internal/infra/wrapper/` package with full context propagation enforcement
- ✅ DB wrapper: `Query`, `QueryRow`, `QueryRowWithCancel`, `Exec` with deadline detection
- ✅ HTTP wrapper: `DoRequest`, `DoRequestWithClient`, `TimeoutTransport` for flexible usage
- ✅ Redis wrapper: `DoRedis`, `DoRedisResult` (generic), `PingRedis` with context check
- ✅ All wrappers implement: early return on cancelled context, deadline preservation, default 30s timeout
- ✅ Deprecated `postgres/timeout.go` with migration guide to new wrapper package
- ✅ 32 unit tests with 96.9% coverage on wrapper package
- ✅ All tests pass with `make verify` - lint and unit tests successful

### File List

**New Files:**
- `internal/infra/wrapper/doc.go` - Package documentation
- `internal/infra/wrapper/db.go` - Database wrapper functions (Query, QueryRow, QueryRowWithCancel, Exec)
- `internal/infra/wrapper/db_test.go` - Database wrapper tests (11 tests)
- `internal/infra/wrapper/http.go` - HTTP wrapper functions (DoRequest, DoRequestWithClient, TimeoutTransport)
- `internal/infra/wrapper/http_test.go` - HTTP wrapper tests (10 tests)
- `internal/infra/wrapper/redis.go` - Redis wrapper functions (DoRedis, DoRedisResult, PingRedis)
- `internal/infra/wrapper/redis_test.go` - Redis wrapper tests (11 tests)

**Modified Files:**
- `internal/infra/postgres/timeout.go` - Added deprecation notice with migration guide
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status to review

## Senior Developer Review (AI)

**Reviewer:** Antigravity (Senior Reviewer Persona)
**Date:** 2025-12-15

**Findings:**
- **CRITICAL:** Resource leak in `db.go` (QueryRow context not cancelled). FIXED.
- **MEDIUM:** Inconsistent Redis timeout handling. FIXED.
- **MEDIUM:** Documentation error in `doc.go`. FIXED.

**Outcome:**
- All issues fixed automatically.
- Tests verified passing.
- Story approved.


## Senior Developer Review (AI) - Round 2

**Reviewer:** Antigravity (Senior Reviewer Persona)
**Date:** 2025-12-15

**Findings:**
- **CRITICAL:** Premature context cancellation in `DoRequest`. FIXED (Applied `cancelBody` wrapper).
- **CRITICAL:** Premature context cancellation in `Query`. FIXED (Applied `cancelRows` wrapper).
- **LOW:** Redis default timeout documentation missing. FIXED.

**Outcome:**
- Complex context safety bugs resolved.
- Validation tests passed.
- Story approved.


## Final Validation (AI)

**Date:** 2025-12-15
**Status:** Verified
**Notes:** Re-ran full validation. All context safety fixes are present and correct. Tests passing. Story marked as DONE.

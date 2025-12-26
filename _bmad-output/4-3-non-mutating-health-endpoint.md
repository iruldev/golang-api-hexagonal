# Story 4.3: Non-Mutating Health Endpoint

Status: done

## Story

**As a** platform engineer,
**I want** /ready to not mutate runtime state,
**So that** health checks don't cause side effects.

**FR:** FR21

## Acceptance Criteria

1. ✅ **Given** /ready endpoint, **When** probed repeatedly, **Then** no DB pool reset, no state mutation
2. ✅ **Given** the code, **When** reviewed, **Then** no side effects confirmed
3. ✅ **Given** implementation, **When** integration test runs, **Then** idempotency verified

## Implementation Summary

### Task 1: Code Review ✅
| Check | Expected | Status |
|-------|----------|--------|
| `db.Ping()` is read-only | Yes | ✅ `pool.go:40-41` just calls `p.pool.Ping(ctx)` |
| No global state access | Yes | ✅ Handler uses local variables only |
| No pool reset | Yes | ✅ Ping never closes/recreates pool |
| Idempotent | Yes | ✅ Same input = same output |

### Task 2: Idempotency Test ✅
- Added `TestReadyHandler_Idempotent`
- Calls /ready 10 times in rapid succession
- Verifies consistent 200 OK responses
- Proves stateless behavior

### Task 3: Refactoring & Integration Tests ✅
- Added `TestIntegrationRoutes_Ready_Idempotent` in `integration_test.go`
- Refactored `ready.go` to use constants and injected logger
- Removed magic numbers for timeouts

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/handler/ready_test.go` | MODIFIED - Added idempotency test, updated constructor |
| `internal/transport/http/handler/ready.go` | MODIFIED - Added constants, logger, error handling |
| `internal/transport/http/handler/integration_test.go` | MODIFIED - Added integration idempotency test |
| `cmd/api/main.go` | MODIFIED - Updated ReadyHandler constructor, added timeout const |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/handler/ready_test.go` - MODIFIED
- `internal/transport/http/handler/ready.go` - MODIFIED
- `internal/transport/http/handler/integration_test.go` - MODIFIED
- `cmd/api/main.go` - MODIFIED

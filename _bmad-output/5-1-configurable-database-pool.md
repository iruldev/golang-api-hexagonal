# Story 5.1: Configurable Database Pool

Status: done

## Story

**As a** platform engineer,
**I want** database pool configurable via environment,
**So that** I can tune performance per deployment.

**FR:** FR36

## Acceptance Criteria

1. ✅ **Given** `DB_POOL_MAX_CONNS`, `DB_POOL_MIN_CONNS`, `DB_POOL_MAX_LIFETIME`, **When** app starts, **Then** pool is configured
2. ✅ **Given** no env vars, **When** app starts, **Then** defaults are used (max=25, min=5, lifetime=1h)
3. ✅ **Given** configuration, **When** unit tests run, **Then** configuration is verified

## Implementation Summary

### Task 1: Add config variables ✅
- Added `DBPoolMaxConns` (default 25)
- Added `DBPoolMinConns` (default 5)
- Added `DBPoolMaxLifetime` (default 1h)

### Task 2: Apply config to pool ✅
- Added `PoolConfig` struct to `postgres/pool.go`
- Modified `NewPool` to accept `PoolConfig`
- Updated `main.go` to pass config values

### Task 3: Add unit tests ✅
- `TestLoad_DBPool_Defaults` - verifies AC#2
- `TestLoad_DBPool_Custom` - verifies AC#1

## Changes

| File | Change |
|------|--------|
| `internal/infra/config/config.go` | MODIFIED - Added DB pool config fields |
| `internal/infra/postgres/pool.go` | MODIFIED - Added PoolConfig, updated NewPool |
| `cmd/api/main.go` | MODIFIED - Passes pool config to newReconnectingDB |
| `internal/infra/config/config_test.go` | MODIFIED - Added pool config tests |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/infra/config/config.go` - MODIFIED
- `internal/infra/postgres/pool.go` - MODIFIED
- `cmd/api/main.go` - MODIFIED
- `internal/infra/config/config_test.go` - MODIFIED
- `internal/infra/config/config_pool_validation_test.go` - NEW
- `internal/infra/postgres/pool_test.go` - NEW

## Senior Developer Review (AI)

**Date:** 2025-12-26
**Reviewer:** AI Agent

**Outcome:** Approved

### Findings
- **Medium:** Missing key validation for pool configuration (Min <= Max, > 0). Fixed by adding validation logic to `config.go`.
- **Medium:** Pool configuration logic in `postgres` package was untestable. Fixed by refactoring `NewPool` to extract configuration logic and adding unit tests in `pool_test.go`.
- **Low:** Missing validation for `DBPoolMaxLifetime` (> 0). Fixed by enforcing positive duration in `config.go`.

### Actions Taken
- Implemented robust validation for `DBPoolMaxConns`, `DBPoolMinConns`.
- Refactored `postgres.NewPool` to use testable `getPGXPoolConfig`.
- Added `internal/infra/config/config_pool_validation_test.go`.
- Added `internal/infra/postgres/pool_test.go`.
- Enforced `DBPoolMaxLifetime > 0` validation.

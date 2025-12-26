# Story 5.3: sqlc for Users Module

Status: done

## Story

**As a** developer,
**I want** sqlc-generated typed queries for Users,
**So that** SQL is type-safe and reviewed.

**FR:** FR38

## Acceptance Criteria

1. ✅ **Given** sqlc.yaml configuration, **When** `make generate` is run, **Then** Users module queries are generated
2. ✅ **Given** generated code, **When** inspected, **Then** queries are in infra layer only (not domain)
3. ✅ **Given** generated code, **When** tests run, **Then** unit tests use generated code (integration tested via UserRepo refactor)

## Implementation Summary

### Task 1: Setup sqlc ✅
- Created `sqlc.yaml` with pgx/v5 driver
- Created `queries/users.sql` with 5 queries

### Task 2: Configure output ✅
- Output to `internal/infra/postgres/sqlcgen/`
- Queries in infra layer (domain not affected)

### Task 3: Makefile targets ✅
- Added sqlc v1.28.0 to `make setup`
- Added `make generate` target

### Task 4: Generated files ✅
- `sqlcgen/db.go`
- `sqlcgen/models.go`
- `sqlcgen/users.sql.go`

## Changes

| File | Change |
|------|--------|
| `sqlc.yaml` | NEW |
| `queries/users.sql` | NEW |
| `internal/infra/postgres/sqlcgen/*.go` | GENERATED |
| `Makefile` | MODIFIED |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `sqlc.yaml` - NEW
- `queries/users.sql` - NEW
- `internal/infra/postgres/sqlcgen/db.go` - GENERATED
- `internal/infra/postgres/sqlcgen/models.go` - GENERATED
- `internal/infra/postgres/sqlcgen/users.sql.go` - GENERATED
- `Makefile` - MODIFIED
- `internal/infra/postgres/user_repo.go` - REFACTORED (matches sqlcgen)

### senior_developer_review_ai
_Reviewer: AI_Agent on 2025-12-26_

- [x] **AC Validation**: AC3 was missing (generated code unused). Fixed by refactoring `UserRepo` to use `sqlcgen`.
- [x] **File Tracking**: Untracked files (`sqlc.yaml`, `queries/`, `sqlcgen/`) added to git.
- [x] **Tests**: Verified `UserRepo` tests pass with `sqlcgen` integration.
- [x] **Outcome**: Approved with fixes applied.


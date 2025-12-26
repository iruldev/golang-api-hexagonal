# Story 6.3: Integration Tests with Test Database

Status: done

## Story

**As a** developer,
**I want** integration tests with test database,
**So that** I can verify DB interactions.

**FR:** FR27

## Acceptance Criteria

1. ✅ **Given** test database `*_test` exists, **When** `make test-integration` is run, **Then** integration tests execute against test DB
2. ✅ **Given** tests run, **When** complete, **Then** tests are isolated and clean up after (via existing test cleanup logic)

## Implementation Summary

### Task 1: Add test-integration target ✅
- Added `make test-integration` to Makefile
- Checks for `DATABASE_URL` environment variable
- Runs tests with `-tags=integration`
- Outputs helpful error message with example DATABASE_URL

### Task 2: Documentation ✅
- Makefile comment explains requirement
- Error message provides example connection string

## Changes

| File | Change |
|------|--------|
| File | Change |
|------|--------|
| `Makefile` | MODIFIED - Added `test-integration` with safety checks, `-race`, and `ARGS` |
| `internal/transport/http/handler/helpers_test.go` | NEW - Extracted shared test mocks and helpers |
| `internal/transport/http/handler/user_test.go` | MODIFIED - Removed extracted helpers |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `Makefile` - MODIFIED
- `internal/transport/http/handler/helpers_test.go` - NEW
- `internal/transport/http/handler/user_test.go` - MODIFIED

## Senior Developer Review (AI)

_Reviewer: BMad (AI)_

### Findings
- **Critical Issue (Build):** `make test-integration` failed because mocks were hidden behind `!integration` build tags in `user_test.go`.
- **Medium Issue (Safety):** Missing safety check to prevent running destructive tests against non-test databases.
- **Low Issue (Quality):** Missing `-race` flag and `ARGS` support.

### Actions Taken
- ✅ **Fixed Critical:** Extracted `MockCreateUserUseCase`, `MockGetUserUseCase`, `MockListUsersUseCase`, and helpers to `helpers_test.go` (shared by unit & integration tests).
- ✅ **Fixed Medium:** Added `Makefile` guard to reject `DATABASE_URL` unless it ends in `_test`.
- ✅ **Fixed Low:** Added `-race` flag and `$(ARGS)` support to `test-integration` target.
- **Status:** Story verified as **done**. Tests passing and build fixed.

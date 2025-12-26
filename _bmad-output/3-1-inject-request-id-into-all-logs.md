
# Story 3.1: Inject request_id into All Logs

**Epic**: Epic 3: Observability & Correlation
**Status**: Done
**Priority**: High

## Goal
As a Developer, I want the system to inject a unique `request_id` into all log entries automatically, so that I can trace a single request's journey through all system components and layers.

## Acceptance Criteria
1. **Given** any request to the API, **When** logs are written, **Then** all log entries include `request_id` field.
2. **Given** a request with propagated context, **When** any downstream component logs, **Then** the log includes the same `request_id` from the original request.
3. **Given** the implementation, **When** unit tests are run, **Then** `request_id` injection is verified.

## Implementation Tasks (Updated)
- [x] Create `LoggerFromContext` helper in `internal/infra/observability/logger.go`
- [x] Implement `LoggerFromContext` logic to extract `request_id` from context
- [x] Refactor `audit_event_repo.go` to use `LoggerFromContext`
- [x] Refactor `get_user.go` (Use Case) to use `LoggerFromContext`
- [x] Refactor `auth.go` (Middleware) to use `LoggerFromContext`
- [x] Create `internal/transport/http/ctxutil` to handle RequestID context keys (Break import cycle)
- [x] Update `middleware` and `handler` packages to use `ctxutil`
- [x] Verify all logs contain `request_id`

## Code Review Findings & Fixes (2024-12-24)

### Re-Review (2024-12-24)
**Status**: PASSED ✅
**Summary**: The re-review confirmed that critical issues identified in the initial review have been fully resolved.
- **AC Compatibility**: Verified that `audit_event_repo.go`, `get_user.go`, and `auth.go` now correctly inject `request_id` using the `LoggerFromContext` helper.
- **Cycle Resolution**: Confirmed that `ctxutil` package successfully resolves the import cycle.
- **Test Coverage**: All unit tests pass, including specific context propagation tests in `logger_test.go`.

### Initial Review Findings
**Issue**: Several key components were logging without `request_id` injection:
- `internal/infra/postgres/audit_event_repo.go`: Used global `slog.Warn`.
- `internal/app/user/get_user.go`: Used a static `uc.logger` initialized at startup.
- `internal/transport/http/middleware/auth.go`: Used a static logger.

**Fix**: Refactored all call sites to use `observability.LoggerFromContext(ctx, logger)`, ensuring dynamic injection of `request_id` from the current context.

**Issue**: Using `middleware.GetRequestID` inside `observability/logger.go` caused a cyclic dependency.
**Fix**: Extracted request ID context handling into a new leaf package `internal/transport/http/ctxutil`.

## Dev Notes
- **`ctxutil` Package**: Created to hold common context keys and accessors to prevent import cycles between middleware and other infrastructure packages.
- **Observability**: `LoggerFromContext` is now the standard way to obtain a logger with request context.

## Changes
| File | Change |
| :--- | :--- |
| `internal/infra/observability/logger.go` | Added `LoggerFromContext` helper |
| `internal/transport/http/ctxutil/request_id.go` | **[NEW]** Context utilities for RequestID |
| `internal/transport/http/middleware/requestid.go` | Refactored to use `ctxutil` |
| `internal/transport/http/middleware/logging.go` | Refactored to use `ctxutil` |
| `internal/infra/postgres/audit_event_repo.go` | **Fixed**: Now uses `LoggerFromContext` |
| `internal/app/user/get_user.go` | **Fixed**: Now uses `LoggerFromContext` |
| `internal/transport/http/middleware/auth.go` | **Fixed**: Now uses `LoggerFromContext` |

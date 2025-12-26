# Story 3.3: Add request_id to RFC7807 Errors

Status: done

## Story

**As a** developer,
**I want** RFC7807 errors to include request_id,
**So that** I can correlate client errors with server logs.

**FR:** FR16

## Acceptance Criteria

1. ✅ **Given** any API error response, **When** the error is formatted, **Then** the response includes `request_id` extension field
2. ✅ **Given** the implementation, **When** unit tests are run, **Then** RFC7807 + request_id injection is verified

## Implementation Summary

### Task 1: Extend ProblemDetail struct ✅
- Added `RequestID string` field with `json:"requestId,omitempty"` tag

### Task 2: Update WriteProblemJSON ✅
- Extract request_id using `ctxutil.GetRequestID(r.Context())`
- Set in `problem.RequestID` before writing response

### Task 3: Update NewValidationProblem ✅
- Added request_id extraction to validation errors too

### Task 4: Unit Tests ✅
- `TestWriteProblemJSON_IncludesRequestID` - verifies request_id present
- `TestWriteProblemJSON_NoRequestID_GracefulDegradation` - verifies graceful handling
- `TestNewValidationProblem_IncludesRequestID` - verifies validation errors

## Dev Notes

All contract tests pass (15 tests).

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/contract/error.go` | Added `RequestID` field, updated functions |
| `internal/transport/http/contract/error_test.go` | Added 3 tests for request_id |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/contract/error.go` - MODIFIED
- `internal/transport/http/contract/error_test.go` - MODIFIED

## Senior Developer Review (AI)

_Reviewer: Code Review Agent on 2025-12-25_

### Findings
- **Medium**: Fallback 500 JSON construction was missing `request_id`.
- **Medium**: Inconsistent naming `requestId` (camelCase) vs `request_id` (snake_case/AC).
- **Low**: Swallowed write errors in `WriteProblemJSON`.
- **Low**: Test coverage gap for `TestWriteProblemJSON` regarding request ID.

### Status
- **Resolution**: All findings fixed automatically.
    - Renamed JSON field to `request_id`.
    - Added `request_id` to fallback 500 JSON string.
    - Secured fallback JSON construction against injection.
    - Updated tests to verify `request_id` (closed coverage gap).
    - Added logging for write errors (handled "swallowed error").
- **Outcome**: Approved.


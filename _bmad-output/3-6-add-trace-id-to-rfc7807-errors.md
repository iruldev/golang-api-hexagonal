# Story 3.6: Add trace_id to RFC7807 Errors

Status: done

## Story

**As a** developer,
**I want** RFC7807 errors to include trace_id,
**So that** I can find distributed traces from error responses.

**FR:** FR45

## Acceptance Criteria

1. ✅ **Given** tracing is enabled, **When** an API error is returned, **Then** the response includes `trace_id` extension
2. ✅ **Given** the implementation, **When** both request_id and trace_id are present, **Then** both fields are included
3. ✅ **Given** tracing is disabled, **When** an API error is returned, **Then** `trace_id` is omitted (graceful degradation)
4. ✅ **Given** the implementation, **When** unit tests are run, **Then** trace_id injection is verified

## Implementation Summary

### Task 1: Extend ProblemDetail struct ✅
- Added `TraceID string` with `json:"trace_id,omitempty"`

### Task 2: Update WriteProblemJSON ✅
- Extract trace_id using `ctxutil.GetTraceID(r.Context())`
- Filter zero trace ID (`EmptyTraceID`)

### Task 3: Update NewValidationProblem ✅
- Added trace_id extraction with zero-ID filtering

### Task 4: Unit Tests ✅
- `TestWriteProblemJSON_IncludesTraceID` - verifies trace_id present
- `TestWriteProblemJSON_ZeroTraceID_Omitted` - verifies graceful degradation
- `TestWriteProblemJSON_BothRequestIDAndTraceID` - verifies both IDs present
- `TestNewValidationProblem_IncludesTraceID` - verifies validation errors

## Dev Notes

All contract tests pass (19 tests).

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/contract/error.go` | Added `TraceID` field, updated functions |
| `internal/transport/http/contract/error_test.go` | Added 4 tests for trace_id |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/contract/error.go` - MODIFIED
- `internal/transport/http/contract/error_test.go` - MODIFIED

### Senior Developer Review (AI)

**Reviewer:** Antigravity on 2025-12-25

**Outcome:** Approved with Fixes

**Findings:**
1.  **High:** `trace_id` was missing from the "fallback" 500 error response (used when JSON marshalling fails). While rare, this is when observability is needed most.
2.  **Low:** Duplicate code for extracting request/trace IDs in `WriteProblemJSON` and `NewValidationProblem`.

**Fixes Applied:**
-   Refactored ID injection into a shared `populateIDs` helper function.
-   Updated `writeProblemJSON` fallback logic to manually append `trace_id` to the emergency JSON payload if present.
-   Verified all tests pass.

### Senior Developer Re-Review (AI)

**Reviewer:** Antigravity on 2025-12-25 (Run 2)

**Outcome:** Approved with Refinements

**Findings:**
1.  **Low:** Magic string `"application/problem+json"` duplicated across source and tests.

**Applies Fixes:**
-   Introduced `ContentTypeProblemJSON` constant.
-   Updated code and tests to use constant.



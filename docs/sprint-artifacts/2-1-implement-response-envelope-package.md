# Story 2.1: Implement Response Envelope Package

Status: Done

## Story

As a developer,
I want a standard Envelope{data, error, meta} response package,
So that all API responses are consistent.

## Acceptance Criteria

1. **Given** any HTTP handler returning a response
   **When** I use the response package
   **Then** the JSON output follows `{data, error, meta}` structure
   **And** `meta.trace_id` is automatically populated from context

2. **Given** a successful response
   **When** the Envelope is rendered
   **Then** `data` contains the payload
   **And** `error` is null/omitted
   **And** `meta.trace_id` is present

3. **Given** an error response
   **When** the Envelope is rendered
   **Then** `data` is null/omitted
   **And** `error.code` uses UPPER_SNAKE format (e.g., `NOT_FOUND`)
   **And** `error.message` provides human-readable description
   **And** `meta.trace_id` is present

4. **Given** pagination is needed
   **When** a list response is returned
   **Then** `meta` includes pagination fields (page, page_size, total)
   **And** this is optional enhancement for list endpoints

## Tasks / Subtasks

- [x] Task 1: Refactor Envelope structure (AC: #1, #2, #3)
  - [x] 1.1 Create new `Envelope` struct with `Data`, `Error`, `Meta` fields
  - [x] 1.2 Create `Meta` struct with `trace_id` (mandatory), optional pagination fields
  - [x] 1.3 Create `ErrorBody` struct with `Code`, `Message`, optional `Hint`
  - [x] 1.4 Update JSON tags to snake_case: `trace_id`, `page_size`

- [x] Task 2: Implement trace_id population from context (AC: #1)
  - [x] 2.1 Create helper function to extract trace_id from context
  - [x] 2.2 Use existing `ctxutil.RequestIDFromContext(ctx)`
  - [x] 2.3 Fallback to "unknown" if no trace_id in context

- [x] Task 3: Update helper functions (AC: #2, #3)
  - [x] 3.1 Refactor `SuccessEnvelope(w, ctx, data)` - with ctx parameter for trace_id
  - [x] 3.2 Refactor `ErrorEnvelope(w, ctx, status, code, message)` - with ctx parameter
  - [x] 3.3 Update convenience functions: `BadRequestCtx`, `NotFoundCtx`, `UnauthorizedCtx`, etc.
  - [x] 3.4 Deprecate old `SuccessResponse`/`ErrorResponse` (kept for backward compatibility)

- [x] Task 4: Update existing usages (AC: #1, #2, #3)
  - [x] 4.1 Updated HealthHandler to use `SuccessEnvelope(w, ctx, data)`
  - [x] 4.2 Updated ReadyzHandler to use `ServiceUnavailableCtx` and `SuccessEnvelope`
  - [x] 4.3 Backward compatibility maintained with deprecated old functions

- [x] Task 5: Add comprehensive tests (AC: All)
  - [x] 5.1 Test Envelope structure serialization
  - [x] 5.2 Test trace_id extraction from context
  - [x] 5.3 Test error code format (UPPER_SNAKE)
  - [x] 5.4 Test meta fields are always present

- [x] Task 6: Documentation and verification
  - [x] 6.1 Package documentation included in envelope.go
  - [x] 6.2 `make lint` passes (0 issues), tests pass
  - [x] 6.3 Existing handlers verified - HealthHandler uses new Envelope format

## Dev Notes

### Existing Implementation Analysis

**Current State (`internal/interface/http/response/`):**
- `SuccessResponse`: `{success: true, data: {...}}`
- `ErrorResponse`: `{success: false, error: {code, message}}`
- Missing: `meta` field, `trace_id`, context parameter

**Target State (from architecture-decisions.md):**
```go
type Envelope struct {
    Data  any            `json:"data,omitempty"`
    Error *ErrorBody     `json:"error,omitempty"`
    Meta  *Meta          `json:"meta,omitempty"`
}
// meta.trace_id is MANDATORY
```

**Gap Analysis:**
| Aspect | Current | Required |
|--------|---------|----------|
| Structure | `{success, data}` / `{success, error}` | `{data, error, meta}` |
| trace_id | Not present | MANDATORY in `meta` |
| Context | Not passed to helpers | Required for trace_id |
| Error codes | `ERR_NOT_FOUND` | `NOT_FOUND` (no ERR_ prefix) |

### Architecture Patterns (From architecture-decisions.md)

**Decision 1: Error Code Registry (Hybrid)**
- Central registry: `internal/domain/errors/codes.go` for public/stable codes
- Public codes: UPPER_SNAKE without prefix (e.g., `NOT_FOUND`, not `ERR_NOT_FOUND`)

**API Response Pattern:**
```go
// Success response
{
  "data": {...},
  "meta": {
    "trace_id": "abc123"
  }
}

// Error response
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Note not found"
  },
  "meta": {
    "trace_id": "abc123"
  }
}
```

### File Structure

```
internal/interface/http/
├── response/
│   ├── response.go       # [REFACTOR] Envelope, Success, Error functions
│   ├── errors.go         # [UPDATE] Error codes without ERR_ prefix
│   ├── mapper.go         # [UPDATE] MapError with context
│   ├── envelope.go       # [NEW] Envelope, Meta, ErrorBody structs
│   ├── response_test.go  # [ENHANCE] Add Envelope tests
│   └── mapper_test.go    # [ENHANCE] Update MapError tests
├── middleware/
│   └── request_id.go     # [REFERENCE] GetRequestID() helper
└── note/
    └── handler.go        # [UPDATE] Pass context to response calls
```

### Critical Context from Middleware

**Existing trace_id source:**
- `internal/interface/http/middleware/` has request ID middleware
- Use existing `GetRequestID(ctx)` or `GetTraceFromContext()`
- If using OpenTelemetry, trace_id may come from `trace.SpanFromContext(ctx)`

### Testing Strategy

1. **Unit Tests:**
   - Test Envelope JSON serialization
   - Test `meta.trace_id` is always populated
   - Test error codes match UPPER_SNAKE format

2. **Integration Verification:**
   - Run existing handler tests to ensure no regression
   - Verify API responses match new format

### NFR Targets

| NFR | Requirement | Verification |
|-----|-------------|--------------|
| FR14 | Response uses Envelope{data, error, meta} | JSON structure test |
| FR15 | meta.trace_id mandatory in all responses | Unit test |
| FR16 | error.code uses public UPPER_SNAKE codes | Format test |

### Previous Story Learnings (from Story 1.6)

- CI validates lint + test, so all changes must pass `make verify`
- Use existing patterns for consistency
- Comprehensive test coverage prevents regression
- Update all affected files in single story to avoid partial migration

### Dependencies

- Story 2.2 (Central Error Code Registry) will build on this by adding typed domain errors
- This story provides foundation for consistent API responses Epic 2

### Critical Points

1. **Context is required:** All response helper functions must accept `context.Context`
2. **trace_id is MANDATORY:** Never return response without `meta.trace_id`
3. **Error code format:** Use `NOT_FOUND`, not `ERR_NOT_FOUND`
4. **Backward compatible:** Consider deprecation period for old functions
5. **JSON snake_case:** All JSON fields must use snake_case

### References

- [Source: docs/epics.md#Story 2.1](file:///docs/epics.md) - FR14, FR15, FR16
- [Source: docs/architecture-decisions.md](file:///docs/architecture-decisions.md) - Decision 1: Error Code Registry, API Patterns
- [Source: project_context.md](file:///project_context.md) - Envelope format, UPPER_SNAKE codes
- [Source: internal/interface/http/response/response.go](file:///internal/interface/http/response/response.go) - Current implementation
- [Source: internal/interface/http/response/errors.go](file:///internal/interface/http/response/errors.go) - Current error codes
- [Source: internal/interface/http/middleware/](file:///internal/interface/http/middleware/) - Request ID context

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 2: API Standards & Response Contract (MVP) - in-progress
- Previous stories: Epic 1 complete (1.1-1.6 done)

### Agent Model Used

gemini-2.5-pro

### Debug Log References

None required.

### Completion Notes List

- Created `envelope.go` with new `Envelope`, `Meta`, `ErrorBody` structs following `{data, error, meta}` format
- Implemented `getTraceID(ctx)` using `ctxutil.RequestIDFromContext()` with "unknown" fallback
- Added context-aware functions: `SuccessEnvelope`, `ErrorEnvelope`, `SuccessEnvelopeWithPagination`
- Added convenience functions: `BadRequestCtx`, `NotFoundCtx`, `UnauthorizedCtx`, `ForbiddenCtx`, `ConflictCtx`, `ValidationErrorCtx`, `InternalServerErrorCtx`, `ServiceUnavailableCtx`, `TimeoutCtx`
- Updated `errors.go` with new UPPER_SNAKE codes (`CodeBadRequest`, `CodeNotFound`, etc.) without ERR_ prefix
- Maintained backward compatibility with deprecated `ErrBadRequest`, `ErrNotFound`, etc.
- Updated `health.go` to use new Envelope functions with context
- Created comprehensive `envelope_test.go` with 8 test functions covering structure serialization, trace_id extraction, error codes, and all convenience functions
- Updated `handlers_test.go` to validate new Envelope format with `data`/`meta` instead of `success`/`data`
- All tests pass, `make lint` shows 0 issues

### File List

New files:
- internal/interface/http/response/envelope.go
- internal/interface/http/response/envelope_test.go

Modified files:
- internal/interface/http/response/errors.go
- internal/interface/http/handlers/health.go
- internal/interface/http/handlers/handlers_test.go
- docs/sprint-artifacts/sprint-status.yaml
- docs/sprint-artifacts/2-1-implement-response-envelope-package.md

## Change Log

- 2025-12-15: Implemented Response Envelope Package (Story 2.1)
  - Created Envelope{data, error, meta} structure with mandatory trace_id
  - Added UPPER_SNAKE error codes (CodeBadRequest, CodeNotFound, etc.)
  - Updated HealthHandler to use new Envelope format
  - Added comprehensive tests for all functionality
- 2025-12-15: Code Review Fixes (AI)
  - Deprecated old `SuccessResponse` functions in `response.go`
  - Updated `ExampleHandler` to use new `SuccessEnvelope` format
  - Updated `ExampleHandler` tests to verify `Meta` field presence and correctness
- 2025-12-15: Code Review Fixes Round 2 (AI)
  - Replaced `log.Printf` with `slog.Error` for better observability
  - Added security warning to `ErrorBody.Hint` field
  - Defined `UnknownTraceID` constant to replace magic string

## Status

Done

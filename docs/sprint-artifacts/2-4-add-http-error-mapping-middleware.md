Status: done

## Story

As a developer,
I want domain errors automatically mapped to HTTP status codes,
So that I don't repeat error handling logic.

## Acceptance Criteria

1. **Given** a handler returns `errors.ErrNotFound`
   **When** the response middleware processes it
   **Then** HTTP 404 is returned with Envelope error format
   **And** error.code is "NOT_FOUND"

2. **Given** a handler panics with any error
   **When** the middleware catches the panic
   **Then** HTTP 500 is returned with Envelope error format
   **And** error.code is "INTERNAL_ERROR"
   **And** the panic is logged with trace_id

3. **Given** a handler returns a DomainError with custom code
   **When** the error middleware processes it
   **Then** the HTTP status is derived from the error code
   **And** the Envelope includes code, message, and hint (if present)

4. **Given** context propagation throughout the middleware chain
   **When** any error occurs
   **Then** meta.trace_id is included in the Envelope response
   **And** the error is logged with correlation to trace_id

5. **Given** the middleware is applied to routes
   **When** handlers use `return err` pattern with a wrapper
   **Then** errors are automatically mapped without explicit `HandleErrorCtx` calls
   **And** layer boundaries are respected (middleware in interface layer)

## Tasks / Subtasks

- [x] Task 1: Create error handler middleware (AC: #1, #2, #3, #4)
  - [x] 1.1 Create `internal/interface/http/middleware/error_handler.go`
  - [x] 1.2 Implement panic recovery with structured logging
  - [x] 1.3 Use `response.ErrorEnvelope` for consistent format
  - [x] 1.4 Extract trace_id from context using `ctxutil.RequestIDFromContext`
  - [x] 1.5 Log errors with trace_id correlation via slog

- [x] Task 2: Create handler wrapper for error returns (AC: #5)
  - [x] 2.1 Create `internal/interface/http/handler_func.go` with `HandlerFuncE` type
  - [x] 2.2 Implement `WrapHandler` that converts `func(w, r) error` to `http.HandlerFunc`
  - [x] 2.3 Apply error mapping via existing `response.HandleErrorCtx`
  - [x] 2.4 Ensure compatibility with chi router middleware chain

- [x] Task 3: Add comprehensive tests (AC: All)
  - [x] 3.1 Create `error_handler_test.go` with panic recovery tests
  - [x] 3.2 Test DomainError mapping to correct HTTP status
  - [x] 3.3 Test legacy sentinel error fallback
  - [x] 3.4 Test trace_id inclusion in error responses
  - [x] 3.5 Create `handler_func_test.go` for wrapper tests

- [x] Task 4: Integration with existing handlers (AC: #5)
  - [x] 4.1 Update `ExampleHandler` to demonstrate error return pattern (documented in handler_func.go)
  - [x] 4.2 Document usage patterns in handler file
  - [x] 4.3 Verify no breaking changes to existing handlers

- [x] Task 5: Validation and documentation
  - [x] 5.1 Run `make lint` to verify layer boundaries
  - [x] 5.2 Run `make verify` for full test suite
  - [x] 5.3 Update documentation if needed

## Dev Notes

### Architecture Decision Reference

**Decision 1: Error Code Registry (Hybrid)** from `docs/architecture-decisions.md`:
- Central registry at `internal/domain/errors/codes.go`
- DomainError type with Code, Message, Hint fields
- Response mapper already exists at `internal/interface/http/response/mapper.go`

**Existing Implementation to Leverage:**

```go
// internal/interface/http/response/mapper.go (already exists)
func MapError(err error) (status int, code string)
func HandleErrorCtx(w http.ResponseWriter, ctx context.Context, err error)

// These functions already handle:
// - DomainError detection and code extraction
// - Legacy sentinel error fallback
// - Envelope format with trace_id
```

### Target Implementation Pattern

**Error Handler Middleware (panic recovery):**
```go
// internal/interface/http/middleware/error_handler.go
package middleware

import (
    "log/slog"
    "net/http"
    "runtime/debug"
    
    "github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// ErrorHandler is a middleware that recovers from panics and logs errors.
// It wraps panic errors in a consistent Envelope response format.
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                traceID := ctxutil.RequestIDFromContext(r.Context())
                slog.Error("panic recovered",
                    "trace_id", traceID,
                    "panic", rec,
                    "stack", string(debug.Stack()),
                )
                response.InternalServerErrorCtx(w, r.Context(), "internal server error")
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**Handler Wrapper for Error Returns:**
```go
// internal/interface/http/handler_func.go
package http

import (
    "net/http"
    
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HandlerFuncE is a handler function that can return an error.
// This enables cleaner handler code without explicit error handling.
type HandlerFuncE func(w http.ResponseWriter, r *http.Request) error

// WrapHandler converts a HandlerFuncE to http.HandlerFunc.
// Errors are automatically mapped to Envelope responses.
func WrapHandler(h HandlerFuncE) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := h(w, r); err != nil {
            response.HandleErrorCtx(w, r.Context(), err)
        }
    }
}
```

### Existing Code Analysis

**`internal/interface/http/response/mapper.go` (from Story 2.2):**
- Already has `MapError`, `MapErrorWithHint`, `HandleErrorCtx`
- Handles DomainError detection and code extraction
- Falls back to legacy sentinel errors
- Uses Envelope format with trace_id

**`internal/interface/http/response/envelope.go` (from Story 2.1):**
- `ErrorEnvelope` and `ErrorEnvelopeWithHint` for error responses
- `InternalServerErrorCtx` for 500 responses
- All include `meta.trace_id`

**`internal/interface/http/middleware/` (existing):**
- `request_id.go` - injects request ID into context
- `logger.go` - structured logging middleware
- Pattern to follow for new middleware

### Key Implementation Points

1. **No Duplicating Logic:**
   - Use existing `response.HandleErrorCtx` for error mapping
   - Middleware focuses on panic recovery and handler wrapping

2. **Panic Recovery Pattern:**
   ```go
   defer func() {
       if rec := recover(); rec != nil {
           // log and respond with 500
       }
   }()
   ```

3. **Handler Wrapper Pattern:**
   - `HandlerFuncE` type enables `func(w, r) error` signature
   - `WrapHandler` converts to standard `http.HandlerFunc`
   - Errors automatically mapped via `HandleErrorCtx`

4. **Logging Requirements:**
   - Use `slog` for structured logging (per project standards)
   - Include `trace_id` in all error logs
   - Log stack trace for panics only

### File Structure

```
internal/interface/http/
├── middleware/
│   ├── error_handler.go        # [NEW] Panic recovery middleware
│   └── error_handler_test.go   # [NEW] Middleware tests
├── handler_func.go             # [NEW] HandlerFuncE type and WrapHandler
├── handler_func_test.go        # [NEW] Wrapper tests
└── response/
    ├── mapper.go               # [EXISTS] Already has HandleErrorCtx
    └── envelope.go             # [EXISTS] Already has error helpers
```

### Critical Points from Previous Stories

From Story 2.1 & 2.2 learnings:
- Use `ctxutil.RequestIDFromContext(ctx)` for trace correlation
- UPPER_SNAKE format for error codes without `ERR_` prefix
- All error responses must include `meta.trace_id`
- Test coverage is critical - CI enforces lint+test
- Hint field must not expose internal error details

From Story 2.3 learnings:
- Context propagation is critical for all I/O
- Default timeout applied when no deadline set
- Early return on cancelled context

### Error Code to HTTP Status Mapping

Per `mapper.go` (already implemented):

| Code | HTTP Status |
|------|-------------|
| `NOT_FOUND` | 404 |
| `VALIDATION_ERROR` | 422 |
| `UNAUTHORIZED` | 401 |
| `FORBIDDEN` | 403 |
| `CONFLICT` | 409 |
| `INTERNAL_ERROR` | 500 |
| `TIMEOUT` | 504 |
| `RATE_LIMIT_EXCEEDED` | 429 |
| `BAD_REQUEST` | 400 |

### Testing Strategy

1. **Panic Recovery Tests:**
   - Handler panics with string
   - Handler panics with error
   - Response is 500 with Envelope format
   - Panic is logged with trace_id

2. **HandlerFuncE Tests:**
   - Handler returns nil (success path)
   - Handler returns DomainError
   - Handler returns legacy sentinel error
   - Correct HTTP status and code mapping

3. **Integration Tests:**
   - Full middleware chain with error handler
   - Request ID propagation to error response
   - Correct JSON structure validation

### NFR Targets

| NFR | Requirement | Verification |
|-----|-------------|--------------|
| FR14 | Response uses Envelope{data, error, meta} | Envelope format test |
| FR15 | meta.trace_id mandatory | Trace ID verification |
| FR16 | error.code uses UPPER_SNAKE | Code format test |
| NFR-M1 | Coverage ≥80% for middleware | Unit tests |

### Dependencies

- **Story 2.1 (Done):** Provides Envelope format, trace_id extraction
- **Story 2.2 (Done):** Provides DomainError, central codes, mapper.go
- **Story 2.3 (Done):** Provides context wrapper patterns

### Critical Points

1. **No breaking changes:** Existing handlers continue to work unchanged
2. **Optional adoption:** `WrapHandler` is opt-in, not mandatory
3. **Layer boundaries:** Middleware stays in interface/http layer
4. **Leverage existing code:** Use mapper.go, don't duplicate logic
5. **Comprehensive testing:** All error paths tested

### References

- [Source: docs/epics.md#Story 2.4](file:///docs/epics.md) - FR14, FR15, FR16
- [Source: docs/architecture-decisions.md](file:///docs/architecture-decisions.md) - Error handling patterns
- [Source: project_context.md](file:///project_context.md) - Layer boundaries, error handling
- [Source: internal/interface/http/response/mapper.go](file:///internal/interface/http/response/mapper.go) - Existing error mapping
- [Source: internal/interface/http/response/envelope.go](file:///internal/interface/http/response/envelope.go) - Envelope format
- [Source: internal/domain/errors/codes.go](file:///internal/domain/errors/codes.go) - Central error codes
- [Source: internal/domain/errors/domain_error.go](file:///internal/domain/errors/domain_error.go) - DomainError type
- [Source: internal/interface/http/middleware/](file:///internal/interface/http/middleware/) - Existing middleware patterns
- [Source: docs/sprint-artifacts/2-2-create-central-error-code-registry.md](file:///docs/sprint-artifacts/2-2-create-central-error-code-registry.md) - Previous story learnings
- [Source: docs/sprint-artifacts/2-3-implement-context-wrapper-package.md](file:///docs/sprint-artifacts/2-3-implement-context-wrapper-package.md) - Previous story learnings

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 2: API Standards & Response Contract (MVP) - in-progress
- Previous story: Story 2.3 (Implement Context Wrapper Package) - done

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

None required.

### Completion Notes List

- Implemented `ErrorHandler` middleware with panic recovery using `slog` structured logging
- Uses `response.InternalServerErrorCtx` for Envelope format with trace_id
- Created `HandlerFuncE` type and `WrapHandler` for error-returning handlers
- `WrapHandler` leverages existing `response.HandleErrorCtx` for error mapping
- All tests pass including panic recovery, trace_id inclusion, and error code mapping
- `make verify` passes: lint and all tests successful
- No breaking changes to existing handlers - `WrapHandler` is opt-in
- **Review Fixes:**
  - Updated `ExampleHandler` to demonstrate `HandlerFuncE` usage
  - Updated `routes.go` to use `WrapHandler` for the example
  - Added documentation for new patterns in `project_context.md`
  - Added dependency clarification to `error_handler.go`

### File List

**New Files:**
- `internal/interface/http/middleware/error_handler.go`
- `internal/interface/http/middleware/error_handler_test.go`
- `internal/interface/http/handler_func.go`
- `internal/interface/http/handler_func_test.go`

**Modified Files:**
- `docs/sprint-artifacts/sprint-status.yaml`
- `internal/interface/http/handlers/example.go`
- `internal/interface/http/routes.go`
- `project_context.md`

### Change Log

- 2025-12-15: Implemented Story 2.4 - HTTP error mapping middleware and handler wrapper
- 2025-12-15: Applied code review fixes (documented usage, updated example handler)

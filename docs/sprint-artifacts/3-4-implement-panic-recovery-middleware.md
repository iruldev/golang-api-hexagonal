# Story 3.4: Implement Panic Recovery Middleware

Status: done

## Story

As a SRE,
I want the system to recover from panics,
So that a single panic doesn't crash the server.

## Acceptance Criteria

### AC1: Return 500 with generic error
**Given** a handler panics
**When** the panic occurs
**Then** response status is 500
**And** response body is generic error (no stack trace)

### AC2: Log panic with stack trace
**Given** a handler panics
**When** the panic occurs
**Then** panic is logged with stack trace

### AC3: Server continues
**Given** a handler panics
**When** recovery completes
**Then** server continues handling other requests

---

## Tasks / Subtasks

- [x] **Task 1: Create recovery middleware** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/middleware/recovery.go` ✅
  - [x] Implement `Recovery(logger *zap.Logger) func(http.Handler) http.Handler` ✅
  - [x] Use defer/recover pattern to catch panics ✅
  - [x] Log panic with zap.Stack() for stack trace ✅
  - [x] Return 500 with generic JSON error body ✅

- [x] **Task 2: Wire middleware into router** (AC: #1, #2, #3)
  - [x] Update `router.go` to add recovery middleware ✅
  - [x] Add `r.Use(middleware.Recovery(logger))` as FIRST middleware ✅
  - [x] Recovery must be first to catch panics from all other middleware ✅

- [x] **Task 3: Create middleware tests** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/middleware/recovery_test.go` ✅
  - [x] Test: panicking handler returns 500 ✅
  - [x] Test: response body is generic error JSON ✅
  - [x] Test: no stack trace in response body ✅
  - [x] Test: panic is logged with stack ✅
  - [x] Test: subsequent requests still work ✅

- [x] **Task 4: Verify implementation** (AC: #1, #2, #3)
  - [x] Run `make test` - all pass ✅ (96.6% middleware coverage)
  - [x] Run `make lint` - 0 issues ✅
  - [x] Fixed errcheck on w.Write in test ✅

---

## Dev Notes

### Recovery Middleware Pattern

```go
// internal/interface/http/middleware/recovery.go
package middleware

import (
    "encoding/json"
    "net/http"

    "go.uber.org/zap"
)

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
    Error string `json:"error"`
}

// Recovery middleware recovers from panics and returns 500.
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    // Log panic with stack trace
                    logger.Error("panic recovered",
                        zap.Any("error", err),
                        zap.String("request_id", GetRequestID(r.Context())),
                        zap.Stack("stacktrace"),
                    )

                    // Return generic 500 error
                    w.Header().Set("Content-Type", "application/json")
                    w.WriteHeader(http.StatusInternalServerError)
                    json.NewEncoder(w).Encode(ErrorResponse{
                        Error: "internal server error",
                    })
                }
            }()

            next.ServeHTTP(w, r)
        })
    }
}
```

### Router Integration

```go
// internal/interface/http/router.go
func NewRouter(cfg *config.Config) chi.Router {
    logger, err := observability.NewLogger(&cfg.Log, cfg.App.Env)
    if err != nil {
        log.Printf("Failed to initialize logger, using nop: %v", err)
        logger = observability.NewNopLogger()
    }

    r := chi.NewRouter()

    // Global middleware (order matters!)
    r.Use(middleware.Recovery(logger))    // Story 3.4 - FIRST to catch all panics
    r.Use(middleware.RequestID)           // Story 3.2
    r.Use(middleware.Logging(logger))     // Story 3.3

    // API v1 routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/health", handlers.HealthHandler)
    })

    return r
}
```

### Testing Strategy

```go
// internal/interface/http/middleware/recovery_test.go
func TestRecovery_Returns500(t *testing.T) {
    logger := zap.NewNop()

    handler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        panic("test panic")
    }))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRecovery_GenericError(t *testing.T) {
    logger := zap.NewNop()

    handler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        panic("secret error message")
    }))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    body := rec.Body.String()
    assert.Contains(t, body, "internal server error")
    assert.NotContains(t, body, "secret")
    assert.NotContains(t, body, "stacktrace")
}
```

### Architecture Compliance

**Layer:** `internal/interface/http/middleware`
**Pattern:** Recovery middleware using defer/recover
**Logging:** Uses zap.Stack() for stack trace capture

### Previous Story Learnings

From **Story 3.3** code review:
- ✅ Fixed field naming for semantic accuracy
- ✅ Comprehensive test coverage
- ✅ Proper middleware ordering

### Dependencies

**Existing:** zap (from Story 3.3)

### Middleware Order

1. **Recovery** (catches panics from all below)
2. RequestID (adds request ID)
3. Logging (logs with request ID)

This order ensures:
- Panics are caught and logged with request ID
- All middleware panics are recovered

### References

- [Source: docs/epics.md#Story-3.4]
- [Story 3.3 - Logging Middleware](file:///docs/sprint-artifacts/3-3-add-http-request-logging-middleware.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fourth story in Epic 3: HTTP API Core.
Adds crash protection to HTTP server.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Implementation completed: 2025-12-11
  - Created recovery.go (44 lines) with defer/recover and ErrorResponse
  - Created recovery_test.go (132 lines) with 6 comprehensive tests
  - Wired middleware as FIRST in router.go middleware chain
  - Fixed errcheck lint issue on w.Write in test
  - Coverage: 96.6% (middleware package), Lint: 0 issues

### File List

Files created:
- `internal/interface/http/middleware/recovery.go` - Recovery middleware (44 lines)
- `internal/interface/http/middleware/recovery_test.go` - Middleware tests (132 lines)

Files modified:
- `internal/interface/http/router.go` - Wired recovery middleware first (42 lines)
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

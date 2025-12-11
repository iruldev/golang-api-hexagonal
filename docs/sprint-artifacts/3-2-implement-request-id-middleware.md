# Story 3.2: Implement Request ID Middleware

Status: Ready for Review

## Story

As a SRE,
I want unique request IDs for each request,
So that I can trace requests across systems.

## Acceptance Criteria

### AC1: Generate UUID when no X-Request-ID header
**Given** a request without `X-Request-ID` header
**When** the request is processed
**Then** a UUID request ID is generated
**And** `X-Request-ID` header is set in response

### AC2: Use existing ID when X-Request-ID header provided
**Given** a request with `X-Request-ID` header
**When** the request is processed
**Then** the provided ID is used
**And** same ID is returned in response

---

## Tasks / Subtasks

- [x] **Task 1: Add UUID dependency** (AC: #1)
  - [x] Run `go get github.com/google/uuid` ✅ v1.6.0
  - [x] Verify uuid is added to go.mod ✅

- [x] **Task 2: Create request ID middleware** (AC: #1, #2)
  - [x] Create `internal/interface/http/middleware/requestid.go` ✅
  - [x] Implement `RequestID` middleware function ✅
  - [x] Check for existing `X-Request-ID` header ✅
  - [x] Generate UUID if header not present ✅
  - [x] Store request ID in context for handlers ✅
  - [x] Set `X-Request-ID` header in response ✅

- [x] **Task 3: Wire middleware into router** (AC: #1, #2)
  - [x] Update `router.go` to use middleware ✅
  - [x] Add `r.Use(middleware.RequestID)` before routes ✅
  - [x] Update TODO comment for remaining stories ✅

- [x] **Task 4: Create context helper** (AC: #1, #2)
  - [x] Create `GetRequestID(ctx context.Context) string` ✅
  - [x] Returns empty string if not in context ✅

- [x] **Task 5: Create middleware tests** (AC: #1, #2)
  - [x] Create `internal/interface/http/middleware/requestid_test.go` ✅
  - [x] Test: no header → UUID generated and returned ✅
  - [x] Test: header present → same ID returned ✅
  - [x] Test: context contains request ID ✅
  - [x] Test: unique per request ✅
  - [x] Test: empty without middleware ✅

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass ✅ (100% coverage)
  - [x] Run `make lint` - 0 issues ✅

---

## Dev Notes

### Request ID Middleware Pattern

```go
// internal/interface/http/middleware/requestid.go
package middleware

import (
    "context"
    "net/http"

    "github.com/google/uuid"
)

// RequestIDHeader is the header key for request ID.
const RequestIDHeader = "X-Request-ID"

// requestIDKey is the context key for request ID.
type requestIDKey struct{}

// RequestID middleware generates or uses existing request ID.
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get(RequestIDHeader)
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Set in response header
        w.Header().Set(RequestIDHeader, requestID)

        // Store in context
        ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey{}).(string); ok {
        return id
    }
    return ""
}
```

### Router Integration

```go
// internal/interface/http/router.go
func NewRouter(cfg *config.Config) chi.Router {
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.RequestID)

    // API v1 routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/health", handlers.HealthHandler)
    })

    return r
}
```

### Testing Strategy

```go
// internal/interface/http/middleware/requestid_test.go
func TestRequestID_GeneratesUUID(t *testing.T) {
    handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify context has request ID
        id := GetRequestID(r.Context())
        assert.NotEmpty(t, id)
    }))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    // Verify response header
    assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestRequestID_UsesExisting(t *testing.T) {
    existingID := "test-request-id-123"
    
    handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := GetRequestID(r.Context())
        assert.Equal(t, existingID, id)
    }))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("X-Request-ID", existingID)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, existingID, rec.Header().Get("X-Request-ID"))
}
```

### Architecture Compliance

**Layer:** `internal/interface/http/middleware` (allowed: chi, stdlib, uuid)
**Pattern:** chi middleware pattern with context propagation
**Context:** Request ID stored in context for downstream handlers

### Previous Story Learnings

From **Story 3.1** code review:
- ✅ Document future use of parameters with comments
- ✅ Handle all errors (no unhandled returns)
- ✅ Maintain 100% middleware coverage

### Dependencies

**New:** `github.com/google/uuid`
**Existing:** chi/v5 (from Story 3.1)

### File Structure After Implementation

```
internal/interface/http/
├── middleware/
│   ├── requestid.go       # Request ID middleware (NEW)
│   └── requestid_test.go  # Middleware tests (NEW)
├── router.go              # Updated to use middleware
├── router_test.go         # Updated tests
└── handlers/
    └── health.go
```

### References

- [Source: docs/epics.md#Story-3.2]
- [chi middleware documentation](https://github.com/go-chi/chi#middleware-handlers)
- [Story 3.1 - Router setup](file:///docs/sprint-artifacts/3-1-setup-chi-router-with-versioned-api.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Second story in Epic 3: HTTP API Core.
First middleware to be added to router.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Implementation completed: 2025-12-11
  - Added uuid v1.6.0 dependency
  - Created requestid.go (43 lines) with middleware and GetRequestID helper
  - Created requestid_test.go (99 lines) with 5 comprehensive tests
  - Wired middleware into router.go with r.Use(middleware.RequestID)
  - Coverage: 100% (middleware package), Lint: 0 issues

### File List

Files created:
- `internal/interface/http/middleware/requestid.go` - Request ID middleware (43 lines)
- `internal/interface/http/middleware/requestid_test.go` - Middleware tests (99 lines)

Files modified:
- `go.mod` - Added uuid dependency
- `go.sum` - Updated with uuid dependencies
- `internal/interface/http/router.go` - Wired middleware (32 lines)
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

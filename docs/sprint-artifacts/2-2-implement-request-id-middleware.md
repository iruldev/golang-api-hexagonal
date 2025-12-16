# Story 2.2: Implement Request ID Middleware

Status: done

## Story

**As a** developer,
**I want** every request to have a unique request_id,
**So that** I can trace a single request through all log entries.

## Acceptance Criteria

1. **Given** the service receives an HTTP request without `X-Request-ID` header
   **When** the request is processed
   **Then** a unique `requestId` is generated (opaque random, 16 bytes hex)
   **And** `requestId` is injected into request context
   **And** `requestId` appears in all log entries for that request
   **And** `requestId` is returned in response header `X-Request-ID`

2. **Given** the service receives an HTTP request WITH `X-Request-ID` header
   **When** the request is processed
   **Then** the provided `requestId` is used (passthrough)
   **And** the same value is returned in response header `X-Request-ID`

*Covers: FR13, FR15*

## Tasks / Subtasks

- [x] Task 1: Create Request ID middleware (AC: #1, #2)
  - [x] Create `internal/transport/http/middleware/requestid.go`
  - [x] Check for existing `X-Request-ID` header (passthrough if exists)
  - [x] Generate new request ID if not present (16 bytes hex = 32 chars)
  - [x] Use `crypto/rand` for secure random generation
  - [x] Inject request ID into request context
  - [x] Set `X-Request-ID` response header

- [x] Task 2: Create context helper functions (AC: #1)
  - [x] Create `GetRequestID(ctx context.Context) string` function
  - [x] Create context key type for request ID
  - [x] Export helper for use by logging middleware

- [x] Task 3: Update logging middleware to include requestId (AC: #1)
  - [x] Modify `logging.go` to call `GetRequestID(ctx)`
  - [x] Add `requestId` field to request completion log
  - [x] Ensure requestId appears in all request-scoped logs

- [x] Task 4: Wire middleware into router (AC: #1, #2)
  - [x] Add RequestID middleware to router BEFORE logging middleware
  - [x] Middleware order: RequestID → Logging → Recoverer → routes

- [x] Task 5: Write unit tests (AC: #1, #2)
  - [x] Test request without X-Request-ID generates new ID
  - [x] Test request with X-Request-ID uses provided value
  - [x] Test response header contains request ID
  - [x] Test GetRequestID returns correct value from context
  - [x] Test generated ID is 32 hex characters

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Middleware is in **Transport layer** (`internal/transport/http/middleware/`):
- ✅ ALLOWED: domain, chi, stdlib
- ❌ FORBIDDEN: pgx, direct infra imports (except through injection)

### Request ID Pattern [Source: docs/architecture.md, docs/project-context.md]

```go
// internal/transport/http/middleware/requestid.go
package middleware

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "net/http"
)

type contextKey string

const requestIDKey contextKey = "requestId"

const headerXRequestID = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get(headerXRequestID)
        
        if requestID == "" {
            requestID = generateRequestID()
        }
        
        // Set response header
        w.Header().Set(headerXRequestID, requestID)
        
        // Inject into context
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}

func generateRequestID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b) // 32 hex characters
}
```

### Logging Integration [Source: Story 2.1]

Update logging middleware to include requestId:

```go
// In logging.go, update the log call
logger.Info("request completed",
    "method", r.Method,
    "route", routePattern,
    "status", ww.status,
    "duration_ms", time.Since(start).Milliseconds(),
    "bytes", ww.bytes,
    "requestId", GetRequestID(r.Context()), // Add this line
)
```

### Middleware Order [Source: docs/architecture.md]

Chi middleware executes in order of registration. Required order:

```go
r := chi.NewRouter()
r.Use(middleware.RequestID)      // 1. Generate/passthrough request ID FIRST
r.Use(middleware.RequestLogger)  // 2. Logging (needs requestId in context)
r.Use(chiMiddleware.Recoverer)   // 3. Panic recovery
// ... routes
```

### Previous Story Learnings [Source: Story 2.1]

**Files created in Story 2.1:**
- `internal/infra/observability/logger.go` - Logger initialization
- `internal/transport/http/middleware/logging.go` - Request logging middleware
- `internal/transport/http/middleware/response_wrapper.go` - Response wrapper

**Key patterns established:**
- Middleware uses `http.HandlerFunc` pattern
- Response wrapper captures status/bytes
- Logger is injected into middleware
- Route pattern extracted via `chi.RouteContext(r.Context()).RoutePattern()`

**Router signature:**
```go
func NewRouter(logger *slog.Logger, ...) chi.Router
```

## Technical Requirements

- **Go version:** 1.23+ [Source: docs/project-context.md]
- **Request ID format:** 16 bytes hex encoded = 32 characters
- **Random source:** `crypto/rand` (cryptographically secure)

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Request ID must be in response header `X-Request-ID`
- Request ID must be in request context for logging
- Use passthrough pattern for existing request IDs (distributed tracing)
- 16 bytes random = 32 hex chars is standard

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-16)

### Agent Model Used

Gemini 2.5

### Debug Log References

N/A - Implementation completed without issues.

### Completion Notes List

- ✅ Task 1: Created `requestid.go` with RequestID middleware
  - Passthrough for existing X-Request-ID header
  - 16-byte hex ID generation using `crypto/rand`
  - Context injection and response header setting
- ✅ Task 2: Created `GetRequestID(ctx)` helper function with custom context key type
- ✅ Task 3: Updated `logging.go` to use custom `GetRequestID` instead of chi's `GetReqID`
- ✅ Task 4: Wired middleware into `router.go` with correct order: RequestID → Logging → RealIP → Recoverer
- ✅ Task 5: Created 9 comprehensive unit tests covering:
  - New ID generation, passthrough, response headers
  - Context helper edge cases (empty, wrong type)
  - Hex format verification, uniqueness
  - Multiple request handling

All tests pass (23 middleware tests total). Build succeeds.

### File List

- `internal/transport/http/middleware/requestid.go` [NEW]
- `internal/transport/http/middleware/requestid_test.go` [NEW]
- `internal/transport/http/middleware/logging.go` [MODIFIED]
- `internal/transport/http/middleware/logging_test.go` [MODIFIED]
- `internal/transport/http/router.go` [MODIFIED]
- `docs/sprint-artifacts/sprint-status.yaml` [MODIFIED]

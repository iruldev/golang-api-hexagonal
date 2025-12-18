# Story 5.1: Implement Request Validation Middleware

Status: done

## Story

As a **developer**,
I want **automatic request payload validation with size limits**,
so that **invalid requests are rejected before reaching handlers**.

## Acceptance Criteria

1. **Given** a request with JSON body decoded into DTO struct, **When** validation runs using `go-playground/validator` tags, **Then** if validation fails, HTTP 400 is returned with RFC 7807 error format **And** `validationErrors` array contains field names and messages

2. **Given** request body exceeds size limit (configurable via `MAX_REQUEST_SIZE`, default 1MB), **When** request is received, **Then** HTTP 413 Request Entity Too Large is returned **And** request body is NOT fully read into memory

3. **Given** invalid JSON syntax in request body, **When** decode fails, **Then** HTTP 400 is returned with RFC 7807 error **And** Code="VALIDATION_ERROR"

4. **And** validation happens after JSON decode, before handler logic

5. **And** existing validation functions in `contract/validation.go` continue to work

*Covers: FR25-27*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 5.1".
- Existing validation infrastructure is in `internal/transport/http/contract/validation.go`.
- RFC 7807 error handling is in `internal/transport/http/contract/error.go`.
- Router with middleware stack is in `internal/transport/http/router.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Add MAX_REQUEST_SIZE to config (AC: #2)
  - [x] 1.1 Update `internal/infra/config/config.go` to add `MaxRequestSize int64` field
  - [x] 1.2 Set default to 1048576 (1MB) via envconfig tag
  - [x] 1.3 Document in `.env.example`

- [x] Task 2: Implement RequestBodyLimiter middleware (AC: #2)
  - [x] 2.1 Create `internal/transport/http/middleware/body_limiter.go`
  - [x] 2.2 Use `http.MaxBytesReader` to limit request body size
  - [x] 2.3 Return HTTP 413 with RFC 7807 error when limit exceeded
  - [x] 2.4 Middleware accepts configurable size limit
  - [x] 2.5 Add request-too-large error code + HTTP mapping

- [x] Task 3: Create enhanced validation helpers (AC: #1, #3, #4)
  - [x] 3.1 Review existing `contract.ValidateRequestBody()` function
  - [x] 3.2 Ensure `WriteValidationError()` uses RFC 7807 format
  - [x] 3.3 Verify camelCase field naming in validation errors

- [x] Task 4: Integrate middleware into router (AC: #2)
  - [x] 4.1 Add `BodyLimiter` middleware to router middleware stack
  - [x] 4.2 Update `NewRouter` signature if needed to accept config
  - [x] 4.3 Pass MaxRequestSize from config (main + tests)
  - [x] 4.4 Position middleware AFTER RealIP, BEFORE route handlers

- [x] Task 5: Write unit tests (AC: all)
  - [x] 5.1 Create `internal/transport/http/middleware/body_limiter_test.go`
  - [x] 5.2 Test request body within limit → passes through
  - [x] 5.3 Test request body exceeds limit → 413 + RFC 7807
  - [x] 5.4 Test configurable limit values
  - [x] 5.5 Test existing validation continues to work

- [x] Task 6: Verify layer compliance (AC: all)
  - [x] 6.1 Run `make lint` to verify depguard rules pass
  - [x] 6.2 Run `make test` to ensure all tests pass
  - [ ] 6.3 Run `make ci` to verify full local CI passes (blocked: clean working tree required)

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Epic 4 (Complete):**
| File | Description |
|------|-------------|
| `internal/transport/http/contract/validation.go` | `Validate()`, `DecodeAndValidateJSON()`, `ValidateRequestBody()` |
| `internal/transport/http/contract/error.go` | `WriteProblemJSON()`, `WriteValidationError()`, RFC 7807 types |
| `internal/transport/http/middleware/requestid.go` | Existing middleware pattern |
| `internal/transport/http/middleware/logging.go` | Existing middleware pattern |
| `internal/transport/http/router.go` | Middleware stack ordering |
| `internal/infra/config/config.go` | envconfig struct |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/body_limiter.go` | Request body size limiter middleware |
| `internal/transport/http/middleware/body_limiter_test.go` | Unit tests for body limiter |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/infra/config/config.go` | Add `MaxRequestSize` field |
| `internal/transport/http/router.go` | Add BodyLimiter middleware to stack |
| `.env.example` | Document `MAX_REQUEST_SIZE` |
| `docs/sprint-artifacts/sprint-status.yaml` | Sync story/epic statuses |
| `docs/sprint-artifacts/epic-4-retro-2025-12-18.md` | Added retrospective notes |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in middleware: chi, stdlib (net/http), app (for error types), contract
❌ FORBIDDEN: pgx, uuid, direct infra imports (except config)
```

**Middleware Layer Rules:**
- Accept configuration via constructor parameters
- Use `http.MaxBytesReader` for body limiting (stdlib)
- Return RFC 7807 errors via `contract.WriteProblemJSON()`
- Chi middleware pattern: `func(next http.Handler) http.Handler`

### Request Body Limiter Implementation Pattern

```go
// internal/transport/http/middleware/body_limiter.go
package middleware

import (
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// BodyLimiter returns middleware that limits request body size.
// If limit is exceeded, returns HTTP 413 with RFC 7807 error.
func BodyLimiter(maxBytes int64) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Only limit requests with body (POST, PUT, PATCH)
            if r.ContentLength > maxBytes {
                contract.WriteProblemJSON(w, r, &app.AppError{
                    Op:      "BodyLimiter",
                    Code:    app.CodeRequestTooLarge,
                    Message: "Request body too large",
                })
                return
            }
            
            // Wrap body with MaxBytesReader to enforce limit during read
            r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
            next.ServeHTTP(w, r)
        })
    }
}
```

### Error Code for 413 Response

Add to `internal/app/errors.go`:

```go
const CodeRequestTooLarge = "REQUEST_TOO_LARGE"
```

Add mapping in `internal/transport/http/contract/error.go`:

```go
case app.CodeRequestTooLarge:
    return http.StatusRequestEntityTooLarge
```

### Config Enhancement Pattern

```go
// internal/infra/config/config.go (add to existing struct)
type Config struct {
    // ... existing fields ...
    
    // MaxRequestSize is the maximum request body size in bytes.
    // Default: 1MB (1048576 bytes)
    MaxRequestSize int64 `envconfig:"MAX_REQUEST_SIZE" default:"1048576"`
}
```

### Router Integration Pattern

```go
// internal/transport/http/router.go - updated middleware stack
func NewRouter(
    logger *slog.Logger,
    tracingEnabled bool,
    metricsReg *prometheus.Registry,
    httpMetrics metrics.HTTPMetrics,
    healthHandler, readyHandler stdhttp.Handler,
    userHandler UserRoutes,
    maxRequestSize int64, // NEW
) chi.Router {
    r := chi.NewRouter()
    
    // Middleware stack (order matters!):
    // 1. RequestID: Generate/passthrough request ID FIRST
    // 2. Tracing: Create spans and propagate trace context
    // 3. Metrics: Record request counts and durations
    // 4. Logging: Needs requestId and traceId in context
    // 5. RealIP: Extract real IP from headers
    // 6. BodyLimiter: Limit request body size
    // 7. Recoverer: Panic recovery
    r.Use(middleware.RequestID)
    if tracingEnabled {
        r.Use(middleware.Tracing)
    }
    r.Use(middleware.Metrics(httpMetrics))
    r.Use(middleware.RequestLogger(logger))
    r.Use(chiMiddleware.RealIP)
    r.Use(middleware.BodyLimiter(maxRequestSize)) // NEW
    r.Use(chiMiddleware.Recoverer)
    
    // ... rest of router setup ...
}
```

### Test Pattern

```go
//go:build !integration

package middleware

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

func TestBodyLimiter_WithinLimit(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    limiter := BodyLimiter(1024) // 1KB limit
    wrapped := limiter(handler)
    
    body := strings.Repeat("a", 500) // 500 bytes
    req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    req.ContentLength = int64(len(body))
    
    rr := httptest.NewRecorder()
    wrapped.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
}

func TestBodyLimiter_ExceedsLimit(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    limiter := BodyLimiter(100) // 100 byte limit
    wrapped := limiter(handler)
    
    body := strings.Repeat("a", 500) // 500 bytes
    req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    req.ContentLength = int64(len(body))
    
    rr := httptest.NewRecorder()
    wrapped.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
    
    var problem contract.ProblemDetail
    err := json.Unmarshal(rr.Body.Bytes(), &problem)
    require.NoError(t, err)
    assert.Equal(t, "REQUEST_TOO_LARGE", problem.Code)
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci

# Manual verification
# Test within limit (should succeed)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","firstName":"John","lastName":"Doe"}'

# Test exceeds limit (should return 413)
# First set MAX_REQUEST_SIZE=100 in environment, then:
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","firstName":"John","lastName":"Doe"}'
```

### References

- [Source: docs/epics.md#Story 5.1] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#API Design Standards] - RFC 7807 error format
- [Source: docs/project-context.md#Transport Layer] - Layer rules
- [Source: internal/transport/http/contract/validation.go] - Existing validation functions
- [Source: internal/transport/http/middleware/requestid.go] - Middleware pattern example

### Learnings from Epic 4 Retrospective

**Critical Patterns to Follow:**
1. **Import Cycle Prevention:** Use interface-based injection if needed
2. **Middleware Ordering:** Position is critical, see router.go comments
3. **RFC 7807 Enforcement:** All errors use `WriteProblemJSON()`
4. **Config via envconfig:** Use default tags, fail-fast validation

**From Story 4.6:**
- Validation happens in handler after decode
- `ValidateRequestBody()` already handles both decode and validation
- Error responses use `application/problem+json` content type

**Clock Interface Recommendation (from retro):**
- If any time-dependent logic is needed, use injected `now()` function
- Not needed for this story, but pattern established for Story 5.2

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 5.1 acceptance criteria
- `docs/architecture.md` - Middleware patterns, error handling
- `docs/project-context.md` - Transport layer conventions
- `docs/sprint-artifacts/epic-4-retro-2025-12-18.md` - Epic 4 learnings
- `internal/transport/http/middleware/*.go` - Existing middleware patterns
- `internal/transport/http/contract/validation.go` - Validation infrastructure

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- ✅ `MaxRequestSize` config field implemented with default 1MB and validation
- ✅ `CodeRequestTooLarge` error code added to `internal/app/errors.go`
- ✅ HTTP 413 status mapping added to `contract/error.go` with RFC 7807 support
- ✅ `BodyLimiter` middleware implemented using `http.MaxBytesReader` with content-length fast path
- ✅ Middleware drains over-limit bodies and rehydrates request body for downstream handlers (bounded to maxBytes+1)
- ✅ Router updated with `BodyLimiter` middleware positioned after RealIP, before Recoverer
- ✅ 4 unit tests covering: within limit, content-length exceeds, streaming exceeds, disabled
- ✅ `.env.example` documented with `MAX_REQUEST_SIZE` configuration
- ⚠️ Lint done (`make lint`), tests done (`go test ./...`); CI pending (blocked by dirty working tree requirement in make ci)

### File List

**Created:**
- `internal/transport/http/middleware/body_limiter.go`
- `internal/transport/http/middleware/body_limiter_test.go`

**Modified:**
- `internal/infra/config/config.go` - Added `MaxRequestSize` field with validation
- `internal/app/errors.go` - Added `CodeRequestTooLarge` constant
- `internal/transport/http/contract/error.go` - Added 413 status mapping and title
- `internal/transport/http/router.go` - Added `BodyLimiter` to middleware stack, `maxRequestSize` parameter
- `.env.example` - Documented `MAX_REQUEST_SIZE` environment variable
- `cmd/api/main.go` - Pass `MaxRequestSize` to `NewRouter`

# Story 4.4: Implement RFC 7807 Error Response Mapper

Status: done

## Story

As a **developer**,
I want **standardized error responses following RFC 7807**,
so that **all handlers can return consistent error format**.

## Acceptance Criteria

1. **Given** the transport layer, **When** I view `internal/transport/http/contract/error.go`, **Then** `ProblemDetail` struct exists with fields:
   - `Type` (string, URL)
   - `Title` (string)
   - `Status` (int)
   - `Detail` (string)
   - `Instance` (string)
   - `Code` (string, extension)
   - `ValidationErrors` ([]ValidationError, optional)

2. **Given** an AppError with Code="USER_NOT_FOUND", **When** error mapper processes it, **Then** HTTP status 404 is mapped **And** response Content-Type is `application/problem+json`

3. **Given** an AppError with Code="VALIDATION_ERROR", **When** error mapper processes it, **Then** HTTP status 400 is mapped **And** validationErrors array is populated

4. **Given** an unknown/internal error, **When** error mapper processes it, **Then** HTTP status 500 is returned **And** Code="INTERNAL_ERROR" **And** error details are NOT exposed (no stack trace, DB error)

5. **And** error code to HTTP status mapping is centralized

*Covers: FR60-63*

## Source of Truth (Important)

- The canonical requirements for this story are in `docs/epics.md` under "Story 4.4".
- The AppError contract is in `internal/app/errors.go`.
- If any snippet in this story conflicts with architecture.md, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create ProblemDetail Struct (AC: #1)
  - [x] 1.1 Create `internal/transport/http/contract/error.go`
  - [x] 1.2 Define `ProblemDetail` struct with all RFC 7807 fields + extensions
  - [x] 1.3 Define `ValidationError` struct with `Field` and `Message` fields
  - [x] 1.4 Add JSON tags with camelCase: `type`, `title`, `status`, `detail`, `instance`, `code`, `validationErrors`
  - [x] 1.5 Add `omitempty` to optional fields (ValidationErrors)

- [x] Task 2: Implement Error Code to HTTP Status Mapping (AC: #2, #4, #5)
  - [x] 2.1 Create centralized mapping function `mapCodeToStatus(code string) int`
  - [x] 2.2 Map `app.CodeUserNotFound` → HTTP 404
  - [x] 2.3 Map `app.CodeEmailExists` → HTTP 409 (Conflict)
  - [x] 2.4 Map `app.CodeValidationError` → HTTP 400
  - [x] 2.5 Map `app.CodeInternalError` → HTTP 500
  - [x] 2.6 Default unknown codes → HTTP 500

- [x] Task 3: Implement Error Response Writer (AC: #2, #3, #4)
  - [x] 3.1 Create `WriteProblemJSON(w http.ResponseWriter, r *http.Request, err error)` function
  - [x] 3.2 Extract AppError using `errors.As()`
  - [x] 3.3 Build `ProblemDetail` with proper Type URL, Title, Status, Detail
  - [x] 3.4 Set `Content-Type: application/problem+json`
  - [x] 3.5 For internal errors, use generic message (don't expose internal details)
  - [x] 3.6 Set `Instance` to request path `r.URL.Path`

- [x] Task 4: Implement Validation Error Helper (AC: #3)
  - [x] 4.1 Create `NewValidationProblem(r *http.Request, validationErrors []ValidationError) *ProblemDetail`
  - [x] 4.2 Helper populates Type, Title, Status=400, Code="VALIDATION_ERROR"
  - [x] 4.3 Helper includes validationErrors array

- [x] Task 5: Create Problem Type URL Constants (AC: #1)
  - [x] 5.1 Define base URL constant (e.g., `https://api.example.com/problems/`)
  - [x] 5.2 Define type URLs: `validation-error`, `not-found`, `conflict`, `internal-error`
  - [x] 5.3 Create helper `problemTypeURL(slug string) string`

- [x] Task 6: Write Unit Tests for Error Mapper (AC: all)
  - [x] 6.1 Create `internal/transport/http/contract/error_test.go`
  - [x] 6.2 Test USER_NOT_FOUND → 404 + correct ProblemDetail
  - [x] 6.3 Test VALIDATION_ERROR → 400 + validationErrors populated
  - [x] 6.4 Test EMAIL_EXISTS → 409
  - [x] 6.5 Test INTERNAL_ERROR → 500 + generic message (no internal details)
  - [x] 6.6 Test unknown error type → 500 + generic message
  - [x] 6.7 Test Content-Type header is `application/problem+json`

- [x] Task 7: Verify Layer Compliance (AC: all)
  - [x] 7.1 Run `make lint` to verify no depguard violations
  - [x] 7.2 Confirm contract package imports only: domain, app, stdlib
  - [x] 7.3 Run `make test` to ensure all unit tests pass
  - [x] 7.4 Run `make coverage` to verify coverage threshold

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Story 4.1 + 4.2 + 4.3 (Complete):**
| File | Description |
|------|-------------|
| `internal/app/errors.go` | AppError struct with Op, Code, Message, Err fields |
| `internal/app/errors.go` | Error codes: `CodeUserNotFound`, `CodeEmailExists`, `CodeValidationError`, `CodeInternalError` |
| `internal/transport/http/handler/health.go` | Existing handler pattern (JSON encoding) |
| `internal/transport/http/contract/.keep` | Empty directory (placeholder) |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/contract/error.go` | ProblemDetail, ValidationError, error mapper |
| `internal/transport/http/contract/error_test.go` | Unit tests for error mapper |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in transport/contract: domain, app, stdlib
❌ FORBIDDEN: pgx, infra imports
```

**Transport Layer Rules:**
- ONLY place that knows HTTP status codes
- Map `AppError.Code` → HTTP status + RFC 7807
- Use `application/problem+json` content type for errors

### RFC 7807 Problem Details Format

```go
// internal/transport/http/contract/error.go
package contract

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
)

// ProblemBaseURL is the base URL for problem type URIs.
const ProblemBaseURL = "https://api.example.com/problems/"

// ProblemDetail represents an RFC 7807 Problem Details response.
type ProblemDetail struct {
    Type             string            `json:"type"`
    Title            string            `json:"title"`
    Status           int               `json:"status"`
    Detail           string            `json:"detail"`
    Instance         string            `json:"instance"`
    Code             string            `json:"code"`
    ValidationErrors []ValidationError `json:"validationErrors,omitempty"`
}

// ValidationError represents a single field validation error.
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### Error Code to HTTP Status Mapping

```go
// mapCodeToStatus maps AppError.Code to HTTP status code.
func mapCodeToStatus(code string) int {
    switch code {
    case app.CodeUserNotFound:
        return http.StatusNotFound // 404
    case app.CodeEmailExists:
        return http.StatusConflict // 409
    case app.CodeValidationError:
        return http.StatusBadRequest // 400
    case app.CodeInternalError:
        return http.StatusInternalServerError // 500
    default:
        return http.StatusInternalServerError // 500
    }
}
```

### Error Response Writer Pattern

```go
// WriteProblemJSON writes an RFC 7807 error response.
func WriteProblemJSON(w http.ResponseWriter, r *http.Request, err error) {
    var appErr *app.AppError
    if !errors.As(err, &appErr) {
        // Unknown error → internal error (don't expose details)
        appErr = &app.AppError{
            Op:      "unknown",
            Code:    app.CodeInternalError,
            Message: "An internal error occurred",
            Err:     err,
        }
    }
    
    status := mapCodeToStatus(appErr.Code)
    
    problem := ProblemDetail{
        Type:     problemTypeURL(codeToTypeSlug(appErr.Code)),
        Title:    codeToTitle(appErr.Code),
        Status:   status,
        Detail:   safeDetail(appErr),
        Instance: r.URL.Path,
        Code:     appErr.Code,
    }
    
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(problem)
}

// safeDetail returns a safe error message (no internal details for 5xx).
func safeDetail(appErr *app.AppError) string {
    if mapCodeToStatus(appErr.Code) >= 500 {
        return "An internal error occurred"
    }
    return appErr.Message
}

// codeToTitle returns a human-readable title for the error code.
func codeToTitle(code string) string {
    switch code {
    case app.CodeUserNotFound:
        return "User Not Found"
    case app.CodeEmailExists:
        return "Email Already Exists"
    case app.CodeValidationError:
        return "Validation Error"
    case app.CodeInternalError:
        return "Internal Server Error"
    default:
        return "Internal Server Error"
    }
}

// codeToTypeSlug maps AppError.Code to a stable RFC 7807 type slug.
func codeToTypeSlug(code string) string {
    switch code {
    case app.CodeValidationError:
        return "validation-error"
    case app.CodeUserNotFound:
        return "not-found"
    case app.CodeEmailExists:
        return "conflict"
    default:
        return "internal-error"
    }
}

// problemTypeURL returns the RFC 7807 type URL.
func problemTypeURL(slug string) string {
    return ProblemBaseURL + slug
}
```

### Validation Error Helper

```go
// NewValidationProblem creates a ProblemDetail for validation errors.
func NewValidationProblem(r *http.Request, validationErrors []ValidationError) *ProblemDetail {
    return &ProblemDetail{
        Type:             ProblemBaseURL + "validation-error",
        Title:            "Validation Error",
        Status:           http.StatusBadRequest,
        Detail:           "One or more fields failed validation",
        Instance:         r.URL.Path,
        Code:             app.CodeValidationError,
        ValidationErrors: validationErrors,
    }
}

// WriteValidationError writes a validation error response.
func WriteValidationError(w http.ResponseWriter, r *http.Request, validationErrors []ValidationError) {
    problem := NewValidationProblem(r, validationErrors)
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(problem)
}
```

### Test Pattern

```go
//go:build !integration

package contract

import (
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
)

func TestWriteProblemJSON(t *testing.T) {
    tests := []struct {
        name           string
        err            error
        wantStatus     int
        wantCode       string
        wantTitle      string
        wantNoInternal bool // verify internal details not exposed
    }{
        {
            name: "USER_NOT_FOUND maps to 404",
            err: &app.AppError{
                Op:      "GetUser",
                Code:    app.CodeUserNotFound,
                Message: "User not found",
            },
            wantStatus: http.StatusNotFound,
            wantCode:   app.CodeUserNotFound,
            wantTitle:  "User Not Found",
        },
        {
            name: "INTERNAL_ERROR hides details",
            err: &app.AppError{
                Op:      "CreateUser",
                Code:    app.CodeInternalError,
                Message: "database connection failed: SQLSTATE 42P01", // sensitive
            },
            wantStatus:     http.StatusInternalServerError,
            wantCode:       app.CodeInternalError,
            wantTitle:      "Internal Server Error",
            wantNoInternal: true,
        },
        {
            name:           "unknown error becomes INTERNAL_ERROR",
            err:            errors.New("something went wrong"),
            wantStatus:     http.StatusInternalServerError,
            wantCode:       app.CodeInternalError,
            wantNoInternal: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
            rec := httptest.NewRecorder()
            
            WriteProblemJSON(rec, req, tt.err)
            
            assert.Equal(t, tt.wantStatus, rec.Code)
            assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
            // ... parse JSON and verify fields
        })
    }
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run coverage check (domain + app must be ≥ 80%)
make coverage

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci
```

### References

- [Source: docs/epics.md#Story 4.4] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Error Format] - RFC 7807 Problem Details format
- [Source: docs/architecture.md#API Design Standards] - Error response structure
- [Source: docs/project-context.md#Error Handling] - Layered error strategy
- [Source: docs/project-context.md#Transport Layer] - Handler rules
- [Source: internal/app/errors.go] - AppError struct and error codes

### Learnings from Story 4.1 + 4.2 + 4.3

**From Story 4.3:**
- AppError struct with Op, Code, Message, Err fields already exists
- Error codes: CodeUserNotFound, CodeEmailExists, CodeValidationError, CodeInternalError
- Transport layer is the ONLY place that knows HTTP status codes
- Use `errors.As()` to extract typed errors

**From Epic 3 Retrospective:**
- depguard rules enforced - transport layer CANNOT import pgx
- Use table-driven tests with testify assertions
- Run `make ci` before marking story complete

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 4.4 acceptance criteria
- `docs/architecture.md` - RFC 7807 error format + API standards
- `docs/project-context.md` - Transport layer conventions
- `internal/app/errors.go` - AppError struct and error codes

### Agent Model Used

Gemini 2.5

### Debug Log References

N/A

### Completion Notes List

- Implemented RFC 7807 ProblemDetail struct with all required fields and JSON tags
- Created centralized error code to HTTP status mapping (mapCodeToStatus)
- Implemented WriteProblemJSON for generic error responses with internal detail hiding
- Implemented WriteValidationError and NewValidationProblem for validation errors
- Populated `validationErrors` for `VALIDATION_ERROR` using wrapped domain validation errors
- Created comprehensive unit tests (9 test functions)
- All tests pass with race detection
- Lint passes with 0 issues confirming layer compliance
- Added configurable `PROBLEM_BASE_URL` (env) to control RFC 7807 `type` base URL (wired at startup)
- Exposed problem type slug constants: `ProblemTypeValidationErrorSlug`, `ProblemTypeNotFoundSlug`, `ProblemTypeConflictSlug`, `ProblemTypeInternalErrorSlug`

### File List

- `internal/transport/http/contract/error.go` (NEW) - RFC 7807 error response mapper
- `internal/transport/http/contract/error_test.go` (NEW) - Unit tests for error mapper
- `internal/transport/http/contract/.keep` (DELETED) - Removed placeholder file
- `docs/sprint-artifacts/sprint-status.yaml` (MODIFIED) - Story status updated for sprint tracking
- `.env.example` (MODIFIED) - Documented `PROBLEM_BASE_URL`
- `internal/infra/config/config.go` (MODIFIED) - Added `PROBLEM_BASE_URL` config + validation
- `internal/infra/config/config_test.go` (MODIFIED) - Tests for `PROBLEM_BASE_URL`
- `cmd/api/main.go` (MODIFIED) - Wired `PROBLEM_BASE_URL` into `contract.SetProblemBaseURL`

### Change Log

- 2025-12-18: Implemented RFC 7807 Error Response Mapper (Story 4.4)

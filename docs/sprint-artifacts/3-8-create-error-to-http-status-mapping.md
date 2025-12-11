# Story 3.8: Create Error to HTTP Status Mapping

Status: done

## Story

As a developer,
I want application errors mapped to HTTP status codes,
So that clients receive appropriate status codes.

## Acceptance Criteria

### AC1: ErrNotFound → 404
**Given** `ErrNotFound` domain error
**When** handler returns this error
**Then** HTTP status is 404

### AC2: ErrValidation → 422
**Given** `ErrValidation` domain error
**When** handler returns this error
**Then** HTTP status is 422 (aligned with Story 3.7)

### AC3: ErrUnauthorized → 401
**Given** `ErrUnauthorized` domain error
**When** handler returns this error
**Then** HTTP status is 401

---

## Tasks / Subtasks

- [x] **Task 1: Create domain error types** (AC: #1, #2, #3)
  - [x] Create `internal/domain/errors.go`
  - [x] Define `ErrNotFound` sentinel error
  - [x] Define `ErrValidation` sentinel error
  - [x] Define `ErrUnauthorized` sentinel error
  - [x] Define `ErrForbidden`, `ErrConflict`, `ErrInternal`

- [x] **Task 2: Create error mapper** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/response/mapper.go`
  - [x] Implement `MapError(err error) (status int, code string)`
  - [x] Use `errors.Is()` for error comparison
  - [x] Return appropriate HTTP status and error code

- [x] **Task 3: Create HandleError helper** (AC: #1, #2, #3)
  - [x] Implement `HandleError(w, err error)`
  - [x] Maps domain error to HTTP status
  - [x] Returns envelope response format

- [x] **Task 4: Create mapper tests** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/response/mapper_test.go`
  - [x] Test: ErrNotFound returns 404
  - [x] Test: ErrValidation returns 400
  - [x] Test: ErrUnauthorized returns 401
  - [x] Test: Unknown errors return 500

- [x] **Task 5: Create domain error tests** (AC: #1, #2, #3)
  - [x] Create `internal/domain/errors_test.go`
  - [x] Test: errors.Is() works correctly
  - [x] Test: error messages are correct

- [x] **Task 6: Verify implementation** (AC: #1, #2, #3)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Domain Errors

```go
// internal/domain/errors.go
package domain

import "errors"

var (
    // ErrNotFound indicates a resource was not found (HTTP 404).
    ErrNotFound = errors.New("resource not found")

    // ErrValidation indicates invalid input data (HTTP 400).
    ErrValidation = errors.New("validation error")

    // ErrUnauthorized indicates missing/invalid credentials (HTTP 401).
    ErrUnauthorized = errors.New("unauthorized")

    // ErrForbidden indicates insufficient permissions (HTTP 403).
    ErrForbidden = errors.New("forbidden")

    // ErrConflict indicates a conflict with current state (HTTP 409).
    ErrConflict = errors.New("conflict")

    // ErrInternal indicates an internal server error (HTTP 500).
    ErrInternal = errors.New("internal error")
)
```

### Error Mapper

```go
// internal/interface/http/response/mapper.go
package response

import (
    "errors"
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// MapError maps a domain error to HTTP status and error code.
func MapError(err error) (status int, code string) {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound, ErrNotFound
    case errors.Is(err, domain.ErrValidation):
        return http.StatusBadRequest, ErrValidation
    case errors.Is(err, domain.ErrUnauthorized):
        return http.StatusUnauthorized, ErrUnauthorized
    case errors.Is(err, domain.ErrForbidden):
        return http.StatusForbidden, ErrForbidden
    case errors.Is(err, domain.ErrConflict):
        return http.StatusConflict, ErrConflict
    default:
        return http.StatusInternalServerError, ErrInternalServer
    }
}

// HandleError writes an error response based on domain error type.
func HandleError(w http.ResponseWriter, err error) {
    status, code := MapError(err)
    Error(w, status, code, err.Error())
}
```

### Handler Usage

```go
// handlers/user.go
func GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := userService.GetByID(ctx, id)
    if err != nil {
        response.HandleError(w, err)  // Automatically maps domain error
        return
    }
    response.Success(w, user)
}
```

### Architecture Compliance

**Layer:** `internal/domain` (errors), `internal/interface/http/response` (mapper)
**Pattern:** Clean separation of domain errors from HTTP concerns
**Benefit:** Domain layer doesn't know about HTTP, easy to change status codes

### References

- [Source: docs/epics.md#Story-3.8]
- [Story 3.7 - Response Envelope](file:///docs/sprint-artifacts/3-7-implement-response-envelope-pattern.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Eighth and final story in Epic 3: HTTP API Core.
Establishes error mapping from domain to HTTP layer.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/domain/errors.go` - Domain error types
- `internal/domain/errors_test.go` - Domain error tests
- `internal/interface/http/response/mapper.go` - Error mapper
- `internal/interface/http/response/mapper_test.go` - Mapper tests

Files to modify:
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

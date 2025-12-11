# Story 3.7: Implement Response Envelope Pattern

Status: done

## Story

As a developer,
I want consistent response format for success/error,
So that clients can parse responses uniformly.

## Acceptance Criteria

### AC1: Success response format
**Given** a successful request
**When** response is sent
**Then** body is `{"success": true, "data": {...}}`

### AC2: Error response format
**Given** an error occurs
**When** response is sent
**Then** body is `{"success": false, "error": {"code": "ERR_*", "message": "..."}}`

---

## Tasks / Subtasks

- [x] **Task 1: Create response types** (AC: #1, #2)
  - [x] Create `internal/interface/http/response/response.go`
  - [x] Define `SuccessResponse` struct with `interface{}` data
  - [x] Define `ErrorResponse` struct with code and message
  - [x] Define `ErrorDetail` struct for error info

- [x] **Task 2: Create response helper functions** (AC: #1, #2)
  - [x] Implement `JSON(w, status int, data interface{})`
  - [x] Implement `Success(w, data interface{})`
  - [x] Implement `Error(w, status int, code, message string)`
  - [x] Set Content-Type header automatically

- [x] **Task 3: Create error codes** (AC: #2)
  - [x] Create `internal/interface/http/response/errors.go`
  - [x] Define standard error codes: ERR_BAD_REQUEST, ERR_NOT_FOUND, etc.
  - [x] Document error code usage pattern

- [x] **Task 4: Update example handler** (AC: #1)
  - [x] Modify `handlers/example.go` to use response helpers
  - [x] Update health handler to use response pattern
  - [x] Ensure consistent JSON structure

- [x] **Task 5: Create response tests** (AC: #1, #2)
  - [x] Create `internal/interface/http/response/response_test.go`
  - [x] Test: Success response has correct format
  - [x] Test: Error response has correct format
  - [x] Test: Content-Type is application/json

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Response Types

```go
// internal/interface/http/response/response.go
package response

import (
    "encoding/json"
    "net/http"
)

// SuccessResponse represents a successful API response.
type SuccessResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response.
type ErrorResponse struct {
    Success bool        `json:"success"`
    Error   ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Response Helpers

```go
// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Success writes a success response with the given data.
func Success(w http.ResponseWriter, data interface{}) {
    JSON(w, http.StatusOK, SuccessResponse{
        Success: true,
        Data:    data,
    })
}

// Error writes an error response with the given details.
func Error(w http.ResponseWriter, status int, code, message string) {
    JSON(w, status, ErrorResponse{
        Success: false,
        Error: ErrorDetail{
            Code:    code,
            Message: message,
        },
    })
}
```

### Standard Error Codes

```go
// internal/interface/http/response/errors.go
package response

const (
    ErrBadRequest     = "ERR_BAD_REQUEST"
    ErrUnauthorized   = "ERR_UNAUTHORIZED"
    ErrForbidden      = "ERR_FORBIDDEN"
    ErrNotFound       = "ERR_NOT_FOUND"
    ErrConflict       = "ERR_CONFLICT"
    ErrValidation     = "ERR_VALIDATION"
    ErrInternalServer = "ERR_INTERNAL_SERVER"
)
```

### Handler Usage

```go
// handlers/example.go
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{
        "message": "Example handler working correctly",
    }
    response.Success(w, data)
}

// Error example
func FailingHandler(w http.ResponseWriter, r *http.Request) {
    response.Error(w, http.StatusNotFound, response.ErrNotFound, "Resource not found")
}
```

### Architecture Compliance

**Layer:** `internal/interface/http/response`
**Pattern:** Consistent API response structure
**Benefit:** Clients can reliably parse all API responses

### References

- [Source: docs/epics.md#Story-3.7]
- [Story 3.6 - Handler Pattern](file:///docs/sprint-artifacts/3-6-create-handler-registration-pattern.md)
- [Story 3.8 - Error Mapping](file:///docs/sprint-artifacts/3-8-create-error-to-http-status-mapping.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Seventh story in Epic 3: HTTP API Core.
Establishes response envelope pattern for uniform API responses.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/interface/http/response/response.go` - Response types and helpers
- `internal/interface/http/response/errors.go` - Error codes
- `internal/interface/http/response/response_test.go` - Response tests

Files to modify:
- `internal/interface/http/handlers/example.go` - Use response helpers
- `internal/interface/http/handlers/health.go` - Use response helpers
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

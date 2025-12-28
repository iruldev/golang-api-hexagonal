# Story 3.2: App→Transport Error Mapping

Status: done

## Story

As a **developer**,
I want consistent error mapping,
so that domain errors become proper HTTP responses.

## Acceptance Criteria

1. **AC1:** Mapping table: domain error → HTTP status
2. **AC2:** Unit tests cover all error mappings
3. **AC3:** Unknown errors map to 500
4. **AC4:** Mapping preserves error code in response

## Tasks / Subtasks

- [x] Task 1: Update error mapping to use new domain errors (AC: #1, #4)
  - [x] Update `contract/error.go` to import `domain/errors`
  - [x] Map `ErrorCode` from domain errors to HTTP status
  - [x] Ensure error code preserved in ProblemDetail response
- [x] Task 2: Ensure unknown errors map to 500 (AC: #3)
  - [x] Verify `defaultErrorDef` handles unknown errors
  - [x] Test error wrapping scenarios
  - [x] Ensure internal details not exposed for 5xx
- [x] Task 3: Add comprehensive unit tests (AC: #2)
  - [x] Test all error code → status mappings
  - [x] Test unknown error handling
  - [x] Test error code preservation in response

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **FR-21:** Error mapping

### Current State

Existing `internal/transport/http/contract/error.go` already has:

```go
// errorRegistry maps app codes to HTTP statuses
var errorRegistry = map[string]errorDef{
    app.CodeUserNotFound:    {http.StatusNotFound, ...},
    app.CodeEmailExists:     {http.StatusConflict, ...},
    app.CodeValidationError: {http.StatusBadRequest, ...},
    // ... more mappings
}

var defaultErrorDef = errorDef{
    Status: http.StatusInternalServerError,
    Title:  "Internal Server Error",
    // Unknown errors → 500 (AC3 already satisfied)
}
```

### Story 3.1 Context

Story 3.1 introduced `internal/domain/errors` with:
- `DomainError` type with `Code` field
- `ErrorCode` constants like `ErrCodeUserNotFound`
- errors.Is/As support

### Integration Approach

The current implementation uses `app.CodeXxx` strings. Story 3.2 should:
1. Create a mapping from `errors.ErrorCode` → HTTP status
2. Update `WriteProblemJSON` to check for `*errors.DomainError`
3. Extract code from `DomainError.Code` for response

### Proposed Updates

```go
import (
    domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// domainErrorRegistry maps domain error codes to HTTP definitions
var domainErrorRegistry = map[domainerrors.ErrorCode]errorDef{
    domainerrors.ErrCodeUserNotFound:     {http.StatusNotFound, "User Not Found", ProblemTypeNotFoundSlug},
    domainerrors.ErrCodeEmailExists:      {http.StatusConflict, "Email Already Exists", ProblemTypeConflictSlug},
    domainerrors.ErrCodeInvalidEmail:     {http.StatusBadRequest, "Validation Error", ProblemTypeValidationErrorSlug},
    domainerrors.ErrCodeValidation:       {http.StatusBadRequest, "Validation Error", ProblemTypeValidationErrorSlug},
    domainerrors.ErrCodeNotFound:         {http.StatusNotFound, "Not Found", ProblemTypeNotFoundSlug},
    domainerrors.ErrCodeConflict:         {http.StatusConflict, "Conflict", ProblemTypeConflictSlug},
    domainerrors.ErrCodeUnauthorized:     {http.StatusUnauthorized, "Unauthorized", ProblemTypeUnauthorizedSlug},
    domainerrors.ErrCodeForbidden:        {http.StatusForbidden, "Forbidden", ProblemTypeForbiddenSlug},
    domainerrors.ErrCodeInternal:         {http.StatusInternalServerError, "Internal Server Error", ProblemTypeInternalErrorSlug},
}

func getDomainErrorDef(code domainerrors.ErrorCode) errorDef {
    if def, ok := domainErrorRegistry[code]; ok {
        return def
    }
    return defaultErrorDef // AC3: Unknown → 500
}

// In WriteProblemJSON:
var domainErr *domainerrors.DomainError
if errors.As(err, &domainErr) {
    def := getDomainErrorDef(domainErr.Code)
    problem := ProblemDetail{
        Status: def.Status,
        Code:   string(domainErr.Code), // AC4: Preserve code
        // ...
    }
}
```

### Existing Tests

`contract/error_test.go` exists with tests for:
- `TestWriteProblemJSON_UnknownError`
- `TestWriteProblemJSON_*` for various error types

Story 3.2 should add tests for domain error mapping.

### Testing Standards

- Unit tests for all error code mappings
- Table-driven tests for mapping table
- Test unknown/unmapped codes → 500
- Verify error code appears in JSON response

### Previous Story Learnings (Story 3.1)

- Use `errors.As()` for DomainError extraction
- Return `error` interface (not concrete type)
- Check for nil receiver

### References

- [Source: _bmad-output/architecture.md#Error Handling]
- [Source: _bmad-output/epics.md#Story 3.2]
- [Source: _bmad-output/prd.md#FR21]
- [Existing: internal/transport/http/contract/error.go]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List
 
 _Files created/modified during implementation:_
 - [x] `internal/transport/http/contract/error.go` (modify)
 - [x] `internal/transport/http/contract/error_test.go` (add tests)
 - [x] `_bmad-output/sprint-status.yaml` (sync status)
 
 ## Senior Developer Review (AI)
 
 _Reviewer: @bmad-bmm-workflows-code-review on 2025-12-28_
 
 ### Findings
 - **[CRITICAL] Fixed**: `domainErrorRegistry` was defined but unused. Implemented `WriteProblemJSON` logic to prioritize domain errors.
 - **[CRITICAL] Fixed**: Missing unit tests for domain error mappings. Added comprehensive tests in `error_test.go` covering ALL domain error codes (AC2).
 - **[MEDIUM] Fixed**: `WriteProblemJSON` implementation initially lost `ValidationErrors` for domain validation errors. Fixed to populate them correctly.
 - **[MEDIUM] Fixed**: Legacy tests in `error_test.go` were failing because they expected legacy error codes. Updated tests to expect the new stable domain error codes (AC4 compliant).
 - **[MEDIUM] Fixed**: `sprint-status.yaml` was modified but missing from File List. Added.
 - **[MEDIUM] Fixed**: `error_test.go` lacked explicit cases for `InvalidFirstName/LastName`. Also expanded coverage for Audit, Authorization, and Conflict errors.
 
 ### Outcome
 **Approved** with automated fixes applied.

# Story 1.4: Handle UUID v7 Parsing Gracefully

Status: done

## Story

As a **developer**,
I want UUID v7 parsing errors to be handled gracefully,
so that invalid IDs return 400 Bad Request instead of 500.

## Acceptance Criteria

1. **Given** an API request with an invalid UUID format
   **When** the handler parses the UUID
   **Then** the system returns 400 Bad Request with RFC7807 error

2. **And** the error includes descriptive message (not stack trace)

3. **And** unit tests cover valid and invalid UUID scenarios

## Tasks / Subtasks

- [x] Task 1: Verify Current Handler UUID Validation (AC: #1, #2)
  - [x] Confirmed `GetUser` handler validates UUID format and version
  - [x] Confirmed RFC7807 response does not leak stack traces
  - [x] Error messages now field-level: `field: "id"`

- [x] Task 2: Improve Error Consistency (AC: #2)
  - [x] Added field="id" to validation error for API consistency
  - [x] Distinguished messages:
    - Invalid format → "must be a valid UUID"
    - Wrong version → "must be UUID v7 (time-ordered)"

- [x] Task 3: Add Empty ID Edge Case Test (AC: #3)
  - [x] Created `TestUserHandler_GetUser_EmptyID`
  - [x] Test passes with 400 and proper field-level error

- [x] Task 4: Verify Existing Tests (AC: #3)
  - [x] `TestUserHandler_GetUser_InvalidUUID` PASS
  - [x] `TestUserHandler_GetUser_InvalidUUIDVersion` PASS
  - [x] Both assert `application/problem+json` content-type

## Dev Notes

### Updated Handler

```go
// Validate UUID format and version
parsedID, err := uuid.Parse(idParam)
if err != nil {
    contract.WriteValidationError(w, r, []contract.ValidationError{
        {Field: "id", Message: "must be a valid UUID"},
    })
    return
}
if parsedID.Version() != 7 {
    contract.WriteValidationError(w, r, []contract.ValidationError{
        {Field: "id", Message: "must be UUID v7 (time-ordered)"},
    })
    return
}
```

### Improved Response Format

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "code": "VALIDATION_ERROR",
  "validationErrors": [
    {"field": "id", "message": "must be a valid UUID"}
  ]
}
```

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- Build: SUCCESS
- Handler Tests: 3 UUID tests PASS
- Regression: 15 packages ALL PASS

### Completion Notes List

- Refactored UUID validation to use `WriteValidationError` with field-level errors
- Separated format error vs version error for better developer experience
- Added `TestUserHandler_GetUser_EmptyID` edge case test
- All existing tests continue to pass

### File List

- `internal/transport/http/handler/user.go` - MODIFIED (field-level UUID errors, canonical ID usage)
- `internal/transport/http/handler/user_test.go` - MODIFIED (added EmptyID, MixedCase, NilUUID tests)

### Change Log

- 2024-12-24: Improved UUID validation with field-level errors
- 2024-12-24: Separated format vs version error messages
- 2024-12-24: Added empty ID edge case test
- 2024-12-24: [Code Review] Enforced canonical UUID usage in UseCase call
- 2024-12-24: [Code Review] Added mixed-case and nil UUID tests

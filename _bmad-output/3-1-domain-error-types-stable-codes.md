# Story 3.1: Domain Error Types + Stable Codes

Status: done

## Story

As a **developer**,
I want well-defined domain errors,
so that error handling is consistent.

## Acceptance Criteria

1. **AC1:** `internal/domain/errors/` package with error types
2. **AC2:** Error codes format: `ERR_{DOMAIN}_{CODE}` (e.g., `ERR_USER_NOT_FOUND`)
3. **AC3:** `errors.Is` / `errors.As` work correctly
4. **AC4:** Error codes are documented and stable (no breaking changes)

## Tasks / Subtasks

- [x] Task 1: Create errors package structure (AC: #1)
  - [x] Create `internal/domain/errors/` package
  - [x] Define base `DomainError` type with Code field
  - [x] Implement `Error()` method
  - [x] Implement `Unwrap()` for error chaining
- [x] Task 2: Define error codes (AC: #2, #4)
  - [x] Define `ErrorCode` type
  - [x] Create constants: `ERR_USER_NOT_FOUND`, etc.
  - [x] Document all codes in package doc
  - [x] Add code stability comment
- [x] Task 3: Migrate existing errors (AC: #1)
  - [x] Replace sentinel errors in `domain/errors.go`
  - [x] Update all error references in codebase
  - [x] Ensure backward compatibility
- [x] Task 4: Verify errors.Is/As support (AC: #3)
  - [x] Add tests for `errors.Is()` matching
  - [x] Add tests for `errors.As()` type assertion
  - [x] Test error wrapping scenarios

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **FR-20:** Domain error types
- **FR-23:** Stable error codes

### Current State

Existing `internal/domain/errors.go`:
```go
// Sentinel errors - no codes, no structure
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidEmail      = errors.New("invalid email format")
    // ... more sentinel errors
)
```

### Proposed Error Type

```go
package errors

// ErrorCode represents a stable, documented error code.
type ErrorCode string

// Stable error codes - DO NOT CHANGE existing codes.
const (
    // User domain errors
    ErrCodeUserNotFound      ErrorCode = "ERR_USER_NOT_FOUND"
    ErrCodeEmailExists       ErrorCode = "ERR_USER_EMAIL_EXISTS"
    ErrCodeInvalidEmail      ErrorCode = "ERR_USER_INVALID_EMAIL"
    ErrCodeInvalidFirstName  ErrorCode = "ERR_USER_INVALID_FIRST_NAME"
    ErrCodeInvalidLastName   ErrorCode = "ERR_USER_INVALID_LAST_NAME"

    // Audit domain errors
    ErrCodeAuditNotFound     ErrorCode = "ERR_AUDIT_NOT_FOUND"
    ErrCodeInvalidEventType  ErrorCode = "ERR_AUDIT_INVALID_EVENT_TYPE"
    ErrCodeInvalidEntityType ErrorCode = "ERR_AUDIT_INVALID_ENTITY_TYPE"
    ErrCodeInvalidEntityID   ErrorCode = "ERR_AUDIT_INVALID_ENTITY_ID"
)

// DomainError represents a domain-layer error with a stable code.
type DomainError struct {
    Code    ErrorCode
    Message string
    Err     error // wrapped error
}

func (e *DomainError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
    return e.Err
}

// Is implements errors.Is matching by code.
func (e *DomainError) Is(target error) bool {
    if t, ok := target.(*DomainError); ok {
        return e.Code == t.Code
    }
    return false
}
```

### Error Constructors

```go
// User errors
var (
    ErrUserNotFound = &DomainError{
        Code:    ErrCodeUserNotFound,
        Message: "user not found",
    }
    ErrEmailExists = &DomainError{
        Code:    ErrCodeEmailExists,
        Message: "email already exists",
    }
)

// NewUserNotFound creates a new user not found error with context.
func NewUserNotFound(userID string) error {
    return &DomainError{
        Code:    ErrCodeUserNotFound,
        Message: fmt.Sprintf("user %s not found", userID),
    }
}
```

### Testing errors.Is/As

```go
func TestDomainErrorIs(t *testing.T) {
    err := NewUserNotFound("123")
    
    assert.True(t, errors.Is(err, ErrUserNotFound))
    assert.False(t, errors.Is(err, ErrEmailExists))
}

func TestDomainErrorAs(t *testing.T) {
    err := NewUserNotFound("123")
    
    var domainErr *DomainError
    require.True(t, errors.As(err, &domainErr))
    assert.Equal(t, ErrCodeUserNotFound, domainErr.Code)
}

func TestDomainErrorWrapping(t *testing.T) {
    cause := errors.New("db connection failed")
    err := &DomainError{
        Code:    ErrCodeUserNotFound,
        Message: "user lookup failed",
        Err:     cause,
    }
    
    assert.True(t, errors.Is(err, cause))
}
```

### Migration Strategy

1. Create new `internal/domain/errors/` package
2. Define all error types with codes
3. Update `internal/domain/errors.go` to re-export from new package
4. Gradually update references across codebase

### Testing Standards

- Unit tests for all error types
- Test `errors.Is()` and `errors.As()` behavior
- Test error message formatting
- Test wrapping/unwrapping

### Previous Epic Learnings (Epic 2)

- Use table-driven tests for multiple error scenarios
- Test edge cases (nil errors, empty codes)
- Document stability guarantees in code comments

### References

- [Source: _bmad-output/architecture.md#Error Handling]
- [Source: _bmad-output/epics.md#Story 3.1]
- [Source: _bmad-output/prd.md#FR20, FR23]
- [Go errors package](https://pkg.go.dev/errors)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

- Migrated `internal/domain/errors.go` to alias the new `internal/domain/errors` package.
- Implemented `DomainError` with stable codes in `internal/domain/errors`.
- Verified `errors.Is` and `errors.As` support with comprehensive tests.
- All tests passing.

### File List

_Files created/modified during implementation:_
- [x] `internal/domain/errors/errors.go` (new)
- [x] `internal/domain/errors/codes.go` (new)
- [x] `internal/domain/errors/errors_test.go` (new)
- [x] `internal/domain/errors.go` (migrate to re-export)
- [x] `internal/domain/errors_test.go` (new / compatibility tests)

## Senior Developer Review (AI)

_Reviewer: @bmad-bmm-workflows-code-review on 2025-12-28_

### Findings
- **[HIGH] Fixed**: Uncommitted changes in `internal/domain/errors.go` (aliases). Git status showed modifications were not staged. Added to git.
- **[MEDIUM] Fixed**: Constructors and helper functions were returning concrete `*DomainError` types instead of the `error` interface. This is non-idiomatic in Go and risks typed nil issues. Refactored all constructors to return `error`.
- **[MEDIUM] Fixed**: Implementation files in `internal/domain/errors/` were not tracked in git. Added them.
- **[MEDIUM] Fixed**: Missing compatibility tests for `internal/domain/errors.go` aliases. Added `internal/domain/errors_test.go` to verify aliases map correctly to new error types.
- **[LOW] Fixed**: `DomainError.Error()` method lacked nil receiver check, which could cause panics if called on a typed nil. Added check to return "<nil>".
- **[LOW] Fixed**: Missing documentation comments for exported sentinel errors in `codes.go`. Added comments.
- **[LOW] Note**: `internal/domain/errors.go` aliases `ErrEmailAlreadyExists` to `ErrEmailExists`. This is acceptable for backward compatibility.

### Outcome
**Approved** with automated fixes applied.

# Story 2.2: Create Central Error Code Registry

Status: Done

## Story

As a developer,
I want typed domain errors with public codes,
So that errors are consistent and API clients can handle them.

## Acceptance Criteria

1. **Given** `internal/domain/errors/codes.go` with central registry
   **When** I create a domain error with `errors.NewDomain("NOT_FOUND", msg)`
   **Then** the error implements the standard error interface
   **And** the code is UPPER_SNAKE format

2. **Given** the DomainError type
   **When** I use `errors.Is()` or `errors.As()` to check error types
   **Then** it correctly identifies the error type
   **And** I can extract the public code from the error

3. **Given** an HTTP handler receives a domain error
   **When** the error is mapped to an API response
   **Then** `error.code` matches the domain error's public code
   **And** the error follows Envelope format from Story 2.1

4. **Given** a domain-specific error like `note.ErrNoteNotFound`
   **When** the error is created using the central registry
   **Then** it uses appropriate central code (e.g., `NOT_FOUND`)
   **And** maintains domain-specific context in the message

## Tasks / Subtasks

- [x] Task 1: Create central error code registry (AC: #1)
  - [x] 1.1 Create `internal/domain/errors/` package directory
  - [x] 1.2 Create `codes.go` with central error code constants (UPPER_SNAKE format)
  - [x] 1.3 Define base codes: `NOT_FOUND`, `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `CONFLICT`, `INTERNAL_ERROR`, `TIMEOUT` (plus `RATE_LIMIT_EXCEEDED` and `BAD_REQUEST`)
  - [x] 1.4 Add package documentation explaining code naming conventions

- [x] Task 2: Create DomainError type (AC: #1, #2)
  - [x] 2.1 Create `domain_error.go` with `DomainError` struct containing `Code`, `Message`, optional `Hint`
  - [x] 2.2 Implement `error` interface (`Error() string` method)
  - [x] 2.3 Implement `Unwrap() error` for error chain support
  - [x] 2.4 Create `NewDomain(code, message string) *DomainError` constructor
  - [x] 2.5 Create `NewDomainWithHint(code, message, hint string) *DomainError` constructor (plus `NewDomainWithCause`)
  - [x] 2.6 Ensure `errors.Is()` and `errors.As()` work correctly (implemented custom `Is()` method)

- [x] Task 3: Refactor existing domain errors (AC: #4)
  - [x] 3.1 Update `internal/domain/errors.go` to use new DomainError type
  - [x] 3.2 Deprecate old sentinel errors with backward compatibility
  - [x] 3.3 Update `internal/domain/note/errors.go` to use new pattern
  - [x] 3.4 Ensure all error codes follow UPPER_SNAKE convention

- [x] Task 4: Integrate with response package (AC: #3)
  - [x] 4.1 Update `internal/interface/http/response/mapper.go` to detect DomainError
  - [x] 4.2 Extract `Code` from DomainError for `error.code` field
  - [x] 4.3 Use `envelope.go` functions from Story 2.1 for response formatting
  - [x] 4.4 Add fallback for non-DomainError errors (use INTERNAL_ERROR)

- [x] Task 5: Add comprehensive tests (AC: All)
  - [x] 5.1 Create `codes_test.go` verifying all codes are UPPER_SNAKE
  - [x] 5.2 Create `domain_error_test.go` testing error interface, Is(), As()
  - [x] 5.3 Update `mapper_test.go` for DomainError mapping
  - [x] 5.4 Verify backward compatibility with existing code

- [x] Task 6: Documentation and verification
  - [x] 6.1 Add package documentation to `internal/domain/errors/`
  - [x] 6.2 Update `project_context.md` if needed (verified already correct)
  - [x] 6.3 Run `make verify` - all lint + tests must pass

## Dev Notes

### Architecture Decision Reference

**Decision 1: Error Code Registry (Hybrid)** from `docs/architecture-decisions.md`:
- Central registry: `internal/domain/errors/codes.go` for public/stable codes
- Domain-specific: Each domain has own errors with central codes
- Public codes: UPPER_SNAKE without prefix (e.g., `NOT_FOUND`, not `ERR_NOT_FOUND`)

**Target Implementation Pattern:**
```go
// internal/domain/errors/codes.go (public registry)
const (
    CodeNotFound       = "NOT_FOUND"
    CodeValidation     = "VALIDATION_ERROR"
    CodeUnauthorized   = "UNAUTHORIZED"
    CodeForbidden      = "FORBIDDEN"
    CodeConflict       = "CONFLICT"
    CodeInternal       = "INTERNAL_ERROR"
    CodeTimeout        = "TIMEOUT"
    CodeRateLimit      = "RATE_LIMIT_EXCEEDED"
)

// internal/domain/errors/domain_error.go
type DomainError struct {
    Code    string
    Message string
    Hint    string // Optional hint for API clients (use with caution - security!)
    cause   error  // For error chaining
}

func (e *DomainError) Error() string { return e.Message }
func (e *DomainError) Unwrap() error { return e.cause }

func NewDomain(code, message string) *DomainError {
    return &DomainError{Code: code, Message: message}
}

// Usage in domain package
// internal/domain/note/errors.go
var ErrNoteNotFound = errors.NewDomain(codes.CodeNotFound, "note not found")
```

### Existing State Analysis

**Current `internal/domain/errors.go`:**
- Uses sentinel errors with `errors.New()`
- No public codes attached
- Uses `WrapError()` for context

**Current `internal/domain/note/errors.go`:**
- Domain-specific sentinel errors
- No connection to central registry
- No public codes

**Current `internal/interface/http/response/errors.go`:**
- Has `Code*` constants (UPPER_SNAKE) from Story 2.1
- These are HTTP-level codes, need domain-level counterparts

### Migration Path

1. **Phase 1 (This Story):** Create central registry + DomainError type
2. **Phase 2 (Future):** Migrate existing errors to use new pattern
3. **Phase 3 (Future):** Remove deprecated sentinel errors

### Critical Points from Story 2.1

From previous story learnings:
- Use UPPER_SNAKE format WITHOUT `ERR_` prefix (e.g., `NOT_FOUND` not `ERR_NOT_FOUND`)
- Hint field must include security warning (never expose internal details)
- All response helpers use `ctxutil.RequestIDFromContext(ctx)`
- Test coverage is critical - CI enforces lint+test

### File Structure

```
internal/domain/
├── errors/                    # [NEW] Central error package
│   ├── codes.go               # [NEW] Public error code constants
│   ├── codes_test.go          # [NEW] Code format validation tests
│   ├── domain_error.go        # [NEW] DomainError type
│   └── domain_error_test.go   # [NEW] DomainError tests
├── errors.go                  # [REFACTOR] Use DomainError, deprecate old
├── note/
│   └── errors.go              # [REFACTOR] Use central codes
```

### Layer Boundary Rules

Per `project_context.md`:
- `domain` → (nothing) - no external dependencies
- DomainError must stay within domain layer
- Response mapper in interface layer extracts codes

### Testing Strategy

1. **Unit Tests:**
   - All code constants are UPPER_SNAKE
   - `DomainError.Error()` returns message
   - `errors.Is(err, &DomainError{})` works
   - `errors.As(err, &DomainError{})` extracts code

2. **Integration with Response:**
   - DomainError maps to correct Envelope format
   - `error.code` matches domain error code

### NFR Targets

| NFR | Requirement | Verification |
|-----|-------------|--------------|
| FR16 | error.code uses public UPPER_SNAKE codes | Code format test |
| NFR-M1 | Coverage ≥80% for domain | Test coverage |

### Dependencies

- **Story 2.1 (Done):** Provides Envelope format, UPPER_SNAKE codes in response layer
- **Story 2.4 (Future):** Will use this registry for HTTP error mapping middleware

### Critical Points

1. **No breaking changes:** Maintain backward compatibility with existing sentinel errors
2. **Layer boundaries:** DomainError stays in domain layer, code extraction in interface layer
3. **UPPER_SNAKE codes:** Strictly follow format without `ERR_` prefix
4. **Hint security:** Warning about not exposing internal details
5. **Testing:** Comprehensive tests for all new types

### References

- [Source: docs/epics.md#Story 2.2](file:///docs/epics.md) - FR16
- [Source: docs/architecture-decisions.md#Decision 1](file:///docs/architecture-decisions.md) - Hybrid error registry
- [Source: project_context.md](file:///project_context.md) - UPPER_SNAKE codes, layer boundaries
- [Source: internal/domain/errors.go](file:///internal/domain/errors.go) - Current sentinel errors
- [Source: internal/domain/note/errors.go](file:///internal/domain/note/errors.go) - Domain-specific errors
- [Source: internal/interface/http/response/errors.go](file:///internal/interface/http/response/errors.go) - HTTP codes from Story 2.1
- [Source: docs/sprint-artifacts/2-1-implement-response-envelope-package.md](file:///docs/sprint-artifacts/2-1-implement-response-envelope-package.md) - Previous story learnings

## Dev Agent Record

### Context Reference

- Story generated by create-story workflow on 2025-12-15
- Epic 2: API Standards & Response Contract (MVP) - in-progress
- Previous story: Story 2.1 (Implement Response Envelope Package) - done

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

None required.

### Completion Notes List

- Created `internal/domain/errors/` package with central error code registry
- Implemented 9 error codes in UPPER_SNAKE format: `NOT_FOUND`, `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `CONFLICT`, `INTERNAL_ERROR`, `TIMEOUT`, `RATE_LIMIT_EXCEEDED`, `BAD_REQUEST`
- Created `DomainError` type with full support for `errors.Is()`, `errors.As()`, and error chaining via `Unwrap()`
- Added `NewDomain()`, `NewDomainWithHint()`, and `NewDomainWithCause()` constructors
- Added `IsDomainError()` helper function for easy type checking
- Updated `internal/domain/errors.go` with `NewErr*` factory functions and deprecated old sentinel errors
- Updated `internal/domain/note/errors.go` with new error factory functions
- Updated `mapper.go` to detect DomainError first, extract Code, and fallback to legacy errors
- Added `MapErrorWithHint()` and `HandleErrorCtx()` functions for enhanced error handling
- Added deprecation warning exclusion in `policy/golangci.yml` for migration period
- All 21 domain/errors tests pass
- All 46 response package tests pass
- `make verify` passes (lint + all unit tests)

### File List

**New Files:**
- `internal/domain/errors/codes.go` - Central error code constants
- `internal/domain/errors/codes_test.go` - Code format validation tests
- `internal/domain/errors/domain_error.go` - DomainError type
- `internal/domain/errors/domain_error_test.go` - DomainError tests

**Modified Files:**
- `internal/domain/errors.go` - Added NewErr* functions, deprecated old sentinel errors
- `internal/domain/note/errors.go` - Added NewErr* functions, deprecated old sentinel errors
- `internal/interface/http/response/mapper.go` - Added DomainError detection and mapping
- `internal/interface/http/response/mapper_test.go` - Added DomainError mapping tests
- `policy/golangci.yml` - Added deprecation warning exclusion for migration period
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

### Code Review Fixes (2025-12-15)

- **Hardened Error Registry**: Added `allCodes` registry map in `codes.go` and enforced validation in `NewDomain` factory functions (panics on invalid code).
- **Refactored Tests**: Updated `codes_test.go` to iterate over the registry for robust efficient testing, and added `TestNewDomain_InvalidCode` to verify runtime strictness.
- **Improved Observability**: Added CRITICAL logging in `mapper.go` for unmapped domain error codes to prevent silent failures.
- **Documentation Cleanup**: Removed HTTP layer implementation details from domain layer documentation (`codes.go`) to strictly enforce layer boundaries.

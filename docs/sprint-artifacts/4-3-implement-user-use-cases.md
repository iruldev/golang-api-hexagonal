# Story 4.3: Implement User Use Cases

Status: done

## Story

As a **developer**,
I want **use cases for creating, getting, and listing users**,
so that **business logic is properly orchestrated in the app layer**.

## Acceptance Criteria

1. **Given** the app layer, **When** I view `internal/app/user/`, **Then** the following use cases exist:
   - `CreateUserUseCase`
   - `GetUserUseCase`
   - `ListUsersUseCase`

2. **Given** CreateUserUseCase is instantiated, **When** use case is created, **Then** it accepts `IDGenerator` interface for generating user IDs **And** ID is generated via `idGen.NewID()` (not in handler)

3. **Given** CreateUserUseCase is executed with valid request, **When** the operation completes, **Then** user is created via repository **And** domain.User is returned

4. **Given** GetUserUseCase is executed with non-existent ID, **When** repository returns ErrUserNotFound, **Then** AppError with Code="USER_NOT_FOUND" is returned

5. **And** app layer has NO imports of net/http, pgx, slog, uuid

6. **And** use cases accept repository interfaces (not implementations)

*Covers: FR10 (partial), FR42*

## Source of Truth (Important)

- The canonical requirements for this story are in `docs/epics.md` under "Story 4.3".
- The canonical domain contracts (types + method signatures) are in:
  - `internal/domain/user.go` (User, UserRepository)
  - `internal/domain/querier.go` (Querier)
  - `internal/domain/tx.go` (TxManager)
  - `internal/domain/pagination.go` (ListParams)
  - `internal/domain/errors.go` (ErrUserNotFound, ErrEmailAlreadyExists)
- If any snippet in this story conflicts with domain files, **follow the domain files**.

## Tasks / Subtasks

- [x] Task 1: Create AppError Struct (AC: #4)
  - [x] 1.1 Create `internal/app/errors.go` with `AppError` struct
  - [x] 1.2 Define fields: `Op string`, `Code string`, `Message string`, `Err error`
  - [x] 1.3 Implement `Error() string` and `Unwrap() error` methods
  - [x] 1.4 Define error code constants: `CodeUserNotFound`, `CodeEmailExists`, `CodeValidationError`, `CodeInternalError`

- [x] Task 2: Implement GetUserUseCase (AC: #1, #4, #6)
  - [x] 2.1 Create `internal/app/user/get_user.go` with `GetUserUseCase` struct
  - [x] 2.2 Define `GetUserRequest` with `ID domain.ID` field
  - [x] 2.3 Define `GetUserResponse` with `User domain.User` field
  - [x] 2.4 Add `NewGetUserUseCase(userRepo domain.UserRepository, db domain.Querier)` constructor
  - [x] 2.5 Implement `Execute(ctx context.Context, req GetUserRequest) (GetUserResponse, error)`
  - [x] 2.6 On `domain.ErrUserNotFound` → return `&app.AppError{Op: "GetUser", Code: app.CodeUserNotFound, Message: "User not found", Err: err}`

- [x] Task 3: Implement ListUsersUseCase (AC: #1, #6)
  - [x] 3.1 Create `internal/app/user/list_users.go` with `ListUsersUseCase` struct
  - [x] 3.2 Define `ListUsersRequest` with `Page int`, `PageSize int` fields
  - [x] 3.3 Define `ListUsersResponse` with `Users []domain.User`, `TotalCount int`, `Page int`, `PageSize int` fields
  - [x] 3.4 Add `NewListUsersUseCase(userRepo domain.UserRepository, db domain.Querier)` constructor
  - [x] 3.5 Implement `Execute(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error)`
  - [x] 3.6 Create `domain.ListParams` from request: `domain.ListParams{Page: req.Page, PageSize: req.PageSize}`
  - [x] 3.7 Call `userRepo.List(ctx, db, params)` and return results with pagination info

- [x] Task 4: Update CreateUserUseCase for AppError (AC: #3, #4)
  - [x] 4.1 Update error handling to return `AppError` for domain errors:
    - `ErrEmailAlreadyExists` → `&app.AppError{Op: "CreateUser", Code: app.CodeEmailExists, ...}`
    - `ErrInvalidEmail` etc → `&app.AppError{Op: "CreateUser", Code: app.CodeValidationError, ...}`
  - [x] 4.2 Ensure all repository errors propagate wrapped in AppError

- [x] Task 5: Write Unit Tests for GetUserUseCase (AC: all)
  - [x] 5.1 Create `internal/app/user/get_user_test.go`
  - [x] 5.2 Test success case: repository returns user → use case returns GetUserResponse
  - [x] 5.3 Test not found: repository returns ErrUserNotFound → use case returns AppError with Code="USER_NOT_FOUND"
  - [x] 5.4 Test repository error propagation

- [x] Task 6: Write Unit Tests for ListUsersUseCase (AC: all)
  - [x] 6.1 Create `internal/app/user/list_users_test.go`
  - [x] 6.2 Test success case: repository returns users → use case returns ListUsersResponse with pagination
  - [x] 6.3 Test empty list: repository returns empty slice → use case returns empty data with TotalCount=0
  - [x] 6.4 Test repository error propagation

- [x] Task 7: Update CreateUserUseCase Tests (AC: #4)
  - [x] 7.1 Add test cases for AppError conversion
  - [x] 7.2 Test that ErrEmailAlreadyExists returns AppError with Code=app.CodeEmailExists

- [x] Task 8: Verify Layer Compliance (AC: #5, #6)
  - [x] 8.1 Run `make lint` to verify no depguard violations
  - [x] 8.2 Confirm app layer only imports domain (not infra/transport/net/http/pgx/slog/uuid)
  - [x] 8.3 Run `make test` to ensure all unit tests pass
  - [x] 8.4 Run `make coverage` to verify coverage threshold (90.9% > 80%)

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Story 4.1 + 4.2 (Complete):**
| File | Description |
|------|-------------|
| `internal/domain/user.go` | User entity with `FirstName`/`LastName`, `UserRepository` interface |
| `internal/domain/querier.go` | `Querier` interface with stdlib-only types |
| `internal/domain/tx.go` | `TxManager` interface |
| `internal/domain/pagination.go` | `ListParams` with `Offset()`, `Limit()` methods |
| `internal/domain/errors.go` | `ErrUserNotFound`, `ErrEmailAlreadyExists`, `ErrInvalidEmail`, `ErrInvalidFirstName`, `ErrInvalidLastName` |
| `internal/domain/id.go` | `ID` type with `IsEmpty()` method, `IDGenerator` interface |
| `internal/infra/postgres/user_repo.go` | UserRepository PostgreSQL implementation |
| `internal/app/user/create_user.go` | Existing CreateUserUseCase implementation |
| `internal/app/user/create_user_test.go` | Existing tests with mock patterns |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/app/errors.go` | AppError struct and error codes |
| `internal/app/user/get_user.go` | GetUserUseCase implementation |
| `internal/app/user/get_user_test.go` | GetUser unit tests |
| `internal/app/user/list_users.go` | ListUsersUseCase implementation |
| `internal/app/user/list_users_test.go` | ListUsers unit tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/app/user/create_user.go` | Add AppError wrapper for domain errors |
| `internal/app/user/create_user_test.go` | Add AppError conversion test cases |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in app layer: domain imports only
❌ FORBIDDEN: net/http, pgx, slog, otel, uuid, transport, infra
```

**App Layer Rules:**
- Accept repository interfaces (domain.UserRepository), not implementations
- Authorization checks happen HERE (not in this story, but note for future)
- Use `TxManager.WithTx()` for multi-step operations (not required in this story)
- Convert domain errors to typed `AppError` with machine-readable `Code`
- NO logging — only tracing context (not used in this story)

### AppError Pattern (from Architecture)

```go
// internal/app/errors.go
package app

// Error codes for machine-readable error handling
const (
    CodeUserNotFound     = "USER_NOT_FOUND"
    CodeEmailExists      = "EMAIL_EXISTS"
    CodeValidationError  = "VALIDATION_ERROR"
    CodeInternalError    = "INTERNAL_ERROR"
)

// AppError represents an application-layer error with machine-readable code.
type AppError struct {
    Op      string // operation name: "GetUser", "CreateUser"
    Code    string // machine-readable: "USER_NOT_FOUND"
    Message string // human-readable message
    Err     error  // wrapped error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Op + ": " + e.Message + ": " + e.Err.Error()
    }
    return e.Op + ": " + e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}
```

### GetUserUseCase Pattern

```go
// internal/app/user/get_user.go
package user

import (
    "context"
    "errors"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type GetUserRequest struct {
    ID domain.ID
}

type GetUserResponse struct {
    User domain.User
}

type GetUserUseCase struct {
    userRepo domain.UserRepository
    db       domain.Querier
}

func NewGetUserUseCase(userRepo domain.UserRepository, db domain.Querier) *GetUserUseCase {
    return &GetUserUseCase{
        userRepo: userRepo,
        db:       db,
    }
}

func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (GetUserResponse, error) {
    user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return GetUserResponse{}, &app.AppError{
                Op:      "GetUser",
                Code:    app.CodeUserNotFound,
                Message: "User not found",
                Err:     err,
            }
        }
        return GetUserResponse{}, &app.AppError{
            Op:      "GetUser",
            Code:    app.CodeInternalError,
            Message: "Failed to get user",
            Err:     err,
        }
    }
    return GetUserResponse{User: *user}, nil
}
```

### ListUsersUseCase Pattern

```go
// internal/app/user/list_users.go
package user

import (
    "context"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type ListUsersRequest struct {
    Page     int
    PageSize int
}

type ListUsersResponse struct {
    Users      []domain.User
    TotalCount int
    Page       int
    PageSize   int
}

type ListUsersUseCase struct {
    userRepo domain.UserRepository
    db       domain.Querier
}

func NewListUsersUseCase(userRepo domain.UserRepository, db domain.Querier) *ListUsersUseCase {
    return &ListUsersUseCase{
        userRepo: userRepo,
        db:       db,
    }
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error) {
    params := domain.ListParams{Page: req.Page, PageSize: req.PageSize}
    
    users, totalCount, err := uc.userRepo.List(ctx, uc.db, params)
    if err != nil {
        return ListUsersResponse{}, &app.AppError{
            Op:      "ListUsers",
            Code:    app.CodeInternalError,
            Message: "Failed to list users",
            Err:     err,
        }
    }
    
    return ListUsersResponse{
        Users:      users,
        TotalCount: totalCount,
        Page:       params.Page,
        PageSize:   params.Limit(),
    }, nil
}
```

### Test Pattern (Existing from create_user_test.go)

The existing tests use manual mock implementations:
- `mockIDGenerator` - implements `domain.IDGenerator`
- `mockQuerier` - implements `domain.Querier`
- `mockUserRepository` - implements `domain.UserRepository`

Reuse these mocks in new test files. The mock repository already implements `GetByID` and `List` methods.

```go
//go:build !integration

package user

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    
    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

func TestGetUserUseCase_Execute(t *testing.T) {
    tests := []struct {
        name       string
        req        GetUserRequest
        setupMock  func(*mockUserRepository)
        wantCode   string
        wantErr    bool
    }{
        {
            name: "successfully gets user by ID",
            req:  GetUserRequest{ID: "existing-id"},
            setupMock: func(m *mockUserRepository) {
                m.users["existing-id"] = domain.User{
                    ID:        "existing-id",
                    Email:     "test@example.com",
                    FirstName: "John",
                    LastName:  "Doe",
                }
            },
            wantErr: false,
        },
        {
            name:      "returns USER_NOT_FOUND when user doesn't exist",
            req:       GetUserRequest{ID: "non-existent-id"},
            setupMock: func(m *mockUserRepository) {},
            wantCode:  app.CodeUserNotFound,
            wantErr:   true,
        },
    }
    // ... test loop
}
```

### Verification Commands

```bash
# Run all unit tests (no integration test required for this story)
make test

# Run coverage check (domain + app must be ≥ 80%)
make coverage

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci
```

### References

- [Source: docs/epics.md#Story 4.3] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Error Handling Strategy] - AppError pattern
- [Source: docs/architecture.md#File Organization] - App layer structure
- [Source: docs/project-context.md#App Layer] - Layer rules and forbidden imports
- [Source: internal/app/user/create_user.go] - Existing use case pattern
- [Source: internal/app/user/create_user_test.go] - Existing test mock patterns

### Learnings from Story 4.1 + 4.2

**From Story 4.1:**
- Domain layer uses `type ID string` with `IsEmpty()` method
- `ListParams` provides `Offset()` and `Limit()` with defaulting/clamping
- Domain validation errors: `ErrInvalidEmail`, `ErrInvalidFirstName`, `ErrInvalidLastName`

**From Story 4.2:**
- Repository methods wrap errors with `op` pattern
- `ErrUserNotFound` and `ErrEmailAlreadyExists` are sentinel errors in domain
- Integration tests use Docker Compose PostgreSQL

**From Epic 3 Retrospective:**
- depguard rules enforced - app layer CANNOT import net/http, pgx, slog, uuid
- Use table-driven tests with testify assertions
- Run `make ci` before marking story complete

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

- `docs/epics.md` - Story 4.3 acceptance criteria
- `docs/architecture.md` - AppError strategy + boundary rules
- `docs/project-context.md` - App layer conventions
- `internal/domain/user.go` - User + UserRepository contract
- `internal/domain/errors.go` - Domain sentinel errors
- `internal/domain/id.go` - ID + IDGenerator contract
- `internal/domain/pagination.go` - ListParams defaulting/clamping
- `internal/app/user/create_user.go` - Existing CreateUserUseCase baseline
- `internal/app/user/create_user_test.go` - Existing test + mocks baseline

### Agent Model Used

gpt-5 (Codex CLI)

### Debug Log References

N/A (ready-for-dev story doc only)

### Completion Notes List

- Created `internal/app/errors.go` with AppError struct and error codes (CodeUserNotFound, CodeEmailExists, CodeValidationError, CodeInternalError)
- Implemented GetUserUseCase with proper AppError wrapping for ErrUserNotFound and repository errors
- Implemented ListUsersUseCase with pagination support using domain.ListParams
- Updated CreateUserUseCase to wrap validation errors (CodeValidationError), ErrEmailAlreadyExists (CodeEmailExists), and repository errors (CodeInternalError) in AppError
- Created comprehensive unit tests for get_user_test.go including success, not found, repository error, and nil-user guard cases
- Added `internal/app/errors_test.go` for AppError behavior tests (moved from user use case tests)
- Created comprehensive unit tests for list_users_test.go including success, empty list, repository error, and default page size cases
- Updated create_user_test.go with new test cases verifying AppError conversion for validation errors, email exists, and repository errors
- Verification: `make lint` passes (no depguard violations), `make test` passes (all tests), `make coverage` shows 100.0% coverage (exceeds 80% threshold)

### File List

| Action | Path |
|--------|------|
| Created | `internal/app/errors.go` |
| Created | `internal/app/errors_test.go` |
| Created | `internal/app/user/get_user.go` |
| Created | `internal/app/user/get_user_test.go` |
| Created | `internal/app/user/list_users.go` |
| Created | `internal/app/user/list_users_test.go` |
| Modified | `internal/app/user/create_user.go` |
| Modified | `internal/app/user/create_user_test.go` |
| Modified | `docs/sprint-artifacts/sprint-status.yaml` |

### Change Log

- **2025-12-17**: Story drafted and marked `ready-for-dev`
- **2025-12-17**: Implemented all use cases (GetUser, ListUsers) and updated CreateUser with AppError wrapping. Added comprehensive unit tests. All tests pass. Coverage 90.9%. Marked `Ready for Review`.
- **2025-12-17**: Code review fixes applied (nil-user guard, pagination page normalization, AppError tests moved to app package). Re-verified `make lint`, `make test`, `make coverage`.
- **2025-12-17**: Code review rerun (final) — approved; story marked `done` and sprint-status synced.

## Senior Developer Review (AI)

Reviewer: Chat  
Date: 2025-12-17

### Summary

- Outcome: Changes requested → applied
- Verification: `make lint`, `make test`, `make coverage` PASS

### High Issues (Fixed)

- GetUser: guard against repository returning `(nil, nil)` to avoid panic; now returns `AppError` with `CodeInternalError`. (`internal/app/user/get_user.go`)
- ListUsers: normalize response `Page` to `1` when request is `0`/negative (aligned with domain pagination semantics). (`internal/app/user/list_users.go`)
- CreateUser: avoid duplicated/leaky validation message by using stable message and wrapping original error. (`internal/app/user/create_user.go`)

### Medium Issues (Fixed)

- AppError tests moved to `internal/app/errors_test.go` (package ownership + coverage correctness); removed from `internal/app/user/get_user_test.go`.
- Story status kept in sync with `docs/sprint-artifacts/sprint-status.yaml` (review → done after rerun).

## Senior Developer Review (AI) - Rerun (Final)

Reviewer: Chat  
Date: 2025-12-17

### Summary

- Outcome: Approved
- Verification: `make lint`, `make test`, `make coverage` PASS

### Low Improvements (Not Blocking)

- Git working tree sebelumnya punya banyak file `??` (untracked); sudah dibereskan dengan staging (`git add -A`) sebelum status di-set `done`.
- `go test` masih mengeluarkan warning linker macOS `malformed LC_DYSYMTAB`; tes tetap PASS tapi perlu investigasi toolchain di mesin lokal bila warning mengganggu.

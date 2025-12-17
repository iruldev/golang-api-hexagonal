# Story 4.1: Implement User Domain Entity and Repository Interface

Status: done

## Story

As a **developer**,
I want **User entity and repository interface in the domain layer**,
So that **I have a clear contract for user data access that follows hexagonal architecture principles**.

## Acceptance Criteria

1. **Given** the domain layer, **When** I view `internal/domain/user.go`, **Then** User entity exists with fields:
   - `ID` (type `ID` which is `type ID string`)
   - `Email` (string)
   - `FirstName` (string)
   - `LastName` (string)
   - `CreatedAt` (time.Time)
   - `UpdatedAt` (time.Time)

2. **And** `UserRepository` interface is defined with methods:
   - `Create(ctx context.Context, q Querier, user *User) error`
   - `GetByID(ctx context.Context, q Querier, id ID) (*User, error)`
   - `List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)`

3. **And** domain errors are defined:
   - `var ErrUserNotFound = errors.New("user not found")`
   - `var ErrEmailAlreadyExists = errors.New("email already exists")`

4. **And** `IDGenerator` interface is defined:
   - `type IDGenerator interface { NewID() ID }`

5. **And** domain layer has NO external imports (stdlib only)

## Source of Truth (Important)

- This story's signatures and names are the canonical contract for Epic 4 (Users module).
- If any older docs/snippets (including `docs/architecture.md`) show different signatures for `UserRepository` methods, follow this story.

*Covers: FR10 (partial), FR41*

## Tasks / Subtasks

- [x] Task 1: Extend User Entity (AC: #1)
  - [x] 1.1 Modify `internal/domain/user.go` to update User struct with `FirstName`, `LastName` fields (rename `Name` to split fields)
  - [x] 1.2 Ensure User entity uses `ID` type from `id.go`
  - [x] 1.3 Verify all fields have proper time.Time types for timestamps

- [x] Task 2: Update UserRepository Interface (AC: #2)
  - [x] 2.1 Modify UserRepository interface to accept `Querier` parameter in all methods
  - [x] 2.2 Add `List` method with `ListParams` and return `([]User, int, error)` for pagination support
  - [x] 2.3 Ensure interface uses pointer `*User` for consistency
  - [x] 2.4 Remove legacy methods not in AC (#2) (`GetByEmail`, `Update`, `Delete`, etc.) to keep the contract minimal and unambiguous

- [x] Task 3: Add ListParams Type (AC: #2)
  - [x] 3.1 Create `internal/domain/pagination.go` with `ListParams` (`Page`, `PageSize`)
  - [x] 3.2 Add default handling: `DefaultPageSize = 20` and clamp invalid values
  - [x] 3.3 Define paging semantics: `Page` is 1-based, `PageSize` defaults to `DefaultPageSize`, and `List(...)` returns `totalItems` as the second return value (total matching rows, not just returned page size)

- [x] Task 4: Create Querier Interface (AC: #2)
  - [x] 4.1 Create `internal/domain/querier.go` with `Querier` interface
  - [x] 4.2 Define methods: `Exec`, `Query`, `QueryRow` with stdlib-only types
  - [x] 4.3 Keep interface stdlib-only (avoid pgx/pgconn types in signatures)

- [x] Task 5: Create TxManager Interface (AC: #2)
  - [x] 5.1 Create `internal/domain/tx.go` with `TxManager` interface
  - [x] 5.2 Define `WithTx(ctx context.Context, fn func(tx Querier) error) error`

- [x] Task 6: Add IDGenerator Interface (AC: #4)
  - [x] 6.1 Add `IDGenerator` interface to `internal/domain/id.go`
  - [x] 6.2 Define single method: `NewID() ID`

- [x] Task 7: Update Domain Errors (AC: #3)
  - [x] 7.1 Verify `ErrUserNotFound` exists (already present)
  - [x] 7.2 Rename `ErrEmailExists` to `ErrEmailAlreadyExists` for consistency with acceptance criteria
  - [x] 7.3 Keep `ErrInvalidEmail` for validation
  - [x] 7.4 Replace `ErrInvalidUserName` with name-field-specific errors:
    - `var ErrInvalidFirstName = errors.New("invalid first name")`
    - `var ErrInvalidLastName = errors.New("invalid last name")`

- [x] Task 8: Update User Validation Method
  - [x] 8.1 Update `Validate()` method to check `FirstName` and `LastName` (instead of `Name`)
  - [x] 8.2 Return `ErrInvalidFirstName` / `ErrInvalidLastName` for empty/whitespace fields (email still validated first)

- [x] Task 9: Update Unit Tests
  - [x] 9.1 Update `internal/domain/user_test.go` to test new field structure
  - [x] 9.2 Ensure all tests pass with race detection
  - [x] 9.3 Ensure domain/app coverage gate passes (target ≥ 80% per NFRs)

- [x] Task 10: Verify Layer Compliance (AC: #5)
  - [x] 10.1 Run `make lint` to verify no depguard violations
  - [x] 10.2 Confirm domain layer has only stdlib imports

## Dev Notes

### ⚠️ CRITICAL: Existing Code Must Be Modified

**Story 3.2 created sample domain files.** This story EXTENDS them:

| File | Current State | Required Changes |
|------|---------------|------------------|
| `internal/domain/user.go` | Sample entity with `Name` field | Split to `FirstName`/`LastName`, update `UserRepository` signature |
| `internal/domain/id.go` | `type ID string` ✅ | Add `IDGenerator` interface |
| `internal/domain/errors.go` | Has `ErrUserNotFound`, `ErrEmailExists` | Rename to `ErrEmailAlreadyExists` |
| `internal/domain/querier.go` | **Does not exist** | Create new file |
| `internal/domain/tx.go` | **Does not exist** | Create new file |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in domain layer: stdlib only (errors, context, time, strings, fmt)
❌ FORBIDDEN: slog, uuid, pgx, chi, otel, ANY external package
```

**Note on tests:** `internal/domain/**/*_test.go` may use `github.com/stretchr/testify` (allowed by depguard), but production domain files must remain stdlib-only.

**Domain Purity Rules:**
- Use `type ID string` for identifiers — UUID parsing happens at infra boundary
- Define sentinel errors only (no typed errors with HTTP status)
- Interfaces define contracts, implementations live in infra layer
- NO logging in domain — ever

### Querier Interface Design

The `Querier` interface abstracts database operations to work with both connection pool and transactions. **Keep it stdlib-only:**

```go
// internal/domain/querier.go
package domain

import "context"

// Querier is an abstraction for database operations that works with both
// connection pools and transactions. Implementations convert between
// this interface and driver-specific types.
type Querier interface {
    Exec(ctx context.Context, sql string, args ...any) (any, error)
    Query(ctx context.Context, sql string, args ...any) (any, error)
    QueryRow(ctx context.Context, sql string, args ...any) any
}
```

**Note:** The `any` return types allow infra layer to return driver-specific types. Domain layer doesn't need to know the concrete types.

### TxManager Interface Design

```go
// internal/domain/tx.go
package domain

import "context"

// TxManager provides transaction management for use cases that need
// atomicity across multiple repository operations.
type TxManager interface {
    WithTx(ctx context.Context, fn func(tx Querier) error) error
}
```

### UserRepository Interface (Updated)

```go
// internal/domain/user.go
type UserRepository interface {
    Create(ctx context.Context, q Querier, user *User) error
    GetByID(ctx context.Context, q Querier, id ID) (*User, error)
    List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
}
```

### ListParams Design

```go
// internal/domain/pagination.go
package domain

const DefaultPageSize = 20
const MaxPageSize = 100

type ListParams struct {
    Page     int
    PageSize int
}

func (p ListParams) Offset() int {
    page := p.Page
    if page <= 0 {
        page = 1
    }
    return (page - 1) * p.Limit()
}

func (p ListParams) Limit() int {
    if p.PageSize <= 0 {
        return DefaultPageSize
    }
    if p.PageSize > MaxPageSize {
        return MaxPageSize
    }
    return p.PageSize
}
```

**Paging contract:** `List(...)` returns `(items, totalItems, err)` where `totalItems` is the total number of matching rows (for calculating total pages). For MVP, ordering will be defined in the infra repo (see Story 4.2).

### IDGenerator Interface

```go
// Add to internal/domain/id.go
type IDGenerator interface {
    NewID() ID
}
```

**Usage:** App-layer use cases will call `idGen.NewID()` for new entities; the composition root provides a UUID v7-based implementation (and tests can use deterministic fakes).

### Transaction Note (TxManager)

`TxManager` is defined in this story for transaction boundaries, but a single repository call may still use a plain `Querier` directly.
When Story 4.3 introduces multi-step operations (e.g., create user + audit in one unit of work), prefer `TxManager.WithTx()` in the app layer.

### Git Hygiene (Review/Done)

Before setting story status to `done`, ensure the working tree is clean and all new/modified files are staged and committed, so the reviewed state matches what will be pushed/PR’d.

### Project Structure Notes

```
internal/domain/
├── id.go           # ID type + IDGenerator interface
├── id_test.go      # Tests for ID
├── errors.go       # Sentinel errors
├── user.go         # User entity + UserRepository interface
├── user_test.go    # Tests for User
├── querier.go      # NEW: Querier interface (DB abstraction)
├── tx.go           # NEW: TxManager interface
└── pagination.go   # NEW: ListParams struct
```

### References

- [Source: docs/epics.md#Story 4.1] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Repository Pattern] - Querier abstraction design
- [Source: docs/architecture.md#Transaction Handling] - TxManager pattern
- [Source: docs/project-context.md#Domain Layer] - Layer rules and forbidden imports
- [Source: docs/sprint-artifacts/epic-3-retro-2025-12-17.md] - Domain entity patterns from Story 3.2

### Learnings from Previous Stories

**From Epic 3 Retrospective:**
- Domain entity pattern established in Story 3.2 (`user.go`, `id.go`, `errors.go`)
- depguard rules for production vs test files are configured
- Use `type ID string` not `uuid.UUID` in domain layer
- Test files can import testify (configured in `.golangci.yml`)

**From Story 3.2:**
- Sample User entity exists with `Name` field → needs split to `FirstName`/`LastName`
- UserRepository interface exists but signatures differ from Epic 4 requirements
- Unit tests exist and follow table-driven pattern

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude 3.7 Sonnet

### Debug Log References

### Completion Notes List

- ✅ Task 1-8: Extended User entity with `FirstName`/`LastName`, updated `UserRepository` with `Querier` parameter, created `ListParams`, `Querier`, `TxManager`, `IDGenerator` interfaces, updated errors and validation
- ✅ Task 9: Updated `user_test.go` and created `pagination_test.go` - all tests pass; coverage gate passes (≥80% required, current: 100% for domain+app in this run)
- ✅ Task 10: Verified `make lint` passes with no depguard violations - domain layer is stdlib-only
- ✅ Updated app layer (`create_user.go`, `create_user_test.go`) to use new domain signatures
- ✅ Code review fixes applied: synced `docs/architecture.md` and `docs/project-context.md`, improved pagination clamping, simplified test helper; rerun code review report saved

### File List

- `internal/domain/user.go` - Modified: User entity with FirstName/LastName, updated UserRepository interface
- `internal/domain/id.go` - Modified: Added IDGenerator interface
- `internal/domain/errors.go` - Modified: Renamed/added errors (ErrEmailAlreadyExists, ErrInvalidFirstName, ErrInvalidLastName)
- `internal/domain/querier.go` - New: Querier interface for DB abstraction
- `internal/domain/tx.go` - New: TxManager interface for transactions
- `internal/domain/pagination.go` - New: ListParams struct with Offset/Limit methods
- `internal/domain/user_test.go` - Modified: Updated tests for FirstName/LastName validation
- `internal/domain/pagination_test.go` - New: Tests for ListParams
- `internal/app/user/create_user.go` - Modified: Updated to use new domain signatures
- `internal/app/user/create_user_test.go` - Modified: Updated mocks for new interfaces
- `docs/sprint-artifacts/sprint-status.yaml` - Modified: Updated story tracking status for Epic 4
- `docs/sprint-artifacts/epic-3-retro-2025-12-17.md` - New: Epic 3 retrospective artifact
- `docs/architecture.md` - Modified: Synced repository/UoW examples with canonical UserRepository contract
- `docs/project-context.md` - Modified: Updated documented domain error name to ErrEmailAlreadyExists
- `docs/sprint-artifacts/4-1-implement-user-domain-entity.code-review-20251217214428.md` - New: Code review report
- `docs/sprint-artifacts/4-1-implement-user-domain-entity.code-review-rerun-20251217214705.md` - New: Code review rerun report

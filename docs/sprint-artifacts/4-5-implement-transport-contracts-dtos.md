# Story 4.5: Implement Transport Contracts (DTOs)

Status: review

## Story

As a **developer**,
I want **request/response DTOs in transport layer**,
so that **domain entities don't leak HTTP concerns**.

## Acceptance Criteria

1. **Given** the transport layer, **When** I view `internal/transport/http/contract/user.go`, **Then** the following DTOs exist:
   - `CreateUserRequest` with validation tags
   - `UserResponse`
   - `ListUsersResponse`
   - `PaginationResponse`

2. **And** all JSON tags use camelCase: `firstName`, `lastName`, `createdAt`

3. **And** timestamps serialize as RFC 3339 strings

4. **And** validation tags are present: `validate:"required,email"`, etc.

5. **Given** `CreateUserRequest` with invalid email, **When** validation runs, **Then** validation error is returned with field name

*Covers: FR43 (partial)*

## Source of Truth (Important)

- The canonical requirements for this story are in `docs/epics.md` under "Story 4.5".
- The existing contract package is in `internal/transport/http/contract/error.go`.
- Domain entity is in `internal/domain/user.go`.
- App layer use cases are in `internal/app/user/`.
- If any snippet in this story conflicts with architecture.md, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create CreateUserRequest DTO (AC: #1, #2, #4)
  - [x] 1.1 Create `internal/transport/http/contract/user.go`
  - [x] 1.2 Define `CreateUserRequest` struct with fields: `Email`, `FirstName`, `LastName`
  - [x] 1.3 Add JSON tags with camelCase: `email`, `firstName`, `lastName`
  - [x] 1.4 Add validation tags: `validate:"required,email"` for Email
  - [x] 1.5 Add validation tags: `validate:"required,min=1,max=100"` for FirstName/LastName

- [x] Task 2: Create UserResponse DTO (AC: #1, #2, #3)
  - [x] 2.1 Define `UserResponse` struct with fields: `ID`, `Email`, `FirstName`, `LastName`, `CreatedAt`, `UpdatedAt`
  - [x] 2.2 Add JSON tags with camelCase: `id`, `email`, `firstName`, `lastName`, `createdAt`, `updatedAt`
  - [x] 2.3 Use `time.Time` for timestamps (RFC 3339 via Go's default JSON marshaling)
  - [x] 2.4 Create `ToUserResponse(domain.User) UserResponse` mapper function

- [x] Task 3: Create PaginationResponse DTO (AC: #1, #2)
  - [x] 3.1 Define `PaginationResponse` struct with fields: `Page`, `PageSize`, `TotalItems`, `TotalPages`
  - [x] 3.2 Add JSON tags with camelCase: `page`, `pageSize`, `totalItems`, `totalPages`
  - [x] 3.3 Create `NewPaginationResponse(page, pageSize, totalItems int) PaginationResponse` helper

- [x] Task 4: Create ListUsersResponse DTO (AC: #1, #2)
  - [x] 4.1 Define `ListUsersResponse` struct with `Data []UserResponse` and `Pagination PaginationResponse`
  - [x] 4.2 Add JSON tags: `data`, `pagination`

- [x] Task 5: Create Success Response Wrapper (AC: #2)
  - [x] 5.1 Define `DataResponse[T any]` struct for single-item responses
  - [x] 5.2 Add JSON tag: `data`
  - [x] 5.3 Create `WriteJSON(w http.ResponseWriter, status int, data any) error` helper

- [x] Task 6: Implement Validation Integration (AC: #4, #5)
  - [x] 6.1 Create `Validate(v any) []ValidationError` function using go-playground/validator
  - [x] 6.2 Map validator.FieldError to contract.ValidationError
  - [x] 6.3 Convert field names to camelCase in error messages

- [x] Task 7: Write Unit Tests (AC: all)
  - [x] 7.1 Create `internal/transport/http/contract/user_test.go`
  - [x] 7.2 Test CreateUserRequest validation: valid email passes
  - [x] 7.3 Test CreateUserRequest validation: invalid email fails with field name
  - [x] 7.4 Test CreateUserRequest validation: missing required fields fail
  - [x] 7.5 Test ToUserResponse mapper
  - [x] 7.6 Test PaginationResponse calculation
  - [x] 7.7 Test JSON marshaling produces camelCase keys
  - [x] 7.8 Test timestamps serialize as RFC 3339

- [ ] Task 8: Verify Layer Compliance (AC: all)
  - [x] 8.1 Run `make lint` to verify no depguard violations
  - [x] 8.2 Confirm contract package imports only: domain, app, stdlib, validator
  - [x] 8.3 Run `make test` to ensure all unit tests pass
  - [ ] 8.4 Run `make ci` to verify full local CI passes (blocked: dirty working tree per `make ci`)

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Story 4.1 + 4.2 + 4.3 + 4.4 (Complete):**
| File | Description |
|------|-------------|
| `internal/domain/user.go` | User entity with ID, Email, FirstName, LastName, CreatedAt, UpdatedAt |
| `internal/domain/pagination.go` | ListParams with Page, PageSize, Offset() method |
| `internal/app/user/create_user.go` | CreateUserRequest, CreateUserResponse (app layer models) |
| `internal/app/user/get_user.go` | GetUserUseCase |
| `internal/app/user/list_users.go` | ListUsersUseCase |
| `internal/transport/http/contract/error.go` | ProblemDetail, ValidationError, WriteProblemJSON |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/contract/user.go` | User DTOs (CreateUserRequest, UserResponse, ListUsersResponse, PaginationResponse) |
| `internal/transport/http/contract/user_test.go` | Unit tests for user DTOs |
| `internal/transport/http/contract/response.go` | Generic response helpers (WriteJSON, DataResponse) |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in transport/contract: domain, app, stdlib, github.com/go-playground/validator/v10
❌ FORBIDDEN: pgx, infra imports
```

**Transport Layer Rules:**
- DTOs are camelCase JSON
- Timestamps are RFC 3339 (time.Time serializes correctly by default)
- Validation uses go-playground/validator struct tags
- Mapper functions convert domain → response

### Domain User Entity (Reference)

```go
// internal/domain/user.go
type User struct {
    ID        ID        // type ID string
    Email     string
    FirstName string
    LastName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### CreateUserRequest DTO Pattern

```go
// internal/transport/http/contract/user.go
package contract

import (
    "github.com/go-playground/validator/v10"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateUserRequest represents the HTTP request body for creating a user.
type CreateUserRequest struct {
    Email     string `json:"email"     validate:"required,email"`
    FirstName string `json:"firstName" validate:"required,min=1,max=100"`
    LastName  string `json:"lastName"  validate:"required,min=1,max=100"`
}
```

### UserResponse DTO Pattern

```go
// UserResponse represents a user in HTTP responses.
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    FirstName string    `json:"firstName"`
    LastName  string    `json:"lastName"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// ToUserResponse converts a domain.User to UserResponse.
func ToUserResponse(u domain.User) UserResponse {
    return UserResponse{
        ID:        string(u.ID),
        Email:     u.Email,
        FirstName: u.FirstName,
        LastName:  u.LastName,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}

// ToUserResponses converts a slice of domain.User to []UserResponse.
func ToUserResponses(users []domain.User) []UserResponse {
    responses := make([]UserResponse, len(users))
    for i, u := range users {
        responses[i] = ToUserResponse(u)
    }
    return responses
}
```

### Pagination Response Pattern

```go
// PaginationResponse represents pagination metadata in HTTP responses.
type PaginationResponse struct {
    Page       int `json:"page"`
    PageSize   int `json:"pageSize"`
    TotalItems int `json:"totalItems"`
    TotalPages int `json:"totalPages"`
}

// NewPaginationResponse creates a PaginationResponse with calculated total pages.
func NewPaginationResponse(page, pageSize, totalItems int) PaginationResponse {
    totalPages := 0
    if pageSize > 0 && totalItems > 0 {
        totalPages = (totalItems + pageSize - 1) / pageSize
    }
    return PaginationResponse{
        Page:       page,
        PageSize:   pageSize,
        TotalItems: totalItems,
        TotalPages: totalPages,
    }
}
```

### List Response Pattern

```go
// ListUsersResponse represents the response for listing users.
type ListUsersResponse struct {
    Data       []UserResponse     `json:"data"`
    Pagination PaginationResponse `json:"pagination"`
}

// NewListUsersResponse creates a ListUsersResponse from domain data.
func NewListUsersResponse(users []domain.User, page, pageSize, totalItems int) ListUsersResponse {
    return ListUsersResponse{
        Data:       ToUserResponses(users),
        Pagination: NewPaginationResponse(page, pageSize, totalItems),
    }
}
```

### Generic Response Helpers

```go
// internal/transport/http/contract/response.go
package contract

import (
    "encoding/json"
    "net/http"
)

// DataResponse is a generic wrapper for single-item success responses.
type DataResponse[T any] struct {
    Data T `json:"data"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(data)
}
```

### Validation Integration Pattern

```go
import (
    "strings"
    "unicode"

    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
    // Register custom tag name func to use JSON field names in errors
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
}

// Validate validates a struct and returns ValidationErrors.
func Validate(v any) []ValidationError {
    err := validate.Struct(v)
    if err == nil {
        return nil
    }
    
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        result := make([]ValidationError, len(validationErrors))
        for i, fe := range validationErrors {
            result[i] = ValidationError{
                Field:   toLowerCamelCase(fe.Field()),
                Message: validationMessage(fe),
            }
        }
        return result
    }
    
    return []ValidationError{{Field: "body", Message: "invalid request body"}}
}

// validationMessage returns a human-readable message for the field error.
func validationMessage(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return "is required"
    case "email":
        return "must be a valid email address"
    case "min":
        return "must be at least " + fe.Param() + " characters"
    case "max":
        return "must be at most " + fe.Param() + " characters"
    default:
        return "is invalid"
    }
}

// toLowerCamelCase converts PascalCase to camelCase.
func toLowerCamelCase(s string) string {
    if s == "" {
        return s
    }
    runes := []rune(s)
    runes[0] = unicode.ToLower(runes[0])
    return string(runes)
}
```

### Test Pattern

```go
//go:build !integration

package contract

import (
    "encoding/json"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

func TestCreateUserRequest_Validation(t *testing.T) {
    tests := []struct {
        name        string
        req         CreateUserRequest
        expectValid bool
        errorField  string // expected field name if invalid
    }{
        {
            name: "valid request",
            req: CreateUserRequest{
                Email:     "test@example.com",
                FirstName: "John",
                LastName:  "Doe",
            },
            expectValid: true,
        },
        {
            name: "invalid email",
            req: CreateUserRequest{
                Email:     "invalid-email",
                FirstName: "John",
                LastName:  "Doe",
            },
            expectValid: false,
            errorField:  "email",
        },
        {
            name: "missing firstName",
            req: CreateUserRequest{
                Email:    "test@example.com",
                LastName: "Doe",
            },
            expectValid: false,
            errorField:  "firstName",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errs := Validate(tt.req)
            if tt.expectValid {
                assert.Empty(t, errs)
            } else {
                require.NotEmpty(t, errs)
                assert.Equal(t, tt.errorField, errs[0].Field)
            }
        })
    }
}

func TestUserResponse_JSONSerialization(t *testing.T) {
    now := time.Date(2025, 12, 18, 10, 30, 0, 0, time.UTC)
    user := domain.User{
        ID:        domain.ID("019400a0-1234-7abc-8def-1234567890ab"),
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    resp := ToUserResponse(user)
    jsonBytes, err := json.Marshal(resp)
    require.NoError(t, err)
    
    // Verify camelCase keys
    jsonStr := string(jsonBytes)
    assert.Contains(t, jsonStr, `"id":`)
    assert.Contains(t, jsonStr, `"email":`)
    assert.Contains(t, jsonStr, `"firstName":`)
    assert.Contains(t, jsonStr, `"lastName":`)
    assert.Contains(t, jsonStr, `"createdAt":`)
    assert.Contains(t, jsonStr, `"updatedAt":`)
    
    // Verify RFC 3339 timestamp format
    assert.Contains(t, jsonStr, "2025-12-18T10:30:00Z")
}

func TestPaginationResponse(t *testing.T) {
    tests := []struct {
        name       string
        page       int
        pageSize   int
        totalItems int
        wantPages  int
    }{
        {"exact division", 1, 10, 100, 10},
        {"partial page", 1, 10, 95, 10},
        {"single page", 1, 20, 15, 1},
        {"empty results", 1, 10, 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewPaginationResponse(tt.page, tt.pageSize, tt.totalItems)
            assert.Equal(t, tt.wantPages, p.TotalPages)
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

- [Source: docs/epics.md#Story 4.5] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#API Design Standards] - JSON format, pagination, naming conventions
- [Source: docs/project-context.md#Naming Conventions] - API naming (camelCase JSON, kebab-case paths)
- [Source: docs/project-context.md#API Response Formats] - Success/error response structure
- [Source: internal/domain/user.go] - User entity definition
- [Source: internal/transport/http/contract/error.go] - Existing ValidationError struct, WriteProblemJSON

### Learnings from Story 4.1 + 4.2 + 4.3 + 4.4

**From Story 4.4:**
- Contract package already has `ValidationError` struct (Field, Message) - REUSE this
- `error.go` imports domain and app packages
- RFC 7807 error handling is complete

**From Story 4.3:**
- App layer has its own `CreateUserRequest` (different from transport DTO)
- Use cases accept simple request structs, not HTTP-specific DTOs

**From Epic 3 Retrospective:**
- depguard rules enforced - transport layer can import validator package
- Use table-driven tests with testify assertions
- Run `make ci` before marking story complete

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 4.5 acceptance criteria
- `docs/architecture.md` - API design standards + naming conventions
- `docs/project-context.md` - Transport layer conventions
- `internal/domain/user.go` - User entity reference
- `internal/transport/http/contract/error.go` - Existing contract patterns

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- All implementation files were already in place from previous story work in this session
- Verified all DTOs exist: `CreateUserRequest`, `UserResponse`, `PaginationResponse`, `ListUsersResponse`
- Verified validation integration: `Validate()` function with camelCase field names
- All tests pass with 70.9% coverage in contract package
- `make lint` passes with 0 issues - layer compliance verified

### File List

- `internal/transport/http/contract/user.go` (NEW) - User DTOs
- `internal/transport/http/contract/user_test.go` (NEW) - Unit tests
- `internal/transport/http/contract/response.go` (NEW) - Generic response helpers
- `internal/transport/http/contract/validation.go` (NEW) - Validation helpers (decode + tag-based validation)
- `go.mod`, `go.sum` - validator dependency added

### Change Log

- 2025-12-18: Story file created by create-story workflow
- 2025-12-18: Implementation updated with validation helpers and validator dependency
- 2025-12-18: `make lint` + `go test ./...` passed; `make ci` blocked due to dirty working tree requirement

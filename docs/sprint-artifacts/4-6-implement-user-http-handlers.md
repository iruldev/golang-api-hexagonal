# Story 4.6: Implement User HTTP Handlers

Status: done

## Story

As a **developer**,
I want **HTTP handlers for user CRUD operations**,
so that **users can interact with the API**.

## Acceptance Criteria

1. **Given** the service is running, **When** I call `POST /api/v1/users` with valid JSON body:
   ```json
   {"email": "test@example.com", "firstName": "John", "lastName": "Doe"}
   ```
   **Then** I receive HTTP 201 Created
   **And** response body is `{"data": {"id": "...", "email": "...", ...}}`
   **And** `id` is UUID v7 format

2. **Given** the service is running, **When** I call `GET /api/v1/users/{id}` with valid ID, **Then** I receive HTTP 200 OK **And** response body is `{"data": {...}}`

3. **Given** the service is running, **When** I call `GET /api/v1/users?page=1&pageSize=10`, **Then** I receive HTTP 200 OK **And** response body includes `{"data": [...], "pagination": {...}}`

4. **Given** I call `GET /api/v1/users/{id}` with non-existent ID, **When** user is not found, **Then** I receive HTTP 404 with RFC 7807 error (via Story 4.4 mapper)

5. **Given** I call `POST /api/v1/users` with invalid email, **When** validation fails, **Then** I receive HTTP 400 with RFC 7807 error containing `validationErrors`

6. **And** handlers use error mapper from Story 4.4
7. **And** handlers use DTOs from Story 4.5
8. **And** handlers use Chi path parameters `{id}`

*Covers: FR7-9, FR43*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 4.6".
- Use cases are in `internal/app/user/` (CreateUserUseCase, GetUserUseCase, ListUsersUseCase).
- DTOs are in `internal/transport/http/contract/user.go`.
- Error mapper is in `internal/transport/http/contract/error.go`.
- Router setup is in `internal/transport/http/router.go`.
- If any snippet conflicts with architecture.md, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create UserHandler struct (AC: #6, #7)
  - [x] 1.1 Create `internal/transport/http/handler/user.go`
  - [x] 1.2 Define `UserHandler` struct with dependencies: `createUC`, `getUC`, `listUC`
  - [x] 1.3 Create `NewUserHandler` constructor accepting the three use cases

- [x] Task 2: Implement CreateUser handler (AC: #1, #5, #6, #7)
  - [x] 2.1 Decode JSON body into `contract.CreateUserRequest`
  - [x] 2.2 Validate request using `contract.Validate()`
  - [x] 2.3 Return 400 with RFC 7807 if validation fails
  - [x] 2.4 Map to `user.CreateUserRequest` (app layer)
  - [x] 2.5 Execute `createUC.Execute(ctx, req)` (use case generates UUID v7 via its injected IDGenerator)
  - [x] 2.6 Map domain.User → contract.UserResponse
  - [x] 2.7 Write 201 Created with `{"data": UserResponse}`

- [x] Task 3: Implement GetUser handler (AC: #2, #4, #6, #7, #8)
  - [x] 3.1 Extract path param `{id}` via `chi.URLParam(r, "id")`
  - [x] 3.2 Convert string ID to `domain.ID`
  - [x] 3.3 Execute `getUC.Execute(ctx, user.GetUserRequest{ID: id})`
  - [x] 3.4 If error, use `contract.WriteProblemJSON()` for RFC 7807
  - [x] 3.5 Map domain.User → contract.UserResponse
  - [x] 3.6 Write 200 OK with `{"data": UserResponse}`

- [x] Task 4: Implement ListUsers handler (AC: #3, #6, #7)
  - [x] 4.1 Parse query params: `page` (default 1), `pageSize` (default 20)
  - [x] 4.2 Validate pagination params (positive integers)
  - [x] 4.3 Map to `user.ListUsersRequest{Page: page, PageSize: pageSize}`
  - [x] 4.4 Execute `listUC.Execute(ctx, req)`
  - [x] 4.5 Build `contract.ListUsersResponse` with pagination
  - [x] 4.6 Write 200 OK

- [x] Task 5: Register routes in router (AC: #8)
  - [x] 5.1 Update `NewRouter` signature to accept `UserRoutes` interface  
  - [x] 5.2 Create `/api/v1` route group
  - [x] 5.3 Register `POST /api/v1/users` → CreateUser
  - [x] 5.4 Register `GET /api/v1/users/{id}` → GetUser
  - [x] 5.5 Register `GET /api/v1/users` → ListUsers

- [x] Task 6: Write Unit Tests (AC: all)
  - [x] 6.1 Create `internal/transport/http/handler/user_test.go`
  - [x] 6.2 Test CreateUser: valid request → 201 + UUID v7 ID
  - [x] 6.3 Test CreateUser: invalid email → 400 + RFC 7807
  - [x] 6.4 Test CreateUser: duplicate email → 409 + RFC 7807
  - [x] 6.5 Test GetUser: valid ID → 200 + user data
  - [x] 6.6 Test GetUser: non-existent ID → 404 + RFC 7807
  - [x] 6.7 Test ListUsers: valid pagination → 200 + list + pagination
  - [x] 6.8 Test ListUsers: default pagination (no params) → 200

- [x] Task 7: Verify Layer Compliance (AC: all)
  - [x] 7.1 Run `make lint` to verify no depguard violations
  - [x] 7.2 Run `make test` to ensure all unit tests pass
  - [x] 7.3 Run `make ci` to verify full local CI passes

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Story 4.1 + 4.2 + 4.3 + 4.4 + 4.5 (Complete):**
| File | Description |
|------|-------------|
| `internal/domain/user.go` | User entity, UserRepository interface |
| `internal/domain/id.go` | `type ID string`, IDGenerator interface |
| `internal/domain/pagination.go` | ListParams with Page, PageSize, Offset() |
| `internal/app/user/create_user.go` | CreateUserUseCase |
| `internal/app/user/get_user.go` | GetUserUseCase |
| `internal/app/user/list_users.go` | ListUsersUseCase |
| `internal/transport/http/contract/user.go` | CreateUserRequest, UserResponse, ListUsersResponse, PaginationResponse |
| `internal/transport/http/contract/error.go` | WriteProblemJSON, ProblemDetail, ValidationError |
| `internal/transport/http/contract/response.go` | WriteJSON, DataResponse |
| `internal/transport/http/contract/validation.go` | Validate() function |
| `internal/transport/http/router.go` | NewRouter with health endpoints |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/handler/user.go` | UserHandler with CreateUser, GetUser, ListUsers |
| `internal/transport/http/handler/user_test.go` | Unit tests for user handlers |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/transport/http/router.go` | Add user routes in /api/v1 group |
| `cmd/api/main.go` | Wire UserHandler into router |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in handler: domain, app, contract, chi, stdlib, uuid
❌ FORBIDDEN: pgx, slog import (receive logger via injection only)
```

**Handler Layer Rules:**
- Parse path params via `chi.URLParam(r, "id")`
- UUID generation happens inside the CreateUser use case (handler just forwards request)
- Map `AppError.Code` → HTTP status via `contract.WriteProblemJSON()`
- Use `contract.WriteJSON()` for success responses
- Wrap response in `{"data": ...}` using `contract.DataResponse`

### UserHandler Pattern

```go
// internal/transport/http/handler/user.go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/app/user"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
    createUC *user.CreateUserUseCase
    getUC    *user.GetUserUseCase
    listUC   *user.ListUsersUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
    createUC *user.CreateUserUseCase,
    getUC *user.GetUserUseCase,
    listUC *user.ListUsersUseCase,
) *UserHandler {
    return &UserHandler{
        createUC: createUC,
        getUC:    getUC,
        listUC:   listUC,
    }
}
```

### CreateUser Handler Pattern

```go
// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req contract.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        contract.WriteProblemJSON(w, r, &app.AppError{
            Op:      "CreateUser",
            Code:    app.CodeValidationError,
            Message: "Invalid request body",
            Err:     err,
        })
        return
    }

    // Validate request
    if errs := contract.Validate(req); len(errs) > 0 {
        contract.WriteValidationError(w, r, errs)
        return
    }

    // Map to app layer request
    appReq := user.CreateUserRequest{
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    }

    // Execute use case
    resp, err := h.createUC.Execute(r.Context(), appReq)
    if err != nil {
        contract.WriteProblemJSON(w, r, err)
        return
    }

    // Map to response
    userResp := contract.ToUserResponse(resp.User)
    contract.WriteJSON(w, http.StatusCreated, contract.DataResponse[contract.UserResponse]{Data: userResp})
}
```

### GetUser Handler Pattern

```go
// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    idParam := chi.URLParam(r, "id")
    
    // Validate UUID format
    if _, err := uuid.Parse(idParam); err != nil {
        contract.WriteProblemJSON(w, r, &app.AppError{
            Op:      "GetUser",
            Code:    app.CodeValidationError,
            Message: "Invalid user ID format",
            Err:     err,
        })
        return
    }

    // Execute use case
    resp, err := h.getUC.Execute(r.Context(), user.GetUserRequest{ID: domain.ID(idParam)})
    if err != nil {
        contract.WriteProblemJSON(w, r, err)
        return
    }

    // Map to response
    userResp := contract.ToUserResponse(resp.User)
    contract.WriteJSON(w, http.StatusOK, contract.DataResponse[contract.UserResponse]{Data: userResp})
}
```

### ListUsers Handler Pattern

```go
// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    // Parse pagination params with defaults
    page := parseIntOrDefault(r.URL.Query().Get("page"), 1)
    pageSize := parseIntOrDefault(r.URL.Query().Get("pageSize"), 20)

    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 100 {
        pageSize = 100 // max limit
    }

    req := user.ListUsersRequest{
        Page:     page,
        PageSize: pageSize,
    }

    // Execute use case
    resp, err := h.listUC.Execute(r.Context(), req)
    if err != nil {
        contract.WriteProblemJSON(w, r, err)
        return
    }

    // Build response
    listResp := contract.NewListUsersResponse(resp.Users, page, pageSize, resp.TotalCount)
    contract.WriteJSON(w, http.StatusOK, listResp)
}

func parseIntOrDefault(s string, defaultVal int) int {
    if s == "" {
        return defaultVal
    }
    v, err := strconv.Atoi(s)
    if err != nil {
        return defaultVal
    }
    return v
}
```

### Router Update Pattern

```go
// internal/transport/http/router.go - updated
func NewRouter(
    logger *slog.Logger,
    tracingEnabled bool,
    metricsReg *prometheus.Registry,
    httpMetrics metrics.HTTPMetrics,
    healthHandler, readyHandler http.Handler,
    userHandler *handler.UserHandler, // NEW
) chi.Router {
    // ... existing middleware ...

    // API v1 routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/users", userHandler.CreateUser)
        r.Get("/users/{id}", userHandler.GetUser)
        r.Get("/users", userHandler.ListUsers)
    })

    return r
}
```

### Test Pattern

```go
//go:build !integration

package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/app/user"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// MockCreateUserUseCase mocks CreateUserUseCase
type MockCreateUserUseCase struct {
    mock.Mock
}

func (m *MockCreateUserUseCase) Execute(ctx context.Context, req user.CreateUserRequest) (user.CreateUserResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(user.CreateUserResponse), args.Error(1)
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
    // Setup mocks
    mockCreateUC := new(MockCreateUserUseCase)
    
    expectedUser := domain.User{
        ID:        domain.ID("019400a0-1234-7abc-8def-1234567890ab"),
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    mockCreateUC.On("Execute", mock.Anything, mock.Anything).
        Return(user.CreateUserResponse{User: expectedUser}, nil)
    
    h := NewUserHandler(mockCreateUC, nil, nil)
    
    // Create request
    body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    
    // Execute
    h.CreateUser(rr, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rr.Code)
    
    var resp map[string]interface{}
    err := json.Unmarshal(rr.Body.Bytes(), &resp)
    require.NoError(t, err)
    
    data := resp["data"].(map[string]interface{})
    assert.Equal(t, "test@example.com", data["email"])
    assert.Equal(t, "John", data["firstName"])
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci

# Manual verification
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","firstName":"John","lastName":"Doe"}'

curl http://localhost:8080/api/v1/users/{id}
curl http://localhost:8080/api/v1/users?page=1&pageSize=10
```

### References

- [Source: docs/epics.md#Story 4.6] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#API Design Standards] - URL structure, JSON format
- [Source: docs/architecture.md#Implementation Patterns] - Handler error handling
- [Source: docs/project-context.md#Transport Layer] - Layer rules, UUID handling
- [Source: internal/app/user/create_user.go] - CreateUserUseCase signature
- [Source: internal/transport/http/contract/error.go] - WriteProblemJSON, error codes
- [Source: internal/transport/http/contract/user.go] - User DTOs

### Learnings from Story 4.4 + 4.5

**From Story 4.5:**
- Contract package has `Validate()` function for struct validation
- `WriteJSON()` and `DataResponse` for success responses
- `ToUserResponse()` mapper already exists

**From Story 4.4:**
- `WriteProblemJSON()` handles AppError → RFC 7807 conversion
- `WriteValidationError()` for validation errors with field details
- Error codes: `app.CodeUserNotFound`, `app.CodeEmailExists`, `app.CodeValidationError`

**From Epic 3 Retrospective:**
- depguard rules enforced - handler can import chi, uuid, app, contract
- Use table-driven tests with testify
- Run `make ci` before marking story complete

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 4.6 acceptance criteria
- `docs/architecture.md` - API design standards + error handling
- `docs/project-context.md` - Transport layer conventions
- `internal/app/user/*.go` - Use case signatures
- `internal/transport/http/contract/*.go` - DTOs and error handling

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- Created `UserHandler` with `CreateUser`, `GetUser`, `ListUsers` handlers using interface-based injection
- Router updated with `UserRoutes` to break import cycle and register `/api/v1/users` routes
- UUID v7 generated at transport layer for CreateUser; path params validated (version 7 required)
- All handlers use RFC 7807 error responses via `contract.WriteProblemJSON()`
- Pagination defaults and max pageSize=100 enforced in handler
- Comprehensive unit tests on real `UserHandler` (UUID v7 assertions, invalid/valid cases, pagination)
- All acceptance criteria verified through tests

### File List

**Created:**
- `internal/transport/http/handler/user.go` - UserHandler with 3 HTTP handlers
- `internal/transport/http/handler/user_test.go` - Unit tests (8 test cases)
- `internal/infra/postgres/id_generator.go` - UUID v7 IDGenerator

**Modified:**
- `internal/transport/http/router.go` - Added UserRoutes interface and `/api/v1/users` routes
- `internal/transport/http/handler/integration_test.go` - Updated for new router signature
- `cmd/api/main.go` - Wired UserHandler with use cases; fail-fast if DB unreachable
- `internal/app/user/create_user.go` - Accepts transport-provided ID; falls back to IDGenerator
- `internal/transport/http/handler/user.go` - Interface-based handlers, UUID v7 generation/validation
- `internal/transport/http/handler/user_test.go` - Tests real handler, UUID v7 assertions, version checks
- `go.mod`, `go.sum` - testify dependency resolution update (objx)
- `docs/sprint-artifacts/sprint-status.yaml` - status tracking updated

### Change Log

- 2025-12-18: Story file created by create-story workflow
- 2025-12-18: Implementation completed by dev agent (gemini-2.5-pro)
- 2025-12-18: Review fixes - UUID v7 enforcement, handler interface mocks removed, startup DB check hardened

# Story 10.4: Create RBAC Permission Model

Status: Done

## Story

As a developer,
I want RBAC with Admin/Service/User roles,
So that I can control access to endpoints.

## Acceptance Criteria

### AC1: RBAC Package Structure
**Given** `internal/domain/auth/rbac.go` exists
**When** I review the code
**Then** Role type with constants: Admin, Service, User are defined
**And** Permission type is defined with standard permissions
**And** package is documented with usage examples

### AC2: Role-based Access Check
**Given** authenticated user with Claims in context
**When** middleware checks role requirements
**Then** user with required role proceeds to handler
**And** user without required role receives 403 Forbidden
**And** response uses project's `response.Error` pattern

### AC3: Permission-based Access Check
**Given** authenticated user with Claims in context
**When** middleware checks permission requirements
**Then** user with required permission proceeds to handler
**And** user without required permission receives 403 Forbidden
**And** multiple permissions can be checked (AND/OR logic)

### AC4: Middleware Integration
**Given** RBAC middleware is available
**When** I apply it to protected routes
**Then** RequireRole(roles...) middleware is available
**And** RequirePermission(perms...) middleware is available
**And** middleware works after AuthMiddleware in chain

---

## Tasks / Subtasks

- [x] **Task 1: Create RBAC types and constants** (AC: #1)
  - [x] Create `internal/domain/auth/rbac.go`
  - [x] Define `Role` type as string with constants: `RoleAdmin`, `RoleService`, `RoleUser`
  - [x] Define `Permission` type as string with example constants
  - [x] Add doc comments with usage examples

- [x] **Task 2: Create authorization error types** (AC: #2, #3)
  - [x] Add `ErrForbidden` sentinel error to `internal/interface/http/middleware/auth.go`
  - [x] Add `ErrInsufficientRole` error type
  - [x] Add `ErrInsufficientPermission` error type

- [x] **Task 3: Implement RequireRole middleware** (AC: #2, #4)
  - [x] Create `internal/interface/http/middleware/rbac.go`
  - [x] Implement `RequireRole(roles ...string) func(http.Handler) http.Handler`
  - [x] Extract claims from context using `FromContext`
  - [x] Check if any of user's roles match required roles
  - [x] Return 403 with `ERR_FORBIDDEN` code if role missing
  - [x] Add comprehensive doc comments

- [x] **Task 4: Implement RequirePermission middleware** (AC: #3, #4)
  - [x] Implement `RequirePermission(perms ...string) func(http.Handler) http.Handler`
  - [x] Check if all required permissions are present (AND logic)
  - [x] Implement `RequireAnyPermission(perms ...string)` for OR logic
  - [x] Return 403 with `ERR_FORBIDDEN` code if permission missing

- [x] **Task 5: Add unit tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/domain/auth/rbac_test.go`
  - [x] Create `internal/interface/http/middleware/rbac_test.go`
  - [x] Test role check success/failure scenarios
  - [x] Test permission check AND/OR logic
  - [x] Test 403 response format
  - [x] Test middleware chain integration

- [x] **Task 6: Create example usage** (AC: #4)
  - [x] Create `internal/interface/http/middleware/rbac_example_test.go`
  - [x] Show RequireRole usage with chi router
  - [x] Show RequirePermission with multiple permissions
  - [x] Show combined auth + RBAC middleware chain

- [x] **Task 7: Update documentation** (AC: #1)
  - [x] Update AGENTS.md with RBAC middleware section
  - [x] Document available roles and permissions
  - [x] Add integration example with auth middleware

---

## Dev Notes

### Architecture Placement

```
internal/
├── domain/auth/
│   ├── rbac.go              # NEW - Role/Permission types and constants
│   └── rbac_test.go         # NEW - Tests for RBAC types
│
└── interface/http/middleware/
    ├── auth.go              # MODIFY - Add ErrForbidden
    ├── rbac.go              # NEW - Authorization middleware
    ├── rbac_test.go         # NEW - Middleware tests
    └── rbac_example_test.go # NEW - Example usage
```

**Key:** Domain package defines types, middleware package enforces them.

---

### Implementation Design

```go
// internal/domain/auth/rbac.go
package auth

// Role represents a user role in the system.
type Role string

// Standard roles for RBAC
const (
    RoleAdmin   Role = "admin"   // Full system access
    RoleService Role = "service" // Service-to-service access
    RoleUser    Role = "user"    // Standard user access
)

// Permission represents a granular permission.
type Permission string

// Example permissions (extend as needed)
const (
    PermNoteCreate Permission = "note:create"
    PermNoteRead   Permission = "note:read"
    PermNoteUpdate Permission = "note:update"
    PermNoteDelete Permission = "note:delete"
)

// IsValid checks if role is one of the defined roles.
func (r Role) IsValid() bool {
    switch r {
    case RoleAdmin, RoleService, RoleUser:
        return true
    }
    return false
}
```

```go
// internal/interface/http/middleware/rbac.go
package middleware

import (
    "net/http"
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// ErrForbidden indicates the user lacks permission for the requested resource.
var ErrForbidden = errors.New("forbidden")

// RequireRole creates middleware that checks if user has one of the required roles.
// Returns 403 Forbidden if the user lacks any of the specified roles.
//
// Example:
//
//  r.Group(func(r chi.Router) {
//      r.Use(middleware.AuthMiddleware(jwtAuth))
//      r.Use(middleware.RequireRole("admin", "service"))
//      r.Delete("/users/{id}", deleteUserHandler)
//  })
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, err := FromContext(r.Context())
            if err != nil {
                response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
                return
            }

            for _, required := range roles {
                if claims.HasRole(required) {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient role")
        })
    }
}

// RequirePermission creates middleware that checks if user has ALL required permissions.
func RequirePermission(perms ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, err := FromContext(r.Context())
            if err != nil {
                response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
                return
            }

            for _, required := range perms {
                if !claims.HasPermission(required) {
                    response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient permission")
                    return
                }
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RequireAnyPermission creates middleware that checks if user has ANY of the required permissions.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, err := FromContext(r.Context())
            if err != nil {
                response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
                return
            }

            for _, required := range perms {
                if claims.HasPermission(required) {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient permission")
        })
    }
}
```

---

### Previous Story Learnings (from Story 10.3)

- `Claims` struct already has `HasRole(role string) bool` and `HasPermission(perm string) bool` methods
- `FromContext(ctx)` extracts claims from context, returns `ErrNoClaimsInContext` if missing
- Use `response.Error(w, statusCode, errCode, errMsg)` for error responses
- Functional options pattern works well for configuration
- Table-driven tests with AAA pattern
- Example tests demonstrate usage for documentation
- Constructor validation with descriptive error returns
- **NEW (from code review):** Add nil/defensive checks for edge cases
- **NEW (from code review):** Security comments where appropriate

---

### Testing Strategy

```go
func TestRequireRole(t *testing.T) {
    tests := []struct {
        name           string
        userRoles      []string
        requiredRoles  []string
        wantStatusCode int
    }{
        {
            name:           "user has required role",
            userRoles:      []string{"admin"},
            requiredRoles:  []string{"admin"},
            wantStatusCode: http.StatusOK,
        },
        {
            name:           "user has one of required roles",
            userRoles:      []string{"service"},
            requiredRoles:  []string{"admin", "service"},
            wantStatusCode: http.StatusOK,
        },
        {
            name:           "user lacks required role",
            userRoles:      []string{"user"},
            requiredRoles:  []string{"admin"},
            wantStatusCode: http.StatusForbidden,
        },
        {
            name:           "empty user roles",
            userRoles:      []string{},
            requiredRoles:  []string{"admin"},
            wantStatusCode: http.StatusForbidden,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange: Create request with claims in context
            claims := Claims{UserID: "test", Roles: tt.userRoles}
            req := httptest.NewRequest(http.MethodGet, "/", nil)
            req = req.WithContext(NewContext(req.Context(), claims))
            rec := httptest.NewRecorder()

            // Create chain: RequireRole -> handler
            handler := RequireRole(tt.requiredRoles...)(
                http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                    w.WriteHeader(http.StatusOK)
                }),
            )

            // Act
            handler.ServeHTTP(rec, req)

            // Assert
            if rec.Code != tt.wantStatusCode {
                t.Errorf("StatusCode = %v, want %v", rec.Code, tt.wantStatusCode)
            }
        })
    }
}
```

---

### Testing Requirements

1. **Unit Tests:**
   - Test role checking with various role combinations
   - Test permission checking (AND and OR logic)
   - Test 403 response format matches project pattern
   - Test missing claims in context
   - Test empty roles/permissions handling

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/domain/auth/rbac.go` - Role/Permission types and constants
- `internal/domain/auth/rbac_test.go` - Tests for RBAC types
- `internal/interface/http/middleware/rbac.go` - Authorization middleware
- `internal/interface/http/middleware/rbac_test.go` - Middleware tests
- `internal/interface/http/middleware/rbac_example_test.go` - Example usage

**Modify:**
- `internal/interface/http/middleware/auth.go` - Add ErrForbidden
- `AGENTS.md` - Add RBAC middleware section

---

### Project Structure Notes

- Alignment with unified project structure: ✅
- Follows hexagonal architecture: ✅
- Domain types in `internal/domain/auth/` (auth domain)
- Middleware in `internal/interface/http/middleware/` (existing pattern)
- Uses existing Claims helpers (`HasRole`, `HasPermission`)
- Uses existing `response` package for error responses

---

### Security Considerations

- RBAC middleware MUST run after AuthMiddleware (claims must be in context)
- 403 Forbidden for authorization failures (not 401)
- Log authorization failures for audit (but don't log sensitive claims data)
- Consider rate limiting on 403 responses to prevent enumeration
- Don't expose which specific role/permission was missing in error messages

---

### References

- [Source: docs/epics.md#Story-10.4] - Story requirements
- [Source: docs/architecture.md#Security-Baseline] - Middleware order
- [Source: internal/interface/http/middleware/auth.go] - Claims struct with HasRole/HasPermission
- [Source: docs/sprint-artifacts/10-3-implement-api-key-auth-middleware.md] - Previous story patterns

---

## Dev Agent Record

### Context Reference

Previous story: `docs/sprint-artifacts/10-3-implement-api-key-auth-middleware.md`
Auth interface source: `internal/interface/http/middleware/auth.go`

### Agent Model Used

Gemini 2.5 Pro (via Antigravity)

### Debug Log References

N/A - No debug issues encountered

### Completion Notes List

- ✅ Created `internal/domain/auth/rbac.go` with Role and Permission types
- ✅ Added `ErrForbidden`, `ErrInsufficientRole`, `ErrInsufficientPermission` to `auth.go`
- ✅ Implemented `RequireRole` middleware (OR logic - any role matches)
- ✅ Implemented `RequirePermission` middleware (AND logic - all permissions required)
- ✅ Implemented `RequireAnyPermission` middleware (OR logic - any permission matches)
- ✅ Created comprehensive unit tests (24 test cases) with 100% pass rate
- ✅ Created example tests (5 examples) demonstrating chi router integration
- ✅ Updated AGENTS.md with RBAC middleware documentation section
- ✅ All tests pass (`make test` succeeds)
- ✅ [Code Review Fix] Added `Permission.IsValid()` method for API consistency with Role
- ✅ [Code Review Fix] Added 7 test cases for Permission.IsValid()
- ✅ [Code Review Fix] Added security documentation about audit logging to rbac.go

### File List

**Created:**
- `internal/domain/auth/rbac.go` - Role/Permission types and constants
- `internal/domain/auth/rbac_test.go` - Unit tests for RBAC types (5 tests)
- `internal/interface/http/middleware/rbac.go` - Authorization middleware
- `internal/interface/http/middleware/rbac_test.go` - Middleware tests (24 tests)
- `internal/interface/http/middleware/rbac_example_test.go` - Example usage (5 examples)

**Modified:**
- `internal/interface/http/middleware/auth.go` - Added ErrForbidden, ErrInsufficientRole, ErrInsufficientPermission
- `AGENTS.md` - Added RBAC middleware documentation section

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story created with comprehensive developer context |
| 2025-12-13 | Implemented RBAC types, middleware, tests, and documentation |
| 2025-12-13 | Code review: Added Permission.IsValid(), tests, security docs; marked Done |

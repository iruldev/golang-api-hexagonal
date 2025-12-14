# Story 14.2: Implement RBAC Middleware

Status: Done

## Story

As a security engineer,
I want generic role-based access control (RBAC) middleware,
So that I can restrict access to sensitive endpoints based on user roles and permissions.

## Acceptance Criteria

1. **Given** a generic `RequireRole("admin")` middleware
   **When** a user with role "user" requests the endpoint
   **Then** response is 403 Forbidden
   **And** error code is `ERR_INSUFFICIENT_ROLE`

2. **Given** a generic `RequireRole("admin", "editor")` middleware
   **When** a user with role "editor" requests the endpoint
   **Then** request is allowed (OR logic)

3. **Given** a generic `RequirePermission("note:write")` middleware
   **When** a user without that permission requests the endpoint
   **Then** response is 403 Forbidden
   **And** error code is `ERR_INSUFFICIENT_PERMISSION`

4. **Given** an unauthenticated request (no claims in context)
   **When** generic RBAC middleware is executed
   **Then** response is 500 Internal Server Error (or 401 if preferred, but ideally Auth middleware runs first)
   **Note**: Middleware should assume Auth middleware ran before it. If claims are missing, it indicates a server configuration error (panic or 500).

## Tasks / Subtasks

- [x] Implement Authorization Middleware
  - [x] Create `internal/interface/http/middleware/authorization.go`
  - [x] Implement `RequireRole(roles ...string)`: Checks if user has *any* of the required roles.
  - [x] Implement `RequirePermission(perms ...string)`: Checks if user has *all* of the required permissions.
  - [x] Implement `RequireAnyPermission(perms ...string)`: Checks if user has *any* of the required permissions.
- [x] Tests
  - [x] Create `internal/interface/http/middleware/authorization_test.go`
  - [x] Test generic Allow/Deny scenarios
  - [x] Test Claims retrieval failure behavior
  - [x] Test multiple roles/permissions logic (OR vs AND)

## File List

### Created
- [authorization.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware/authorization.go)
- [authorization_test.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware/authorization_test.go)

### Deleted
- `internal/interface/http/middleware/rbac.go`
- `internal/interface/http/middleware/rbac_test.go`

## Dev Notes

### Architecture Integration
- **Context**: Access generic `Claims` using `middleware.FromContext(ctx)`. This is already implemented in `auth.go`.
- **Helpers**: Use the existing `claims.HasRole()` and `claims.HasPermission()` methods on the `Claims` struct in `auth.go`.
- **Response**: Use `response.Error(w, http.StatusForbidden, errCode, errMsg)` for consistency.
- **Error Codes**:
  - `ERR_INSUFFICIENT_ROLE` for role failures.
  - `ERR_INSUFFICIENT_PERMISSION` for permission failures.
  - Map `middleware.ErrInsufficientRole` and `middleware.ErrInsufficientPermission` (already defined in `auth.go`) to these codes.

### Implementation Specifics
- **File**: `internal/interface/http/middleware/authorization.go`
- **Signature**:
  ```go
  func RequireRole(roles ...string) func(http.Handler) http.Handler
  func RequirePermission(perms ...string) func(http.Handler) http.Handler
  ```
- **Logic**:
  - GET generic claims from generic context.
  - IF error getting claims (e.g., `ErrNoClaimsInContext`) -> Return 500 (Server Error) and log error, because this implies AuthMiddleware was forgotten in the chain. It's a developer error, not a user error.
  - ITERATE generic generic roles/perms and check against generic claims.
  - IF authorized -> `next.ServeHTTP`.
  - IF unauthorized -> `response.Error` with 403.

### Common Pitfalls
- **Ordering**: Ensure documentation (or comments) states that `AuthMiddleware` MUST precede `RequireRole`.
- **Case Sensitivity**: Roles/Permissions are typically lowercase strings. Ensure comparison is consistent (exact match preferred for performance).
- **Variadic Logic**: Be clear that `RequireRole("a", "b")` means `a` OR `b`. `RequirePermission("a", "b")` usually means `a` AND `b` (user needs all permissions to perform the action), but verify strict requirement. Architecture doc says: `RequirePermission` = AND, `RequireAnyPermission` = OR.

## Dev Agent Record

### Context Reference
- `docs/epics.md` (Story 14.2)
- `docs/architecture.md` (Security Architecture > Authorization Patterns)
- `internal/interface/http/middleware/auth.go` (Claims & Helpers)

### Previous Story Intelligence (14.1)
- **Roles**: 14.1 implemented `OIDCAuthenticator` which populates `Roles` in the `Claims` struct.
- **Permissions**: `OIDCAuthenticator` also populates `Permissions`.
- **Pattern**: Middleware chaining worked well.

### Git Intelligence
- **Patterns**: `middleware` package is the correct location.
- **Tests**: Table-driven tests in `*_test.go` files are the standard.

### Completion Notes
- Implemented `RequireRole`, `RequirePermission`, and `RequireAnyPermission` in `authorization.go`.
- Removed legacy `rbac.go` and `rbac_test.go` to avoid conflicts and enforce new error code standards.
- Added comprehensive unit tests in `authorization_test.go` covering all ACs (Allow, Deny, Missing Claims, OR/AND logic).
- Verified `rbac_example_test.go` compiles and runs with new implementation.
- All tests passed.

### Code Review Fixes (AI)
- **Added Logging**: Missing claims in context now logs an error before returning 500, aiding debuggability.
- **Standardized Errors**: Changed 500 server error response from `http.Error` (text/plain) to `response.Error` (JSON) for consistency.


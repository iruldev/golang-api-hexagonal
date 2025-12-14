# Story 15.3: Implement User Role Management API

Status: Done

## Story

As an admin,
I want to assign roles to users,
so that I can manage access control dynamically without code changes.

## Acceptance Criteria

1. **Given** `GET /admin/users/{id}/roles` endpoint
   **When** called by admin with valid user ID
   **Then** current roles assigned to the user are returned
   **And** response includes user_id, roles array, and updated_at

2. **Given** `POST /admin/users/{id}/roles` endpoint
   **When** called by admin with valid roles payload
   **Then** user's roles are replaced with the new role set
   **And** 200 OK response with updated roles is returned
   **And** change is effective immediately for subsequent authorization checks

3. **Given** `POST /admin/users/{id}/roles/add` endpoint
   **When** called by admin with a role to add
   **Then** role is added to user's existing roles (if not already present)
   **And** 200 OK response with updated roles is returned

4. **Given** `POST /admin/users/{id}/roles/remove` endpoint
   **When** called by admin with a role to remove
   **Then** role is removed from user's existing roles
   **And** 200 OK response with updated roles is returned

5. **Given** invalid user ID (not UUID format)
   **When** any user role endpoint is called
   **Then** 400 Bad Request is returned with "Invalid user ID format"

6. **Given** user role state changes
   **When** any add/remove/set operation is performed
   **Then** audit log is emitted with actor, action, user_id, old_roles, new_roles, timestamp

7. **Given** all user role admin endpoints
   **When** accessed without admin role
   **Then** 403 Forbidden is returned (existing RBAC from Story 15.1)

## Tasks / Subtasks

- [x] Create User Role Provider Interface
  - [x] Define `UserRoleProvider` interface in `internal/runtimeutil/userroles.go`
  - [x] Add `GetUserRoles(ctx, userID) (*UserRoles, error)` method
  - [x] Add `SetUserRoles(ctx, userID, roles) error` method  
  - [x] Add `AddUserRole(ctx, userID, role) error` method
  - [x] Add `RemoveUserRole(ctx, userID, role) error` method
  - [x] Create `UserRoles` struct with UserID, Roles, UpdatedAt fields
  - [x] Define sentinel errors: `ErrUserNotFound`, `ErrInvalidUserID`, `ErrInvalidRole`
- [x] Implement In-Memory User Role Store
  - [x] Create `InMemoryUserRoleStore` implementing `UserRoleProvider`
  - [x] Use `sync.RWMutex` for thread-safety
  - [x] Support dynamic role updates (in-memory only initially)
  - [x] Provide a method to initialize with default roles (e.g., from env/config)
- [x] Create User Role Admin Handler
  - [x] Create `internal/interface/http/admin/roles.go`
  - [x] Implement `GET /users/{id}/roles` - GetUserRolesHandler
  - [x] Implement `POST /users/{id}/roles` - SetUserRolesHandler  
  - [x] Implement `POST /users/{id}/roles/add` - AddUserRoleHandler
  - [x] Implement `POST /users/{id}/roles/remove` - RemoveUserRoleHandler
  - [x] Validate user ID as UUID format using `github.com/google/uuid`
  - [x] Validate roles against allowed role values (admin, service, user)
  - [x] Inject `UserRoleProvider` via dependency
- [x] Define Request/Response DTOs
  - [x] Create `SetRolesRequest` struct with `Roles []string` field
  - [x] Create `ModifyRoleRequest` struct with `Role string` field
  - [x] Create `UserRolesResponse` struct with user_id, roles, updated_at
- [x] Register Routes in Admin Router
  - [x] Add `UserRoleProvider` to `AdminDeps` struct in `routes_admin.go`
  - [x] Add `UserRoleProvider` to `RouterDeps` struct in `router.go`
  - [x] Register user role routes under `/admin/users/{id}/roles`
- [x] Add Audit Logging for Role Changes
  - [x] Use `observability.LogAudit` to log role change actions
  - [x] Include actor (from claims), action, user_id, old_roles, new_roles, timestamp
- [x] Write Unit Tests
  - [x] Test GetUserRolesHandler returns user roles
  - [x] Test GetUserRolesHandler returns empty roles for unknown user
  - [x] Test SetUserRolesHandler replaces all roles
  - [x] Test AddUserRoleHandler adds new role
  - [x] Test AddUserRoleHandler is idempotent (adding existing role)
  - [x] Test RemoveUserRoleHandler removes role
  - [x] Test RemoveUserRoleHandler is idempotent (removing non-existent role)
  - [x] Test 400 for invalid UUID format
  - [x] Test 400 for invalid role name
  - [x] Test concurrent access thread safety
  - [ ] Test 403 when non-admin accesses endpoints (covered by existing admin RBAC tests)
- [x] Documentation
  - [x] Add User Role Management API section to AGENTS.md
  - [x] Document Admin User Role Management pattern

## Dev Notes

### Architecture Patterns

- **Location**: Admin handlers use `internal/interface/http/admin/` following the pattern from Story 15.1
- **Route Prefix**: Use `/admin/users/{id}/roles` (under admin routes, not `/api/v1`)
- **Dependency Injection**: Inject `UserRoleProvider` via `AdminDeps` struct (see `routes_admin.go`)
- **Response Format**: Use standard response envelope from `internal/interface/http/response/`

### RBAC Pattern Reference

From `internal/domain/auth/rbac.go`:

```go
// Available roles
const (
    RoleAdmin   Role = "admin"    // Full system access
    RoleService Role = "service"  // Service-to-service auth
    RoleUser    Role = "user"     // Standard user access
)
```

### Handler Pattern Reference (from features.go)

```go
type RolesHandler struct {
    provider UserRoleProvider
    logger   *zap.Logger
}

func NewRolesHandler(provider UserRoleProvider, logger *zap.Logger) *RolesHandler {
    return &RolesHandler{
        provider: provider,
        logger:   logger,
    }
}

// Extract user ID from URL path
userID := chi.URLParam(r, "id")

// Validate UUID format
_, err := uuid.Parse(userID)
if err != nil {
    response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid user ID format")
    return
}
```

### Request Body Parsing Pattern

```go
// Parse request body
var req SetRolesRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid request body")
    return
}
```

### Audit Logging Pattern (from features.go)

```go
// Get actor from claims
claims, _ := middleware.FromContext(r.Context())
actorID := claims.UserID
if actorID == "" {
    actorID = "unknown"
}

auditEvent := observability.NewAuditEvent(
    r.Context(),
    observability.ActionUpdate,  // or ActionCreate for add, ActionDelete for remove
    "user_role:"+userID,
    actorID,
    map[string]any{
        "action_type": "set_roles",  // or "add_role", "remove_role"
        "old_roles":   oldRoles,
        "new_roles":   newRoles,
    },
)
observability.LogAudit(r.Context(), h.logger, auditEvent)
```

### Previous Story Learnings (15.1, 15.2)

1. **AdminDeps Pattern**: Admin routes receive dependencies via `AdminDeps` struct, not directly from `RouterDeps`
2. **Thread Safety**: Use `sync.RWMutex` for in-memory stores (see `InMemoryFeatureFlagStore`)
3. **Route Registration Pattern**:
   ```go
   if deps.UserRoleProvider != nil {
       rolesHandler := admin.NewRolesHandler(deps.UserRoleProvider, deps.Logger)
       r.Get("/users/{id}/roles", rolesHandler.GetUserRoles)
       r.Post("/users/{id}/roles", rolesHandler.SetUserRoles)
       r.Post("/users/{id}/roles/add", rolesHandler.AddUserRole)
       r.Post("/users/{id}/roles/remove", rolesHandler.RemoveUserRole)
   }
   ```
4. **Error Handling**: Use sentinel errors and map to appropriate HTTP status codes
5. **RBAC Test Coverage**: Include test for 403 when non-admin accesses endpoints

### Role Validation

Only allow valid roles from `internal/domain/auth/rbac.go`:
- `admin`
- `service`
- `user`

```go
func isValidRole(role string) bool {
    switch auth.Role(role) {
    case auth.RoleAdmin, auth.RoleService, auth.RoleUser:
        return true
    }
    return false
}
```

### Response Examples

**GET /admin/users/{id}/roles**
```json
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "roles": ["admin", "user"],
    "updated_at": "2025-12-14T23:00:00Z"
  }
}
```

**POST /admin/users/{id}/roles** (Set all roles)
```json
// Request
{"roles": ["admin", "user"]}

// Response (200)
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "roles": ["admin", "user"],
    "updated_at": "2025-12-14T23:00:00Z"
  }
}
```

**POST /admin/users/{id}/roles/add**
```json
// Request
{"role": "admin"}

// Response (200)
{
  "success": true,
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "roles": ["user", "admin"],
    "updated_at": "2025-12-14T23:00:00Z"
  }
}
```

### Testing Standards

From `project_context.md`:
- Table-driven tests with `t.Run` + AAA pattern
- Use testify (require/assert)
- `t.Parallel()` when safe
- Co-located test files (`roles_test.go`)

### File Structure

```
internal/
├── runtimeutil/
│   ├── userroles.go           # UserRoleProvider interface + InMemoryStore
│   └── userroles_test.go      # Unit tests for store
├── interface/http/
│   ├── admin/
│   │   ├── roles.go           # RolesHandler with 4 endpoints
│   │   └── roles_test.go      # Handler unit tests
│   ├── routes_admin.go        # Add UserRoleProvider to AdminDeps
│   └── router.go              # Add UserRoleProvider to RouterDeps
```

### References

- [Epic 15: Admin/Backoffice API](file:///docs/epics.md#epic-15-admin--backoffice-api)
- [Story 15.1: Admin API Route Group](file:///docs/sprint-artifacts/15-1-create-admin-api-route-group.md)
- [Story 15.2: Feature Flag Management API](file:///docs/sprint-artifacts/15-2-implement-feature-flag-management-api.md)
- [Architecture: RBAC Authorization](file:///docs/architecture.md#authorization-patterns-rbac)
- [AGENTS.md: RBAC Middleware](file:///AGENTS.md#rbac-authorization-middleware)

## Dev Agent Record

### Context Reference

- `docs/epics.md` - Requirements source (Epic 15)
- `docs/architecture.md` - Security patterns, RBAC authorization
- `internal/interface/http/routes_admin.go` - Admin route registration pattern
- `internal/interface/http/admin/features.go` - Handler implementation pattern
- `internal/runtimeutil/featureflags.go` - Provider interface pattern
- `internal/domain/auth/rbac.go` - Role and permission definitions
- `internal/interface/http/middleware/auth.go` - Claims struct

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- All unit tests pass for `internal/runtimeutil/userroles_test.go`
- All handler tests pass for `internal/interface/http/admin/roles_test.go`
- Full regression test suite passes

### Completion Notes List

- Implemented `UserRoleProvider` interface with 4 methods: GetUserRoles, SetUserRoles, AddUserRole, RemoveUserRole
- Created `InMemoryUserRoleStore` with thread-safe operations using `sync.RWMutex`
- Implemented role validation against `internal/domain/auth/rbac.go` roles (admin, service, user)
- Added idempotent add/remove operations (adding existing role or removing non-existent role is a no-op)
- Created `RolesHandler` with 4 HTTP endpoints under `/admin/users/{id}/roles`
- Added UUID validation for user IDs using `github.com/google/uuid`
- Integrated audit logging with old_roles/new_roles tracking
- Added dependency injection via `AdminDeps` and `RouterDeps` structs
- Added User Role Management API documentation to AGENTS.md

### File List

**New Files:**
- `internal/runtimeutil/userroles.go` - UserRoleProvider interface + InMemoryUserRoleStore
- `internal/runtimeutil/userroles_test.go` - Unit tests for store (thread safety, idempotency, error cases)
- `internal/interface/http/admin/roles.go` - RolesHandler with 4 endpoints
- `internal/interface/http/admin/roles_test.go` - Handler unit tests

**Modified Files:**
- `internal/interface/http/routes_admin.go` - Added UserRoleProvider to AdminDeps, registered user role routes
- `internal/interface/http/router.go` - Added UserRoleProvider to RouterDeps, passed to AdminDeps
- `AGENTS.md` - Added User Role Management API section

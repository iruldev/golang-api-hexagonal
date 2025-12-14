# Story 15.1: Create Admin API Route Group

Status: Done

## Story

As a developer,
I want a separate `/admin` API group,
so that sensitive management endpoints are isolated and protected by role-based access.

## Acceptance Criteria

1. **Given** `/admin` routes are defined
   **When** accessing without authentication
   **Then** 401 Unauthorized is returned

2. **Given** `/admin` routes are defined
   **When** accessing with valid authentication but without `admin` role
   **Then** 403 Forbidden is returned immediately

3. **Given** requests to `/admin` routes
   **When** user has `admin` role
   **Then** request proceeds to the protected handler

4. **Given** `/admin` route group
   **When** routes are registered
   **Then** they are mounted under `/admin` prefix (NOT under `/api/v1`)

5. **Given** the admin route infrastructure
   **When** a developer adds new admin endpoints (future stories)
   **Then** they follow a documented pattern in AGENTS.md

## Tasks / Subtasks

- [x] Create Admin Router Function
  - [x] Add `RegisterAdminRoutes(r chi.Router)` function in `internal/interface/http/routes_admin.go`
  - [x] Include comment documentation explaining admin route registration pattern
- [x] Mount Admin Routes with RBAC
  - [x] In `router.go`, mount `/admin` route group with authentication + RequireRole("admin")
  - [x] Apply global middleware chain (Recovery, RequestID, SecurityHeaders, Metrics, Otel, Logging)
  - [x] Apply AuthMiddleware BEFORE RequireRole
  - [x] Ensure middleware order follows architecture.md security chain
- [x] Implement Placeholder Handler
  - [x] Create `internal/interface/http/admin/handler.go` with stub handlers
  - [x] Add `GET /admin/health` as a simple test endpoint returning admin-specific health info
- [x] Write Unit Tests
  - [x] Test 401 when no token provided
  - [x] Test 403 when valid token but no `admin` role
  - [x] Test 200 when valid token with `admin` role
  - [x] Test middleware order (auth before RBAC)
- [x] Documentation
  - [x] Update AGENTS.md with "Adding Admin Endpoints" guide
  - [x] Add admin routes section to docs/architecture.md (covered in AGENTS.md)

## Dev Notes

### Architecture Patterns

- **Location**: Admin routes use `internal/interface/http/admin/` for handlers, distinct from regular API handlers
- **Route Prefix**: Use `/admin` at root level (NOT `/api/v1/admin`) to clearly separate from versioned API
- **Middleware Chain Order** (from architecture.md):
  1. Recovery
  2. Request ID
  3. Security Headers
  4. Metrics
  5. OTEL
  6. Logging
  7. **Authentication** (CRITICAL: Must be applied BEFORE authorization)
  8. **Authorization** (RequireRole)
  9. Handler

### RBAC Pattern Reference

From `internal/interface/http/middleware/authorization.go`:

```go
// RequireRole checks if authenticated user has at least one of the required roles
// Uses OR logic: user needs ANY of the roles to pass
func RequireRole(roles ...string) func(http.Handler) http.Handler

// Usage pattern for admin routes:
r.Route("/admin", func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(deps.Authenticator))
    r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
    RegisterAdminRoutes(r)
})
```

### Existing Patterns to Follow

From `router.go`:
- `RegisterRoutes(r)` pattern for route registration
- `RouterDeps` struct for dependency injection
- Middleware application order in global chain

From Story 14.2 (RBAC Middleware):
- `middleware.RequireRole()` uses OR logic
- Returns 403 Forbidden when role check fails
- Returns 500 Internal Server Error when claims missing (shouldn't happen if AuthMiddleware applied)

### Important: Authenticator Dependency

The admin routes MUST have `deps.Authenticator != nil` to function. If no authenticator is configured:
- Option A: Skip admin route registration entirely (preferred for dev simplicity)
- Option B: Return 501 Not Implemented for all admin routes

### Response Format

Use standard response envelope from `internal/interface/http/response/`:

```go
response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Admin access required")
```

### Security Considerations

1. **Fail-closed**: No fallback to allow access if auth/RBAC fails
2. **No sensitive data in logs**: Do not log admin actions with PII
3. **Audit logging**: Consider adding audit logs for admin actions (optional for this story, required for future admin stories)

### Router Integration Pattern

```go
// In router.go - Add after API v1 routes
// Admin routes - separated from versioned API (Story 15.1)
if deps.Authenticator != nil {
    r.Route("/admin", func(r chi.Router) {
        r.Use(middleware.AuthMiddleware(deps.Authenticator))
        r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
        RegisterAdminRoutes(r)
    })
}
```

### References

- [Epic 15: Admin/Backoffice API](file:///docs/epics.md#epic-15-admin--backoffice-api)
- [Architecture: RBAC Authorization](file:///docs/architecture.md#authorization-patterns-rbac)
- [AGENTS.md: RBAC Middleware](file:///AGENTS.md#rbac-authorization-middleware)
- [Story 14.2: RBAC Middleware Implementation](file:///docs/sprint-artifacts/14-2-implement-rbac-middleware.md)

## Dev Agent Record

### Context Reference

- `docs/epics.md` - Requirements source (Epic 15)
- `docs/architecture.md` - Security patterns, middleware chain order
- `internal/interface/http/router.go` - Existing router patterns
- `internal/interface/http/middleware/authorization.go` - RBAC middleware
- `internal/domain/auth/rbac.go` - Role constants

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- Created `routes_admin.go` with `RegisterAdminRoutes` function and comprehensive documentation
- Created `admin/handler.go` with `HealthHandler` returning admin-specific health info
- Mounted `/admin` routes in `router.go` with Auth + RBAC middleware (correct order)
- Created 5 unit tests covering 401, 403, 200, middleware order, and handler behavior
- Added "Adding Admin Endpoints" section to AGENTS.md with step-by-step guide
- All 35 test packages pass with no regressions

### File List

- `internal/interface/http/routes_admin.go` [NEW]
- `internal/interface/http/admin/handler.go` [NEW]
- `internal/interface/http/admin/handler_test.go` [NEW]
- `internal/interface/http/router.go` [MODIFIED]
- `AGENTS.md` [MODIFIED]
- `internal/interface/http/router_test.go` [MODIFIED] (Added integration tests)

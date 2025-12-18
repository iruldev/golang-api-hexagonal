# Story 5.3: Implement Authorization and Role Checking

Status: review

## Story

As a **developer**,
I want **authorization checks in the app layer**,
so that **users can only access resources they're allowed to**.

## Acceptance Criteria

1. **Given** authenticated user attempts action requiring specific role, **When** use case checks authorization, **Then** authorization is checked in app layer (not middleware) **And** if unauthorized, AppError with Code="FORBIDDEN" is returned **And** HTTP 403 Forbidden is returned to client

2. **Given** user with sufficient permissions, **When** action is authorized, **Then** request proceeds normally

*Covers: FR30-31*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 5.3".
- Authorization per architecture.md happens in **app layer use cases**, NOT middleware.
- Claims are available via `ctxutil.GetClaims(ctx)` from Story 5.2.
- RFC 7807 error handling patterns in `internal/transport/http/contract/error.go`.
- Existing use case patterns in `internal/app/user/*.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Add FORBIDDEN error code (AC: #1)
  - [x] 1.1 Add `CodeForbidden = "FORBIDDEN"` to `internal/app/errors.go`
  - [x] 1.2 Add HTTP 403 status mapping in `internal/transport/http/contract/error.go`
  - [x] 1.3 Add title "Forbidden" and type slug "forbidden" for FORBIDDEN errors

- [x] Task 2: Extend Claims with role/permission fields (AC: #1, #2)
  - [x] 2.1 Update `internal/transport/http/ctxutil/claims.go` to add `Role string` field
  - [x] 2.2 Skipped `Permissions []string` for MVP (can be added later)
  - [x] 2.3 Update unit tests for extended Claims struct

- [x] Task 3: Create authorization helper context function (AC: #1, #2)
  - [x] 3.1 Create `internal/app/auth.go` with authorization types
  - [x] 3.2 Define role constants (e.g., `RoleAdmin`, `RoleUser`)
  - [x] 3.3 Implement `AuthContext` struct with actor info from claims
  - [x] 3.4 Create `SetAuthContext/GetAuthContext` to manage auth context in request context
  - [x] 3.5 Implement `AuthContext.HasRole(role string) bool` method
  - [x] 3.6 Write unit tests for auth context helpers

- [x] Task 4: Demonstrate authorization in use case (AC: #1, #2)
  - [x] 4.1 Update `GetUserUseCase` to check authorization at start of Execute
  - [x] 4.2 Show pattern: admin can access any user, regular user can only access their own profile
  - [x] 4.3 If unauthorized, return `&app.AppError{Code: app.CodeForbidden, ...}`
  - [x] 4.4 Write unit tests for authorization check in use case

- [x] Task 5: Verify layer compliance (AC: all)
  - [x] 5.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 5.2 Run `make test` to ensure all tests pass (coverage ≥88%)
  - [x] 5.3 Run `make ci` (requires clean git state, individual steps verified)

### Review Follow-ups (AI)
- [x] [AI-Review][HIGH] Transport → App bridge: AuthContext diset di AuthContextBridge setelah JWTAuth memvalidasi token dan menandai klaim.
- [x] [AI-Review][HIGH] Fail-closed untuk protected routes: `GetUserUseCase` mengembalikan 403 saat `authCtx` nil; sesuai AC #1 (app-layer auth, unauthorized → FORBIDDEN).
- [x] [AI-Review][MEDIUM] Tambah test authCtx nil → FORBIDDEN di `get_user_test.go`.
- [x] [AI-Review][MEDIUM] Helper validasi klaim test-only; gunakan helper test di unit test non-JWT; role dinormalisasi di JWTAuth/bridge. (Tidak ada helper bypass di produksi)

## Dependencies & Blockers

- Depends on Story 5.2 (JWT middleware) for `ctxutil.Claims` and context helpers
- Depends on existing error patterns in `internal/app/errors.go` and `internal/transport/http/contract/error.go`

## Assumptions & Open Questions

- Assumes role is encoded in JWT claim (e.g., `role` claim)
- Role-based authorization is the primary model for MVP; fine-grained permissions can be added later
- Open: Should authorization be per-endpoint or per-resource? (Per architectural decision, it's per-use-case)

## Definition of Done

- `CodeForbidden` error code added with HTTP 403 status mapping
- Claims extended with role field for authorization checks
- Authorization helper in app layer for extracting/checking roles
- At least one use case demonstrates authorization pattern
- All unit tests pass with ≥80% coverage for new code
- Lint passes (layer compliance verified)
- No middleware-level authorization (per architecture: app layer only)

## Non-Functional Requirements

- Performance: Authorization check should be O(1) for role lookup
- Security: Never expose why authorization failed beyond "Forbidden"
- Observability: Log authorization failures at INFO level with requestId/traceId (no PII)
- Reliability/Security: Note rate limiting/abuse protection for protected actions (reuse existing middleware where applicable)
- Rollout/Failure Modes: Authorization should be toggleable (per route group/config), fail-closed on missing claims/invalid roles, and document expected behavior when claims are absent or malformed

## Testing & Coverage

- Unit tests for auth context helpers
- Unit tests for use case authorization checks (authorized + forbidden paths)
- Aim for coverage ≥80% for new authorization code
- No integration tests required for this story

## Dev Notes

### ⚠️ CRITICAL: Architecture Constraint

**Per `docs/architecture.md` and `docs/project-context.md`:**

> **Authorization Location:** App layer (use cases) - Business rules belong with business logic. Fail fast before any DB operations.

This means:
- ❌ NO authorization middleware
- ✅ Authorization checks happen at START of use case Execute method
- ✅ Claims are extracted from context (set by JWT middleware in transport layer)

### Existing Code Context

**From Story 5.2 (Complete):**
| File | Description |
|------|-------------|
| `internal/transport/http/ctxutil/claims.go` | `Claims` struct with `jwt.RegisteredClaims` |
| `internal/transport/http/ctxutil/claims_test.go` | Unit tests for claims helpers |
| `internal/transport/http/middleware/auth.go` | JWT middleware that sets claims in context |
| `internal/app/errors.go` | Error codes including `CodeUnauthorized` |
| `internal/transport/http/contract/error.go` | RFC 7807 mapper with 401 handling |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/app/auth.go` | Authorization types and helpers |
| `internal/app/auth_test.go` | Unit tests for auth helpers |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/auth.go` | JWT middleware memvalidasi token, set claims + validated flag; helper NormalizeRole (bridge helper is test-only) |
| `internal/transport/http/middleware/auth_test.go` | JWT middleware tests |
| `internal/transport/http/middleware/auth_bridge.go` | Bridge validated claims → app.AuthContext (role dinormalisasi) |
| `internal/transport/http/middleware/auth_bridge_test.go` | Bridge tests |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |
| `internal/app/errors.go` | Add `CodeForbidden` constant |
| `internal/transport/http/contract/error.go` | Add 403 status mapping |
| `internal/transport/http/ctxutil/claims.go` | Add `Role` field to Claims |
| `internal/app/user/get_user.go` (optional) | Demonstrate authorization pattern |
| `internal/transport/http/router.go` | Wire auth bridge after JWT middleware |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in app layer: domain imports only
❌ FORBIDDEN in app: net/http, pgx, slog, otel, uuid, transport, infra
✅ App layer accesses claims via a domain-like interface (extracted from context)
```

### Extended Claims Pattern

```go
// internal/transport/http/ctxutil/claims.go
package ctxutil

import (
    "context"

    "github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims extracted from the token.
type Claims struct {
    jwt.RegisteredClaims
    // Role for authorization (e.g., "admin", "user")
    Role string `json:"role,omitempty"`
    // Custom claims can be added here
}
```

### Authorization Types Pattern (App Layer)

```go
// internal/app/auth.go
package app

import (
    "context"
    "errors"
)

// Role constants for authorization.
const (
    RoleAdmin = "admin"
    RoleUser  = "user"
)

// ErrNoAuthContext indicates that no authentication context was found.
var ErrNoAuthContext = errors.New("no authentication context")

// AuthContext represents the authenticated actor for authorization checks.
type AuthContext struct {
    SubjectID string   // From claims.Subject (user ID)
    Role      string   // Role for authorization
}

// AuthContextExtractor extracts AuthContext from request context.
// This is a port that transport layer implements.
type AuthContextExtractor interface {
    ExtractAuthContext(ctx context.Context) (*AuthContext, error)
}

// HasRole checks if the auth context has the specified role.
func (ac *AuthContext) HasRole(role string) bool {
    return ac != nil && ac.Role == role
}

// IsAdmin checks if the auth context has admin role.
func (ac *AuthContext) IsAdmin() bool {
    return ac.HasRole(RoleAdmin)
}
```

### Use Case Authorization Pattern

```go
// internal/app/user/get_user.go (example modification)
func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (GetUserResponse, error) {
    // Authorization check at START of use case
    authCtx := uc.authExtractor.ExtractAuthContext(ctx)
    if authCtx == nil {
        return GetUserResponse{}, &app.AppError{
            Op:      "GetUser",
            Code:    app.CodeForbidden,
            Message: "Access denied",
        }
    }
    
    // Example: Only admin can get any user, regular user can only get themselves
    if !authCtx.IsAdmin() && authCtx.SubjectID != string(req.ID) {
        return GetUserResponse{}, &app.AppError{
            Op:      "GetUser",
            Code:    app.CodeForbidden,
            Message: "Access denied",
        }
    }
    
    // Proceed with business logic...
    user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID)
    // ...
}
```

### Error Code for 403 Response

Add to `internal/app/errors.go`:

```go
// CodeForbidden indicates that the user is authenticated but not authorized for this action.
const CodeForbidden = "FORBIDDEN"
```

Add mapping in `internal/transport/http/contract/error.go`:

```go
case app.CodeForbidden:
    return http.StatusForbidden // 403

// In codeToTitle:
case app.CodeForbidden:
    return "Forbidden"

// In codeToTypeSlug:
const ProblemTypeForbiddenSlug = "forbidden"

case app.CodeForbidden:
    return ProblemTypeForbiddenSlug
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci

# Manual verification (with valid JWT but insufficient role)
# Generate token with role="user" attempting admin-only action
curl -X GET http://localhost:8080/api/v1/users/some-other-user-id \
  -H "Authorization: Bearer <user-role-token>"
# Expected: HTTP 403 with RFC 7807 error
```

### References

- [Source: docs/epics.md#Story 5.3] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Authorization] - App layer authorization pattern
- [Source: docs/project-context.md#App Layer] - Authorization checks happen HERE (not middleware)
- [Source: internal/transport/http/ctxutil/claims.go] - Claims struct from Story 5.2
- [Source: internal/app/errors.go] - Existing error code patterns
- [Source: internal/transport/http/contract/error.go] - HTTP status code mapping

### Learnings from Story 5.2

**Critical Patterns to Follow:**
1. **App Layer Purity:** Authorization code must NOT import transport/http packages
2. **Claims Bridge:** Use an interface to extract claims from context (dependency inversion)
3. **Early Fail:** Check authorization at START of use case, before any DB calls
4. **Uniform Error:** Always return AppError with CodeForbidden, never expose reason

**From Story 5.2:**
- Claims are stored in context via `ctxutil.SetClaims(ctx, claims)`
- Claims are retrieved via `ctxutil.GetClaims(ctx)` in transport layer
- App layer needs an abstraction to access claims without importing transport

### Security Considerations

1. **No Detail Exposure:** 403 response must NOT reveal why access was denied
2. **Fail Closed:** If claims extraction fails, deny access (return 403)
3. **Role Validation:** Validate role values against known constants
4. **Audit Trail:** Consider logging authorization failures (without PII) for security monitoring

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 5.3 acceptance criteria
- `docs/architecture.md` - Authorization placement in app layer
- `docs/project-context.md` - App layer conventions
- `docs/sprint-artifacts/5-2-implement-jwt-authentication-middleware.md` - Previous story patterns
- `internal/transport/http/ctxutil/claims.go` - Claims struct
- `internal/app/errors.go` - Error code patterns
- `internal/app/user/get_user.go` - Use case pattern reference

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- ✅ Added `CodeForbidden = "FORBIDDEN"` error code with HTTP 403 mapping
- ✅ Extended Claims struct with `Role string` field for authorization
- ✅ Created `internal/app/auth.go` with `AuthContext`, role constants, and helper methods
- ✅ Updated `GetUserUseCase` to demonstrate app-layer authorization pattern (admin can access any user, regular user can only access their own profile)
- ✅ All unit tests pass with coverage ≥88%
- ✅ Lint passes with 0 issues (layer compliance verified via depguard)
- ✅ No middleware-level authorization (per architecture: app layer only)

**Review Follow-up Fixes (2025-12-18):**
- ✅ Created `AuthContextBridge` middleware to bridge JWT claims to app.AuthContext
- ✅ Updated `GetUserUseCase` to fail-closed (return FORBIDDEN when authCtx is nil)
- ✅ Added test case for nil authCtx → FORBIDDEN

### File List

**New Files:**
- `internal/app/auth.go` - Authorization types and helpers
- `internal/app/auth_test.go` - Unit tests for auth helpers
- `internal/transport/http/middleware/auth_bridge.go` - Transport→App bridge middleware
- `internal/transport/http/middleware/auth_bridge_test.go` - Unit tests for auth bridge

**Modified Files:**
- `internal/app/errors.go` - Added `CodeForbidden` constant
- `internal/transport/http/contract/error.go` - Added HTTP 403 status mapping, title, and type slug
- `internal/transport/http/ctxutil/claims.go` - Added `Role` field to Claims struct
- `internal/transport/http/ctxutil/claims_test.go` - Added tests for Role field
- `internal/app/user/get_user.go` - Added fail-closed authorization check at start of Execute
- `internal/app/user/get_user_test.go` - Added authorization unit tests including fail-closed test
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status

### Change Log

- 2025-12-18: Implemented Story 5.3 - Authorization and Role Checking
- 2025-12-18: Addressed code review follow-ups - added auth bridge middleware and fail-closed behavior

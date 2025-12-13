# Story 10.1: Define Auth Middleware Interface

Status: Done

## Story

As a developer,
I want an auth middleware interface,
So that I can plug in different auth providers.

## Acceptance Criteria

### AC1: Auth Interface Exists
**Given** `internal/interface/http/middleware/auth.go` exists
**When** I review the interface
**Then** interface defines: `Authenticate(r *http.Request) (Claims, error)`
**And** interface is documented with usage examples

### AC2: Claims Struct Defined
**Given** the auth middleware interface
**When** I check the Claims struct
**Then** Claims includes: UserID, Roles, Permissions
**And** struct is JSON-serializable with snake_case
**And** struct includes helper methods (HasRole, HasPermission)

### AC3: Context Integration
**Given** authentication succeeds
**When** Claims are extracted
**Then** Claims are stored in request context
**And** Claims can be retrieved via `auth.FromContext(ctx)`
**And** missing claims returns sentinel error

### AC4: Error Types Defined
**Given** authentication fails
**When** error is returned
**Then** distinct error types exist (ErrUnauthenticated, ErrTokenExpired, ErrTokenInvalid)
**And** errors are sentinel errors for easy checking

---

## Tasks / Subtasks

- [x] **Task 1: Create auth package structure** (AC: #1)
  - [x] Create `internal/interface/http/middleware/auth.go`
  - [x] Define `Authenticator` interface with `Authenticate` method
  - [x] Add comprehensive documentation comments

- [x] **Task 2: Define Claims struct** (AC: #2)
  - [x] Create Claims struct with UserID, Roles, Permissions fields
  - [x] Add JSON tags with snake_case
  - [x] Implement `HasRole(role string) bool` method
  - [x] Implement `HasPermission(perm string) bool` method

- [x] **Task 3: Implement context helpers** (AC: #3)
  - [x] Create `NewContext(ctx, claims)` function
  - [x] Create `FromContext(ctx) (Claims, error)` function
  - [x] Define context key type (unexported)
  - [x] Return `ErrNoClaimsInContext` when claims missing

- [x] **Task 4: Define error types** (AC: #4)
  - [x] Sentinel errors defined in `auth.go` (consolidated)
  - [x] Define: ErrUnauthenticated, ErrTokenExpired, ErrTokenInvalid
  - [x] Define: ErrNoClaimsInContext
  - [x] Document when each error is returned

- [x] **Task 5: Create AuthMiddleware wrapper** (AC: #1, #3)
  - [x] Create `AuthMiddleware(auth Authenticator) func(http.Handler) http.Handler`
  - [x] Call Authenticate, store claims in context on success
  - [x] Return 401 on failure with appropriate error response
  - [x] Path exclusion deferred to route configuration (chi router pattern)

- [x] **Task 6: Add unit tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/interface/http/middleware/auth_test.go`
  - [x] Test Claims HasRole/HasPermission methods
  - [x] Test context helpers (NewContext, FromContext)
  - [x] Test AuthMiddleware with mock Authenticator
  - [x] Test error responses (401 for auth failures)

- [x] **Task 7: Create example usage** (AC: #1)
  - [x] Create `internal/interface/http/middleware/auth_example_test.go`
  - [x] Show interface implementation example
  - [x] Show middleware registration example
  - [x] Show claims extraction in handler

- [x] **Task 8: Update documentation** (AC: #1)
  - [x] Update AGENTS.md with auth middleware pattern
  - [x] Document middleware integration in routes

---

## Dev Notes

### Architecture Placement

```
internal/interface/http/middleware/
├── logging.go           # Existing
├── tracing.go           # Existing
├── requestid.go         # Existing
├── recovery.go          # Existing
├── auth.go              # NEW - Interface + Claims + Context
├── auth_test.go         # NEW
└── errors.go            # Auth-related errors (or in auth.go)
```

**Key:** Auth middleware follows same pattern as other middleware in `internal/interface/http/middleware/`.

---

### Interface Design

```go
// internal/interface/http/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
)

// Authenticator defines the interface for authentication providers.
// Implementations may use JWT, API keys, sessions, or other mechanisms.
type Authenticator interface {
    // Authenticate validates credentials from the request and returns claims.
    // Returns ErrUnauthenticated if authentication fails.
    // Returns ErrTokenExpired if token is valid but expired.
    // Returns ErrTokenInvalid if token format/signature is invalid.
    Authenticate(r *http.Request) (Claims, error)
}

// Claims represents authenticated user information.
type Claims struct {
    UserID      string   `json:"user_id"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    // Optional fields for extension
    Metadata    map[string]string `json:"metadata,omitempty"`
}

// HasRole checks if the claims include the specified role.
func (c Claims) HasRole(role string) bool {
    for _, r := range c.Roles {
        if r == role {
            return true
        }
    }
    return false
}

// HasPermission checks if the claims include the specified permission.
func (c Claims) HasPermission(perm string) bool {
    for _, p := range c.Permissions {
        if p == perm {
            return true
        }
    }
    return false
}
```

---

### Context Helpers

```go
// Context key type (unexported)
type contextKey string

const claimsKey contextKey = "auth_claims"

// NewContext returns a new context with the given claims.
func NewContext(ctx context.Context, claims Claims) context.Context {
    return context.WithValue(ctx, claimsKey, claims)
}

// FromContext extracts claims from context.
// Returns ErrNoClaimsInContext if claims are not present.
func FromContext(ctx context.Context) (Claims, error) {
    claims, ok := ctx.Value(claimsKey).(Claims)
    if !ok {
        return Claims{}, ErrNoClaimsInContext
    }
    return claims, nil
}
```

---

### Error Types

```go
// errors.go or in auth.go
import "errors"

var (
    // ErrUnauthenticated indicates authentication failed (invalid credentials).
    ErrUnauthenticated = errors.New("unauthenticated")
    
    // ErrTokenExpired indicates the token has expired.
    ErrTokenExpired = errors.New("token expired")
    
    // ErrTokenInvalid indicates the token format or signature is invalid.
    ErrTokenInvalid = errors.New("token invalid")
    
    // ErrNoClaimsInContext indicates claims were not found in context.
    ErrNoClaimsInContext = errors.New("no claims in context")
)
```

---

### Middleware Pattern

```go
// AuthMiddleware creates authentication middleware using the provided Authenticator.
// Authenticated claims are stored in the request context.
func AuthMiddleware(auth Authenticator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, err := auth.Authenticate(r)
            if err != nil {
                // Map error to response
                switch {
                case errors.Is(err, ErrTokenExpired):
                    httpx.WriteError(w, http.StatusUnauthorized, "ERR_TOKEN_EXPIRED", "Token has expired")
                case errors.Is(err, ErrTokenInvalid):
                    httpx.WriteError(w, http.StatusUnauthorized, "ERR_TOKEN_INVALID", "Invalid token")
                default:
                    httpx.WriteError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required")
                }
                return
            }
            
            // Store claims in context
            ctx := NewContext(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

### Middleware Order (from architecture.md)

```
1. Recovery → 2. Request ID → 3. OTEL → 4. Logging → 5. Auth → 6. Handler
```

Auth middleware is position 5, after logging so request details are captured even for auth failures.

---

### Usage Example

```go
// In router setup
func SetupRoutes(r chi.Router, auth Authenticator) {
    // Public routes (no auth required)
    r.Get("/healthz", healthHandler)
    
    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.AuthMiddleware(auth))
        r.Get("/api/v1/notes", noteHandler.List)
    })
}

// In handler
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
    claims, err := middleware.FromContext(r.Context())
    if err != nil {
        // Should not happen if middleware is applied
        httpx.WriteError(w, 500, "ERR_INTERNAL", "Internal error")
        return
    }
    
    // Use claims.UserID for filtering
    notes, _ := h.usecase.ListByUser(r.Context(), claims.UserID)
    httpx.WriteSuccess(w, notes)
}
```

---

### Previous Epic Learnings

**From Epic 9 (Async & Reliability):**
- Interface-based design enables testing with mock implementations
- Context helpers follow standard Go patterns
- Sentinel errors enable error checking with `errors.Is()`
- Table-driven tests with AAA pattern work well
- Example tests document usage

**From architecture.md:**
- Middleware order: Recovery → RequestID → OTEL → Logging → Auth → Handler
- Error codes: ERR_UNAUTHORIZED (401), ERR_FORBIDDEN (403)
- Response envelope pattern for errors

---

### Testing Strategy

```go
// Test Claims helper methods
func TestClaims_HasRole(t *testing.T) {
    tests := []struct {
        name     string
        claims   Claims
        role     string
        expected bool
    }{
        {"has role", Claims{Roles: []string{"admin", "user"}}, "admin", true},
        {"no role", Claims{Roles: []string{"user"}}, "admin", false},
        {"empty roles", Claims{Roles: nil}, "admin", false},
    }
    // ...
}

// Test AuthMiddleware
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        mockAuth       Authenticator
        expectedStatus int
    }{
        {"success", &mockAuth{claims: validClaims}, 200},
        {"unauthorized", &mockAuth{err: ErrUnauthenticated}, 401},
        {"expired", &mockAuth{err: ErrTokenExpired}, 401},
    }
    // ...
}
```

---

### Testing Requirements

1. **Unit Tests:**
   - Test Claims struct methods (HasRole, HasPermission)
   - Test context helpers (NewContext, FromContext, missing claims)
   - Test AuthMiddleware with mock Authenticator
   - Test error responses match expected status codes

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/interface/http/middleware/auth.go` - Interface, Claims, context helpers
- `internal/interface/http/middleware/auth_test.go` - Unit tests
- `internal/interface/http/middleware/auth_example_test.go` - Example usage

**Modify:**
- `AGENTS.md` - Add auth middleware pattern section

---

### Project Structure Notes

- Alignment with unified project structure: ✅
- Follows hexagonal architecture: ✅
- Middleware in `internal/interface/http/middleware/` (consistent with existing)
- Uses existing httpx response helpers

---

### References

- [Source: docs/epics.md#Epic-10-Story-10.1] - Story requirements
- [Source: docs/architecture.md#Security-Baseline] - Middleware order
- [Source: docs/architecture.md#Error-Codes] - Error code mapping
- [Source: internal/interface/http/middleware/] - Existing middleware patterns
- [Source: docs/sprint-artifacts/9-4-add-idempotency-key-pattern.md] - Interface patterns from prior story

---

## Dev Agent Record

### Context Reference

Previous stories: 
- `docs/sprint-artifacts/9-4-add-idempotency-key-pattern.md` (interface patterns)
- `docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md` (context helpers)

Architecture: `docs/architecture.md`
Extension Interfaces: `docs/architecture.md#Extension-Interfaces`

### Agent Model Used

Claude 3.5 Sonnet (via Gemini)

### Debug Log References

### Completion Notes List

- ✅ Implemented `Authenticator` interface with comprehensive documentation
- ✅ Created `Claims` struct with `HasRole()` and `HasPermission()` helpers
- ✅ Implemented context helpers (`NewContext`, `FromContext`) following Go patterns
- ✅ Defined 4 sentinel errors for auth failure scenarios
- ✅ Created `AuthMiddleware` wrapper with JSON error response envelope
- ✅ 15+ unit tests covering all components (all passing)
- ✅ 5 example tests demonstrating usage patterns
- ✅ Updated AGENTS.md with auth middleware section
- ℹ️ Consolidated errors in `auth.go` instead of separate `errors.go` (simpler for consumers)
- ℹ️ Path exclusion handled via chi router groups (standard pattern)

### File List

**Created:**
- `internal/interface/http/middleware/auth.go`
- `internal/interface/http/middleware/auth_test.go`
- `internal/interface/http/middleware/auth_example_test.go`

**Modified:**
- `AGENTS.md`
- `docs/sprint-artifacts/sprint-status.yaml`

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story created with comprehensive developer context |
| 2025-12-13 | Implemented all tasks, all tests pass, marked Ready for Review |
| 2025-12-13 | Code review completed - 5 issues fixed (use response package, add Content-Type test, remove custom contains, fix placeholder). Story marked Done. |

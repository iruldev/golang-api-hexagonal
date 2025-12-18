# Story 5.2: Implement JWT Authentication Middleware

Status: done

## Story

As a **developer**,
I want **JWT authentication middleware with deterministic time handling**,
so that **I can protect endpoints and test reliably**.

## Acceptance Criteria

1. **Given** a request to protected endpoint without `Authorization` header, **When** middleware processes the request, **Then** HTTP 401 Unauthorized is returned **And** RFC 7807 error with Code="UNAUTHORIZED"

2. **Given** a request with invalid JWT token (malformed, wrong signature), **When** middleware validates the token, **Then** HTTP 401 Unauthorized is returned **And** Code="UNAUTHORIZED" (no detail exposed to client)

3. **Given** middleware with injected `Clock` / `Now func() time.Time`, **When** validating claim `exp`, **Then** expiry decision uses injected `now()`, NOT `time.Now()` directly

4. **Given** token with `exp` < `now()`, **When** request is processed, **Then** HTTP 401 Unauthorized is returned **And** Code="UNAUTHORIZED"

5. **Given** a request with valid JWT token, **When** middleware validates the token, **Then** claims are extracted and stored in request context **And** request proceeds to handler **And** claims are accessible via `ctxutil.GetClaims(ctx)`

6. **And** JWT secret is loaded from environment variable `JWT_SECRET`

7. **And** supported algorithms: HS256

*Covers: FR28-29*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 5.2".
- JWT library: `github.com/golang-jwt/jwt/v5` (per architecture.md)
- Existing middleware patterns in `internal/transport/http/middleware/`.
- RFC 7807 error handling in `internal/transport/http/contract/error.go`.
- Router with middleware stack in `internal/transport/http/router.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Add JWT configuration to config (AC: #6)
  - [x] 1.1 Update `internal/infra/config/config.go` to add `JWTSecret string` field (required for protected mode)
  - [x] 1.2 Add `JWTEnabled bool` field (default: false) so authentication can be toggled
  - [x] 1.3 Document in `.env.example`
  - [x] 1.4 Add validation: if `JWTEnabled=true`, `JWTSecret` must be set

- [x] Task 2: Add UNAUTHORIZED error code (AC: #1, #2, #4)
  - [x] 2.1 Add `CodeUnauthorized = "UNAUTHORIZED"` to `internal/app/errors.go`
  - [x] 2.2 Add HTTP 401 status mapping in `internal/transport/http/contract/error.go`
  - [x] 2.3 Add title and type slug for UNAUTHORIZED errors

- [x] Task 3: Create claims types and context helpers (AC: #5)
  - [x] 3.1 Create `internal/shared/ctxutil/claims.go`
  - [x] 3.2 Define `Claims` struct with standard JWT claims + custom claims
  - [x] 3.3 Implement `SetClaims(ctx, claims)` and `GetClaims(ctx)` functions
  - [x] 3.4 Write unit tests for context helpers

- [x] Task 4: Implement JWT Authentication Middleware (AC: #1-5, #7)
  - [x] 4.1 Create `internal/transport/http/middleware/auth.go`
  - [x] 4.2 Implement `JWTAuth(secret []byte, now func() time.Time)` middleware factory
  - [x] 4.3 Extract token from `Authorization: Bearer <token>` header
  - [x] 4.4 Validate token using `golang-jwt/jwt/v5` with HS256 algorithm
  - [x] 4.5 Use injected `now` function for expiry validation
  - [x] 4.6 Store claims in context on success
  - [x] 4.7 Return 401 with RFC 7807 error on failure (no detail exposed)

- [x] Task 5: Write unit tests (AC: all)
  - [x] 5.1 Create `internal/transport/http/middleware/auth_test.go`
  - [x] 5.2 Test missing Authorization header → 401
  - [x] 5.3 Test malformed token → 401
  - [x] 5.4 Test invalid signature → 401
  - [x] 5.5 Test expired token (mock time) → 401
  - [x] 5.6 Test valid token → claims in context, handler called
  - [x] 5.7 Test `GetClaims(ctx)` returns correct claims

- [x] Task 6: Integrate middleware into router (AC: all)
  - [x] 6.1 Update router to optionally apply JWTAuth middleware to protected routes
  - [x] 6.2 Create route group for protected endpoints (`/api/v1/*` except public)
  - [x] 6.3 Pass JWT config from main.go to router
  - [x] 6.4 Note: Keep existing user routes available for testing without auth initially

- [x] Task 7: Verify layer compliance (AC: all)
  - [x] 7.1 Run `make lint` to verify depguard rules pass
  - [x] 7.2 Run `make test` to ensure all tests pass
  - [x] 7.3 Run `make ci` to verify full local CI passes

## Dependencies & Blockers
- Depends on existing error mapping patterns in `internal/transport/http/contract/error.go` (no blocking changes expected)
- Router must remain compatible with current unauthenticated routes; rollout should allow toggling via config (`JWTEnabled`)

## Assumptions & Open Questions
- Assumes all protected endpoints will sit under a clear route group (e.g., `/api/v1` minus public health/docs)
- Assumes no multi-tenant claim parsing beyond standard claims; confirm if `sub`/`aud` custom validation is required
- Open: Should we enforce `aud`/`iss` validation now or leave for a later story?

## Definition of Done
- JWT middleware wired and optional via config (`JWTEnabled`/`JWTSecret`); defaults keep current behavior (auth off)
- RFC 7807 responses for all auth failures with `CodeUnauthorized`
- Claims set/retrievable via `ctxutil.GetClaims` in protected handlers
- Unit tests cover happy path + failure modes; integration unaffected when auth is disabled
- Docs: `.env.example` updated; router behavior documented in code comments

## Non-Functional Requirements
- Performance: middleware should add minimal overhead (single parse/validate per request)
- Observability: no PII in logs; log only metadata if added later (requestId/traceId), avoid reasons for auth failure
- Rollout: feature-flag via `JWTEnabled`; safe to disable without code changes
- Reliability/Security: consider rate limiting/abuse protection for protected routes and note expected handling for oversized tokens/headers and clock skew tolerance

## Testing & Coverage
- Unit tests for middleware and context helpers (missing header, malformed, bad signature, expired, valid)
- Aim for coverage ≥80% for new middleware and ctxutil packages
- No integration tests required for this story, but ensure existing suites still pass

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Epic 4 & Story 5.1 (Complete):**
| File | Description |
|------|-------------|
| `internal/transport/http/contract/error.go` | `WriteProblemJSON()`, `mapCodeToStatus()`, RFC 7807 types |
| `internal/transport/http/middleware/requestid.go` | Middleware pattern reference (context key pattern) |
| `internal/transport/http/middleware/body_limiter.go` | Middleware with config injection pattern |
| `internal/app/errors.go` | Error codes and `AppError` type |
| `internal/infra/config/config.go` | envconfig struct pattern |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/auth.go` | JWT authentication middleware |
| `internal/transport/http/middleware/auth_test.go` | Unit tests for auth middleware |
| `internal/transport/http/ctxutil/claims.go` | Claims context helpers (transport scope) |
| `internal/transport/http/ctxutil/claims_test.go` | Unit tests for claims helpers |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/infra/config/config.go` | Add `JWTSecret`, `JWTEnabled` fields |
| `internal/app/errors.go` | Add `CodeUnauthorized` constant |
| `internal/transport/http/contract/error.go` | Add 401 status mapping |
| `internal/transport/http/router.go` | Add optional auth middleware to protected routes |
| `.env.example` | Document JWT env vars |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in middleware: chi, stdlib (net/http), app (for error types), contract, golang-jwt/jwt/v5
❌ FORBIDDEN: pgx, uuid (direct import), direct infra imports (except config)
✅ ALLOWED in transport/ctxutil: JWT-specific claims helpers live in transport scope
```

**Middleware Layer Rules:**
- Accept configuration via constructor parameters
- Inject `now func() time.Time` for deterministic time testing
- Use `http.Handler` chi middleware pattern
- Return RFC 7807 errors via `contract.WriteProblemJSON()`

### JWT Library Usage Pattern

```go
// internal/transport/http/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    
    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// JWTAuth returns middleware that validates JWT tokens.
// The now function is injected for deterministic time testing.
func JWTAuth(secret []byte, now func() time.Time) func(http.Handler) http.Handler {
    parser := jwt.NewParser(
        jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
        jwt.WithTimeFunc(now), // Inject time function for exp/nbf validation
    )
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                writeUnauthorized(w, r)
                return
            }
            
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
                writeUnauthorized(w, r)
                return
            }
            
            tokenString := parts[1]
            
            // Parse and validate token
            claims := &ctxutil.Claims{}
            token, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
                return secret, nil
            })
            
            if err != nil || !token.Valid {
                writeUnauthorized(w, r)
                return
            }
            
            // Store claims in context
            ctx := ctxutil.SetClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
    contract.WriteProblemJSON(w, r, &app.AppError{
        Op:      "JWTAuth",
        Code:    app.CodeUnauthorized,
        Message: "Unauthorized",
    })
}
```

### Claims Context Pattern

```go
// internal/transport/http/ctxutil/claims.go
package ctxutil

import (
    "context"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

// claimsKey is the context key for storing JWT claims.
type claimsKey struct{}

// Claims represents JWT claims extracted from the token.
type Claims struct {
    jwt.RegisteredClaims
    // Custom claims can be added here
    // UserID string `json:"userId,omitempty"`
}

// SetClaims stores claims in context.
func SetClaims(ctx context.Context, claims *Claims) context.Context {
    return context.WithValue(ctx, claimsKey{}, claims)
}

// GetClaims retrieves claims from context.
// Returns nil if no claims are present.
func GetClaims(ctx context.Context) *Claims {
    if claims, ok := ctx.Value(claimsKey{}).(*Claims); ok {
        return claims
    }
    return nil
}
```

### Error Code for 401 Response

Add to `internal/app/errors.go`:

```go
const CodeUnauthorized = "UNAUTHORIZED"
```

Add mapping in `internal/transport/http/contract/error.go`:

```go
case app.CodeUnauthorized:
    return http.StatusUnauthorized // 401

// In codeToTitle:
case app.CodeUnauthorized:
    return "Unauthorized"

// In codeToTypeSlug:
case app.CodeUnauthorized:
    return "unauthorized"
```

### Config Enhancement Pattern

```go
// internal/infra/config/config.go (add to existing struct)
type Config struct {
    // ... existing fields ...
    
    // JWT Authentication
    // JWTEnabled enables JWT authentication for protected endpoints.
    JWTEnabled bool   `envconfig:"JWT_ENABLED" default:"false"`
    // JWTSecret is the secret key for JWT signing (required if JWTEnabled=true).
    JWTSecret  string `envconfig:"JWT_SECRET"`
}

// In Validate():
if c.JWTEnabled && strings.TrimSpace(c.JWTSecret) == "" {
    return fmt.Errorf("JWT_ENABLED is true but JWT_SECRET is empty")
}
```

### Test Pattern with Mock Time

```go
//go:build !integration

package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

var testSecret = []byte("test-secret-key-32-bytes-long!!")

func TestJWTAuth_MissingHeader(t *testing.T) {
    fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
    nowFunc := func() time.Time { return fixedNow }
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Error("handler should not be called")
    })
    
    middleware := JWTAuth(testSecret, nowFunc)
    wrapped := middleware(handler)
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    rr := httptest.NewRecorder()
    
    wrapped.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTAuth_ValidToken(t *testing.T) {
    fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
    nowFunc := func() time.Time { return fixedNow }
    
    // Create valid token
    claims := &ctxutil.Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(fixedNow.Add(1 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(fixedNow),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(testSecret)
    
    var gotClaims *ctxutil.Claims
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        gotClaims = ctxutil.GetClaims(r.Context())
        w.WriteHeader(http.StatusOK)
    })
    
    middleware := JWTAuth(testSecret, nowFunc)
    wrapped := middleware(handler)
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+tokenString)
    rr := httptest.NewRecorder()
    
    wrapped.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
    require.NotNil(t, gotClaims)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
    fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
    nowFunc := func() time.Time { return fixedNow }
    
    // Create expired token (expired 1 hour ago)
    claims := &ctxutil.Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(fixedNow.Add(-1 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(fixedNow.Add(-2 * time.Hour)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(testSecret)
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Error("handler should not be called for expired token")
    })
    
    middleware := JWTAuth(testSecret, nowFunc)
    wrapped := middleware(handler)
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    req.Header.Set("Authorization", "Bearer "+tokenString)
    rr := httptest.NewRecorder()
    
    wrapped.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusUnauthorized, rr.Code)
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

# Manual verification with curl
# Generate a test token (use jwt.io or a test script)
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <valid-token>"

# Test without header → 401
curl -X GET http://localhost:8080/api/v1/users

# Test with invalid token → 401
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer invalid-token"
```

### References

- [Source: docs/epics.md#Story 5.2] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Authentication & Security] - JWT middleware pattern
- [Source: docs/architecture.md#API Design Standards] - RFC 7807 error format
- [Source: docs/project-context.md#Transport Layer] - Layer rules
- [Source: internal/transport/http/middleware/requestid.go] - Middleware pattern example
- [Source: internal/transport/http/middleware/body_limiter.go] - Config injection pattern

### Learnings from Story 5.1

**Critical Patterns to Follow:**
1. **Time Injection:** Use `now func() time.Time` param for deterministic testing (from Epic 4 retro)
2. **Middleware Ordering:** Position auth middleware AFTER body limiter, BEFORE route handlers
3. **RFC 7807 Enforcement:** All errors use `WriteProblemJSON()` with proper code
4. **No Detail Exposure:** 401 responses must NOT expose why authentication failed

**From Story 5.1:**
- `BodyLimiter` pattern shows config injection via middleware factory
- Error codes added to `app/errors.go`, mappings to `contract/error.go`
- Middleware ordering documented in `router.go`

### Security Considerations

1. **Secret Key Length:** JWT secret should be at least 32 bytes for HS256
2. **Algorithm Restriction:** Only accept HS256 to prevent algorithm confusion attacks
3. **No Error Details:** Never expose why authentication failed (timing attacks, enumeration)
4. **Constant-Time Compare:** jwt library handles this internally
5. **Token Location:** Bearer token in Authorization header only (not query params)

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 5.2 acceptance criteria
- `docs/architecture.md` - JWT middleware patterns, error handling
- `docs/project-context.md` - Transport layer conventions
- `docs/sprint-artifacts/5-1-implement-request-validation-middleware.md` - Previous story patterns
- `internal/transport/http/middleware/*.go` - Existing middleware patterns
- `internal/transport/http/contract/error.go` - RFC 7807 infrastructure

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- Implemented JWT authentication middleware with deterministic time handling via injected `now` function
- Added `JWTEnabled` and `JWTSecret` config fields with validation (feature flag pattern)
- Added `CodeUnauthorized` error code with HTTP 401 status mapping and RFC 7807 compliance
- Created `internal/shared/ctxutil` package with type-safe Claims context helpers (100% coverage)
- JWT middleware validates HS256 tokens only (prevents algorithm confusion attacks)
- Router conditionally applies auth middleware to `/api/v1/*` routes when `JWTEnabled=true`
- All unit tests pass with 88.2% middleware coverage
- Lint passes with 0 issues (layer compliance verified via depguard)

### File List

**New Files:**
- `internal/shared/ctxutil/claims.go`
- `internal/shared/ctxutil/claims_test.go`
- `internal/transport/http/middleware/auth.go`
- `internal/transport/http/middleware/auth_test.go`

**Modified Files:**
- `internal/infra/config/config.go` - Added JWTEnabled, JWTSecret fields with validation
- `internal/app/errors.go` - Added CodeUnauthorized constant
- `internal/transport/http/contract/error.go` - Added 401 mapping, unauthorized type slug
- `internal/transport/http/router.go` - Added JWTConfig type, conditional auth middleware
- `cmd/api/main.go` - Pass JWTConfig to router
- `internal/transport/http/handler/integration_test.go` - Added JWTConfig parameter
- `.env.example` - Documented JWT_ENABLED and JWT_SECRET vars
- `go.mod`, `go.sum` - Added github.com/golang-jwt/jwt/v5 dependency

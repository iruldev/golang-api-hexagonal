# Story 10.2: Implement JWT Auth Middleware

Status: Done

## Story

As a developer,
I want JWT authentication middleware,
So that I can validate JWT tokens.

## Acceptance Criteria

### AC1: JWT Authenticator Implementation
**Given** `internal/interface/http/middleware/jwt.go` exists
**When** I review the implementation
**Then** `JWTAuthenticator` implements `Authenticator` interface
**And** constructor accepts secret key and optional config
**And** implementation is documented with usage examples

### AC2: Authorization Header Validation
**Given** request has valid JWT in Authorization header
**When** the middleware processes the request
**Then** claims are extracted and added to context
**And** "Bearer" prefix is stripped correctly
**And** empty/missing header returns ErrUnauthenticated

### AC3: Token Validation
**Given** a JWT token is provided
**When** the token is validated
**Then** invalid signature returns 401 with ErrTokenInvalid
**And** expired token returns 401 with ErrTokenExpired (specific error)
**And** malformed token returns 401 with ErrTokenInvalid
**And** "none" algorithm is rejected

### AC4: Claims Mapping
**Given** a valid JWT is parsed
**When** claims are extracted
**Then** standard claims (sub, exp, iat) are handled
**And** custom claims map to middleware.Claims struct
**And** roles/permissions from JWT map to Claims.Roles/Permissions

---

## Tasks / Subtasks

- [x] **Task 1: Create jwt.go with JWTAuthenticator** (AC: #1)
  - [x] Create `internal/interface/http/middleware/jwt.go`
  - [x] Define `JWTConfig` struct (SecretKey, Issuer, Audience optional)
  - [x] Define `JWTAuthenticator` struct implementing `Authenticator`
  - [x] Create `NewJWTAuthenticator(secretKey []byte, opts ...JWTOption) *JWTAuthenticator`
  - [x] Add comprehensive doc comments with usage examples

- [x] **Task 2: Implement Authenticate method** (AC: #2, #3, #4)
  - [x] Extract Authorization header, strip "Bearer " prefix
  - [x] Return `ErrUnauthenticated` for missing/empty header
  - [x] Parse token using `golang-jwt/jwt/v5`
  - [x] Validate signature with HMAC-SHA256 (HS256)
  - [x] Reject "none" algorithm explicitly
  - [x] Return `ErrTokenExpired` for expired tokens
  - [x] Return `ErrTokenInvalid` for malformed/bad signature
  - [x] Extract and map claims to `middleware.Claims`

- [x] **Task 3: Implement claims mapping** (AC: #4)
  - [x] Map `sub` claim to Claims.UserID
  - [x] Map `roles` claim (array) to Claims.Roles
  - [x] Map `permissions` claim (array) to Claims.Permissions
  - [ ] Handle custom claims mapping via config option (optional) - *Descoped: standard claims sufficient for v1*

- [x] **Task 4: Add unit tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/interface/http/middleware/jwt_test.go`
  - [x] Test valid token parsing and claims extraction
  - [x] Test expired token returns ErrTokenExpired
  - [x] Test invalid signature returns ErrTokenInvalid
  - [x] Test missing Authorization header returns ErrUnauthenticated
  - [x] Test malformed token (not JWT format)
  - [x] Test "none" algorithm rejection
  - [x] Test "Bearer " prefix handling

- [x] **Task 5: Create example usage** (AC: #1)
  - [x] Create `internal/interface/http/middleware/jwt_example_test.go`
  - [x] Show JWTAuthenticator instantiation
  - [x] Show integration with AuthMiddleware
  - [x] Show token generation for testing

- [x] **Task 6: Update documentation** (AC: #1)
  - [x] Update AGENTS.md with JWT middleware usage
  - [x] Add JWT config environment variables to docs

---

## Dev Notes

### Architecture Placement

```
internal/interface/http/middleware/
├── auth.go              # Interface, Claims, AuthMiddleware (from 10.1)
├── auth_test.go         # Tests for interface (from 10.1)
├── auth_example_test.go # Examples (from 10.1)
├── jwt.go               # NEW - JWTAuthenticator implementation
├── jwt_test.go          # NEW - JWT tests
├── jwt_example_test.go  # NEW - JWT examples
├── logging.go           # Existing
├── recovery.go          # Existing
└── ...
```

**Key:** JWTAuthenticator implements the `Authenticator` interface from `auth.go`.

---

### Implementation Design

```go
// internal/interface/http/middleware/jwt.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds JWT authenticator configuration.
type JWTConfig struct {
    SecretKey []byte
    Issuer    string // Optional: validate iss claim
    Audience  string // Optional: validate aud claim
}

// JWTOption configures the JWT authenticator.
type JWTOption func(*JWTConfig)

// WithIssuer validates the token issuer.
func WithIssuer(issuer string) JWTOption {
    return func(c *JWTConfig) {
        c.Issuer = issuer
    }
}

// JWTAuthenticator validates JWT tokens from Authorization header.
type JWTAuthenticator struct {
    config JWTConfig
}

// NewJWTAuthenticator creates a new JWT authenticator.
//
// Usage:
//
//  jwtAuth := middleware.NewJWTAuthenticator(
//      []byte(os.Getenv("JWT_SECRET")),
//      middleware.WithIssuer("my-app"),
//  )
//  r.Use(middleware.AuthMiddleware(jwtAuth))
func NewJWTAuthenticator(secretKey []byte, opts ...JWTOption) *JWTAuthenticator {
    config := JWTConfig{SecretKey: secretKey}
    for _, opt := range opts {
        opt(&config)
    }
    return &JWTAuthenticator{config: config}
}

// Authenticate implements Authenticator interface.
func (a *JWTAuthenticator) Authenticate(r *http.Request) (Claims, error) {
    // 1. Extract Authorization header
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return Claims{}, ErrUnauthenticated
    }

    // 2. Strip "Bearer " prefix
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return Claims{}, ErrUnauthenticated
    }
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    // 3. Parse and validate token
    token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
        // Validate algorithm - only allow HMAC
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, ErrTokenInvalid
        }
        return a.config.SecretKey, nil
    })

    if err != nil {
        // Map jwt library errors to our sentinel errors
        if errors.Is(err, jwt.ErrTokenExpired) {
            return Claims{}, ErrTokenExpired
        }
        return Claims{}, ErrTokenInvalid
    }

    if !token.Valid {
        return Claims{}, ErrTokenInvalid
    }

    // 4. Extract claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return Claims{}, ErrTokenInvalid
    }

    return mapJWTClaims(claims), nil
}

// mapJWTClaims converts JWT claims to middleware.Claims.
func mapJWTClaims(jwtClaims jwt.MapClaims) Claims {
    c := Claims{
        Metadata: make(map[string]string),
    }

    // Map standard claims
    if sub, ok := jwtClaims["sub"].(string); ok {
        c.UserID = sub
    }

    // Map roles (array)
    if roles, ok := jwtClaims["roles"].([]interface{}); ok {
        for _, r := range roles {
            if role, ok := r.(string); ok {
                c.Roles = append(c.Roles, role)
            }
        }
    }

    // Map permissions
    if perms, ok := jwtClaims["permissions"].([]interface{}); ok {
        for _, p := range perms {
            if perm, ok := p.(string); ok {
                c.Permissions = append(c.Permissions, perm)
            }
        }
    }

    return c
}
```

---

### Dependency

**Required:** `github.com/golang-jwt/jwt/v5`

```bash
go get github.com/golang-jwt/jwt/v5
```

---

### Previous Story Learnings (from Story 10.1)

- `Authenticator` interface is defined in `auth.go` with `Authenticate(r *http.Request) (Claims, error)`
- Sentinel errors: `ErrUnauthenticated`, `ErrTokenExpired`, `ErrTokenInvalid`, `ErrNoClaimsInContext`
- `Claims` struct: UserID (string), Roles ([]string), Permissions ([]string), Metadata (map[string]string)
- `AuthMiddleware(auth Authenticator)` wraps any Authenticator and stores claims in context
- Use `response` package for error responses (not manual JSON encoding)
- Table-driven tests with AAA pattern work well
- Example tests for documentation

---

### Testing Strategy

```go
// Test helper - generate test JWT
func generateTestJWT(claims jwt.MapClaims, secret []byte) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(secret)
    return tokenString
}

func TestJWTAuthenticator_Authenticate(t *testing.T) {
    secret := []byte("test-secret-key-at-least-32-bytes")
    auth := NewJWTAuthenticator(secret)

    tests := []struct {
        name        string
        authHeader  string
        wantErr     error
        wantUserID  string
    }{
        {
            name:       "valid token",
            authHeader: "Bearer " + generateTestJWT(jwt.MapClaims{
                "sub": "user-123",
                "exp": time.Now().Add(time.Hour).Unix(),
            }, secret),
            wantErr:    nil,
            wantUserID: "user-123",
        },
        {
            name:       "missing header",
            authHeader: "",
            wantErr:    ErrUnauthenticated,
        },
        {
            name:       "expired token",
            authHeader: "Bearer " + generateTestJWT(jwt.MapClaims{
                "sub": "user-123",
                "exp": time.Now().Add(-time.Hour).Unix(),
            }, secret),
            wantErr:    ErrTokenExpired,
        },
        {
            name:       "invalid signature",
            authHeader: "Bearer " + generateTestJWT(jwt.MapClaims{
                "sub": "user-123",
                "exp": time.Now().Add(time.Hour).Unix(),
            }, []byte("wrong-secret")),
            wantErr:    ErrTokenInvalid,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/test", nil)
            if tt.authHeader != "" {
                req.Header.Set("Authorization", tt.authHeader)
            }

            claims, err := auth.Authenticate(req)
            // Assert...
        })
    }
}
```

---

### Testing Requirements

1. **Unit Tests:**
   - Test all error scenarios (missing header, expired, invalid signature, malformed)
   - Test claims extraction (sub, roles, permissions)
   - Test "Bearer" prefix handling
   - Test algorithm validation (reject "none")

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/interface/http/middleware/jwt.go` - JWTAuthenticator implementation
- `internal/interface/http/middleware/jwt_test.go` - Unit tests
- `internal/interface/http/middleware/jwt_example_test.go` - Example usage

**Modify:**
- `AGENTS.md` - Add JWT middleware section
- `go.mod` - Add jwt/v5 dependency

---

### Project Structure Notes

- Alignment with unified project structure: ✅
- Follows hexagonal architecture: ✅
- JWTAuthenticator in middleware package implements interface from same package
- Uses existing `response` package for error responses

---

### Security Considerations

- Secret key should be at least 32 bytes for HMAC-SHA256
- Never log token contents or secret key
- Validate algorithm to prevent "none" algorithm attack
- Consider token issuer/audience validation for production

---

### References

- [Source: docs/epics.md#Story-10.2] - Story requirements
- [Source: docs/architecture.md#Security-Baseline] - Middleware order
- [Source: internal/interface/http/middleware/auth.go] - Authenticator interface
- [Source: docs/sprint-artifacts/10-1-define-auth-middleware-interface.md] - Previous story patterns
- [Pkg: github.com/golang-jwt/jwt/v5] - JWT library

---

## Dev Agent Record

### Context Reference

Previous story: `docs/sprint-artifacts/10-1-define-auth-middleware-interface.md`
Interface source: `internal/interface/http/middleware/auth.go`

### Agent Model Used

Claude 3.5 Sonnet (via Gemini)

### Debug Log References

### Completion Notes List

- ✅ Implemented `JWTAuthenticator` with HMAC-SHA256 validation
- ✅ Created `JWTConfig` struct with SecretKey, Issuer, Audience fields
- ✅ Implemented functional options pattern (WithIssuer, WithAudience)
- ✅ Authenticate method extracts/validates JWT from Authorization header
- ✅ Proper error mapping: ErrUnauthenticated, ErrTokenExpired, ErrTokenInvalid
- ✅ "none" algorithm explicitly rejected for security
- ✅ Claims mapping: sub→UserID, roles→Roles, permissions→Permissions
- ✅ 34 unit tests covering all AC scenarios (all passing)
- ✅ 5 example tests demonstrating usage patterns
- ✅ AGENTS.md updated with JWT middleware section and env variables
- ✅ Added `golang-jwt/jwt/v5` dependency
- ✅ [Code Review] Added secret key length validation (≥32 bytes) with `ErrSecretKeyTooShort`
- ✅ [Code Review] Added 4 new tests for secret key validation (nil, empty, short, exact 32 bytes)
- ✅ [Code Review] Added defensive programming comment for token.Valid check
- ✅ [Code Review] Updated AGENTS.md with error return signature

### File List

**Created:**
- `internal/interface/http/middleware/jwt.go`
- `internal/interface/http/middleware/jwt_test.go`
- `internal/interface/http/middleware/jwt_example_test.go`

**Modified:**
- `AGENTS.md`
- `go.mod` (added jwt/v5 dependency)
- `go.sum` (updated)

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story created with comprehensive developer context |
| 2025-12-13 | Implemented all tasks, all tests pass (60 middleware tests), marked Ready for Review |
| 2025-12-13 | Code Review: Added secret key validation, 4 new tests, fixed Task 3.4 status, 68 tests passing |

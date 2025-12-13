# Story 10.3: Implement API Key Auth Middleware

Status: Done

## Story

As a developer,
I want API key authentication middleware,
So that I can support service-to-service auth.

## Acceptance Criteria

### AC1: API Key Authenticator Implementation
**Given** `internal/interface/http/middleware/apikey.go` exists
**When** I review the implementation
**Then** `APIKeyAuthenticator` implements `Authenticator` interface
**And** constructor accepts a key validator function/interface
**And** implementation is documented with usage examples

### AC2: X-API-Key Header Validation
**Given** request has valid API key in `X-API-Key` header
**When** the middleware processes the request
**Then** service identity is added to context via Claims
**And** empty/missing header returns `ErrUnauthenticated`
**And** header name is configurable (default: `X-API-Key`)

### AC3: Key Validation
**Given** an API key is provided
**When** the key is validated
**Then** invalid key returns 401 with `ErrTokenInvalid`
**And** revoked key returns 401 with appropriate error
**And** key lookup is pluggable (env, DB, external)

### AC4: Service Identity Mapping
**Given** a valid API key is validated
**When** service identity is extracted
**Then** Claims.UserID is set to service identifier
**And** Claims.Roles includes "service" role
**And** Optional metadata (permissions, scopes) is mapped

---

## Tasks / Subtasks

- [x] **Task 1: Create apikey.go with APIKeyAuthenticator** (AC: #1)
  - [x] Create `internal/interface/http/middleware/apikey.go`
  - [x] Define `APIKeyConfig` struct (HeaderName, Validator)
  - [x] Define `KeyValidator` interface: `Validate(ctx, key string) (*KeyInfo, error)`
  - [x] Define `KeyInfo` struct: ServiceID, Roles, Permissions, Metadata
  - [x] Define `APIKeyAuthenticator` struct implementing `Authenticator`
  - [x] Create `NewAPIKeyAuthenticator(validator KeyValidator, opts ...APIKeyOption) (*APIKeyAuthenticator, error)`
  - [x] Add comprehensive doc comments with usage examples

- [x] **Task 2: Implement Authenticate method** (AC: #2, #3, #4)
  - [x] Extract API key from configurable header (default: `X-API-Key`)
  - [x] Return `ErrUnauthenticated` for missing/empty header
  - [x] Call validator to validate key
  - [x] Return `ErrTokenInvalid` for invalid keys
  - [x] Map `KeyInfo` to `middleware.Claims`
  - [x] Set "service" role in Claims.Roles

- [x] **Task 3: Create pluggable validators** (AC: #3)
  - [x] Implement `EnvKeyValidator` - validates against env var (`API_KEYS`)
  - [x] Implement `MapKeyValidator` - validates against in-memory map (for testing)
  - [x] Document interface for DB/external validators

- [x] **Task 4: Add unit tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/interface/http/middleware/apikey_test.go`
  - [x] Test valid key extraction and claims mapping
  - [x] Test missing X-API-Key header returns ErrUnauthenticated
  - [x] Test invalid key returns ErrTokenInvalid
  - [x] Test custom header name option
  - [x] Test EnvKeyValidator with mock env
  - [x] Test MapKeyValidator

- [x] **Task 5: Create example usage** (AC: #1)
  - [x] Create `internal/interface/http/middleware/apikey_example_test.go`
  - [x] Show APIKeyAuthenticator instantiation with EnvKeyValidator
  - [x] Show APIKeyAuthenticator with custom MapKeyValidator
  - [x] Show integration with AuthMiddleware
  - [x] Show custom header name configuration

- [x] **Task 6: Update documentation** (AC: #1)
  - [x] Update AGENTS.md with API Key middleware section
  - [x] Add API_KEYS environment variable documentation

---

## Dev Notes

### Architecture Placement

```
internal/interface/http/middleware/
├── auth.go              # Interface, Claims, AuthMiddleware (from 10.1)
├── auth_test.go         # Tests for interface (from 10.1)
├── jwt.go               # JWTAuthenticator (from 10.2)
├── jwt_test.go          # JWT tests (from 10.2)
├── apikey.go            # NEW - APIKeyAuthenticator implementation
├── apikey_test.go       # NEW - API Key tests
├── apikey_example_test.go # NEW - API Key examples
├── logging.go           # Existing
├── recovery.go          # Existing
└── ...
```

**Key:** APIKeyAuthenticator implements the `Authenticator` interface from `auth.go`, same pattern as JWTAuthenticator.

---

### Implementation Design

```go
// internal/interface/http/middleware/apikey.go
package middleware

import (
    "context"
    "net/http"
)

// DefaultAPIKeyHeader is the default header name for API key.
const DefaultAPIKeyHeader = "X-API-Key"

// KeyInfo contains information about a validated API key.
type KeyInfo struct {
    ServiceID   string            // Unique service identifier
    Roles       []string          // Assigned roles (e.g., "service", "admin")
    Permissions []string          // Assigned permissions
    Metadata    map[string]string // Additional metadata
}

// KeyValidator validates API keys and returns service information.
// Implementations can use environment variables, database, or external services.
type KeyValidator interface {
    // Validate checks if the API key is valid and returns service info.
    // Returns nil KeyInfo and error if key is invalid.
    Validate(ctx context.Context, key string) (*KeyInfo, error)
}

// APIKeyConfig holds API key authenticator configuration.
type APIKeyConfig struct {
    HeaderName string       // Header to read key from (default: X-API-Key)
    Validator  KeyValidator // Key validator implementation
}

// APIKeyOption configures the API key authenticator.
type APIKeyOption func(*APIKeyConfig)

// WithHeaderName sets a custom header name for the API key.
func WithHeaderName(name string) APIKeyOption {
    return func(c *APIKeyConfig) {
        c.HeaderName = name
    }
}

// APIKeyAuthenticator validates API keys from request headers.
// It implements the Authenticator interface.
//
// Example usage:
//
//  validator := middleware.NewEnvKeyValidator("API_KEYS")
//  apiAuth, err := middleware.NewAPIKeyAuthenticator(validator)
//  if err != nil {
//      log.Fatal(err)
//  }
//  r.Use(middleware.AuthMiddleware(apiAuth))
type APIKeyAuthenticator struct {
    config APIKeyConfig
}

// NewAPIKeyAuthenticator creates a new API key authenticator.
func NewAPIKeyAuthenticator(validator KeyValidator, opts ...APIKeyOption) (*APIKeyAuthenticator, error) {
    if validator == nil {
        return nil, errors.New("validator is required")
    }
    config := APIKeyConfig{
        HeaderName: DefaultAPIKeyHeader,
        Validator:  validator,
    }
    for _, opt := range opts {
        opt(&config)
    }
    return &APIKeyAuthenticator{config: config}, nil
}

// Authenticate implements the Authenticator interface.
func (a *APIKeyAuthenticator) Authenticate(r *http.Request) (Claims, error) {
    apiKey := r.Header.Get(a.config.HeaderName)
    if apiKey == "" {
        return Claims{}, ErrUnauthenticated
    }

    keyInfo, err := a.config.Validator.Validate(r.Context(), apiKey)
    if err != nil {
        return Claims{}, ErrTokenInvalid
    }

    return mapKeyInfoToClaims(keyInfo), nil
}

func mapKeyInfoToClaims(info *KeyInfo) Claims {
    roles := info.Roles
    if roles == nil {
        roles = []string{"service"} // Default role for API key auth
    }
    return Claims{
        UserID:      info.ServiceID,
        Roles:       roles,
        Permissions: info.Permissions,
        Metadata:    info.Metadata,
    }
}
```

---

### Validator Implementations

```go
// EnvKeyValidator validates API keys from environment variable.
// Keys are stored as comma-separated "key:service_id" pairs.
//
// Example: API_KEYS="abc123:svc-payments,xyz789:svc-inventory"
type EnvKeyValidator struct {
    keys map[string]string // key -> serviceID
}

// NewEnvKeyValidator creates a validator from environment variable.
func NewEnvKeyValidator(envVar string) *EnvKeyValidator {
    v := &EnvKeyValidator{keys: make(map[string]string)}
    rawKeys := os.Getenv(envVar)
    if rawKeys != "" {
        for _, pair := range strings.Split(rawKeys, ",") {
            parts := strings.SplitN(pair, ":", 2)
            if len(parts) == 2 {
                v.keys[parts[0]] = parts[1]
            }
        }
    }
    return v
}

func (v *EnvKeyValidator) Validate(ctx context.Context, key string) (*KeyInfo, error) {
    serviceID, ok := v.keys[key]
    if !ok {
        return nil, ErrTokenInvalid
    }
    return &KeyInfo{
        ServiceID: serviceID,
        Roles:     []string{"service"},
    }, nil
}

// MapKeyValidator validates against an in-memory map (for testing).
type MapKeyValidator struct {
    Keys map[string]*KeyInfo
}

func (v *MapKeyValidator) Validate(ctx context.Context, key string) (*KeyInfo, error) {
    info, ok := v.Keys[key]
    if !ok {
        return nil, ErrTokenInvalid
    }
    return info, nil
}
```

---

### Previous Story Learnings (from Story 10.2)

- `Authenticator` interface is defined in `auth.go` with `Authenticate(r *http.Request) (Claims, error)`
- Sentinel errors: `ErrUnauthenticated`, `ErrTokenExpired`, `ErrTokenInvalid`, `ErrNoClaimsInContext`
- `Claims` struct: UserID (string), Roles ([]string), Permissions ([]string), Metadata (map[string]string)
- `AuthMiddleware(auth Authenticator)` wraps any Authenticator and stores claims in context
- Use `response` package for error responses (not manual JSON encoding)
- Table-driven tests with AAA pattern work well
- Example tests for documentation
- **NEW (from code review):** Constructor should validate inputs and return error (e.g., `NewAPIKeyAuthenticator(validator, opts...) (*APIKeyAuthenticator, error)`)
- **NEW (from code review):** Add ErrSecretKeyTooShort-style error when validator is nil

---

### Testing Strategy

```go
func TestAPIKeyAuthenticator_Authenticate(t *testing.T) {
    validator := &MapKeyValidator{
        Keys: map[string]*KeyInfo{
            "valid-key": {ServiceID: "svc-test", Roles: []string{"service"}},
        },
    }
    auth, err := NewAPIKeyAuthenticator(validator)
    if err != nil {
        t.Fatalf("failed to create authenticator: %v", err)
    }

    tests := []struct {
        name       string
        header     string
        headerVal  string
        wantErr    error
        wantUserID string
    }{
        {
            name:       "valid key",
            header:     "X-API-Key",
            headerVal:  "valid-key",
            wantErr:    nil,
            wantUserID: "svc-test",
        },
        {
            name:      "missing header",
            header:    "",
            headerVal: "",
            wantErr:   ErrUnauthenticated,
        },
        {
            name:      "invalid key",
            header:    "X-API-Key",
            headerVal: "wrong-key",
            wantErr:   ErrTokenInvalid,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/test", nil)
            if tt.header != "" {
                req.Header.Set(tt.header, tt.headerVal)
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
   - Test all error scenarios (missing header, invalid key)
   - Test claims extraction from KeyInfo
   - Test custom header name option
   - Test EnvKeyValidator parsing
   - Test MapKeyValidator behavior

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/interface/http/middleware/apikey.go` - APIKeyAuthenticator implementation
- `internal/interface/http/middleware/apikey_test.go` - Unit tests
- `internal/interface/http/middleware/apikey_example_test.go` - Example usage

**Modify:**
- `AGENTS.md` - Add API Key middleware section

---

### Project Structure Notes

- Alignment with unified project structure: ✅
- Follows hexagonal architecture: ✅
- APIKeyAuthenticator in middleware package implements interface from same package
- Uses existing `response` package for error responses
- Follows same pattern as JWTAuthenticator from Story 10.2

---

### Security Considerations

- API keys should be stored securely (env vars, secrets manager)
- Never log API keys
- Consider rate limiting for API key endpoints
- Support key rotation (multiple valid keys per service)
- Consider key expiry for long-running services

---

### References

- [Source: docs/epics.md#Story-10.3] - Story requirements
- [Source: docs/architecture.md#Security-Baseline] - Middleware order
- [Source: internal/interface/http/middleware/auth.go] - Authenticator interface
- [Source: internal/interface/http/middleware/jwt.go] - JWT implementation pattern
- [Source: docs/sprint-artifacts/10-2-implement-jwt-auth-middleware.md] - Previous story patterns

---

## Dev Agent Record

### Context Reference

Previous story: `docs/sprint-artifacts/10-2-implement-jwt-auth-middleware.md`
Interface source: `internal/interface/http/middleware/auth.go`

### Agent Model Used

Gemini (via Claude)

### Debug Log References

### Completion Notes List

- ✅ Implemented `APIKeyAuthenticator` with pluggable `KeyValidator` interface
- ✅ Created `APIKeyConfig` struct with HeaderName, Validator fields
- ✅ Implemented functional options pattern (`WithHeaderName`)
- ✅ Authenticate method extracts/validates API key from configurable header
- ✅ Proper error mapping: ErrUnauthenticated, ErrTokenInvalid
- ✅ Claims mapping: ServiceID→UserID, default "service" role
- ✅ Built-in validators: `EnvKeyValidator`, `MapKeyValidator`
- ✅ 25+ unit tests covering all AC scenarios (all passing)
- ✅ 6 example tests demonstrating usage patterns
- ✅ AGENTS.md updated with API Key middleware section
- ✅ ErrValidatorRequired sentinel error for nil validator
- ✅ All tests pass with 90.7% overall coverage
- ✅ [CODE REVIEW] Added nil KeyInfo defensive check to prevent panics
- ✅ [CODE REVIEW] Added SECURITY WARNING comments about not logging API keys
- ✅ [CODE REVIEW] Consistent Metadata map initialization (matching JWT pattern)
- ✅ [CODE REVIEW] Fixed AGENTS.md unclosed code fence
- ✅ [CODE REVIEW] Added `KeyCount()` method to EnvKeyValidator for debugging
- ✅ [CODE REVIEW] Added test for nil KeyInfo edge case

### File List

**Created:**
- `internal/interface/http/middleware/apikey.go`
- `internal/interface/http/middleware/apikey_test.go`
- `internal/interface/http/middleware/apikey_example_test.go`

**Modified:**
- `AGENTS.md`
- `docs/sprint-status.yaml`

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story created with comprehensive developer context |
| 2025-12-13 | Implemented all tasks, all tests pass (25+ API key tests), marked Ready for Review |
| 2025-12-13 | Code review completed: Fixed 6 issues (3 MEDIUM, 3 LOW), marked Done |

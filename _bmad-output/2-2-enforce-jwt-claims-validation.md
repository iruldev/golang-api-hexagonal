# Story 2.2: Enforce JWT Claims Validation

Status: done

## Story

As a **security engineer**,
I want JWT validation to require exp and validate iss/aud claims,
so that expired or misrouted tokens are rejected.

## Acceptance Criteria

1. **Given** a JWT without `exp` claim
   **When** the token is validated
   **Then** the request is rejected with 401 Unauthorized

2. **Given** JWT with non-matching `iss` or `aud`
   **When** the token is validated
   **Then** the request is rejected with 401 Unauthorized

3. **And** clock skew is configurable via `JWT_CLOCK_SKEW`

4. **And** unit tests cover all claim validation scenarios

## Tasks / Subtasks

- [x] Task 1: Add exp Required Enforcement (AC: #1)
  - [x] Added `jwt.WithExpirationRequired()` to parser options
  - [x] Tokens without exp now rejected with 401

- [x] Task 2: Add iss/aud Validation (AC: #2)
  - [x] Added `Issuer` and `Audience` fields to `JWTConfig`
  - [x] Added `jwt.WithIssuer()` and `jwt.WithAudience()` to parser
  - [x] Refactored `JWTAuth()` to accept `JWTAuthConfig` struct

- [x] Task 3: Add Clock Skew Config (AC: #3)
  - [x] Added `ClockSkew` field to `JWTConfig`
  - [x] Added `jwt.WithLeeway()` to parser options
  - [x] Default is 0 (no tolerance)

- [x] Task 4: Add Unit Tests (AC: #4)
  - [x] `TestJWTAuth_MissingExp` - Token without exp → 401
  - [x] `TestJWTAuth_WrongIssuer` - Wrong issuer → 401
  - [x] `TestJWTAuth_WrongAudience` - Wrong audience → 401
  - [x] `TestJWTAuth_ClockSkew` - Expired within/beyond skew
  - [x] `TestJWTAuth_CorrectIssuerPasses` - Correct issuer → 200

## Dev Notes

### New JWTAuthConfig Struct

```go
type JWTAuthConfig struct {
    Secret    []byte
    Logger    *slog.Logger
    Now       func() time.Time
    Issuer    string         // Optional: validate iss claim
    Audience  string         // Optional: validate aud claim
    ClockSkew time.Duration  // Optional: tolerance for expired tokens
}
```

### Updated Parser Options

```go
parserOptions := []jwt.ParserOption{
    jwt.WithValidMethods([]string{AllowedAlgorithm}),
    jwt.WithExpirationRequired(), // AC #1
    jwt.WithTimeFunc(cfg.Now),
}
if cfg.Issuer != "" {
    parserOptions = append(parserOptions, jwt.WithIssuer(cfg.Issuer))
}
if cfg.Audience != "" {
    parserOptions = append(parserOptions, jwt.WithAudience(cfg.Audience))
}
if cfg.ClockSkew > 0 {
    parserOptions = append(parserOptions, jwt.WithLeeway(cfg.ClockSkew))
}
```

### JWTConfig Extended (router.go)

| Field | Type | Description |
|-------|------|-------------|
| `Issuer` | string | Expected issuer claim |
| `Audience` | string | Expected audience claim |
| `ClockSkew` | time.Duration | Tolerance for expired tokens |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- Build: SUCCESS
- New Tests: 6 PASS (`MissingExp`, `WrongIssuer`, `WrongAudience`, `ClockSkew/2`, `CorrectIssuer`)
- Regression: 15 packages ALL PASS

### Completion Notes List

- Refactored `JWTAuth()` to accept `JWTAuthConfig` struct
- Added `jwt.WithExpirationRequired()` for AC #1
- Added `jwt.WithIssuer()` and `jwt.WithAudience()` for AC #2
- Added `jwt.WithLeeway()` for AC #3
- Extended `JWTConfig` in `router.go` with Issuer/Audience/ClockSkew
- Updated router call site to pass new config fields
- Added test helpers `testJWTAuthConfig()` and `testJWTAuthConfigWith()`
- Added 6 new test cases for claims validation

### File List

- `internal/transport/http/middleware/auth.go` - MODIFIED (refactored JWTAuth with config struct)
- `internal/transport/http/middleware/auth_test.go` - MODIFIED (added 6 new tests)
- `internal/transport/http/router.go` - MODIFIED (extended JWTConfig, updated call site)
- `internal/infra/config/config.go` - MODIFIED (added JWTIssuer, JWTAudience, JWTClockSkew)
- `cmd/api/main.go` - MODIFIED (wired JWT config fields)

### Change Log

- 2024-12-24: Refactored JWTAuth to use JWTAuthConfig struct
- 2024-12-24: Added exp required, iss, aud, leeway parser options
- 2024-12-24: Extended JWTConfig with Issuer/Audience/ClockSkew fields
- 2024-12-24: Added 6 new unit tests for claims validation
- 2024-12-24: Added configuration support and wiring for validation claims (Review Fix)

```
# Story 5.5: Implement Rate Limiting Middleware

Status: done

## Story

As a **developer**,
I want **rate limiting on API endpoints with per-user and per-IP support**,
so that **the service is protected from abuse**.

## Acceptance Criteria

1. **Given** request without JWT (unauthenticated), **When** rate limit is calculated, **Then** limiter key = resolved client IP

2. **Given** request with valid JWT, **When** rate limit is calculated, **Then** limiter key = `claims.userId` (per-user rate limiting)

3. **Given** service behind reverse proxy with `TRUST_PROXY=true`, **When** client IP is resolved, **Then** IP is extracted from `X-Forwarded-For` / `X-Real-IP`

4. **Given** `TRUST_PROXY=false` (default), **When** client IP is resolved, **Then** IP is taken from `RemoteAddr`

5. **Given** client exceeds the rate limit, **When** request is processed, **Then** HTTP 429 Too Many Requests is returned, **And** `Retry-After` header indicates when to retry, **And** RFC 7807 error with Code="RATE_LIMIT_EXCEEDED"

6. **And** rate limiting uses `go-chi/httprate`
7. **And** rate limit values are configurable via `RATE_LIMIT_RPS` environment variable

*Covers: FR33-34*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 5.5".
- Middleware patterns established in `internal/transport/http/middleware/` (e.g., `security.go`, `requestid.go`).
- RFC 7807 error mapping in `internal/transport/http/contract/error.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Add `go-chi/httprate` dependency (AC: #6)
  - [x] 1.1 Run `go get github.com/go-chi/httprate@latest`
  - [x] 1.2 Verify import in go.mod

- [x] Task 2: Add `RATE_LIMIT_EXCEEDED` error code (AC: #5)
  - [x] 2.1 Add `CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"` to `internal/app/errors.go`
  - [x] 2.2 Update `mapCodeToStatus` in `internal/transport/http/contract/error.go` to map to HTTP 429
  - [x] 2.3 Update `codeToTitle` and `codeToTypeSlug` for rate limit error
  - [x] 2.4 Add `ProblemTypeRateLimitSlug = "rate-limit-exceeded"` constant

- [x] Task 3: Add rate limit configuration (AC: #3, #4, #7)
  - [x] 3.1 Add `RateLimitRPS int` to config struct in `internal/infra/config/config.go`
  - [x] 3.2 Add `TrustProxy bool` to config struct
  - [x] 3.3 Set sensible defaults: `RATE_LIMIT_RPS` = 100, `TRUST_PROXY` = false
  - [x] 3.4 Update `.env.example` with new environment variables

- [x] Task 4: Create rate limiting middleware (AC: #1, #2, #3, #4, #5)
  - [x] 4.1 Create `internal/transport/http/middleware/ratelimit.go`
  - [x] 4.2 Implement `RateLimiter(rps int, trustProxy bool) func(next http.Handler) http.Handler`
  - [x] 4.3 Implement key function that checks for JWT claims first, falls back to IP
  - [x] 4.4 Implement IP resolution with proxy trust logic
  - [x] 4.5 Return RFC 7807 error response on rate limit exceeded with `Retry-After` header
  - [x] 4.6 Use httprate's built-in 429 handling or customize error response
  - Added RateLimitWindow constant and input validation for config (AI Review Fixes)

- [x] Task 5: Integrate middleware into router (AC: all)
  - [x] 5.1 Update `internal/transport/http/router.go` to accept rate limit config
  - [x] 5.2 Add `RateLimiter` middleware after JWT auth (so claims are available)
  - [x] 5.3 Apply rate limiting to `/api/v1` routes

- [x] Task 6: Write unit tests (AC: all)
  - [x] 6.1 Create `internal/transport/http/middleware/ratelimit_test.go`
  - [x] 6.2 Test unauthenticated requests use IP-based key
  - [x] 6.3 Test authenticated requests use user ID key
  - [x] 6.4 Test `TRUST_PROXY=true` extracts IP from `X-Forwarded-For`
  - [x] 6.5 Test `TRUST_PROXY=false` uses `RemoteAddr`
  - [x] 6.6 Test rate limit exceeded returns 429 with RFC 7807 format
  - [x] 6.7 Test `Retry-After` header is present
  - [x] 6.8 Achieve ≥80% coverage for new code (90.3% achieved)

- [x] Task 7: Verify layer compliance (AC: all)
  - [x] 7.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 7.2 Run `make test` to ensure all tests pass
  - [x] 7.3 Run `make ci` for full verification (mod-tidy check expected to fail on working tree)

- [x] Review Follow-ups (AI)
  - [x] [AI-Review][Medium] Add Config Validation for RateLimitRPS (Fixed)
  - [x] [AI-Review][Medium] Use Configurable/Constant Time Window (Fixed)
  - [x] [AI-Review][Low] Use Dynamic Retry-After Header (Fixed)
  - [x] [AI-Review][Low] Use Dynamic Retry-After Header (Fixed)

- [x] Review Follow-ups (Round 2)
  - [x] [AI-Review][Medium] Add Tests for RateLimitRPS Validation (Fixed)
  - [x] [AI-Review][Low] Add Tests for RateLimitExceeded Contract (Fixed)
  - [x] [AI-Review][Medium] Documented Proxy Fallback Behavior (Verified in Tests)

- [x] Review Follow-ups (Round 3)
  - [x] [AI-Review] Final Verification: Lint & Tests Passed (Clean Review)
## Dependencies & Blockers

- Depends on Story 5.2 (JWT Authentication) for claims extraction from context
- Uses established middleware patterns from Epic 2 and Epic 5
- go-chi/httprate is a new dependency that must be added

## Assumptions & Open Questions

- Assumes in-memory rate limiting is sufficient for MVP (no Redis needed)
- httprate handles window-based limiting; default is sliding window
- Consider: Should different endpoints have different rate limits? (AC suggests single global RPS)
- Open: Should rate limit apply to health/ready/metrics endpoints? (No per AC - only API routes)

## Definition of Done

- Rate limiting middleware created in `internal/transport/http/middleware/ratelimit.go`
- httprate dependency added to go.mod
- `RATE_LIMIT_EXCEEDED` error code added and mapped to HTTP 429
- Configuration via `RATE_LIMIT_RPS` and `TRUST_PROXY` environment variables
- Middleware integrated into router for `/api/v1` routes
- Uses user ID for authenticated requests, IP for unauthenticated
- RFC 7807 error response with `Retry-After` header on 429
- Unit tests pass with ≥80% coverage for new code
- Lint passes (layer compliance verified)

## Non-Functional Requirements

- Performance: httprate is efficient (in-memory sliding window)
- Security: Per-user limiting prevents abuse across sessions
- Observability: No specific logging required for middleware itself (429s will be logged by request logging middleware)
- Reliability: Middleware should never panic

## Testing & Coverage

- Unit tests for rate limit middleware
- Test IP-based keying (unauthenticated)
- Test user ID-based keying (authenticated with JWT claims)
- Test proxy trust settings (X-Forwarded-For vs RemoteAddr)
- Test 429 response format (RFC 7807 + Retry-After)
- Test with high request volume to trigger limit
- Aim for coverage ≥80% for new rate limiting code

## Dev Notes

### ⚠️ CRITICAL: httprate Library Usage

The `go-chi/httprate` package provides flexible rate limiting middleware for Chi routers. Key features:
- Sliding window algorithm
- Custom key functions for per-user/per-IP limiting
- Configurable limits per window

### Existing Code Context

**From Previous Stories (Middleware Patterns):**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/security.go` | Environment variable parsing patterns |
| `internal/transport/http/middleware/auth.go` | JWT claims extraction from context |
| `internal/transport/http/middleware/requestid.go` | Simple middleware pattern |
| `internal/transport/http/router.go` | Router and middleware wiring |
| `internal/transport/http/contract/error.go` | RFC 7807 error mapping |
| `internal/app/errors.go` | Error codes |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/ratelimit.go` | Rate limiting middleware |
| `internal/transport/http/middleware/ratelimit_test.go` | Unit tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/app/errors.go` | Add `CodeRateLimitExceeded` |
| `internal/transport/http/contract/error.go` | Map rate limit code to 429 |
| `internal/transport/http/router.go` | Add rate limiter to middleware chain |
| `internal/infra/config/config.go` | Add rate limit config |
| `go.mod` | Add httprate dependency |
| `.env.example` | Add new env vars |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in transport/middleware: stdlib, chi, httprate, domain
❌ FORBIDDEN: pgx, direct infra imports (except observability for tracing)
UUID v7 boundary rule: Middleware MUST NOT generate IDs
```

### Rate Limiting Implementation Pattern

```go
// internal/transport/http/middleware/ratelimit.go
package middleware

import (
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/go-chi/httprate"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
    RequestsPerSecond int  // Requests allowed per second
    TrustProxy        bool // Trust X-Forwarded-For/X-Real-IP headers
}

// RateLimiter returns middleware that limits requests per key (user ID or IP).
// For authenticated requests, limits are per-user (claims.userId).
// For unauthenticated requests, limits are per-IP.
func RateLimiter(cfg RateLimitConfig) func(http.Handler) http.Handler {
    return httprate.Limit(
        cfg.RequestsPerSecond,
        time.Second,
        httprate.WithKeyFuncs(keyFunc(cfg.TrustProxy)),
        httprate.WithLimitHandler(rateLimitExceededHandler),
    )
}

// keyFunc returns the rate limit key based on JWT claims or IP.
func keyFunc(trustProxy bool) httprate.KeyFunc {
    return func(r *http.Request) (string, error) {
        // Try to get user ID from JWT claims first
        if claims := GetClaimsFromContext(r.Context()); claims != nil {
            if claims.UserID != "" {
                return "user:" + claims.UserID, nil
            }
        }
        
        // Fallback to IP-based limiting
        ip := resolveClientIP(r, trustProxy)
        return "ip:" + ip, nil
    }
}

// resolveClientIP extracts the client IP address.
func resolveClientIP(r *http.Request, trustProxy bool) string {
    if trustProxy {
        // Check X-Forwarded-For first
        if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
            // Take first IP in comma-separated list
            if idx := strings.Index(xff, ","); idx != -1 {
                return strings.TrimSpace(xff[:idx])
            }
            return strings.TrimSpace(xff)
        }
        // Check X-Real-IP
        if xri := r.Header.Get("X-Real-IP"); xri != "" {
            return strings.TrimSpace(xri)
        }
    }
    
    // Use RemoteAddr (strip port)
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return ip
}

// rateLimitExceededHandler handles 429 responses with RFC 7807 format.
func rateLimitExceededHandler(w http.ResponseWriter, r *http.Request) {
    // Set Retry-After header (use 1 second as window is 1s)
    w.Header().Set("Retry-After", "1")
    
    // Write RFC 7807 error response
    // Use contract.WriteProblemJSON or inline the error
}
```

### Claims Access Pattern

The JWT middleware stores claims in context. Access them in rate limiter:

```go
// From internal/transport/http/middleware/auth.go
type Claims struct {
    UserID string
    // ... other fields
}

// GetClaimsFromContext extracts JWT claims from request context.
// Returns nil if no claims are present (unauthenticated request).
func GetClaimsFromContext(ctx context.Context) *Claims {
    claims, _ := ctx.Value(claimsContextKey).(*Claims)
    return claims
}
```

### Middleware Placement in Router

Rate limiting should run AFTER JWT auth (so claims are available):

```go
// internal/transport/http/router.go
r.Route("/api/v1", func(r chi.Router) {
    // JWT auth first (populates context with claims)
    if jwtConfig.Enabled {
        r.Use(middleware.JWTAuth(jwtConfig.Secret, jwtConfig.Now))
        r.Use(middleware.AuthContextBridge)
    }
    
    // Rate limiting second (uses claims for key if available)
    r.Use(middleware.RateLimiter(rateLimitConfig))
    
    // Routes...
    r.Post("/users", userHandler.CreateUser)
    // ...
})
```

### Error Code and Mapping

Add to `internal/app/errors.go`:
```go
const CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
```

Add to `internal/transport/http/contract/error.go`:
```go
const ProblemTypeRateLimitSlug = "rate-limit-exceeded"

// In mapCodeToStatus:
case app.CodeRateLimitExceeded:
    return http.StatusTooManyRequests // 429

// In codeToTitle:
case app.CodeRateLimitExceeded:
    return "Too Many Requests"

// In codeToTypeSlug:
case app.CodeRateLimitExceeded:
    return ProblemTypeRateLimitSlug
```

### Configuration

Add to `internal/infra/config/config.go`:
```go
type Config struct {
    // ... existing fields
    RateLimitRPS int  `envconfig:"RATE_LIMIT_RPS" default:"100"`
    TrustProxy   bool `envconfig:"TRUST_PROXY" default:"false"`
}
```

### Verification Commands

```bash
# Install httprate dependency
go get github.com/go-chi/httprate@latest

# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci

# Manual verification
# Start server and hit endpoint repeatedly
for i in {1..110}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/users; done
# Expected: First ~100 return 200, then 429s start appearing
```

### References

- [Source: docs/epics.md#Story 5.5] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Rate Limiting] - AR-9: go-chi/httprate
- [Source: docs/project-context.md#Transport Layer] - Middleware conventions
- [go-chi/httprate GitHub](https://github.com/go-chi/httprate)
- [Source: internal/transport/http/middleware/auth.go] - Claims context pattern

### Learnings from Previous Stories

**Critical Patterns to Follow:**
1. **Middleware Placement:** Rate limiting should run AFTER JWT auth to access claims
2. **Environment Variables:** Follow pattern in `security.go` for env parsing
3. **Error Responses:** Use existing RFC 7807 mapping pattern in `contract/error.go`
4. **Testing:** Test with httptest recorder, verify headers and status codes

**From Story 5.2 (JWT Middleware):**
- Claims are stored in context using `ctxutil.GetClaims()` pattern
- Middleware returns early with RFC 7807 error on failure

**From Story 5.4 (Security Headers):**
- Environment variable parsing with parseBool helper
- Simple middleware structure

### Security Considerations

1. **Proxy Trust:** Only enable `TRUST_PROXY=true` when behind a trusted reverse proxy
2. **User ID Key:** Per-user limiting prevents abuse across different IPs for same user
3. **IP Spoofing:** Without `TRUST_PROXY`, attacker cannot spoof IP via headers
4. **Burst Handling:** httprate uses sliding window which handles bursts gracefully

### Error Handling Alignment

- Rate limit errors use RFC 7807 format consistent with other errors
- `Retry-After` header provides client guidance per HTTP 429 spec
- Error message should not leak internal rate limit implementation details

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 5.5 acceptance criteria
- `docs/architecture.md` - AR-9 httprate requirement
- `docs/project-context.md` - Transport layer conventions
- `docs/sprint-artifacts/5-4-implement-security-headers-middleware.md` - Previous story patterns
- `internal/transport/http/middleware/auth.go` - Claims context pattern
- `internal/transport/http/contract/error.go` - RFC 7807 mapping
- `internal/app/errors.go` - Error codes

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- ✅ Added httprate v0.15.0 dependency with transitive deps (xxh3, cpuid)
- ✅ Implemented `RateLimiter` middleware with configurable RPS and proxy trust
- ✅ Key function checks JWT claims first (per-user), falls back to IP (per-IP)
- ✅ IP resolution supports X-Forwarded-For and X-Real-IP headers when TRUST_PROXY=true
- ✅ 429 responses include Retry-After header and RFC 7807 format
- ✅ Config fields added: RateLimitRPS (default 100), TrustProxy (default false)
- ✅ Router updated with RateLimitConfig parameter and middleware placement after JWT auth
- ✅ Comprehensive unit tests cover all 7 acceptance criteria (90.3% coverage)
- ✅ Lint passes with 0 issues - layer compliance verified

### File List

**New Files:**
- `internal/transport/http/middleware/ratelimit.go`
- `internal/transport/http/middleware/ratelimit_test.go`

**Modified Files:**
- `go.mod` - Added httprate dependency
- `go.sum` - Updated with new dependency hashes
- `internal/app/errors.go` - Added CodeRateLimitExceeded constant
- `internal/transport/http/contract/error.go` - Added rate limit error mapping and slug
- `internal/infra/config/config.go` - Added RateLimitRPS and TrustProxy fields
- `internal/transport/http/router.go` - Added RateLimitConfig and middleware wiring
- `cmd/api/main.go` - Updated NewRouter call with RateLimitConfig
- `internal/transport/http/handler/integration_test.go` - Updated NewRouter calls
- `.env.example` - Added RATE_LIMIT_RPS and TRUST_PROXY documentation
- `docs/sprint-artifacts/sprint-status.yaml` - Status updates

### Change Log

- 2025-12-18: Implemented Story 5.5 - Rate Limiting Middleware
  - Added go-chi/httprate v0.15.0 for sliding window rate limiting
  - Per-user rate limiting for authenticated requests (uses JWT sub claim)
  - Per-IP rate limiting for unauthenticated requests
  - Configurable via RATE_LIMIT_RPS and TRUST_PROXY environment variables
  - RFC 7807 compliant 429 responses with Retry-After header
  - Rate limiting applied to /api/v1 routes after JWT auth middleware

# Story 14.3: Add Security Headers Middleware

Status: Done

## Story

As a SRE,
I want security headers added to all HTTP responses,
so that I can protect the application and users against common attacks like XSS, clickjacking, MIM sniffing, and data injection.

## Acceptance Criteria

1. **Given** any HTTP response from the server (success or error)
   **When** I inspect the response headers
   **Then** `X-Content-Type-Options: nosniff` is present
   **And** `X-Frame-Options: DENY` is present (or `SAMEORIGIN` if justified, but default to DENY)
   **And** `X-XSS-Protection: 1; mode=block` is present
   **And** `Referrer-Policy: strict-origin-when-cross-origin` is present
   **And** `Content-Security-Policy` is present with at least `default-src 'self'` configuration

2. **Given** the Security Headers middleware is configured
   **When** it is registered in the middleware chain
   **Then** it allows other middleware (like CORS) to function correctly
   **And** it does not interfere with the response body

## Tasks / Subtasks

- [x] Implement Security Middleware
  - [x] Create `internal/interface/http/middleware/security.go`
  - [x] Define `SecurityHeaders` middleware function following the standard `func(http.Handler) http.Handler` signature
  - [x] Set the required headers on the `ResponseWriter` before calling `next.ServeHTTP`
- [x] Register Middleware
  - [x] Update `internal/interface/http/router.go`
  - [x] Add `SecurityHeaders` to the global middleware stack
  - [x] Ensure correct ordering (likely early in the chain, e.g., after RequestID/Logging)
- [x] Testing
  - [x] Create `internal/interface/http/middleware/security_test.go`
  - [x] Write unit test `TestSecurityHeaders` verifying all headers are present on the response
  - [x] Verify functionality with a mock handler

## Dev Notes

### Architecture Patterns
- **Middleware Pattern**: Follow the established pattern in `internal/interface/http/middleware/`.
- **Naming**: File should be `security.go`. Function can be `SecurityHeaders`.
- **Location**: `internal/interface/http/middleware/` is the correct package.
- **Dependencies**: Standard `net/http` package. No external security libraries needed for basic headers.

### Source Tree Components
- `internal/interface/http/middleware/security.go` (New)
- `internal/interface/http/middleware/security_test.go` (New)
- `internal/interface/http/router.go` (Modify)

### Testing Standards
- Use `httptest` to record responses.
- Use `testify/assert` to check header values.
- Table-driven tests if testing multiple configurations (though this is simple enough for a single comprehensive test).

### Project Structure Notes
- Matches strict project structure.
- No `common` or `utils` packages.

## References
- [Epic 14: Advanced Security](file:///docs/epics.md)
- [Architecture: Security Baseline](file:///docs/architecture.md#security-baseline)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)

## Dev Agent Record

### Context Reference
- `docs/epics.md`
- `docs/architecture.md`
- `docs/sprint-artifacts/14-2-implement-rbac-middleware.md` (Context on previous middleware work)

### Agent Model Used
- Antigravity

### Completion Notes List
- Implemented `SecurityHeaders` middleware in `internal/interface/http/middleware/security.go`.
- Added comprehensive unit tests in `internal/interface/http/middleware/security_test.go`.
- Registered middleware in `internal/interface/http/router.go` after `RequestID` middleware.
- Verified all security headers are set correctly.
- Added `Strict-Transport-Security` and `Permissions-Policy` headers following code review.
- Added integration test `internal/interface/http/router_test.go` to verify middleware registration.

### File List
- `internal/interface/http/middleware/security.go`
- `internal/interface/http/middleware/security_test.go`
- `internal/interface/http/router.go`
- `internal/interface/http/router_test.go`

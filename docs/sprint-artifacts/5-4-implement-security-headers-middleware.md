# Story 5.4: Implement Security Headers Middleware

Status: done

## Story

As a **developer**,
I want **security headers in all HTTP responses**,
so that **common web vulnerabilities are mitigated**.

## Acceptance Criteria

1. **Given** any HTTP response from the service (success or error), **When** response is sent, **Then** the following headers are present:
   - `X-Content-Type-Options: nosniff`
   - `X-Frame-Options: DENY`
   - `X-XSS-Protection: 1; mode=block`
   - `Strict-Transport-Security: max-age=31536000; includeSubDomains` (when HTTPS)
   - `Content-Security-Policy: default-src 'none'` (API-appropriate)
   - `Referrer-Policy: strict-origin-when-cross-origin`

2. **Given** security headers middleware is configured, **When** any request is processed, **Then** headers are applied via global middleware (first in chain)

3. **Given** an error response (4xx, 5xx), **When** response is sent, **Then** security headers are present on error responses too

*Covers: FR32*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 5.4".
- Middleware patterns established in `internal/transport/http/middleware/` (e.g., `requestid.go`).
- OWASP security headers best practices.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create security headers middleware (AC: #1, #2)
  - [x] 1.1 Create `internal/transport/http/middleware/security.go`
  - [x] 1.2 Implement `SecureHeaders(next http.Handler) http.Handler` middleware function
  - [x] 1.3 Set all required security headers in response before calling next handler
  - [x] 1.4 Add HSTS header with `includeSubDomains` directive

- [x] Task 2: Make HSTS configurable for HTTPS detection (AC: #1)
  - [x] 2.1 Add `HSTSEnabled` option or detect scheme from request/config
  - [x] 2.2 Only add `Strict-Transport-Security` header when appropriate (HTTPS or behind TLS proxy)
  - [x] 2.3 Consider `HSTS_ENABLED` environment variable or auto-detect via `X-Forwarded-Proto`

- [x] Task 3: Integrate middleware into router (AC: #2)
  - [x] 3.1 Update `internal/transport/http/router.go` to add SecureHeaders as first middleware
  - [x] 3.2 Verify middleware is applied before all other middleware

- [x] Task 4: Write unit tests (AC: #1, #2, #3)
  - [x] 4.1 Create `internal/transport/http/middleware/security_test.go`
  - [x] 4.2 Test all security headers are present on successful responses
  - [x] 4.3 Test security headers are present on error responses (4xx, 5xx)
  - [x] 4.4 Test HSTS header behavior based on configuration
  - [x] 4.5 Achieve ≥80% coverage for new code

- [x] Task 5: Verify layer compliance (AC: all)
  - [x] 5.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 5.2 Run `make test` to ensure all tests pass
  - [x] 5.3 Run `make ci` for full verification

### Review Follow-ups (AI)

- [x] [AI-Review][High] Tambahkan konfigurasi eksplisit HSTS (flag/env atau opsi) sesuai Task 2.1–2.3; saat ini hanya auto-detect TLS/X-Forwarded-Proto tanpa toggle (internal/transport/http/middleware/security.go:66-90).
- [x] [AI-Review][Medium] Perbaiki deteksi HTTPS case-insensitive dan dukung r.URL.Scheme agar HSTS selalu terpasang di lingkungan proxy yang mengirim `X-Forwarded-Proto: HTTPS` (internal/transport/http/middleware/security.go:75-90).
- [x] [AI-Review][Medium] Dokumentasikan perubahan staged yang tidak tercantum di File List: `docs/sprint-artifacts/5-3-implement-authorization-middleware.md`.

## Dependencies & Blockers

- No specific dependencies on previous stories (can be implemented independently)
- Uses established middleware patterns from Epic 2 (Story 2.2 RequestID middleware)

## Assumptions & Open Questions

- Assumes HSTS should only be enabled when running behind HTTPS/TLS termination
- Consider: Should CSP be more restrictive or configurable? (Current: `default-src 'none'` is API-appropriate)
- Open: Should we add `Permissions-Policy` header? (Not in AC, skip for MVP)

## Definition of Done

- Security headers middleware created in `internal/transport/http/middleware/security.go`
- All 6 required security headers are set on every response
- HSTS only applied when HTTPS is detected/configured
- Middleware integrated as first in router chain
- Unit tests pass with ≥80% coverage for new code
- Lint passes (layer compliance verified)
- Headers present on both success and error responses

## Non-Functional Requirements

- Performance: Minimal overhead (just header setting, O(1))
- Security: Headers must be present on ALL responses including errors
- Observability: No specific logging required for this middleware
- Reliability: Middleware should never panic

## Testing & Coverage

- Unit tests for security middleware
- Test success responses (200, 201)
- Test error responses (400, 404, 500) including HTTPS/HSTS paths
- Test HSTS with/without HTTPS detection (TLS, forwarded proto incl. comma-separated, URL scheme) and env toggle on/off
- Aim for coverage ≥80% for new security code
- Latest run: `go test ./...` (pass), `make lint` (0 issues), `ALLOW_DIRTY=1 make ci` (pass; clean-tree guard bypassed, no code/test failures)

## Dev Notes

### ⚠️ CRITICAL: Security Headers Must Be First in Chain

The security headers middleware should be the **first** middleware applied so headers are present even if later middleware or handlers fail.

```go
// internal/transport/http/router.go
r.Use(middleware.SecureHeaders) // FIRST - ensures headers on ALL responses
r.Use(middleware.RequestID)
r.Use(middleware.Logging(logger))
// ... other middleware
```

### Existing Code Context

**From Epic 2 (Middleware Patterns):**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/requestid.go` | Simple middleware pattern reference |
| `internal/transport/http/middleware/logging.go` | Middleware with dependencies |
| `internal/transport/http/router.go` | Router and middleware wiring |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/transport/http/middleware/security.go` | Security headers middleware |
| `internal/transport/http/middleware/security_test.go` | Unit tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/transport/http/router.go` | Add SecureHeaders to middleware chain |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in transport/middleware: stdlib, chi, domain
❌ FORBIDDEN: pgx, direct infra imports (except observability for tracing)
UUID v7 boundary rule: IDs are generated/parsed only at transport/infra boundaries per project-context; middleware MUST NOT generate IDs
```

### Security Headers Implementation Pattern

```go
// internal/transport/http/middleware/security.go
package middleware

import (
    "net/http"
    "os"
    "strings"
)

// SecureHeaders adds OWASP-recommended security headers to all responses.
// This middleware should be applied first in the middleware chain.
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set security headers BEFORE calling next
        h := w.Header()
        
        // Prevent MIME type sniffing
        h.Set("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        h.Set("X-Frame-Options", "DENY")
        
        // XSS protection (legacy but still recommended)
        h.Set("X-XSS-Protection", "1; mode=block")
        
        // Content Security Policy (restrictive for API)
        h.Set("Content-Security-Policy", "default-src 'none'")
        
        // Referrer policy
        h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // HSTS - when request appears to be over HTTPS or explicitly enabled
        if shouldAddHSTS(r) {
            h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        
        next.ServeHTTP(w, r)
    })
}

// HSTS toggle:
// - HSTS_ENABLED=true/1/on/yes -> force add HSTS
// - HSTS_ENABLED=false/0/off/no -> suppress HSTS only when request is NOT HTTPS; on HTTPS it remains enabled to satisfy AC
// HTTPS detection:
// - r.TLS != nil
// - X-Forwarded-Proto (comma-separated, case-insensitive; first value used)
// - r.URL.Scheme
```

### HSTS Detection Logic

HSTS should be enabled when:
1. Request is over direct TLS (`r.TLS != nil`)
2. Behind reverse proxy with `X-Forwarded-Proto: https` (comma-separated supported; first value)
3. `r.URL.Scheme == "https"`
4. Explicit enable via `HSTS_ENABLED=true/on/1/yes`; explicit disable via `HSTS_ENABLED=false/off/0/no` only suppresses HSTS on plain HTTP (HTTPS still forces HSTS)

```go
// Check for HTTPS in various ways
isHTTPS := r.TLS != nil ||
           strings.EqualFold(firstForwardedProto(r.Header.Get("X-Forwarded-Proto")), "https") ||
           strings.EqualFold(r.URL.Scheme, "https")

// Override
if env := os.Getenv("HSTS_ENABLED"); env != "" {
    if isTrue(env) { add HSTS }
    if isFalse(env) { skip HSTS }
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

# Manual verification
curl -I http://localhost:8080/health
# Expected headers in response:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# X-XSS-Protection: 1; mode=block
# Content-Security-Policy: default-src 'none'
# Referrer-Policy: strict-origin-when-cross-origin
```

### References

- [Source: docs/epics.md#Story 5.4] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Security Headers] - NFR13: Security headers present in all responses
- [Source: docs/project-context.md#Transport Layer] - Middleware conventions
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [Source: internal/transport/http/middleware/requestid.go] - Middleware pattern reference

### Learnings from Previous Stories

**Critical Patterns to Follow:**
1. **Middleware Placement:** Security headers should be FIRST in chain
2. **Response Coverage:** Headers must be on success AND error responses
3. **HTTPS Detection:** Support both direct TLS and reverse proxy scenarios
4. **Testing:** Test with httptest recorder to verify headers

**From Story 2.2 (RequestID Middleware):**
- Simple middleware pattern: `func Middleware(next http.Handler) http.Handler`
- Set response headers before `next.ServeHTTP(w, r)`
- Test using `httptest.NewRecorder()` and check `w.Header().Get()`

### Security Considerations

1. **Header Order:** Set headers BEFORE calling next handler
2. **Error Responses:** Chi's error handling will still have headers set
3. **HSTS Caution:** Only enable HSTS when actually behind HTTPS to avoid breaking HTTP development
4. **CSP for APIs:** `default-src 'none'` is appropriate for JSON APIs (no HTML/scripts)
5. **Auth/Authorization Ordering:** SecurityHeaders must run before auth/authorization middleware to ensure headers are present even on 401/403/500 responses
6. **Feature Flag / Failure Mode:** Allow disabling HSTS/security headers via config/env flag (e.g., `HSTS_ENABLED`) for local/dev; middleware should degrade gracefully without panicking

### Error Handling Alignment

- Error responses continue to use existing RFC 7807 mapping in handlers; middleware only enriches headers and must not change status/payload semantics

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 5.4 acceptance criteria
- `docs/architecture.md` - NFR13 security headers requirement  
- `docs/project-context.md` - Transport layer conventions
- `docs/sprint-artifacts/5-3-implement-authorization-middleware.md` - Previous story patterns
- `internal/transport/http/middleware/requestid.go` - Middleware pattern reference
- `internal/transport/http/router.go` - Router middleware wiring

### Agent Model Used

Gemini 2.5 Pro (Antigravity)

### Debug Log References

N/A

### Completion Notes List

- Created security headers middleware with OWASP-recommended headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, CSP, Referrer-Policy)
- Implemented HSTS auto-detection via r.TLS and X-Forwarded-Proto header
- Integrated SecureHeaders as first middleware in router chain (before RequestID)
- Comprehensive unit tests with 88.9% coverage
- All lint checks pass (0 issues)
- Headers are set BEFORE calling next handler to ensure presence even on errors/panics

### File List

- internal/transport/http/middleware/security.go (NEW)
- internal/transport/http/middleware/security_test.go (NEW)
- internal/transport/http/router.go (MODIFIED)
- docs/sprint-artifacts/sprint-status.yaml (MODIFIED)
- docs/sprint-artifacts/5-3-implement-authorization-middleware.md (STAGED DOC CHANGE; previously not listed)

### Change Log

- 2025-12-18: Implemented security headers middleware with all 6 OWASP headers, HSTS auto-detection, router integration, and unit tests (88.9% coverage)
- 2025-12-18: Noted staged documentation change outside this story scope: `docs/sprint-artifacts/5-3-implement-authorization-middleware.md`

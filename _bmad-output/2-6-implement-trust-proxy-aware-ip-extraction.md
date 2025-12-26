# Story 2.6: Implement TRUST_PROXY-Aware IP Extraction

Status: done

## Story

**As a** security engineer,
**I want** real IP extraction only when TRUST_PROXY=true,
**So that** IP spoofing via X-Forwarded-For is prevented.

**FR:** FR11, FR49, FR50

## Acceptance Criteria

1. **Given** `TRUST_PROXY=false`
   **When** request has X-Forwarded-For header
   **Then** the header is ignored, remote address is used

2. **Given** `TRUST_PROXY=true`
   **When** request has X-Forwarded-For header
   **Then** Chi RealIP middleware extracts real IP
   **And** integration test verifies both scenarios

## Tasks / Subtasks

- [x] Task 1: Make RealIP middleware conditional
  - [x] Use existing `rateLimitConfig.TrustProxy` in `NewRouter()`
  - [x] Wrap `chiMiddleware.RealIP` in conditional: only apply when `TrustProxy=true`
  - [x] Update middleware stack comment to document behavior

- [x] Task 2: Add integration tests
  - [x] `TestNewRouter_TrustProxyFalse_IgnoresXFF` - TRUST_PROXY=false ignores X-Forwarded-For
  - [x] `TestNewRouter_TrustProxyTrue_UsesXFF` - TRUST_PROXY=true extracts forwarded IP
  - [x] Verify behavior consistent with rate limiter's `extractClientIP()`

- [x] Task 3: Update documentation
  - [x] Confirm .env.example documents TRUST_PROXY (already done)
  - [x] Security note implicit in code comments

## Dev Notes

### Implementation

**File:** `internal/transport/http/router.go` line 90-94

```go
// Story 2.6: Only trust proxy headers (X-Forwarded-For, X-Real-IP) when explicitly configured.
// This prevents IP spoofing when not behind a trusted proxy.
if rateLimitConfig.TrustProxy {
    r.Use(chiMiddleware.RealIP)
}
```

### Security Context

| Scenario | Before | After |
|----------|--------|-------|
| `TRUST_PROXY=false` + XFF | RealIP trusts header ❌ | RemoteAddr used ✅ |
| `TRUST_PROXY=true` + XFF | RealIP trusts header ✅ | XFF IP used ✅ |

### Test Results

```
=== RUN   TestNewRouter_TrustProxyFalse_IgnoresXFF
--- PASS: TestNewRouter_TrustProxyFalse_IgnoresXFF (0.00s)
=== RUN   TestNewRouter_TrustProxyTrue_UsesXFF
--- PASS: TestNewRouter_TrustProxyTrue_UsesXFF (0.00s)
```

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Debug Log References
- Router tests: PASS
- Full regression: 16 packages PASS

### Completion Notes List
- Made `chiMiddleware.RealIP` conditional on `rateLimitConfig.TrustProxy`
- Added 2 unit tests to verify both scenarios
- Rate limiter already correct (uses `extractClientIP()` with same logic)
- .env.example already documents TRUST_PROXY
- **Code Review Fixes (AI):**
  - Reordered middleware: `RealIP` now runs *before* `RequestLogger` and `Metrics` to ensure correct IP logging.
  - Simplified `RateLimiter`: Removed redundant XFF parsing; now relies on global `RealIP` middleware normalization.
  - Added access logging to `NewInternalRouter`.

### File List
- `internal/transport/http/router.go` - MODIFIED (conditional RealIP)
- `internal/transport/http/router_test.go` - MODIFIED (added 2 tests)

### Change Log
- 2024-12-24: Implemented conditional RealIP middleware in router.go
- 2024-12-24: Added TestNewRouter_TrustProxyFalse_IgnoresXFF and TestNewRouter_TrustProxyTrue_UsesXFF
- 2024-12-24: Verified full regression passes (16 packages)
- 2024-12-24: Applied Code Review fixes (middleware ordering, rate limiter simplification)

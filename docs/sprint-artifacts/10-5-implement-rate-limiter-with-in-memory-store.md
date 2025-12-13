# Story 10.5: Implement Rate Limiter with In-Memory Store

Status: Done

## Story

As a developer,
I want in-memory rate limiting,
So that I can protect endpoints from abuse.

## Acceptance Criteria

### AC1: In-Memory Rate Limiter Implementation
**Given** `internal/interface/http/middleware/ratelimit.go` exists
**When** I review the code
**Then** token bucket algorithm is implemented
**And** in-memory store uses `sync.Map` or equivalent for thread-safety
**And** limiter implements `runtimeutil.RateLimiter` interface

### AC2: Rate Limit Response Handling
**Given** rate limit is exceeded
**When** client makes a request
**Then** response status is 429 Too Many Requests
**And** `Retry-After` header is set with seconds until limit reset
**And** response uses project's `response.Error` pattern with `ERR_RATE_LIMITED`

### AC3: Configurable Per-Endpoint Limits
**Given** rate limiter middleware is configured
**When** applied to different endpoints
**Then** limits are configurable per endpoint (e.g., `/api/v1/notes` = 100/min)
**And** global default limit can be set
**And** configuration uses functional options pattern

### AC4: Middleware Integration
**Given** RateLimitMiddleware is available
**When** I apply it to protected routes
**Then** middleware extracts key from request (IP, UserID, or custom)
**And** works with existing auth middleware chain
**And** key extractor is pluggable (functional option)

---

## Tasks / Subtasks

- [x] **Task 1: Implement Token Bucket Algorithm** (AC: #1)
  - [x] Create `internal/interface/http/middleware/ratelimit.go`
  - [x] Implement `TokenBucket` struct with rate, capacity, tokens, lastRefill
  - [x] Implement `Allow()` method with token consumption logic
  - [x] Add thread-safe token refill based on elapsed time
  - [x] Add doc comments with algorithm explanation

- [x] **Task 2: Create In-Memory Rate Limiter** (AC: #1, #3)
  - [x] Implement `InMemoryRateLimiter` struct with `sync.Map` store
  - [x] Implement `runtimeutil.RateLimiter` interface (Allow, Limit)
  - [x] Add bucket cleanup goroutine for expired entries
  - [x] Support configurable default rate and cleanup interval
  - [x] Implement functional options: `WithDefaultRate()`, `WithCleanupInterval()`

- [x] **Task 3: Implement Rate Limit Middleware** (AC: #2, #4)
  - [x] Create `RateLimitMiddleware(limiter RateLimiter, opts ...Option)` function
  - [x] Extract key from request (default: IP address)
  - [x] Call `limiter.Allow(ctx, key)` and handle result
  - [x] Return 429 with `Retry-After` header on limit exceeded
  - [x] Use `response.Error(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED", msg)`

- [x] **Task 4: Add Configurable Key Extraction** (AC: #4)
  - [x] Implement `WithKeyExtractor(func(*http.Request) string)` option
  - [x] Implement default extractors: `IPKeyExtractor`, `UserIDKeyExtractor`
  - [x] `UserIDKeyExtractor` uses claims from context if available
  - [x] Add fallback logic if extraction fails

- [x] **Task 5: Add Unit Tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/interface/http/middleware/ratelimit_test.go`
  - [x] Test token bucket algorithm (refill, consumption, exhaustion)
  - [x] Test 429 response with Retry-After header
  - [x] Test per-endpoint configuration
  - [x] Test key extractors (IP, UserID)
  - [x] Test concurrent access thread-safety

- [x] **Task 6: Create Example Usage** (AC: #4)
  - [x] Create `internal/interface/http/middleware/ratelimit_example_test.go`
  - [x] Show basic rate limit middleware usage
  - [x] Show per-endpoint configuration with chi router
  - [x] Show combined auth + rate limit middleware chain

- [x] **Task 7: Update Documentation** (AC: #3)
  - [x] Update AGENTS.md with rate limiting section
  - [x] Document configuration options
  - [x] Add integration examples

---

## Dev Notes

### Architecture Placement

```
internal/
├── runtimeutil/
│   └── ratelimiter.go        # EXISTING - Interface definition (FR45)
│
└── interface/http/middleware/
    ├── ratelimit.go          # NEW - Middleware + InMemory implementation
    ├── ratelimit_test.go     # NEW - Unit tests
    └── ratelimit_example_test.go # NEW - Example usage
```

**Key:** Implementation in middleware package uses interface from runtimeutil.

---

### Implementation Design

```go
// Token Bucket Algorithm
type TokenBucket struct {
    rate       float64   // tokens per second
    capacity   float64   // max tokens (burst size)
    tokens     float64   // current tokens
    lastRefill time.Time // last refill timestamp
    mu         sync.Mutex
}

func (b *TokenBucket) Allow() bool {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // Refill tokens based on time elapsed
    now := time.Now()
    elapsed := now.Sub(b.lastRefill).Seconds()
    b.tokens = min(b.capacity, b.tokens + elapsed*b.rate)
    b.lastRefill = now
    
    // Consume token if available
    if b.tokens >= 1 {
        b.tokens -= 1
        return true
    }
    return false
}
```

```go
// In-Memory Rate Limiter
type InMemoryRateLimiter struct {
    buckets     sync.Map  // map[string]*TokenBucket
    defaultRate runtimeutil.Rate
    cleanupOnce sync.Once
}

func NewInMemoryRateLimiter(opts ...Option) *InMemoryRateLimiter {
    limiter := &InMemoryRateLimiter{
        defaultRate: runtimeutil.NewRate(100, time.Minute), // 100 req/min default
    }
    for _, opt := range opts {
        opt(limiter)
    }
    limiter.startCleanup()
    return limiter
}

func (l *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
    bucket := l.getOrCreateBucket(key)
    return bucket.Allow(), nil
}
```

```go
// Middleware
func RateLimitMiddleware(limiter runtimeutil.RateLimiter, opts ...MiddlewareOption) func(http.Handler) http.Handler {
    cfg := defaultConfig()
    for _, opt := range opts {
        opt(&cfg)
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := cfg.keyExtractor(r)
            
            allowed, err := limiter.Allow(r.Context(), key)
            if err != nil {
                // Log error but allow request (fail-open)
                next.ServeHTTP(w, r)
                return
            }
            
            if !allowed {
                w.Header().Set("Retry-After", strconv.Itoa(cfg.retryAfterSeconds))
                response.Error(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED", "Rate limit exceeded")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

---

### Previous Story Learnings (from Story 10.4)

- Use functional options pattern for configuration (consistent with JWT/APIKey)
- Add comprehensive doc comments with usage examples
- Implement interface from designated package (`runtimeutil.RateLimiter`)
- Return appropriate HTTP status codes (429 for rate limit, not 401/403)
- Add `Permission.IsValid()` type methods for consistency (learned from code review)
- Security: Don't reveal internal state in error messages
- Table-driven tests with AAA pattern
- Example tests demonstrate chi router integration

---

### Existing Interface (from runtimeutil/ratelimiter.go)

```go
type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    Limit(ctx context.Context, key string, rate Rate) error
}

type Rate struct {
    Limit  int           // requests allowed
    Period time.Duration // time window
}
```

**Note:** Must implement this existing interface for consistency.

---

### Key Extractor Functions

```go
// Built-in key extractors
func IPKeyExtractor(r *http.Request) string {
    // Check X-Forwarded-For first (behind proxy)
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        return strings.Split(xff, ",")[0]
    }
    // Fall back to RemoteAddr
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    return host
}

func UserIDKeyExtractor(r *http.Request) string {
    claims, err := FromContext(r.Context())
    if err != nil {
        return IPKeyExtractor(r) // Fallback to IP
    }
    return claims.UserID
}
```

---

### Testing Strategy

```go
func TestTokenBucket_Allow(t *testing.T) {
    tests := []struct {
        name       string
        rate       float64 // tokens/sec
        capacity   float64
        requests   int
        expectPass []bool
    }{
        {
            name:       "within limit",
            rate:       10, // 10/sec
            capacity:   10,
            requests:   5,
            expectPass: []bool{true, true, true, true, true},
        },
        {
            name:       "exceeds limit",
            rate:       2,
            capacity:   2,
            requests:   3,
            expectPass: []bool{true, true, false},
        },
    }
    // ...
}
```

---

### Retry-After Header Calculation

```go
// Calculate seconds until next token is available
func (b *TokenBucket) RetryAfter() int {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if b.tokens >= 1 {
        return 0
    }
    // Time to generate 1 token
    tokensNeeded := 1.0 - b.tokens
    seconds := tokensNeeded / b.rate
    return int(math.Ceil(seconds))
}
```

---

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithDefaultRate(rate)` | 100 req/min | Default rate for new keys |
| `WithCleanupInterval(d)` | 5 min | Interval for cleaning expired buckets |
| `WithKeyExtractor(fn)` | IP address | Function to extract key from request |
| `WithRetryAfterSeconds(n)` | 60 | Seconds to return in Retry-After |

---

### File List

**Create:**
- `internal/interface/http/middleware/ratelimit.go` - Middleware + InMemory implementation
- `internal/interface/http/middleware/ratelimit_test.go` - Unit tests
- `internal/interface/http/middleware/ratelimit_example_test.go` - Example usage

**Modify:**
- `AGENTS.md` - Add rate limiting documentation section

---

### Testing Requirements

1. **Unit Tests:**
   - Test token bucket: within limit, exceeds limit, refill after time
   - Test InMemoryRateLimiter: Allow, Limit, concurrent access
   - Test middleware: 429 response, Retry-After header
   - Test key extractors: IP, UserID, custom

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### Security Considerations

- Fail-open: If rate limiter errors, allow request (don't block service)
- Don't expose internal bucket state in error messages
- Consider memory limits for in-memory store (cleanup old buckets)
- Use mutex for thread-safety
- X-Forwarded-For spoofing: Document risk when using IP extraction

---

### References

- [Source: docs/epics.md#Story-10.5] - Story requirements
- [Source: internal/runtimeutil/ratelimiter.go] - RateLimiter interface (FR45)
- [Source: docs/sprint-artifacts/10-4-create-rbac-permission-model.md] - Previous story patterns
- [Source: internal/interface/http/middleware/auth.go] - Middleware patterns

---

## Dev Agent Record

### Context Reference

Previous story: `docs/sprint-artifacts/10-4-create-rbac-permission-model.md`
RateLimiter interface: `internal/runtimeutil/ratelimiter.go`

### Agent Model Used

Gemini 2.5 Pro (via Antigravity)

### Debug Log References

N/A - No debug issues encountered

### Completion Notes List

- ✅ Created `internal/interface/http/middleware/ratelimit.go` with TokenBucket, InMemoryRateLimiter, RateLimitMiddleware
- ✅ TokenBucket implements token bucket algorithm with thread-safe refill logic
- ✅ InMemoryRateLimiter implements `runtimeutil.RateLimiter` interface (Allow, Limit, RetryAfter)
- ✅ Added automatic bucket cleanup goroutine with configurable TTL and interval
- ✅ Implemented functional options: WithDefaultRate, WithCleanupInterval, WithBucketTTL, WithKeyExtractor, WithRetryAfterSeconds
- ✅ RateLimitMiddleware returns 429 with Retry-After header on limit exceeded
- ✅ Implemented IPKeyExtractor (X-Forwarded-For, X-Real-IP, RemoteAddr fallback)
- ✅ Implemented UserIDKeyExtractor with IP fallback for unauthenticated users
- ✅ Created 24 unit tests covering all acceptance criteria (100% pass rate)
- ✅ Created 5 example tests demonstrating chi router integration
- ✅ Updated AGENTS.md with comprehensive rate limiting documentation
- ✅ All tests pass (`make test` succeeds)

### File List

**Created:**
- `internal/interface/http/middleware/ratelimit.go` - TokenBucket, InMemoryRateLimiter, RateLimitMiddleware
- `internal/interface/http/middleware/ratelimit_test.go` - Unit tests (24 tests)
- `internal/interface/http/middleware/ratelimit_example_test.go` - Example tests (5 examples)

**Modified:**
- `AGENTS.md` - Added Rate Limiting Middleware section with usage examples

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story created with comprehensive developer context |
| 2025-12-13 | Implemented TokenBucket, InMemoryRateLimiter, RateLimitMiddleware, tests, documentation; marked Done |
| 2025-12-13 | **Code Review:** Fixed race condition in `bucketEntry.lastAccess` - added mutex protection in `Allow()` and `cleanup()` |

---

## Senior Developer Review (AI)

**Reviewer:** Gemini 2.5 Pro (via Antigravity)  
**Date:** 2025-12-13

### Issues Found & Fixed

| Severity | Issue | Fix Applied |
|----------|-------|-------------|
| HIGH | Race condition in `Allow()` - `entry.lastAccess` written without mutex | ✅ Added `mu sync.Mutex` to `bucketEntry`, protected writes in `Allow()` and reads in `cleanup()` |
| MEDIUM | Story status mismatch (Done vs review) | ✅ Updated sprint-status.yaml to done |

### Verification

- ✅ `make test` passes with `-race` flag
- ✅ `TestInMemoryRateLimiter_Concurrent` passes
- ✅ All 24 unit tests pass

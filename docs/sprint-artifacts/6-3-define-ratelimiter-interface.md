# Story 6.3: Define RateLimiter Interface

Status: done

## Story

As a developer,
I want a RateLimiter interface abstraction,
So that I can implement rate limiting with different backends.

## Acceptance Criteria

### AC1: RateLimiter interface defined
**Given** `internal/runtimeutil/ratelimiter.go` exists
**When** I review the interface
**Then** methods include: Allow(key), Limit(key, rate)
**And** interface is compatible with middleware usage

---

## Tasks / Subtasks

- [x] **Task 1: Define RateLimiter interface** (AC: #1)
  - [x] Create RateLimiter interface with Allow, Limit methods
  - [x] Add context parameter for cancellation
  - [x] Define Rate struct for limits

- [x] **Task 2: Add NopRateLimiter for testing** (AC: #1)
  - [x] Create no-op implementation that always allows

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### RateLimiter Interface

```go
// internal/runtimeutil/ratelimiter.go
package runtimeutil

import (
    "context"
    "time"
)

// Rate defines a rate limit configuration.
type Rate struct {
    Limit  int           // Number of requests allowed
    Period time.Duration // Time period for the limit
}

// RateLimiter defines rate limiting abstraction.
// Compatible with middleware usage for HTTP rate limiting.
type RateLimiter interface {
    // Allow checks if the request should be allowed for the given key.
    // Returns true if allowed, false if rate limited.
    Allow(ctx context.Context, key string) (bool, error)

    // Limit sets the rate limit for the given key.
    // Use this to configure per-user or per-IP limits.
    Limit(ctx context.Context, key string, rate Rate) error
}
```

### Middleware Usage

```go
func RateLimitMiddleware(limiter runtimeutil.RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := r.RemoteAddr // or user ID
            allowed, err := limiter.Allow(r.Context(), key)
            if err != nil || !allowed {
                http.Error(w, "rate limited", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Architecture Compliance

**Layer:** `internal/runtimeutil/`
**Pattern:** Interface abstraction for rate limiting
**Benefit:** Swappable backends (Redis, in-memory, token bucket)

### References

- [Source: docs/epics.md#Story-6.3]
- [Story 6.2 - Cache Interface](file:///docs/sprint-artifacts/6-2-define-cache-interface.md)

---

## Dev Agent Record

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/runtimeutil/ratelimiter.go` - RateLimiter interface

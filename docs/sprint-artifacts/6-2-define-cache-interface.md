# Story 6.2: Define Cache Interface

Status: done

## Story

As a developer,
I want a Cache interface abstraction,
So that I can plug in Redis, Memcached, or in-memory cache.

## Acceptance Criteria

### AC1: Cache interface defined
**Given** `internal/runtimeutil/cache.go` exists
**When** I review the interface
**Then** methods include: Get, Set, Delete, Exists
**And** interface documentation explains usage

---

## Tasks / Subtasks

- [x] **Task 1: Define Cache interface** (AC: #1)
  - [x] Create Cache interface with Get, Set, Delete, Exists methods
  - [x] Add TTL support for Set method
  - [x] Add context parameter for cancellation

- [x] **Task 2: Add documentation** (AC: #1)
  - [x] Document interface usage patterns
  - [x] Document how to implement Redis adapter

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Cache Interface

```go
// internal/runtimeutil/cache.go
package runtimeutil

import (
    "context"
    "time"
)

// Cache defines caching abstraction for swappable implementations.
// Use for Redis, Memcached, or in-memory cache.
type Cache interface {
    // Get retrieves a value by key. Returns ErrCacheMiss if not found.
    Get(ctx context.Context, key string) ([]byte, error)

    // Set stores a value with optional TTL. Zero TTL means no expiration.
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

    // Delete removes a value by key.
    Delete(ctx context.Context, key string) error

    // Exists checks if a key exists.
    Exists(ctx context.Context, key string) (bool, error)
}

// ErrCacheMiss indicates the key was not found in cache.
var ErrCacheMiss = errors.New("cache: key not found")
```

### Architecture Compliance

**Layer:** `internal/runtimeutil/`
**Pattern:** Interface abstraction for external services
**Benefit:** Swappable cache backends (Redis, Memcached, in-memory)

### References

- [Source: docs/epics.md#Story-6.2]
- [Story 6.1 - Logger Interface](file:///docs/sprint-artifacts/6-1-define-logger-interface.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Second story in Epic 6: Extension Interfaces.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/runtimeutil/cache.go` - Cache interface

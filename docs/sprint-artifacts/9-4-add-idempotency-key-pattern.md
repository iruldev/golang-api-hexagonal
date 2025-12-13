# Story 9.4: Add Idempotency Key Pattern

Status: Done

## Story

As a developer,
I want job idempotency built-in,
So that duplicate jobs don't cause data corruption.

## Acceptance Criteria

### AC1: Idempotency Package Exists
**Given** `internal/worker/idempotency/` package exists
**When** I import the package
**Then** idempotency utilities are available
**And** package is documented with usage examples

### AC2: Idempotency Key Enforcement
**Given** I enqueue a job with idempotency key
**When** a duplicate job with same key is enqueued
**Then** the duplicate is deduplicated (skipped)
**And** original job result is returned
**And** deduplication is logged

### AC3: Configurable Deduplication Window
**Given** I configure idempotency with TTL
**When** TTL expires
**Then** same idempotency key can be used again
**And** default TTL is reasonable (e.g., 24 hours)
**And** TTL is configurable per-job

### AC4: Redis Storage Backend
**Given** Redis is available
**When** idempotency keys are stored
**Then** Redis is used for key storage
**And** atomic SET NX EX pattern is used
**And** keys are prefixed for namespacing

---

## Tasks / Subtasks

- [x] **Task 1: Create idempotency package structure** (AC: #1)
  - [x] Create `internal/worker/idempotency/` directory
  - [x] Create `idempotency.go` with core types and functions
  - [x] Document package with usage examples in code comments

- [x] **Task 2: Implement Store interface** (AC: #4)
  - [x] Define `Store` interface (in `store.go`)
  - [x] Create `RedisStore` implementation using existing `internal/infra/redis` client
  - [x] Use atomic `SET key value NX EX ttl` pattern
  - [x] Use consistent key prefix (e.g., `idempotency:`)

- [x] **Task 3: Implement idempotency key generation** (AC: #2)
  - [x] Support custom key generation via KeyExtractor function
  - [x] Key is string-based for Redis compatibility
  - [x] Empty key = no idempotency (process normally)

- [x] **Task 4: Implement deduplication check** (AC: #2, #3)
  - [x] Create `Check(ctx, key, ttl) (bool, error)` function in RedisStore
  - [x] Return `true` if new (first occurrence), `false` if duplicate
  - [x] Handle Redis errors gracefully (fail-open or strict mode)
  - [x] Log deduplication events

- [x] **Task 5: Create middleware/wrapper for asynq handlers** (AC: #2)
  - [x] Create `IdempotentHandler` wrapper function
  - [x] Extract idempotency key from task payload via KeyExtractor
  - [x] Skip processing if duplicate detected
  - [x] Return `nil` for duplicates (don't retry already-processed work)

- [x] **Task 6: Add configuration options** (AC: #3)
  - [x] Create functional options pattern (WithFailMode, WithLogger, etc.)
  - [x] Default TTL: 24 hours (DefaultTTL constant)
  - [x] Default prefix: `idempotency:` (DefaultKeyPrefix constant)
  - [x] FailMode enum (FailOpen, FailClosed)

- [x] **Task 7: Create unit tests** (AC: #1, #2, #3, #4)
  - [x] Create `internal/worker/idempotency/idempotency_test.go`
  - [x] Test store interface with MockStore
  - [x] Test Redis store with testcontainers
  - [x] Test deduplication logic
  - [x] Test TTL expiration
  - [x] Test handler wrapper
  - [x] Test fail-open mode when Redis unavailable
  - [x] Test fail-closed mode when Redis unavailable

- [x] **Task 8: Create example usage** (AC: #1)
  - [x] Create `internal/worker/idempotency/example_test.go`
  - [x] Show basic idempotency usage
  - [x] Show custom TTL usage
  - [x] Show handler wrapper usage
  - [x] Show key extraction strategies

- [x] **Task 9: Update documentation** (AC: #1)
  - [x] Update `docs/async-jobs.md` with Idempotency Pattern section
  - [x] Add comparison with `asynq.Unique()`
  - [x] Document when-to-use guidelines
  - [x] Add observability section (metrics, logging)

---

## Dev Notes

### Package Location Decision

**Location:** `internal/worker/idempotency/` (separate package)

**Rationale:** Unlike fire-and-forget, scheduled, and fanout (which are _patterns_ for enqueuing/processing), idempotency is a _cross-cutting concern_ that:
- Has its own storage backend (Redis)
- Defines interfaces (`IdempotencyStore`)
- Can be used independently of job patterns
- Has standalone types (`Options`, `FailMode`)

This is similar to how `internal/observability/` is separate from `internal/interface/http/`.

---

### When to Use: asynq.Unique() vs idempotency Package

| Scenario | Use `asynq.Unique()` | Use `idempotency` package |
|----------|---------------------|---------------------------|
| **Prevent duplicate enqueue** (same task not queued twice) | ✅ Yes | ❌ No |
| **Handler-level idempotency** (task re-delivered via retry) | ❌ No | ✅ Yes |
| **Result caching** (return cached result for duplicates) | ❌ No | ✅ Yes |
| **Custom key logic** (e.g., `orderID + productID`) | ❌ No | ✅ Yes |
| **Fail mode control** (fail-open vs fail-closed) | ❌ No | ✅ Yes |

**Use Together:** For maximum protection, combine both:
```go
// Enqueue with asynq.Unique (prevents duplicate enqueue)
task := asynq.NewTask("order:create", payload)
client.Enqueue(task, asynq.Unique(24*time.Hour))

// Handler wrapped with idempotency (prevents duplicate processing on retry)
srv.HandleFunc("order:create", 
    idempotency.IdempotentHandler(store, extractKey, 24*time.Hour, orderHandler.Handle))
```

---

### Key Components

**1. Store Interface:**
```go
type Store interface {
    // Check returns true if this is the first time seeing the key (new).
    Check(ctx context.Context, key string, ttl time.Duration) (isNew bool, err error)
    
    // Optional: Store and retrieve results
    StoreResult(ctx context.Context, key string, result []byte, ttl time.Duration) error
    GetResult(ctx context.Context, key string) ([]byte, bool, error)
}
```

**2. Handler Wrapper Signature:**
```go
func IdempotentHandler(
    store Store,
    keyExtractor func(*asynq.Task) string,
    ttl time.Duration,
    handler func(context.Context, *asynq.Task) error,
    opts ...Option,
) func(context.Context, *asynq.Task) error
```

---

### Key Extraction Strategies

The `keyExtractor` function determines what makes a task "unique". Common strategies:

| Strategy | Key Format | Use Case |
|----------|------------|----------|
| **Payload-based** | `{field1}:{field2}` | Order creation: `orderID:productID` |
| **Task-type + ID** | `{taskType}:{entityID}` | Note archive: `note:archive:uuid-123` |
| **Hash-based** | `sha256(payload)` | Generic tasks where entire payload matters |
| **Business key** | Custom | Domain-specific: `invoice:2024-001` |

**Examples:**
```go
// Payload-based: Extract orderID from JSON payload
func extractOrderKey(t *asynq.Task) string {
    var p struct { OrderID string `json:"order_id"` }
    json.Unmarshal(t.Payload(), &p)
    return fmt.Sprintf("order:%s", p.OrderID)
}

// Task-type + entity ID
func extractNoteArchiveKey(t *asynq.Task) string {
    var p NoteArchivePayload
    json.Unmarshal(t.Payload(), &p)
    return fmt.Sprintf("%s:%s", t.Type(), p.NoteID)
}

// Hash-based for generic tasks
func extractHashKey(t *asynq.Task) string {
    h := sha256.Sum256(t.Payload())
    return fmt.Sprintf("%s:%x", t.Type(), h[:8])
}
```

---

### Redis Key Design

| Pattern | Example | TTL |
|---------|---------|-----|
| Idempotency check | `idempotency:order:create:abc-123` | 24h |
| Idempotency result | `idempotency:result:order:create:abc-123` | 24h |

---

### Configuration

```go
type Options struct {
    TTL       time.Duration // Default: 24h
    KeyPrefix string        // Default: "idempotency:"
    FailMode  FailMode      // Default: FailOpen
    Logger    *zap.Logger   // For logging duplicate detection
}

type FailMode int
const (
    FailOpen   FailMode = iota // On Redis error, process anyway (safe default)
    FailClosed                  // On Redis error, return error (strict mode)
)
```

---

### Worker Integration

Add to `cmd/worker/main.go`:

```go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker/idempotency"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)

func main() {
    // ... existing setup ...
    
    // Create idempotency store using existing Redis client
    idempotencyStore := idempotency.NewRedisStore(redisClient, "idempotency:")
    
    // Create handlers
    noteHandler := tasks.NewNoteArchiveHandler(logger, noteRepo)
    
    // Wrap with idempotency
    idempotentNoteHandler := idempotency.IdempotentHandler(
        idempotencyStore,
        func(t *asynq.Task) string {
            var p tasks.NoteArchivePayload
            if err := json.Unmarshal(t.Payload(), &p); err != nil {
                return "" // No idempotency key → process normally
            }
            return fmt.Sprintf("note:archive:%s", p.NoteID)
        },
        24*time.Hour,
        noteHandler.Handle,
    )
    
    // Register handler
    srv.HandleFunc(tasks.TypeNoteArchive, idempotentNoteHandler)
    
    // ... rest of handlers ...
}
```

---

### Observability

**Logging:** Deduplication events are logged at DEBUG level:
```
DEBUG duplicate task skipped {"idempotency_key": "order:create:abc-123", "task_type": "order:create"}
```

**Redis errors** are logged at WARN level (fail-open) or ERROR level (fail-closed):
```
WARN idempotency check failed, processing anyway {"error": "redis: connection refused"}
```

**Metrics (optional enhancement):** If adding metrics, use these names:
```
idempotency_check_total{result="new|duplicate|error"}
idempotency_check_duration_seconds
```

---

### Fail Mode Testing

```go
func TestRedisStore_Check_FailOpen_WhenRedisUnavailable(t *testing.T) {
    // Arrange: Create store with invalid Redis connection
    store := idempotency.NewRedisStore(brokenRedisClient, "test:")
    store.SetFailMode(idempotency.FailOpen)
    
    // Act: Check should fail but return "new" (fail-open)
    isNew, err := store.Check(ctx, "test-key", time.Hour)
    
    // Assert: Returns new=true (process the task), no error
    assert.NoError(t, err)
    assert.True(t, isNew)
}

func TestRedisStore_Check_FailClosed_WhenRedisUnavailable(t *testing.T) {
    // Arrange: Create store with invalid Redis connection
    store := idempotency.NewRedisStore(brokenRedisClient, "test:")
    store.SetFailMode(idempotency.FailClosed)
    
    // Act: Check should fail and return error
    isNew, err := store.Check(ctx, "test-key", time.Hour)
    
    // Assert: Returns error (don't process, retry later)
    assert.Error(t, err)
    assert.False(t, isNew)
}

func TestRedisStore_Check_TTLExpiration(t *testing.T) {
    // Arrange
    store := idempotency.NewRedisStore(redisClient, "test:")
    key := "ttl-test-key"
    
    // Act: First check
    isNew1, _ := store.Check(ctx, key, 100*time.Millisecond)
    assert.True(t, isNew1) // First occurrence
    
    // Act: Immediate second check
    isNew2, _ := store.Check(ctx, key, 100*time.Millisecond)
    assert.False(t, isNew2) // Duplicate
    
    // Wait for TTL to expire
    time.Sleep(200 * time.Millisecond)
    
    // Act: Third check after TTL
    isNew3, _ := store.Check(ctx, key, 100*time.Millisecond)
    assert.True(t, isNew3) // New again after TTL
}
```

---

### Previous Story Learnings

**From Story 9-1 (Fire-and-Forget):**
- Use `tasks.TaskEnqueuer` interface for DI compatibility
- Document in `docs/async-jobs.md`
- Table-driven tests with AAA pattern

**From Story 9-2 (Scheduled):**
- Keep pattern implementation simple and focused
- Document in code comments with examples

**From Story 9-3 (Fanout):**
- Register functions should return errors for validation
- Add input validation (empty strings, nil values)
- Provide complete worker integration examples in docs
- Thread-safety with sync.RWMutex where needed

**From Story 8-1 (Redis Integration):**
- Use existing Redis client from `internal/infra/redis/`
- Follow connection pooling patterns already established

---

### Testing Requirements

1. **Unit Tests:**
   - Test Store interface with mock
   - Test RedisStore with testcontainers
   - Test Check function (new, duplicate, TTL)
   - Test IdempotentHandler wrapper
   - Test fail modes (open/closed)
   - Test key extractor functions
   - Test fail-open when Redis unavailable
   - Test fail-closed when Redis unavailable
   - Test TTL expiration behavior

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/worker/idempotency/idempotency.go` - Core types, options, constants
- `internal/worker/idempotency/store.go` - IdempotencyStore interface
- `internal/worker/idempotency/redis_store.go` - Redis implementation
- `internal/worker/idempotency/handler.go` - IdempotentHandler wrapper
- `internal/worker/idempotency/idempotency_test.go` - Core tests
- `internal/worker/idempotency/redis_store_test.go` - Redis-specific tests
- `internal/worker/idempotency/example_test.go` - Usage examples

**Modify:**
- `docs/async-jobs.md` - Add Idempotency Pattern section

---

### References

- [Source: docs/epics.md#Epic-9-Story-9.4] - Story requirements
- [Source: docs/architecture.md#Extension-Interfaces] - Interface patterns
- [Source: docs/async-jobs.md] - Existing async documentation
- [Source: internal/worker/patterns/] - Related job patterns
- [Source: internal/infra/redis/] - Existing Redis client setup (Story 8-1)
- [Source: docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md]
- [Source: docs/sprint-artifacts/9-2-implement-scheduled-job-pattern-with-cron.md]
- [Source: docs/sprint-artifacts/9-3-implement-fanout-job-pattern.md]

---

## Dev Agent Record

### Context Reference

Previous stories: 
- `docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md`
- `docs/sprint-artifacts/9-2-implement-scheduled-job-pattern-with-cron.md`
- `docs/sprint-artifacts/9-3-implement-fanout-job-pattern.md`

Architecture: `docs/architecture.md`
Async patterns: `docs/async-jobs.md`
Redis setup: `internal/infra/redis/` (from Story 8-1)

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- ✅ Created `internal/worker/idempotency/` package with 4 source files
- ✅ Implemented Store interface with Check, StoreResult, GetResult methods
- ✅ Implemented RedisStore with atomic SET NX EX pattern
- ✅ Implemented IdempotentHandler wrapper with fail-open/closed modes
- ✅ Created comprehensive unit tests with MockStore 
- ✅ Created Redis integration tests with testcontainers
- ✅ Created example_test.go with usage documentation
- ✅ Updated docs/async-jobs.md with Idempotency Pattern section
- ✅ All tests pass (make test) - 100% success
- ✅ **Code Review Fixes Applied** (6 issues found, all fixed)

### Change Log

| Date | Changes |
|------|---------|
| 2025-12-13 | Story drafted with comprehensive developer context |
| 2025-12-13 | **Implementation Complete:** All 9 tasks implemented and tested |
| 2025-12-13 | **Story Validation:** Applied 9 improvements from quality review |
|            | - Added package location rationale for `/idempotency/` vs `/patterns/` |
|            | - Added decision matrix: asynq.Unique() vs idempotency package |
|            | - Added key extraction strategies with examples |
|            | - Added complete worker integration example for `cmd/worker/main.go` |
|            | - Added fail mode test cases (fail-open, fail-closed, TTL expiration) |
|            | - Added observability section with logging and metrics guidance |
|            | - Referenced existing Redis client from Story 8-1 |
|            | - Consolidated file list (removed duplicate in Dev Agent Record) |
|            | - Streamlined code snippets for token efficiency |
| 2025-12-13 | **Code Review Complete:** 6 issues fixed |
|            | - Fixed: Interface name aligned (`Store` vs `IdempotencyStore`) in docs |
|            | - Fixed: Added Redis key format documentation in `async-jobs.md` |
|            | - Fixed: Added nil parameter validation in `IdempotentHandler` |
|            | - Fixed: Clarified dual FailMode configuration relationship |
|            | - Added: 3 new tests for nil validation (PanicsOnNilStore, etc.) |
|            | - Note: Files should be committed to git |

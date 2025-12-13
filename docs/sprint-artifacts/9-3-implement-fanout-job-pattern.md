# Story 9.3: Implement Fanout Job Pattern

Status: Done

## Story

As a developer,
I want a fanout pattern for broadcasting to multiple handlers,
So that I can implement event-driven workflows.

## Acceptance Criteria

### AC1: Fanout Pattern Exists
**Given** `internal/worker/patterns/fanout.go` exists
**When** I publish a fanout event
**Then** multiple handlers receive the event
**And** each handler processes independently

### AC2: Handler Isolation
**Given** a fanout event is published
**When** one handler fails
**Then** failure doesn't affect other handlers
**And** each handler retries independently
**And** failure is logged with handler context

### AC3: Fanout Configuration
**Given** I define a fanout event
**When** I register handlers
**Then** handlers can be added/removed dynamically
**And** each handler can have different queue priorities
**And** handler registration is type-safe

### AC4: Pattern Documentation
**Given** the fanout pattern
**When** I check the documentation
**Then** usage examples exist in `docs/async-jobs.md`
**And** event-driven workflow examples are provided
**And** when-to-use guidelines are documented

---

## Tasks / Subtasks

- [x] **Task 1: Create fanout pattern package** (AC: #1, #3)
  - [x] Create `internal/worker/patterns/fanout.go`
  - [x] Define `FanoutEvent` struct with event type, payload, metadata
  - [x] Define `FanoutHandler` type with handler ID and function
  - [x] Define `FanoutRegistry` for handler registration
  - [x] Document the pattern in code comments

- [x] **Task 2: Implement FanoutRegistry** (AC: #3)
  - [x] Create registry for managing event handlers
  - [x] Implement `Register(eventType, handlerID, handlerFunc)` method
  - [x] Implement `Handlers(eventType)` to get registered handlers
  - [x] Implement `Unregister(eventType, handlerID)` for dynamic removal
  - [x] Make registry thread-safe with mutex

- [x] **Task 3: Implement Fanout function** (AC: #1, #2)
  - [x] Create `Fanout(ctx, enqueuer, logger, event)` function
  - [x] Enqueue separate task for each registered handler
  - [x] Use unique task type per handler: `fanout:{eventType}:{handlerID}`
  - [x] Ensure handler isolation - one failure doesn't affect others
  - [x] Log enqueue results for each handler

- [x] **Task 4: Implement handler task wrapper** (AC: #1, #2)
  - [x] Create `FanoutDispatcher` wrapper that routes to correct handler
  - [x] Pass event payload to handler function
  - [x] Ensure errors are isolated per handler

- [x] **Task 5: Add queue priority support** (AC: #3)
  - [x] Allow per-handler queue configuration
  - [x] Default to `default` queue
  - [x] Support `asynq.Option` overrides per handler

- [x] **Task 6: Create example usage** (AC: #1, #4)
  - [x] Create `internal/worker/patterns/fanout_example_test.go`
  - [x] Show event definition example
  - [x] Show handler registration example
  - [x] Show fanout publishing example
  - [x] Document common use cases (user events, order events)

- [x] **Task 7: Unit tests** (AC: #1, #2, #3)
  - [x] Create `internal/worker/patterns/fanout_test.go`
  - [x] Test registry registration/unregistration
  - [x] Test fanout to multiple handlers
  - [x] Test handler isolation (one failure, others continue)
  - [x] Test queue priority per handler
  - [x] Test thread-safety of registry

- [x] **Task 8: Update documentation** (AC: #4)
  - [x] Update `docs/async-jobs.md` with Fanout Pattern section
  - [x] Add event-driven workflow architecture diagram
  - [x] Add comparison: fanout vs fire-and-forget vs scheduled
  - [x] Document handler registration patterns

---

## Dev Notes

### Architecture Placement

```
internal/worker/
├── server.go
├── client.go
├── middleware.go
├── metrics_middleware.go
├── tasks/
│   ├── types.go
│   ├── note_archive.go
│   ├── cleanup_old_notes.go
│   └── enqueue.go
└── patterns/
    ├── fireandforget.go
    ├── fireandforget_test.go
    ├── scheduled.go
    ├── scheduled_test.go
    ├── fanout.go              [NEW]
    ├── fanout_test.go         [NEW]
    └── fanout_example_test.go [NEW]
```

**Key:** Fanout pattern in `internal/worker/patterns/`, follows same structure as fire-and-forget and scheduled patterns.

---

### Fanout Pattern Design

```go
// internal/worker/patterns/fanout.go
package patterns

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/hibiken/asynq"
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
    "go.uber.org/zap"
)

// FanoutEvent represents an event to be broadcast to multiple handlers.
type FanoutEvent struct {
    // Type is the event type identifier (e.g., "user:created", "order:completed")
    Type string `json:"type"`
    
    // Payload is the event data
    Payload json.RawMessage `json:"payload"`
    
    // Metadata contains optional event metadata (timestamp, source, correlation_id)
    Metadata map[string]string `json:"metadata,omitempty"`
    
    // Timestamp is when the event was created
    Timestamp time.Time `json:"timestamp"`
}

// FanoutHandlerFunc is the signature for fanout event handlers.
type FanoutHandlerFunc func(ctx context.Context, event FanoutEvent) error

// FanoutHandler represents a registered handler for fanout events.
type FanoutHandler struct {
    ID      string           // Unique handler identifier
    Handler FanoutHandlerFunc
    Queue   string           // Target queue (default: "default")
    Opts    []asynq.Option   // Additional asynq options
}

// FanoutRegistry manages registration of fanout handlers.
type FanoutRegistry struct {
    mu       sync.RWMutex
    handlers map[string][]FanoutHandler // eventType -> handlers
}

// NewFanoutRegistry creates a new fanout registry.
func NewFanoutRegistry() *FanoutRegistry {
    return &FanoutRegistry{
        handlers: make(map[string][]FanoutHandler),
    }
}

// Register adds a handler for an event type.
func (r *FanoutRegistry) Register(eventType, handlerID string, fn FanoutHandlerFunc, opts ...asynq.Option) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    handler := FanoutHandler{
        ID:      handlerID,
        Handler: fn,
        Queue:   worker.QueueDefault, // Default queue
        Opts:    opts,
    }
    r.handlers[eventType] = append(r.handlers[eventType], handler)
}

// Handlers returns all handlers for an event type.
func (r *FanoutRegistry) Handlers(eventType string) []FanoutHandler {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.handlers[eventType]
}

// Unregister removes a handler for an event type.
func (r *FanoutRegistry) Unregister(eventType, handlerID string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    handlers := r.handlers[eventType]
    for i, h := range handlers {
        if h.ID == handlerID {
            r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
            return
        }
    }
}

// Fanout broadcasts an event to all registered handlers.
// Each handler receives a separate task and processes independently.
// Failures in one handler don't affect others.
//
// Returns []error containing any enqueue failures. Callers should check:
//   - len(errors) == 0: All handlers enqueued successfully
//   - len(errors) > 0: Some handlers failed to enqueue (partial success)
//   - len(errors) == len(handlers): Complete failure
//
// Note: Timestamp is auto-set to time.Now().UTC() if zero.
func Fanout(
    ctx context.Context,
    enqueuer tasks.TaskEnqueuer,
    registry *FanoutRegistry,
    logger *zap.Logger,
    event FanoutEvent,
) []error {
    // Auto-set timestamp if not provided
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now().UTC()
    }

    handlers := registry.Handlers(event.Type)
    if len(handlers) == 0 {
        logger.Warn("no handlers registered for event",
            zap.String("event_type", event.Type),
        )
        return nil
    }

    var errors []error
    for _, h := range handlers {
        taskType := fmt.Sprintf("fanout:%s:%s", event.Type, h.ID)
        payload, err := json.Marshal(event)
        if err != nil {
            errors = append(errors, fmt.Errorf("marshal event for handler %s: %w", h.ID, err))
            continue
        }

        task := asynq.NewTask(taskType, payload, h.Opts...)
        opts := []asynq.Option{asynq.Queue(h.Queue)}
        opts = append(opts, h.Opts...)

        info, err := enqueuer.Enqueue(ctx, task, opts...)
        if err != nil {
            logger.Error("fanout enqueue failed",
                zap.Error(err),
                zap.String("event_type", event.Type),
                zap.String("handler_id", h.ID),
            )
            errors = append(errors, err)
            continue
        }

        logger.Debug("fanout task enqueued",
            zap.String("task_id", info.ID),
            zap.String("event_type", event.Type),
            zap.String("handler_id", h.ID),
            zap.String("queue", info.Queue),
        )
    }

    return errors
}
```

---

### Fanout Use Cases

| Scenario | Event Type | Handlers | Description |
|----------|------------|----------|-------------|
| User created | `user:created` | SendWelcomeEmail, CreateDefaultSettings, NotifyAdmin | Multiple actions on user signup |
| Order completed | `order:completed` | SendReceipt, UpdateInventory, NotifyShipping | Order workflow orchestration |
| Note archived | `note:archived` | UpdateSearchIndex, NotifyOwner, CreateBackup | Note lifecycle events |
| Payment received | `payment:received` | UpdateBalance, SendReceipt, RecordAnalytics | Payment processing |

---

### Handler Isolation Design

Each handler is completely isolated:

1. **Separate Task:** Each handler gets its own asynq task
2. **Independent Retry:** Each task retries according to its own configuration
3. **Separate Queue:** Handlers can run on different queues
4. **Independent Failure:** One handler failure doesn't affect others

```
         ┌──────────────────────────────────────────────────┐
         │                 Fanout Event                     │
         │            event.Type: "user:created"            │
         └──────────────────────────────────────────────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
           ▼                   ▼                   ▼
    ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
    │ Handler A    │   │ Handler B    │   │ Handler C    │
    │ WelcomeEmail │   │ CreateSettings│  │ NotifyAdmin  │
    │ Queue: crit  │   │ Queue: default│  │ Queue: low   │
    └──────────────┘   └──────────────┘   └──────────────┘
           │                   │                   │
           │ ❌ Fails          │ ✅ Success        │ ✅ Success
           │ (retries)         │                   │
           ▼                   ▼                   ▼
    ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
    │ Retry logic  │   │ Complete     │   │ Complete     │
    │ (independent)│   │              │   │              │
    └──────────────┘   └──────────────┘   └──────────────┘
```

---

### Comparison: Job Patterns

| Aspect | Fire-and-Forget | Scheduled | Fanout |
|--------|-----------------|-----------|--------|
| Trigger | Immediate, caller-driven | Cron schedule | Event-driven |
| Handlers | Single | Single | Multiple |
| Isolation | Caller isolated | N/A | Handler isolated |
| Default Queue | low | default | per-handler |
| Use case | Non-critical background | Periodic tasks | Event workflows |

---

### Handler-Side Architecture (Worker Processing)

**Critical:** The fanout pattern requires coordination between the enqueue side (API) and process side (Worker).

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ENQUEUE SIDE (API)                                │
│                                                                             │
│   FanoutRegistry ──────► Fanout() ──────► Redis Queue                       │
│   (stores handlers)      (enqueues fanout:event:handlerID tasks)            │
└──────────────────────────────────────────────────────────────|──────────────┘
                                                               │
                                                               ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PROCESS SIDE (Worker)                             │
│                                                                             │
│   Worker Server ◄─────── Redis Queue                                        │
│        │                                                                    │
│        ▼                                                                    │
│   FanoutRegistry ──────► FanoutHandler.Handle(ctx, task)                    │
│   (SAME registry)        (looks up handler by ID, deserializes event)       │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Implementation Details:**

1. **Shared Registry:** The `FanoutRegistry` must be available to both API (for enqueue) and Worker (for processing). Create it once in a shared location (e.g., `internal/worker/patterns/registry.go`) or pass it during initialization.

2. **Worker Handler Registration:** Register a generic handler in `cmd/worker/main.go`:
   ```go
   // Register fanout dispatcher that routes to specific handlers
   dispatcher := patterns.NewFanoutDispatcher(registry, logger)
   srv.HandleFunc("fanout:*", dispatcher.Handle) // Wildcard pattern
   ```

3. **FanoutDispatcher:** (Add to `fanout.go`)
   ```go
   type FanoutDispatcher struct {
       registry *FanoutRegistry
       logger   *zap.Logger
   }
   
   func (d *FanoutDispatcher) Handle(ctx context.Context, t *asynq.Task) error {
       // Task type format: "fanout:{eventType}:{handlerID}"
       parts := strings.Split(t.Type(), ":")
       if len(parts) < 3 {
           return fmt.Errorf("invalid fanout task type: %s: %w", t.Type(), asynq.SkipRetry)
       }
       eventType, handlerID := parts[1], parts[2]
       
       // Find handler
       handlers := d.registry.Handlers(eventType)
       for _, h := range handlers {
           if h.ID == handlerID {
               var event FanoutEvent
               if err := json.Unmarshal(t.Payload(), &event); err != nil {
                   return fmt.Errorf("unmarshal event: %w: %w", err, asynq.SkipRetry)
               }
               return h.Handler(ctx, event)
           }
       }
       return fmt.Errorf("handler %s not found for event %s: %w", handlerID, eventType, asynq.SkipRetry)
   }
   ```

4. **Alternative:** Register each handler explicitly at worker startup instead of wildcard.

---

### Metrics and Observability

**Metrics:** Captured by existing worker `MetricsMiddleware` when handlers execute (worker-side).

- Each handler appears as separate task type: `fanout:{eventType}:{handlerID}`
- Metrics include: `job_processed_total`, `job_duration_seconds`
- No separate enqueue-side fanout metrics (enqueue errors are logged)

---

### Previous Story Learnings

**From Story 9-1 (Fire-and-Forget):**
- Use `tasks.TaskEnqueuer` interface for DI compatibility
- Patterns go in `internal/worker/patterns/`
- Document in `docs/async-jobs.md`
- Table-driven tests with AAA pattern
- Example files provide usage documentation
- Add 5-second timeout for goroutines (avoid leaks)
- Use goroutine for non-blocking behavior

**From Story 9-2 (Scheduled):**
- Keep pattern implementation simple and focused
- Use robfig/cron for expression parsing
- Document in code comments with examples
- Create dedicated test file with comprehensive coverage
- Update async-jobs.md with new pattern section
- Include comparison tables

**From Code Reviews:**
- Use proper interface-based signatures
- Clarify metrics responsibilities (worker-side vs enqueue-side)
- Keep story and code in sync
- Test edge cases (empty handlers, invalid payloads)

---

### Testing Strategy

```go
func TestFanoutRegistry_Register(t *testing.T) {
    // Test registration and retrieval
}

func TestFanoutRegistry_Unregister(t *testing.T) {
    // Test removal of handlers
}

func TestFanoutRegistry_ThreadSafety(t *testing.T) {
    // Test concurrent registration/lookup
    registry := NewFanoutRegistry()
    var wg sync.WaitGroup
    
    // Concurrent writes
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            registry.Register("test", fmt.Sprintf("handler-%d", n), 
                func(_ context.Context, _ FanoutEvent) error { return nil })
        }(i)
    }
    
    // Concurrent reads while writing
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = registry.Handlers("test")
        }()
    }
    
    wg.Wait()
    assert.GreaterOrEqual(t, len(registry.Handlers("test")), 1)
}

func TestFanout_MultipleHandlers(t *testing.T) {
    // Test fanout to multiple handlers
}

func TestFanout_HandlerIsolation(t *testing.T) {
    // Test that one failure doesn't affect others
}

func TestFanout_QueuePerHandler(t *testing.T) {
    // Test different queues per handler
}

func TestFanout_NoHandlers(t *testing.T) {
    // Test graceful handling when no handlers registered
}
```

---

### Testing Requirements

1. **Unit Tests:**
   - Test registry registration/unregistration
   - Test fanout to multiple handlers
   - Test handler isolation (one failure, others continue)
   - Test queue priority per handler
   - Test thread-safety of registry
   - Test graceful handling of no handlers

2. **Coverage:** Match project standards (≥80%)

3. **Run:** `make test` must pass

---

### File List

**Create:**
- `internal/worker/patterns/fanout.go`
- `internal/worker/patterns/fanout_test.go`
- `internal/worker/patterns/fanout_example_test.go`

**Modify:**
- `docs/async-jobs.md` - Add Fanout Pattern section

---

### Project Structure Notes

- Alignment with unified project structure: ✅
- Follows hexagonal architecture: ✅
- Patterns in `internal/worker/patterns/` (consistent with 9-1, 9-2)
- Uses `tasks.TaskEnqueuer` interface for DI

---

### References

- [Source: docs/epics.md#Epic-9-Story-9.3] - Story requirements
- [Source: docs/architecture.md#Extension-Interfaces] - Interface patterns
- [Source: docs/async-jobs.md] - Existing async documentation
- [Source: internal/worker/patterns/fireandforget.go] - Fire-and-forget pattern reference
- [Source: internal/worker/patterns/scheduled.go] - Scheduled pattern reference
- [Source: docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md] - Previous story learnings
- [Source: docs/sprint-artifacts/9-2-implement-scheduled-job-pattern-with-cron.md] - Previous story learnings

---

## Dev Agent Record

### Context Reference

Previous stories: 
- `docs/sprint-artifacts/9-1-implement-fire-and-forget-job-pattern.md`
- `docs/sprint-artifacts/9-2-implement-scheduled-job-pattern-with-cron.md`

Architecture: `docs/architecture.md`
Async patterns: `docs/async-jobs.md`

### Agent Model Used

Gemini 2.5 (Antigravity)

### Debug Log References

### Completion Notes List

- Implemented complete fanout pattern with FanoutEvent, FanoutRegistry, FanoutHandler types
- Created Fanout function that enqueues separate tasks for each handler
- Created FanoutDispatcher for worker-side routing
- Fixed task type parsing to handle event types with colons (e.g., user:created)
- Added RegisterWithQueue for per-handler queue configuration
- Implemented thread-safe registry with sync.RWMutex
- Added 25 unit tests covering all functionality
- Created 6 example tests demonstrating usage patterns
- Updated async-jobs.md with ~130 lines of Fanout Pattern documentation

### File List

**Created:**
- `internal/worker/patterns/fanout.go`
- `internal/worker/patterns/fanout_test.go`
- `internal/worker/patterns/fanout_example_test.go`

**Modified:**
- `docs/async-jobs.md` - Added Fanout Pattern section

### Change Log

| Date | Changes |
|------|--------|
| 2025-12-13 | Implemented Fanout Job Pattern with FanoutEvent, FanoutRegistry, Fanout function, and FanoutDispatcher |
| 2025-12-13 | Created 25 unit tests with comprehensive coverage |
| 2025-12-13 | Created 6 example tests demonstrating usage patterns |
| 2025-12-13 | Fixed task type parsing for event types containing colons |
| 2025-12-13 | Updated docs/async-jobs.md with Fanout Pattern documentation |
| 2025-12-13 | **Code Review Fixes Applied:** |
|            | - Added input validation (ErrEmptyEventType, ErrEmptyHandlerID, ErrNilHandler) |
|            | - Added duplicate handler detection (ErrDuplicateHandler) |
|            | - Register functions now return error for proper validation |
|            | - Updated 30+ test calls to handle new error returns |
|            | - Updated all example tests for new API |
|            | - Added 4 new validation tests and 1 duplicate registration test |
|            | - Fixed unused variables in TestFanout_PartialFailure |
|            | - Enhanced docs/async-jobs.md with complete worker integration example |

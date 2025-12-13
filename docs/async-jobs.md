# Async Job Patterns

This document provides comprehensive documentation for implementing async jobs using the Asynq worker infrastructure. It covers worker architecture, task handler patterns, enqueueing strategies, and observability.

---

## Worker Infrastructure Overview

### Entry Points

The application has two separate entry points:

| Entry Point | Location | Description |
|-------------|----------|-------------|
| **HTTP API** | `cmd/api/main.go` | Handles synchronous HTTP requests |
| **Worker** | `cmd/worker/main.go` | Processes async background jobs |

> **IMPORTANT:** The Worker is a **secondary entry point**, NOT a new architectural layer. It follows the same hexagonal architecture: task handlers call usecases, which use repositories.

### Worker Package Structure

```
internal/worker/
├── server.go              # Worker server configuration with queue priorities
├── client.go              # Task enqueueing client for use in usecases
├── middleware.go          # Recovery, Tracing, Logging middleware
├── metrics_middleware.go  # Prometheus metrics middleware
└── tasks/
    ├── types.go           # Task type constants (e.g., TypeNoteArchive)
    ├── note_archive.go    # Sample task handler (reference implementation)
    ├── note_archive_test.go
    └── enqueue.go         # TaskEnqueuer interface for usecase layer
```

### Server (`internal/worker/server.go`)

The server wraps `asynq.Server` and provides:

```go
// NewServer creates a worker server with Redis connection and config.
func NewServer(redisOpt asynq.RedisClientOpt, cfg config.AsynqConfig) *Server

// HandleFunc registers a task handler for a task type.
func (s *Server) HandleFunc(pattern string, handler func(context.Context, *asynq.Task) error)

// Use adds middleware to the server.
func (s *Server) Use(mws ...asynq.MiddlewareFunc)

// Start begins processing tasks (blocking).
func (s *Server) Start() error

// Shutdown gracefully stops the server.
func (s *Server) Shutdown()
```

**Default Configuration:**
- Concurrency: 10 workers
- Shutdown timeout: 30 seconds
- Retry delay: Asynq defaults (exponential backoff)

### Client (`internal/worker/client.go`)

The client wraps `asynq.Client` for enqueueing tasks:

```go
// NewClient creates a new Asynq client with Redis connection.
func NewClient(redisOpt asynq.RedisClientOpt) *Client

// Enqueue adds a task to the specified queue with options.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)

// EnqueueCritical adds a task to the critical queue (weight 6).
func (c *Client) EnqueueCritical(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)

// EnqueueDefault adds a task to the default queue (weight 3).
func (c *Client) EnqueueDefault(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)

// EnqueueLow adds a task to the low priority queue (weight 1).
func (c *Client) EnqueueLow(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)

// Close closes the client connection.
func (c *Client) Close() error
```

---

## Queue Priority System

The worker uses weighted priority queues:

| Queue | Constant | Weight | Use Case |
|-------|----------|--------|----------|
| `critical` | `worker.QueueCritical` | 6 | Time-sensitive operations (e.g., password reset emails) |
| `default` | `worker.QueueDefault` | 3 | Standard background tasks (e.g., note archival) |
| `low` | `worker.QueueLow` | 1 | Non-urgent work (e.g., cleanup, analytics) |

### How Weights Work

Weights determine the **probability** that a task from each queue will be processed next:

- When all queues have pending tasks, the worker processes approximately **6 critical tasks for every 3 default tasks and 1 low task**
- Empty queues are skipped; tasks are always processed when available
- Higher weight = higher processing priority

### Queue Selection Guidelines

| Scenario | Recommended Queue |
|----------|-------------------|
| User-facing operations (email, notifications) | `critical` |
| Business logic (archival, synchronization) | `default` |
| Analytics, reports, batch processing | `low` |

---

## Fire-and-Forget Pattern

The Fire-and-Forget pattern is for non-critical background tasks where the caller:
- **Doesn't need to wait** for completion
- **Isn't affected** by task failures
- Treats work as **best-effort**

### When to Use Fire-and-Forget

| Scenario | Use Fire-and-Forget? | Reasoning |
|----------|---------------------|-----------|
| Analytics events | ✅ Yes | Non-critical, caller shouldn't wait |
| Cleanup tasks | ✅ Yes | Background work |
| Cache warming | ✅ Yes | Best-effort optimization |
| Audit logging | ✅ Yes | Not user-facing |
| Password reset email | ❌ No | Critical - need confirmation |
| Payment processing | ❌ No | Critical - need result |
| User-visible notifications | ⚠️ Maybe | Depends on importance |

### Fire-and-Forget vs Standard Enqueue

| Aspect | Fire-and-Forget | Standard Enqueue |
|--------|-----------------|------------------|
| Caller blocks? | No (returns immediately) | No (but gets task info) |
| Error propagation | Errors logged, not returned | Errors returned to caller |
| Default queue | `low` | Caller specifies |
| Task info returned | No | Yes (TaskInfo) |
| Use case | Non-critical work | Important async work |

### Usage

```go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
)

// Basic fire-and-forget (uses low queue by default)
task, _ := tasks.NewNoteArchiveTask(noteID)
patterns.FireAndForget(ctx, client, logger, task)

// Override to default queue if more urgent
patterns.FireAndForget(ctx, client, logger, task, asynq.Queue(worker.QueueDefault))

// With additional options
patterns.FireAndForget(ctx, client, logger, task,
    asynq.Queue(worker.QueueLow),
    asynq.MaxRetry(1),
)
```

### Usecase Integration Example

```go
type NoteUsecase struct {
    repo     note.Repository
    enqueuer tasks.TaskEnqueuer  // *worker.Client implements this
    logger   *zap.Logger
}

func (u *NoteUsecase) Update(ctx context.Context, noteID uuid.UUID, data UpdateData) error {
    // 1. Critical business logic (must succeed)
    if err := u.repo.Update(ctx, noteID, data); err != nil {
        return err
    }

    // 2. Fire-and-forget for non-critical follow-up
    task, err := tasks.NewNoteArchiveTask(noteID)
    if err != nil {
        u.logger.Warn("failed to create audit task", zap.Error(err))
        return nil  // Don't fail main operation
    }

    // Returns immediately - main operation not blocked
    patterns.FireAndForget(ctx, u.enqueuer, u.logger, task)

    return nil
}
```

### Error Handling

Fire-and-Forget handles errors internally:

1. **Enqueue error:** Logged at ERROR level, not propagated to caller
2. **Task execution error:** Handled by worker (retry/SkipRetry as configured)
3. **Context cancellation:** Enqueue may fail, logged

```go
// Caller is isolated from errors
patterns.FireAndForget(ctx, client, logger, task)
// This line executes immediately, regardless of enqueue success
```

### Key Files

| Component | Path |
|-----------|------|
| Pattern implementation | `internal/worker/patterns/fireandforget.go` |
| Unit tests | `internal/worker/patterns/fireandforget_test.go` |
| Examples | `internal/worker/patterns/fireandforget_example_test.go` |

---

## Scheduled Job Pattern

The Scheduled Job pattern is for periodic tasks that run on a cron schedule:
- **Scheduler** enqueues tasks on schedule (separate process from worker)
- **Worker** processes the enqueued tasks
- Uses standard 5-field cron expressions
- All times are in **UTC**

### When to Use Scheduled Jobs

| Scenario | Use Scheduled? | Reasoning |
|----------|---------------|-----------|
| Daily database cleanup | ✅ Yes | Predictable timing, runs once per day |
| Weekly report generation | ✅ Yes | Business-driven schedule |
| Hourly health checks | ✅ Yes | Regular intervals |
| Cache invalidation | ✅ Yes | Periodic refresh |
| Real-time event response | ❌ No | Use Fire-and-Forget |
| User-triggered actions | ❌ No | Use standard enqueue |

### Scheduler Architecture

```
                    ┌─────────────┐
                    │   Redis     │
                    │   (Queue)   │
                    └─────────────┘
                          ▲
              Enqueue     │     Dequeue
              on cron     │     & Process
         ┌────────────────┴────────────────┐
         │                                 │
┌────────┴────────┐              ┌─────────┴────────┐
│    Scheduler    │              │      Worker      │
│cmd/scheduler    │              │  cmd/worker      │
│                 │              │                  │
│ Cron: 0 0 * * * │              │ Handle:          │
│ → Enqueue task  │              │ cleanup:old_notes│
└─────────────────┘              └──────────────────┘
```

**Key:** Scheduler and Worker are SEPARATE processes. Scheduler enqueues, Worker processes.

### Cron Expression Format

Asynq uses standard 5-field cron expressions:

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sun=0)
│ │ │ │ │
* * * * *
```

**Common Examples:**

| Schedule | Expression | Description |
|----------|------------|-------------|
| Every minute | `* * * * *` | Runs every minute |
| Every hour | `0 * * * *` | Runs at minute 0 of every hour |
| Daily at midnight | `0 0 * * *` | Runs at 00:00 UTC |
| Every Monday 9am | `0 9 * * 1` | Runs Monday at 09:00 UTC |
| First of month | `0 0 1 * *` | Runs at midnight on 1st |
| Every 5 minutes | `*/5 * * * *` | Runs every 5 minutes |
| Weekdays 6pm | `0 18 * * 1-5` | Runs Mon-Fri at 18:00 UTC |

### Usage

```go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)

// Define scheduled jobs
jobs := []patterns.ScheduledJob{
    {
        Cronspec:    "0 0 * * *",  // Daily at midnight UTC
        Task:        cleanupTask,
        Description: "Daily cleanup of old notes",
    },
    {
        Cronspec:    "0 * * * *",  // Every hour
        Task:        healthTask,
        Description: "Hourly health check",
    },
}

// Register with scheduler
entryIDs, err := patterns.RegisterScheduledJobs(scheduler, jobs, logger)
```

### Timezone Handling

- Scheduler runs in **UTC** by default
- All cron expressions are evaluated in UTC
- Set explicitly in scheduler options:

```go
loc, _ := time.LoadLocation("UTC")
scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
    Location: loc,
})
```

### Missed Job Behavior

Asynq's default behavior:
- **Does NOT catch up:** Missed executions are skipped
- **Jobs persist:** Scheduler state survives restarts
- **Unique tasks:** Use `asynq.Unique(duration)` to prevent duplicates

For critical scheduled jobs:
- Consider shorter intervals with idempotency
- Use Asynq's task uniqueness feature

### Running the Scheduler

```bash
# Start dependencies (Redis, etc.)
docker compose up -d

# Terminal 1: Start worker (processes tasks)
make worker

# Terminal 2: Start scheduler (enqueues tasks on schedule)
make scheduler
```

### Key Files

| Component | Path |
|-----------|------|
| Scheduler entry point | `cmd/scheduler/main.go` |
| Pattern implementation | `internal/worker/patterns/scheduled.go` |
| Unit tests | `internal/worker/patterns/scheduled_test.go` |
| Examples | `internal/worker/patterns/scheduled_example_test.go` |
| Sample scheduled task | `internal/worker/tasks/cleanup_old_notes.go` |

---

## Fanout Pattern

The Fanout pattern is for event-driven workflows where a single event triggers multiple independent handlers:
- Each handler processes the event **independently**
- Failures in one handler **don't affect others**
- Each handler can have **different queue priorities**

### When to Use Fanout

| Scenario | Use Fanout? | Reasoning |
|----------|-------------|-----------|
| User signup → email, settings, notify | ✅ Yes | Multiple independent actions |
| Order completed → receipt, inventory, shipping | ✅ Yes | Event-driven orchestration |
| Note archived → search index, notify, backup | ✅ Yes | Parallel processing |
| Single background task | ❌ No | Use Fire-and-Forget |
| Periodic cleanup | ❌ No | Use Scheduled Job |
| Sequential workflow | ❌ No | Use saga pattern |

### Fanout Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ENQUEUE SIDE (API)                                │
│                                                                             │
│   FanoutRegistry ──────► Fanout() ──────► Redis Queue                       │
│   (stores handlers)      (enqueues fanout:event:handlerID tasks)            │
└──────────────────────────────────────────────────────|──────────────────────┘
                                                       │
                                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PROCESS SIDE (Worker)                             │
│                                                                             │
│   Worker Server ◄─────── Redis Queue                                        │
│        │                                                                    │
│        ▼                                                                    │
│   FanoutDispatcher ──────► Handler Function                                 │
│   (routes by handlerID)   (processes event)                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Usage

```go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
)

// 1. Create registry and register handlers
registry := patterns.NewFanoutRegistry()

registry.Register("user:created", "welcome-email", func(ctx context.Context, event patterns.FanoutEvent) error {
    // Send welcome email
    return nil
})

registry.Register("user:created", "default-settings", func(ctx context.Context, event patterns.FanoutEvent) error {
    // Create default user settings
    return nil
})

// Use different queue for critical handlers
registry.RegisterWithQueue("user:created", "notify-admin", 
    func(ctx context.Context, event patterns.FanoutEvent) error {
        return nil
    },
    worker.QueueCritical,
)

// 2. Publish fanout event
event := patterns.FanoutEvent{
    Type:    "user:created",
    Payload: json.RawMessage(`{"user_id": "123", "email": "user@example.com"}`),
    Metadata: map[string]string{"trace_id": "abc-123"},
}

errors := patterns.Fanout(ctx, client, registry, logger, event)
if len(errors) > 0 {
    // Handle partial failures (some handlers may have failed to enqueue)
}
```

### Worker Registration

The fanout pattern requires both **enqueue-side** (API) and **process-side** (Worker) registration.

#### Complete Worker Setup

Add the following to `cmd/worker/main.go`:

```go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker/patterns"
)

func main() {
    // ... existing config and logger setup ...

    // Create worker server
    srv := worker.NewServer(redisOpt, cfg.Asynq)

    // === FANOUT PATTERN SETUP ===
    
    // 1. Create shared registry (define once, use everywhere)
    fanoutRegistry := patterns.NewFanoutRegistry()
    
    // 2. Register handlers for each event type
    _ = fanoutRegistry.Register("user:created", "welcome-email", func(ctx context.Context, event patterns.FanoutEvent) error {
        // Handler logic: send welcome email
        return nil
    })
    
    _ = fanoutRegistry.Register("user:created", "default-settings", func(ctx context.Context, event patterns.FanoutEvent) error {
        // Handler logic: create default user settings
        return nil
    })
    
    // 3. Create dispatcher
    fanoutDispatcher := patterns.NewFanoutDispatcher(fanoutRegistry, logger)
    
    // 4. Register each fanout task type with the worker
    // Since asynq doesn't support wildcards, register each event:handler combination
    for _, eventType := range []string{"user:created", "order:completed", "note:archived"} {
        for _, h := range fanoutRegistry.Handlers(eventType) {
            taskType := fmt.Sprintf("fanout:%s:%s", eventType, h.ID)
            srv.HandleFunc(taskType, fanoutDispatcher.Handle)
        }
    }

    // === OTHER HANDLERS ===
    // ... register other task handlers ...
    
    // Start server
    if err := srv.Start(); err != nil {
        logger.Fatal("Worker error", zap.Error(err))
    }
}
```

#### Shared Registry Pattern

For production use, create a shared registry in a separate file:

```go
// internal/worker/patterns/registry.go
package patterns

import (
    "context"
)

// DefaultRegistry is a shared fanout registry for the application.
var DefaultRegistry = NewFanoutRegistry()

// RegisterDefaultHandlers registers all fanout handlers.
// Call this from both API and Worker initialization.
func RegisterDefaultHandlers() {
    _ = DefaultRegistry.Register("user:created", "welcome-email", handleWelcomeEmail)
    _ = DefaultRegistry.Register("user:created", "default-settings", handleDefaultSettings)
    // ... more handlers ...
}

func handleWelcomeEmail(ctx context.Context, event FanoutEvent) error {
    // Implementation
    return nil
}

func handleDefaultSettings(ctx context.Context, event FanoutEvent) error {
    // Implementation
    return nil
}
```

### Handler Isolation

Each handler is completely isolated:

1. **Separate Task:** Each handler gets its own asynq task (`fanout:eventType:handlerID`)
2. **Independent Retry:** Each task retries according to its own configuration
3. **Separate Queue:** Handlers can run on different queues
4. **Independent Failure:** One handler failure doesn't affect others

### Comparison: Job Patterns

| Aspect | Fire-and-Forget | Scheduled | Fanout |
|--------|-----------------|-----------|--------|
| Trigger | Immediate, caller-driven | Cron schedule | Event-driven |
| Handlers | Single | Single | Multiple |
| Isolation | Caller isolated | N/A | Handler isolated |
| Default Queue | low | default | per-handler |
| Use case | Non-critical background | Periodic tasks | Event workflows |

### Key Files

| Component | Path |
|-----------|------|
| Pattern implementation | `internal/worker/patterns/fanout.go` |
| Unit tests | `internal/worker/patterns/fanout_test.go` |
| Examples | `internal/worker/patterns/fanout_example_test.go` |

---

## Task Handler Patterns

### Task Type Naming Convention

Define task types in `internal/worker/tasks/types.go`:

```go
package tasks

const (
    TypeNoteArchive = "note:archive"
    // Convention: {domain}:{action}
    // Examples:
    // TypeEmailSend       = "email:send"
    // TypeReportGenerate  = "report:generate"
    // TypeUserCleanup     = "user:cleanup"
)
```

### Handler Pattern (Reference: `note_archive.go`)

Each task handler follows this pattern:

#### 1. Typed Payload Struct

```go
// NoteArchivePayload is the typed payload for note archive tasks.
type NoteArchivePayload struct {
    NoteID uuid.UUID `json:"note_id"`
}
```

#### 2. Task Constructor

```go
// NewNoteArchiveTask creates a new note archive task with default options.
func NewNoteArchiveTask(noteID uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal(NoteArchivePayload{NoteID: noteID})
    if err != nil {
        return nil, fmt.Errorf("marshal note archive payload: %w", err)
    }
    return asynq.NewTask(TypeNoteArchive, payload, asynq.MaxRetry(3)), nil
}
```

#### 3. Handler Struct with Dependency Injection

```go
// NoteArchiveHandler handles note archive tasks.
type NoteArchiveHandler struct {
    logger *zap.Logger
    // Future: inject repository, usecase, etc.
}

// NewNoteArchiveHandler creates a handler with injected dependencies.
func NewNoteArchiveHandler(logger *zap.Logger) *NoteArchiveHandler {
    return &NoteArchiveHandler{logger: logger}
}
```

#### 4. Handle Method with Validation

```go
// Handle processes note archive tasks.
func (h *NoteArchiveHandler) Handle(ctx context.Context, t *asynq.Task) error {
    taskID, _ := asynq.GetTaskID(ctx)

    // 1. Unmarshal payload
    var p NoteArchivePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        h.logger.Error("invalid payload",
            zap.Error(err),
            zap.String("task_type", TypeNoteArchive),
            zap.String("task_id", taskID),
        )
        return fmt.Errorf("unmarshal payload: %v: %w", err, asynq.SkipRetry)
    }

    // 2. Validate payload
    if p.NoteID == uuid.Nil {
        h.logger.Error("missing note_id",
            zap.String("task_type", TypeNoteArchive),
            zap.String("task_id", taskID),
        )
        return fmt.Errorf("note_id is required: %w", asynq.SkipRetry)
    }

    // 3. Process task (call usecase/repository)
    h.logger.Info("archiving note",
        zap.String("task_type", TypeNoteArchive),
        zap.String("task_id", taskID),
        zap.String("note_id", p.NoteID.String()),
    )

    // 4. Return nil on success, error to retry
    return nil
}
```

---

## Error Handling Patterns

### Skip Retry (Validation/Permanent Errors)

Use `asynq.SkipRetry` for errors that won't be fixed by retrying:

```go
// Bad payload - retrying won't help
return fmt.Errorf("invalid payload: %w", asynq.SkipRetry)

// Missing required field
return fmt.Errorf("note_id is required: %w", asynq.SkipRetry)

// Business rule violation
return fmt.Errorf("note already archived: %w", asynq.SkipRetry)
```

### Allow Retry (Transient Errors)

Return regular errors for transient failures:

```go
// Database connection issue - will be retried
return fmt.Errorf("database connection failed: %w", err)

// External service timeout - will be retried
return fmt.Errorf("email service unavailable: %w", err)
```

### Retry Configuration

Set retry limits in task constructor:

```go
asynq.NewTask(TypeNoteArchive, payload,
    asynq.MaxRetry(3),              // Maximum 3 retries
    asynq.Timeout(30*time.Second),  // Task timeout
    asynq.Deadline(time.Now().Add(1*time.Hour)), // Must complete by
)
```

---

## Enqueueing from Usecases

### TaskEnqueuer Interface

Usecases should depend on the interface, not the concrete client:

```go
// internal/worker/tasks/enqueue.go
package tasks

// TaskEnqueuer defines interface for enqueueing tasks from usecase layer.
type TaskEnqueuer interface {
    Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
```

### Usecase Integration Example

```go
// internal/usecase/note/usecase.go
type NoteUsecase struct {
    repo      note.Repository
    enqueuer  tasks.TaskEnqueuer
    logger    *zap.Logger
}

func NewNoteUsecase(repo note.Repository, enqueuer tasks.TaskEnqueuer, logger *zap.Logger) *NoteUsecase {
    return &NoteUsecase{repo: repo, enqueuer: enqueuer, logger: logger}
}

func (u *NoteUsecase) Archive(ctx context.Context, noteID uuid.UUID) error {
    // 1. Business logic
    note, err := u.repo.Get(ctx, noteID)
    if err != nil {
        return err
    }

    // 2. Create and enqueue task
    task, err := tasks.NewNoteArchiveTask(noteID)
    if err != nil {
        return fmt.Errorf("create archive task: %w", err)
    }

    info, err := u.enqueuer.Enqueue(ctx, task, asynq.Queue(worker.QueueDefault))
    if err != nil {
        return fmt.Errorf("enqueue archive task: %w", err)
    }

    u.logger.Info("note archive task enqueued",
        zap.String("note_id", noteID.String()),
        zap.String("task_id", info.ID),
    )

    return nil
}
```

### Wiring in Application Entry Point

Wire the `TaskEnqueuer` in `cmd/api/main.go`:

```go
package main

import (
    "github.com/hibiken/asynq"
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
    noteUsecase "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
)

func main() {
    // ... config and logger setup ...

    // Create Redis options for asynq client
    redisOpt := asynq.RedisClientOpt{
        Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    }

    // Create worker client (implements tasks.TaskEnqueuer)
    workerClient := worker.NewClient(redisOpt)
    defer workerClient.Close()

    // Inject into usecase
    noteUC := noteUsecase.NewNoteUsecase(noteRepo, workerClient, logger)

    // ... rest of app setup (handlers, router, etc.) ...
}
```

> **Note:** The `worker.Client` struct implements the `tasks.TaskEnqueuer` interface, so it can be directly injected into usecases.

---

## Worker Registration

Register handlers in `cmd/worker/main.go`:

```go
package main

import (
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)

func main() {
    // ... config and logger setup ...

    // Create worker server
    srv := worker.NewServer(redisOpt, cfg.Asynq)

    // Add middleware (ORDER MATTERS!)
    srv.Use(
        worker.RecoveryMiddleware(logger),  // 1st: Catch panics
        worker.TracingMiddleware(),          // 2nd: Create OTEL spans
        worker.MetricsMiddleware(),          // 3rd: Record Prometheus metrics
        worker.LoggingMiddleware(logger),    // 4th: Log with context
    )

    // Register task handlers
    noteArchiveHandler := tasks.NewNoteArchiveHandler(logger)
    srv.HandleFunc(tasks.TypeNoteArchive, noteArchiveHandler.Handle)

    // Start server (blocking)
    if err := srv.Start(); err != nil {
        logger.Fatal("Worker error", zap.Error(err))
    }
}
```

---

## Observability

### Middleware Chain

> **CRITICAL:** Middleware order matters!

| Order | Middleware | Purpose |
|-------|------------|---------|
| 1st | `RecoveryMiddleware` | Catch panics, prevent worker crash |
| 2nd | `TracingMiddleware` | Create OpenTelemetry spans |
| 3rd | `MetricsMiddleware` | Record Prometheus metrics |
| 4th | `LoggingMiddleware` | Structured logging with context |

### Prometheus Metrics

From `internal/observability/metrics.go`:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `job_processed_total` | Counter | `task_type`, `queue`, `status` | Total job executions (status=success\|failed) |
| `job_duration_seconds` | Histogram | `task_type`, `queue` | Job execution duration |

### Useful PromQL Queries

```promql
# Job success rate (5m window)
sum(rate(job_processed_total{status="success"}[5m])) / sum(rate(job_processed_total[5m]))

# Job latency p95 by task type
histogram_quantile(0.95, sum(rate(job_duration_seconds_bucket[5m])) by (le, task_type))

# Failed jobs by task type
sum(rate(job_processed_total{status="failed"}[5m])) by (task_type)

# Job processing rate
sum(rate(job_processed_total[5m])) by (task_type, queue)
```

### Grafana Dashboard

Pre-configured dashboard at `deploy/grafana/dashboards/service.json` includes:
- Job Processing Rate panel
- Job Duration panel
- Failure rate tracking

### Structured Logging

All task handlers should use structured logging with consistent fields:

```go
h.logger.Info("processing task",
    zap.String("task_type", TypeNoteArchive),
    zap.String("task_id", taskID),
    zap.String("note_id", p.NoteID.String()),
)
```

---

## Checklist: Creating a New Job

Follow these steps to create a new async job:

### Step 1: Define Task Type Constant

**File:** `internal/worker/tasks/types.go`

```go
const (
    TypeNoteArchive  = "note:archive"
    TypeEmailSend    = "email:send"  // NEW
)
```

### Step 2: Create Task File

**File:** `internal/worker/tasks/email_send.go`

### Step 3: Define Typed Payload

```go
type EmailSendPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

### Step 4: Create Task Constructor

```go
func NewEmailSendTask(to, subject, body string) (*asynq.Task, error) {
    payload, err := json.Marshal(EmailSendPayload{To: to, Subject: subject, Body: body})
    if err != nil {
        return nil, fmt.Errorf("marshal email send payload: %w", err)
    }
    return asynq.NewTask(TypeEmailSend, payload, asynq.MaxRetry(5)), nil
}
```

### Step 5: Create Handler Struct

```go
type EmailSendHandler struct {
    logger      *zap.Logger
    emailClient EmailClient  // Inject dependencies
}

func NewEmailSendHandler(logger *zap.Logger, emailClient EmailClient) *EmailSendHandler {
    return &EmailSendHandler{logger: logger, emailClient: emailClient}
}
```

### Step 6: Implement Handle Method

```go
func (h *EmailSendHandler) Handle(ctx context.Context, t *asynq.Task) error {
    taskID, _ := asynq.GetTaskID(ctx)

    var p EmailSendPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal payload: %v: %w", err, asynq.SkipRetry)
    }

    // Validation
    if p.To == "" {
        return fmt.Errorf("email 'to' is required: %w", asynq.SkipRetry)
    }

    // Process
    if err := h.emailClient.Send(ctx, p.To, p.Subject, p.Body); err != nil {
        return fmt.Errorf("send email: %w", err)  // Will retry
    }

    h.logger.Info("email sent",
        zap.String("task_id", taskID),
        zap.String("to", p.To),
    )
    return nil
}
```

### Step 7: Register Handler

**File:** `cmd/worker/main.go`

```go
emailHandler := tasks.NewEmailSendHandler(logger, emailClient)
srv.HandleFunc(tasks.TypeEmailSend, emailHandler.Handle)
```

### Step 8: Add Unit Tests

**File:** `internal/worker/tasks/email_send_test.go`

> **Reference:** See `internal/worker/tasks/note_archive_test.go` for the canonical test example.

```go
func TestEmailSendHandler_Handle(t *testing.T) {
    tests := []struct {
        name      string
        payload   []byte
        wantErr   bool
        skipRetry bool
    }{
        {
            name:    "valid payload",
            payload: []byte(`{"to":"test@example.com","subject":"Test","body":"Hello"}`),
            wantErr: false,
        },
        {
            name:      "missing to",
            payload:   []byte(`{"subject":"Test","body":"Hello"}`),
            wantErr:   true,
            skipRetry: true,
        },
        {
            name:      "invalid json",
            payload:   []byte(`{invalid}`),
            wantErr:   true,
            skipRetry: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            logger := zap.NewNop()
            mockClient := &MockEmailClient{}
            h := NewEmailSendHandler(logger, mockClient)
            task := asynq.NewTask(TypeEmailSend, tt.payload)

            // Act
            err := h.Handle(context.Background(), task)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                if tt.skipRetry {
                    require.ErrorIs(t, err, asynq.SkipRetry)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Quick Reference

### File Locations

| Component | Path |
|-----------|------|
| Task types | `internal/worker/tasks/types.go` |
| Task handlers | `internal/worker/tasks/{name}.go` |
| Task tests | `internal/worker/tasks/{name}_test.go` |
| Worker main | `cmd/worker/main.go` |
| Worker server | `internal/worker/server.go` |
| Worker client | `internal/worker/client.go` |
| Middleware | `internal/worker/middleware.go`, `metrics_middleware.go` |
| Metrics | `internal/observability/metrics.go` |

### Imports

```go
import (
    "github.com/hibiken/asynq"
    "github.com/iruldev/golang-api-hexagonal/internal/worker"
    "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
)
```

### Common Task Options

```go
asynq.MaxRetry(3)                           // Max 3 retries
asynq.Timeout(30 * time.Second)             // 30s timeout
asynq.Deadline(time.Now().Add(1 * time.Hour)) // Must complete in 1 hour
asynq.Queue(worker.QueueCritical)           // Critical queue
asynq.ProcessIn(5 * time.Minute)            // Delay processing by 5 min
asynq.Unique(1 * time.Hour)                 // Dedupe for 1 hour
```

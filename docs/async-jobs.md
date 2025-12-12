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

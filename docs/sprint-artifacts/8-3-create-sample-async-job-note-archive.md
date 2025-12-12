# Story 8.3: Create Sample Async Job (Note Archive)

Status: done

## Story

As a developer,
I want an example async job implementation,
So that I can understand job patterns.

## Acceptance Criteria

### AC1: Task Handler Exists
**Given** `internal/worker/tasks/note_archive.go` exists
**When** I review the code
**Then** job payload is typed and validated
**And** job handler follows error handling patterns

### AC2: Error Handling Patterns
**Given** a task with invalid payload
**When** task handler processes it
**Then** validation errors skip retry (SkipRetry)
**And** transient errors allow retry
**And** errors are logged with context

### AC3: Usecase Integration
**Given** note_archive task handler
**When** job is enqueued from usecase layer
**Then** task can be created using typed payload
**And** usecase can enqueue via worker.Client dependency

---

## Tasks / Subtasks

- [x] **Task 1: Create task type constants** (AC: #1)
  - [x] Create `internal/worker/tasks/types.go`
  - [x] Define `TypeNoteArchive = "note:archive"` constant
  - [x] Define other task type naming patterns for future reference

- [x] **Task 2: Create NoteArchive task handler** (AC: #1, #2)
  - [x] Create `internal/worker/tasks/note_archive.go`
  - [x] Define `NoteArchivePayload` struct with `NoteID uuid.UUID`
  - [x] Create `NewNoteArchiveTask(noteID uuid.UUID)` constructor
  - [x] Create `Handle(ctx, task)` handler method
  - [x] Validate payload on unmarshal
  - [x] Return `asynq.SkipRetry` for validation errors
  - [x] Add structured logging with zap

- [x] **Task 3: Register handler in worker** (AC: #1)
  - [x] Update `cmd/worker/main.go` to register note_archive handler
  - [x] Import tasks package
  - [x] Add `srv.HandleFunc(tasks.TypeNoteArchive, noteArchiveHandler.Handle)`

- [x] **Task 4: Create task enqueueing interface** (AC: #3)
  - [x] Create `internal/worker/tasks/enqueue.go`
  - [x] Define `TaskEnqueuer` interface for usecase layer dependency

- [x] **Task 5: Unit tests** (AC: #1, #2)
  - [x] Create `internal/worker/tasks/note_archive_test.go`
  - [x] Test valid payload processing
  - [x] Test invalid payload returns SkipRetry
  - [x] Test NewNoteArchiveTask creates correct task

---

## Dev Notes

### Architecture Placement

```
internal/worker/tasks/
├── types.go               # Task type constants
├── note_archive.go        # Sample task handler
├── note_archive_test.go   # Unit tests
└── enqueue.go             # TaskEnqueuer interface
```

**Key:** Task handlers are in `internal/worker/tasks/`, separate from worker infrastructure in `internal/worker/`.

---

### Task Type Naming Convention

Use colon-separated naming: `{domain}:{action}`

```go
// internal/worker/tasks/types.go
const (
    TypeNoteArchive = "note:archive"
    // Future patterns:
    // TypeEmailSend = "email:send"
    // TypeReportGenerate = "report:generate"
)
```

---

### NoteArchive Handler Pattern

```go
// internal/worker/tasks/note_archive.go
package tasks

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
    "github.com/hibiken/asynq"
    "go.uber.org/zap"
)

// NoteArchivePayload is the typed payload for note archive tasks.
type NoteArchivePayload struct {
    NoteID uuid.UUID `json:"note_id"`
}

// NewNoteArchiveTask creates a new note archive task with default options.
func NewNoteArchiveTask(noteID uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal(NoteArchivePayload{NoteID: noteID})
    if err != nil {
        return nil, fmt.Errorf("marshal note archive payload: %w", err)
    }
    return asynq.NewTask(TypeNoteArchive, payload, asynq.MaxRetry(3)), nil
}

// NoteArchiveHandler handles note archive tasks.
// Use NewNoteArchiveHandler to create with injected dependencies.
type NoteArchiveHandler struct {
    logger *zap.Logger
    // Future: inject repository for actual archive operation
}

// NewNoteArchiveHandler creates a handler with injected logger.
func NewNoteArchiveHandler(logger *zap.Logger) *NoteArchiveHandler {
    return &NoteArchiveHandler{logger: logger}
}

// Handle processes note archive tasks.
func (h *NoteArchiveHandler) Handle(ctx context.Context, t *asynq.Task) error {
    taskID, _ := asynq.GetTaskID(ctx)

    var p NoteArchivePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        h.logger.Error("invalid payload", zap.Error(err), zap.String("task_id", taskID))
        return fmt.Errorf("unmarshal payload: %w: %w", err, asynq.SkipRetry)
    }

    if p.NoteID == uuid.Nil {
        h.logger.Error("missing note_id", zap.String("task_id", taskID))
        return fmt.Errorf("note_id is required: %w", asynq.SkipRetry)
    }

    h.logger.Info("archiving note",
        zap.String("task_id", taskID),
        zap.String("note_id", p.NoteID.String()),
    )

    // TODO: Implement actual archive logic when business requirement exists
    // For sample, just log success
    h.logger.Info("note archived successfully",
        zap.String("task_id", taskID),
        zap.String("note_id", p.NoteID.String()),
    )

    return nil
}
```

**Key changes from original pattern:**
- Handler is a struct with injected logger (not using `zap.L()`)
- `asynq.MaxRetry(3)` added to task constructor
- Constructor pattern enables future repository injection

---

### Error Handling Patterns

**Skip Retry for validation errors:**
```go
// Invalid payload - don't retry
return fmt.Errorf("validation error: %w", asynq.SkipRetry)

// Transient error - do retry (default behavior)
return fmt.Errorf("database connection failed: %w", err)
```

---

### Worker Registration

```go
// cmd/worker/main.go updates:
import "github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"

// In main():
noteArchiveHandler := tasks.NewNoteArchiveHandler(logger)
srv.HandleFunc(tasks.TypeNoteArchive, noteArchiveHandler.Handle)
```

---

### TaskEnqueuer Interface (Optional for usecase integration)

```go
// internal/worker/tasks/enqueue.go
package tasks

import (
    "context"
    "github.com/hibiken/asynq"
)

// TaskEnqueuer defines interface for enqueueing tasks from usecase layer.
type TaskEnqueuer interface {
    Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
```

This allows usecases to depend on interface, not concrete worker.Client.

---

### Previous Story Learnings (8-2)

From Story 8-2 Worker Infrastructure:
- Worker package at `internal/worker/`
- Server has `HandleFunc()` method for registration
- Middleware uses zap logger - use zap in task handlers too
- Queue helpers: `EnqueueCritical()`, `EnqueueDefault()`, `EnqueueLow()`
- OTEL tracing middleware already captures task spans
- Code review fixed: use `time.Second` not magic numbers, use `span.SetStatus()`

---

### Asynq Best Practices Applied

- **Typed Payloads:** Define struct for each task type
- **SkipRetry:** Return `asynq.SkipRetry` for validation errors
- **MaxRetry:** Set explicit retry limits in task constructor
- **Idempotency:** Handlers should be idempotent (same ID processed twice = same result)
- **Structured Logging:** Use zap with task_id, task_type fields
- **Error Context:** Wrap errors with context using `%w`
- **Dependency Injection:** Handler struct with constructor for testability

---

### Testing Requirements

Per AGENTS.md:
- ≥70% test coverage required
- Table-driven tests with AAA pattern
- Use mocks for external dependencies

```go
// internal/worker/tasks/note_archive_test.go
func TestNoteArchiveHandler_Handle(t *testing.T) {
    tests := []struct {
        name       string
        payload    []byte
        wantErr    bool
        skipRetry  bool
    }{
        {
            name:    "valid payload",
            payload: []byte(`{"note_id":"550e8400-e29b-41d4-a716-446655440000"}`),
            wantErr: false,
        },
        {
            name:      "invalid json",
            payload:   []byte(`{invalid}`),
            wantErr:   true,
            skipRetry: true,
        },
        {
            name:      "nil note_id",  
            payload:   []byte(`{"note_id":"00000000-0000-0000-0000-000000000000"}`),
            wantErr:   true,
            skipRetry: true,
        },
    }
    // ... test implementation
}
```

---

### File List

**Create:**
- `internal/worker/tasks/types.go`
- `internal/worker/tasks/note_archive.go`
- `internal/worker/tasks/note_archive_test.go`
- `internal/worker/tasks/enqueue.go`

**Modify:**
- `cmd/worker/main.go` - Register note_archive handler

---

## Dev Agent Record

### Agent Model Used
{{agent_model_name_version}}

### Completion Notes
- Handler logs success but doesn't perform actual DB operation (sample only)
- Actual archive logic can be added when note archiving business requirement exists
- TaskEnqueuer interface enables usecase layer to enqueue without direct worker dependency
- Handler uses struct pattern with injected logger for testability

---

## Quality Validation (AI)

**Validation Date:** 2025-12-12
**Improvements Applied:** 4

### Changes Made
1. **Logger Injection:** Changed from `zap.L()` global to injected `*zap.Logger` via constructor
2. **MaxRetry Option:** Added `asynq.MaxRetry(3)` to `NewNoteArchiveTask()`
3. **Handler Factory:** Changed to struct pattern with `NewNoteArchiveHandler()` constructor
4. **Test Examples:** Added table-driven test structure per AGENTS.md requirements


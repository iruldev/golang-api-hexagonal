# Story 8.8: Document Async Job Patterns

Status: done

## Story

As a developer,
I want comprehensive documentation on async job patterns,
So that I can implement new jobs correctly and consistently.

## Acceptance Criteria

### AC1: Async Job Documentation Exists
**Given** `docs/async-jobs.md` exists
**When** I need to create a new async job
**Then** I can follow the step-by-step guide
**And** I understand the worker infrastructure
**And** I know the error handling patterns

### AC2: Copy Pattern Documentation
**Given** the async job documentation
**When** I want to create a new job type
**Then** I can copy and modify the sample NoteArchive pattern
**And** the checklist covers all required steps

---

## Tasks / Subtasks

- [x] **Task 1: Create async job documentation file** (AC: #1)
  - [x] Create `docs/async-jobs.md`
  - [x] Document worker infrastructure overview (entry point, server, client)
  - [x] Explain queue priority system with weight meanings (critical=6, default=3, low=1)
  - [x] Document Server and Client usage patterns with function signatures

- [x] **Task 2: Document task handler patterns** (AC: #1, #2)
  - [x] Document task type naming convention (`{domain}:{action}`)
  - [x] Document typed payload pattern with validation
  - [x] Document handler struct pattern with dependency injection
  - [x] Document error handling (SkipRetry vs retry)
  - [x] Reference NoteArchive as example: `internal/worker/tasks/note_archive.go`

- [x] **Task 3: Document enqueueing patterns** (AC: #1)
  - [x] Document TaskEnqueuer interface usage
  - [x] Document queue helper methods with signatures:
    - `EnqueueCritical(ctx, task) (*TaskInfo, error)`
    - `EnqueueDefault(ctx, task) (*TaskInfo, error)`  
    - `EnqueueLow(ctx, task) (*TaskInfo, error)`
  - [x] Document task options (MaxRetry, Timeout, Deadline)
  - [x] Show usecase layer integration example

- [x] **Task 4: Create new job checklist** (AC: #2)
  - [x] Step 1: Define task type constant in `internal/worker/tasks/types.go`
  - [x] Step 2: Create typed payload struct in `internal/worker/tasks/{name}.go`
  - [x] Step 3: Create task constructor function (e.g., `NewXxxTask()`)
  - [x] Step 4: Create handler struct with constructor (e.g., `XxxHandler`, `NewXxxHandler()`)
  - [x] Step 5: Implement `Handle(ctx, task) error` method with validation
  - [x] Step 6: Register handler in `cmd/worker/main.go`
  - [x] Step 7: Add unit tests in `internal/worker/tasks/{name}_test.go`

- [x] **Task 5: Document observability** (AC: #1)
  - [x] Document middleware chain order (critical for correct behavior):
    1. `RecoveryMiddleware` - catch panics first
    2. `TracingMiddleware` - create spans
    3. `MetricsMiddleware` - record metrics
    4. `LoggingMiddleware` - log with context
  - [x] Document metrics: `job_processed_total`, `job_duration_seconds`
  - [x] Document structured logging patterns with zap
  - [x] Reference Grafana dashboard: `deploy/grafana/dashboards/service.json`

- [x] **Task 6: Update ARCHITECTURE.md** (AC: #1)
  - [x] Add Worker section under **cmd/** documentation (NOT as a new layer)
  - [x] Clarify: Worker is a **secondary entry point** (`cmd/worker/`), not a fifth architectural layer
  - [x] Document worker package responsibilities in Infrastructure Layer section
  - [x] Add cross-reference to `docs/async-jobs.md`

---

## Dev Notes

### Architecture Placement

> **IMPORTANT:** The Worker is a **secondary entry point** (`cmd/worker/main.go`), NOT a new architectural layer. It follows the same hexagonal architecture: task handlers call usecases, which use repositories.

```
cmd/
├── api/
│   └── main.go            # HTTP API entry point
└── worker/
    └── main.go            # Worker entry point (registers handlers, starts server)

internal/worker/
├── server.go              # Worker server with queue priorities
├── client.go              # Task enqueueing client
├── middleware.go          # Recovery, Tracing, Logging middleware
├── metrics_middleware.go  # Prometheus metrics middleware
└── tasks/
    ├── types.go           # Task type constants (e.g., TypeNoteArchive)
    ├── note_archive.go    # Sample task handler (reference implementation)
    ├── note_archive_test.go
    └── enqueue.go         # TaskEnqueuer interface for usecase layer
```

---

### Task Type Constants

From `internal/worker/tasks/types.go`:
```go
const (
    TypeNoteArchive = "note:archive"
    // Naming convention: {domain}:{action}
    // Examples:
    // TypeEmailSend = "email:send"
    // TypeReportGenerate = "report:generate"
)
```

---

### Queue Priority System

From `internal/worker/server.go` - weights determine processing ratio:

| Queue | Weight | Meaning |
|-------|--------|---------|
| `critical` | 6 | 6x more likely to be processed than `low` |
| `default` | 3 | Standard priority for most jobs |
| `low` | 1 | Background/non-urgent tasks |

When all queues have tasks, the worker will process ~6 critical tasks for every 1 low task.

---

### Client Queue Helper Methods

From `internal/worker/client.go`:
```go
// Enqueue adds a task to the specified queue with options.
func (c *Client) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)

// EnqueueCritical adds a task to the critical queue (weight 6).
func (c *Client) EnqueueCritical(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)

// EnqueueDefault adds a task to the default queue (weight 3).
func (c *Client) EnqueueDefault(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)

// EnqueueLow adds a task to the low priority queue (weight 1).
func (c *Client) EnqueueLow(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)
```

---

### Middleware Chain Order

**CRITICAL:** Middleware order matters! From `cmd/worker/main.go`:
```go
srv.Use(
    worker.RecoveryMiddleware(logger),  // 1st: Catch panics (must be first)
    worker.TracingMiddleware(),          // 2nd: Create OTEL spans
    worker.MetricsMiddleware(),          // 3rd: Record Prometheus metrics
    worker.LoggingMiddleware(logger),    // 4th: Log with context (must be last)
)
```

---

### Handler Pattern Reference

See `internal/worker/tasks/note_archive.go` for complete implementation. Key patterns:
- **Typed Payload:** Define struct with JSON tags
- **Constructor:** `NewXxxTask()` returns `*asynq.Task` with options like `asynq.MaxRetry(3)`
- **Handler Struct:** Inject dependencies via constructor, not globals
- **Validation:** Return `asynq.SkipRetry` for invalid payloads (don't retry)
- **Logging:** Use injected `*zap.Logger` with structured fields

---

### Error Handling Patterns

```go
// Skip Retry - for validation/permanent errors (e.g., bad payload):
return fmt.Errorf("invalid payload: %w", asynq.SkipRetry)

// Allow Retry - for transient errors (e.g., network issues):
return fmt.Errorf("database connection failed: %w", err)
```

---

### Metrics Available

From `internal/observability/metrics.go`:
- `job_processed_total{task_type, queue, status}` - Counter (status=success|failed)
- `job_duration_seconds{task_type, queue}` - Histogram

Grafana dashboard: `deploy/grafana/dashboards/service.json` (Job Processing Rate and Job Duration panels)

---

### Previous Story Learnings

**From Story 8-2 (Worker Infrastructure):** Middleware uses zap logger, queue helpers use constants, OTEL tracing captures task spans.

**From Story 8-3 (NoteArchive):** Handler struct with DI, `MaxRetry(3)` in constructor, table-driven tests per AGENTS.md.

**From Story 8-4 (Job Observability):** Metrics middleware pattern, single counter with status label (cleaner than separate counters).

---

### File List

**Create:**
- `docs/async-jobs.md`

**Modify:**
- `ARCHITECTURE.md` - Add Entry Points section and Worker Infrastructure section
- `docs/sprint-artifacts/sprint-status.yaml` - Update story status
- `docs/sprint-artifacts/8-8-document-async-job-patterns.md` - Story file updates

---

### Testing Requirements

Documentation-only story - no unit tests required. Verify documentation accuracy by cross-referencing with actual implementation files.

---

## Dev Agent Record

### Agent Model Used
Claude 3.5 Sonnet (Anthropic)

### Completion Notes
- Created comprehensive `docs/async-jobs.md` covering all async job patterns
- Documentation includes: worker infrastructure, queue priorities, task handler patterns, error handling, enqueueing from usecases, observability, and step-by-step checklist
- Updated `ARCHITECTURE.md` with Entry Points and Worker Infrastructure sections
- Emphasized that Worker is a secondary entry point, NOT a fifth architectural layer
- All task types, queue weights, middleware order, and patterns documented with code examples
- Cross-referenced Grafana dashboard for metrics visualization
- Documentation-only story - no code tests needed, verified accuracy against actual implementation files

---

## Change Log

| Date | Changes |
|------|---------|
| 2025-12-12 | Created comprehensive async job documentation in `docs/async-jobs.md` |
| 2025-12-12 | Updated `ARCHITECTURE.md` with Entry Points and Worker Infrastructure sections |
| 2025-12-12 | Story completed - all 6 tasks finished |
| 2025-12-13 | Code review fixes: Added wiring pattern, cross-reference to test example, updated File List |

---

## Senior Developer Review (AI)

**Review Date:** 2025-12-13
**Outcome:** ✅ Approved (with fixes applied)

### Action Items
- [x] [MEDIUM] Add usecase wiring pattern documentation in `docs/async-jobs.md`
- [x] [MEDIUM] Update File List to include all modified files
- [x] [LOW] Add cross-reference to `note_archive_test.go` in test checklist

### Fixes Applied
1. Added "Wiring in Application Entry Point" section showing how to wire `TaskEnqueuer` in `cmd/api/main.go`
2. Updated File List to include `sprint-status.yaml` and story file itself
3. Added reference to canonical test example `note_archive_test.go`

---

## Quality Validation (AI)

**Validation Date:** 2025-12-12
**Improvements Applied:** 8

### Changes Made
1. **CRITICAL:** Added emphasis that Worker is secondary entry point, not a new architectural layer
2. Added complete Client queue helper method signatures
3. Added middleware chain documentation with order importance note
4. Added `types.go` task type constant example
5. Added queue weight explanation table
6. Simplified NoteArchive code to reference pattern, not duplicate code
7. Added file paths to checklist steps for clarity
8. Condensed testing requirements section

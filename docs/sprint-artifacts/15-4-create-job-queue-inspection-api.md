# Story 15.4: Create Job Queue Inspection API

Status: Done

## Story

As a support engineer,
I want to inspect job queue status via API,
So that I can verify background processing health.

## Acceptance Criteria

1. **Given** `GET /admin/queues/stats` endpoint
   **When** called by admin
   **Then** active, pending, and failed job counts are returned for each queue
   **And** response includes queue names (critical, default, low) with their stats

2. **Given** Redis is connected and healthy
   **When** `GET /admin/queues/stats` is called
   **Then** response includes aggregate stats: total_enqueued, total_active, total_pending, total_failed, total_completed
   **And** per-queue breakdown is included

3. **Given** `GET /admin/queues/{queue}/jobs` endpoint
   **When** called by queue name (critical, default, low)
   **Then** list of jobs in the queue is returned with pagination
   **And** each job includes task_id, type, payload preview, state, created_at, enqueued_at

4. **Given** `GET /admin/queues/{queue}/failed` endpoint
   **When** called by admin
   **Then** failed jobs for the specified queue are returned
   **And** each failed job includes task_id, type, error_message, failed_at, retry_count

5. **Given** `DELETE /admin/queues/{queue}/failed/{task_id}` endpoint
   **When** called by admin with valid task_id
   **Then** the failed task is deleted from the dead queue
   **And** 200 OK is returned with deletion confirmation

6. **Given** `POST /admin/queues/{queue}/failed/{task_id}/retry` endpoint
   **When** called by admin
   **Then** the failed task is requeued for retry
   **And** 200 OK is returned with new task info

7. **Given** invalid queue name (not critical/default/low)
   **When** any queue endpoint is called
   **Then** 400 Bad Request is returned with "Invalid queue name"

8. **Given** all job queue admin endpoints
   **When** accessed without admin role
   **Then** 403 Forbidden is returned (existing RBAC from Story 15.1)

## Tasks / Subtasks

- [x] Create Queue Inspector Interface
  - [x] Define `QueueInspector` interface in `internal/runtimeutil/queueinspector.go`
  - [x] Add `GetQueueStats(ctx) (*QueueStats, error)` method
  - [x] Add `GetJobsInQueue(ctx, queueName, page, pageSize) (*JobList, error)` method
  - [x] Add `GetFailedJobs(ctx, queueName, page, pageSize) (*FailedJobList, error)` method
  - [x] Add `DeleteFailedJob(ctx, queueName, taskID) error` method
  - [x] Add `RetryFailedJob(ctx, queueName, taskID) (*JobInfo, error)` method
  - [x] Create `QueueStats` struct with queue-level and aggregate stats
  - [x] Create `JobInfo`, `FailedJobInfo`, `JobList`, `FailedJobList` structs
  - [x] Define sentinel errors: `ErrInvalidQueue`, `ErrTaskNotFound`
- [x] Implement Asynq Queue Inspector
  - [x] Create `AsynqQueueInspector` implementing `QueueInspector` in `internal/worker/inspector.go`
  - [x] Use `asynq.Inspector` to retrieve queue statistics
  - [x] Use `inspector.GetQueueInfo(queueName)` for per-queue stats
  - [x] Use `inspector.ListPendingTasks()`, `inspector.ListActiveTasks()` for job lists
  - [x] Use `inspector.ListArchivedTasks()` for failed/dead jobs
  - [x] Use `inspector.DeleteTask()` and `inspector.RunTask()` for job management
  - [x] Support pagination via asynq's task list options
- [x] Create Queue Admin Handler
  - [x] Create `internal/interface/http/admin/queues.go`
  - [x] Implement `GET /queues/stats` - GetQueueStatsHandler
  - [x] Implement `GET /queues/{queue}/jobs` - ListJobsHandler
  - [x] Implement `GET /queues/{queue}/failed` - ListFailedJobsHandler
  - [x] Implement `DELETE /queues/{queue}/failed/{task_id}` - DeleteFailedJobHandler
  - [x] Implement `POST /queues/{queue}/failed/{task_id}/retry` - RetryFailedJobHandler
  - [x] Validate queue name against known queues (critical, default, low)
  - [x] Inject `QueueInspector` via dependency
- [x] Define Request/Response DTOs
  - [x] Create `QueueStatsResponse` with aggregate and per-queue stats
  - [x] Create `JobListResponse` with pagination info
  - [x] Create `FailedJobListResponse` with pagination info
  - [x] Create `JobInfoResponse` for single job details
- [x] Register Routes in Admin Router
  - [x] Add `QueueInspector` to `AdminDeps` struct in `routes_admin.go`
  - [x] Add `QueueInspector` to `RouterDeps` struct in `router.go`
  - [x] Register queue routes under `/admin/queues`
- [x] Write Unit Tests
  - [x] Test GetQueueStatsHandler returns aggregate stats
  - [x] Test ListJobsHandler returns job list with pagination
  - [x] Test ListFailedJobsHandler returns failed jobs
  - [x] Test DeleteFailedJobHandler removes task
  - [x] Test RetryFailedJobHandler requeues task
  - [x] Test 400 for invalid queue name
  - [x] Test 404 for task not found
  - [x] Test 403 when non-admin accesses endpoints (covered by existing admin RBAC tests)
- [x] Documentation
  - [x] Add Job Queue Inspection API section to AGENTS.md
  - [x] Document Admin Queue Management pattern

## Dev Notes

### Architecture Patterns

- **Location**: Admin handlers use `internal/interface/http/admin/` following the pattern from Stories 15.1-15.3
- **Route Prefix**: Use `/admin/queues` (under admin routes, not `/api/v1`)
- **Dependency Injection**: Inject `QueueInspector` via `AdminDeps` struct (see `routes_admin.go`)
- **Response Format**: Use standard response envelope from `internal/interface/http/response/`

### Asynq Inspector API Reference

The `asynq.Inspector` provides comprehensive queue inspection capabilities:

```go
import "github.com/hibiken/asynq"

// Create inspector with same Redis connection as worker
inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: "localhost:6379"})

// Get queue statistics
queues, _ := inspector.Queues()  // Returns []string of queue names
info, _ := inspector.GetQueueInfo(queueName)  // Returns *QueueInfo

// QueueInfo contains:
// - Size (total tasks in queue)
// - Active (currently processing)
// - Pending (waiting to be processed)
// - Scheduled (delayed tasks)
// - Retry (tasks waiting for retry)
// - Archived (failed/dead tasks)
// - Completed (successfully processed)
// - Processed (total processed)
// - Failed (total failed)
```

### List Tasks API

```go
// List pending tasks with pagination
tasks, err := inspector.ListPendingTasks(
    queueName,
    asynq.PageSize(20),
    asynq.Page(1),
)

// List archived (dead/failed) tasks
archivedTasks, err := inspector.ListArchivedTasks(
    queueName,
    asynq.PageSize(20),
    asynq.Page(1),
)

// Each task info contains:
// - ID
// - Type
// - Payload
// - State
// - Queue
// - MaxRetry
// - Retried
// - LastErr (for failed tasks)
// - LastFailedAt
// - NextProcessAt
```

### Task Management API

```go
// Delete a failed task
err := inspector.DeleteTask(queueName, taskID)

// Retry a failed task (move from archived to pending)
err := inspector.RunTask(queueName, taskID)

// Run all archived tasks in a queue
count, err := inspector.RunAllArchivedTasks(queueName)
```

### Queue Name Validation

Only allow valid queue names from `internal/worker/server.go`:

```go
// Queue priority constants from worker/server.go
const (
    QueueCritical = "critical"
    QueueDefault  = "default"
    QueueLow      = "low"
)

func isValidQueue(queue string) bool {
    switch queue {
    case worker.QueueCritical, worker.QueueDefault, worker.QueueLow:
        return true
    }
    return false
}
```

### Handler Pattern Reference (from roles.go)

```go
type QueuesHandler struct {
    inspector QueueInspector
    logger    *zap.Logger
}

func NewQueuesHandler(inspector QueueInspector, logger *zap.Logger) *QueuesHandler {
    return &QueuesHandler{
        inspector: inspector,
        logger:    logger,
    }
}

// Extract queue name from URL path
queueName := chi.URLParam(r, "queue")

// Validate queue name
if !isValidQueue(queueName) {
    response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid queue name")
    return
}
```

### Previous Story Learnings (15.1, 15.2, 15.3)

1. **AdminDeps Pattern**: Admin routes receive dependencies via `AdminDeps` struct, not directly from `RouterDeps`
2. **Thread Safety**: Use appropriate synchronization for concurrent operations
3. **Route Registration Pattern**:
   ```go
   if deps.QueueInspector != nil {
       queuesHandler := admin.NewQueuesHandler(deps.QueueInspector, deps.Logger)
       r.Get("/queues/stats", queuesHandler.GetQueueStats)
       r.Get("/queues/{queue}/jobs", queuesHandler.ListJobs)
       r.Get("/queues/{queue}/failed", queuesHandler.ListFailedJobs)
       r.Delete("/queues/{queue}/failed/{task_id}", queuesHandler.DeleteFailedJob)
       r.Post("/queues/{queue}/failed/{task_id}/retry", queuesHandler.RetryFailedJob)
   }
   ```
4. **Error Handling**: Use sentinel errors and map to appropriate HTTP status codes
5. **Pagination**: Use `page` and `page_size` query parameters (default 20, max 100)

### Pagination Query Parameters

Follow the pagination pattern from architecture.md:

```go
// Parse pagination params
page := 1
pageSize := 20

if p := r.URL.Query().Get("page"); p != "" {
    page, _ = strconv.Atoi(p)
}
if ps := r.URL.Query().Get("page_size"); ps != "" {
    pageSize, _ = strconv.Atoi(ps)
    if pageSize > 100 {
        pageSize = 100
    }
}
```

### Response Examples

**GET /admin/queues/stats**
```json
{
  "success": true,
  "data": {
    "aggregate": {
      "total_enqueued": 150,
      "total_active": 5,
      "total_pending": 100,
      "total_scheduled": 20,
      "total_retry": 10,
      "total_archived": 15,
      "total_completed": 5000,
      "total_processed": 5025,
      "total_failed": 25
    },
    "queues": [
      {
        "name": "critical",
        "size": 50,
        "active": 2,
        "pending": 30,
        "scheduled": 5,
        "retry": 3,
        "archived": 10,
        "completed": 1000,
        "processed": 1015,
        "failed": 15
      },
      {
        "name": "default",
        "size": 80,
        "active": 3,
        "pending": 60,
        "scheduled": 10,
        "retry": 5,
        "archived": 5,
        "completed": 3500,
        "processed": 3510,
        "failed": 10
      },
      {
        "name": "low",
        "size": 20,
        "active": 0,
        "pending": 10,
        "scheduled": 5,
        "retry": 2,
        "archived": 0,
        "completed": 500,
        "processed": 500,
        "failed": 0
      }
    ]
  }
}
```

**GET /admin/queues/default/jobs?page=1&page_size=10**
```json
{
  "success": true,
  "data": {
    "jobs": [
      {
        "task_id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "note:archive",
        "payload_preview": "{\"note_id\":\"abc123\"}",
        "state": "pending",
        "queue": "default",
        "max_retry": 3,
        "retried": 0,
        "created_at": "2025-12-14T23:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 60,
      "total_pages": 6
    }
  }
}
```

**GET /admin/queues/critical/failed**
```json
{
  "success": true,
  "data": {
    "failed_jobs": [
      {
        "task_id": "650e8400-e29b-41d4-a716-446655440001",
        "type": "note:export",
        "payload_preview": "{\"note_id\":\"xyz789\"}",
        "error_message": "connection timeout",
        "failed_at": "2025-12-14T22:30:00Z",
        "retry_count": 3,
        "max_retry": 3
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 10,
      "total_pages": 1
    }
  }
}
```

**POST /admin/queues/critical/failed/{task_id}/retry**
```json
{
  "success": true,
  "data": {
    "message": "Task requeued for retry",
    "task_id": "650e8400-e29b-41d4-a716-446655440001",
    "queue": "critical"
  }
}
```

### Testing Standards

From `project_context.md`:
- Table-driven tests with `t.Run` + AAA pattern
- Use testify (require/assert)
- `t.Parallel()` when safe
- Co-located test files (`queues_test.go`)

### File Structure

```
internal/
├── runtimeutil/
│   ├── queueinspector.go           # QueueInspector interface + types
│   └── queueinspector_test.go      # Unit tests (if needed for mock)
├── worker/
│   ├── inspector.go                # AsynqQueueInspector implementation
│   └── inspector_test.go           # Unit tests for inspector
├── interface/http/
│   ├── admin/
│   │   ├── queues.go               # QueuesHandler with 5 endpoints
│   │   └── queues_test.go          # Handler unit tests
│   ├── routes_admin.go             # Add QueueInspector to AdminDeps
│   └── router.go                   # Add QueueInspector to RouterDeps
```

### References

- [Epic 15: Admin/Backoffice API](file:///docs/epics.md#epic-15-admin--backoffice-api)
- [Story 15.1: Admin API Route Group](file:///docs/sprint-artifacts/15-1-create-admin-api-route-group.md)
- [Story 15.2: Feature Flag Management API](file:///docs/sprint-artifacts/15-2-implement-feature-flag-management-api.md)
- [Story 15.3: User Role Management API](file:///docs/sprint-artifacts/15-3-implement-user-role-management-api.md)
- [Story 8.2: Setup Asynq Worker Infrastructure](file:///docs/sprint-artifacts/8-2-setup-asynq-worker-infrastructure.md)
- [Worker Server](file:///internal/worker/server.go) - Queue constants
- [Worker Client](file:///internal/worker/client.go) - Client pattern
- [Asynq Inspector API](https://github.com/hibiken/asynq#inspector) - Official documentation

## Dev Agent Record

### Context Reference

- `docs/epics.md` - Requirements source (Epic 15, Story 15.4)
- `docs/architecture.md` - Security patterns, RBAC authorization, pagination
- `internal/interface/http/routes_admin.go` - Admin route registration pattern
- `internal/interface/http/admin/roles.go` - Handler implementation pattern
- `internal/worker/server.go` - Queue constants (critical, default, low)
- `internal/worker/client.go` - Asynq client pattern

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- Created `QueueInspector` interface in `internal/runtimeutil/queueinspector.go` with types for queue stats, job info, pagination
- Implemented `AsynqQueueInspector` in `internal/worker/inspector.go` using asynq.Inspector API
- Created `QueuesHandler` in `internal/interface/http/admin/queues.go` with 5 endpoints:
  - GET /admin/queues/stats - aggregate and per-queue statistics
  - GET /admin/queues/{queue}/jobs - list pending jobs with pagination
  - GET /admin/queues/{queue}/failed - list archived/failed jobs with pagination
  - DELETE /admin/queues/{queue}/failed/{task_id} - delete failed task
  - POST /admin/queues/{queue}/failed/{task_id}/retry - retry failed task
- Added `QueueInspector` to `AdminDeps` and `RouterDeps` in `routes_admin.go` and `router.go`
- Wrote 11 unit tests covering all endpoints, pagination, error handling, and validation
- Added Job Queue Inspection API section to AGENTS.md with interface definition, endpoints, and wiring examples
- All 8 acceptance criteria satisfied; all regression tests pass

### File List

- [internal/runtimeutil/queueinspector.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/runtimeutil/queueinspector.go)
- [internal/worker/inspector.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/worker/inspector.go)
- [internal/interface/http/admin/queues.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/admin/queues.go)
- [internal/interface/http/admin/queues_test.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/admin/queues_test.go)
- [internal/interface/http/routes_admin.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/routes_admin.go)
- [internal/interface/http/router.go](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/internal/interface/http/router.go)
- [AGENTS.md](file:///Users/khoirulsetyonugroho/Development/go-workspace/src/github.com/iruldev/golang-api-hexagonal/AGENTS.md)

## Senior Developer Review (AI)

**Date**: 2025-12-15
**Result**: Approved with AI Fixes

### Findings
- **High Severity**: `GetQueueStats` previously swallowed errors from the infrastructure, leading to misleading zero-value stats. Fixed to return errors properly.
- **Medium Severity**: Unit tests passed `nil` for logger, risking panics. Fixed to use `zap.NewNop()`.
- **Low Severity**: `IsValidQueue` used duplicated list. Refactored to iterate over `ValidQueues`.

### Actions Taken
- [x] Fixed `GetQueueStats` error handling
- [x] Refactored `IsValidQueue` in `internal/worker/inspector.go`
- [x] Updated `queues_test.go` to use safe no-op logger

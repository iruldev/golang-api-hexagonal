# Story 8.4: Add Job Observability (Metrics + Logging)

Status: done

## Story

As a SRE,
I want job execution metrics and structured logging,
So that I can monitor background job health.

## Acceptance Criteria

### AC1: Prometheus Metrics
**Given** asynq worker is running
**When** jobs are processed
**Then** metrics include: `job_processed_total{status=success|failed}`, `job_duration_seconds`
**And** metrics are labeled by `task_type` and `queue`

> **Note:** Using single counter with `status` label instead of separate counters (cleaner pattern)

### AC2: Structured Logging (Already Implemented)
**Given** asynq worker is running
**When** jobs are processed
**Then** logs include: job_type, job_id, duration, status
**Note:** This is already implemented in `LoggingMiddleware` from Story 8-2

---

## Tasks / Subtasks

- [x] **Task 1: Add job metrics to observability package** (AC: #1)
  - [x] Update `internal/observability/metrics.go`
  - [x] Add `JobProcessedTotal` counter with labels `task_type`, `queue`, `status`
  - [x] Add `JobDurationSeconds` histogram with labels `task_type`, `queue`

- [x] **Task 2: Create MetricsMiddleware for worker** (AC: #1)
  - [x] Create `internal/worker/metrics_middleware.go`
  - [x] Implement `MetricsMiddleware()` recording counters and histograms
  - [x] Use `QueueDefault` constant (queue info not available in handler context)
  - [x] Following same pattern as HTTP middleware

- [x] **Task 3: Register metrics middleware in worker** (AC: #1)
  - [x] Update `cmd/worker/main.go`
  - [x] Add `worker.MetricsMiddleware()` to middleware chain (after tracing, before logging)

- [x] **Task 4: Unit tests for metrics middleware** (AC: #1)
  - [x] Create `internal/worker/metrics_middleware_test.go`
  - [x] Test metrics are recorded on success (status="success")
  - [x] Test metrics are recorded on failure (status="failed")
  - [x] Use `prometheus.NewRegistry()` for test isolation

- [x] **Task 5: Verify logging (already done)** (AC: #2)
  - [x] Confirm `LoggingMiddleware` logs job_type, job_id, duration, status
  - [x] Already implemented in Story 8-2

---

## Dev Notes

### Architecture Placement

```
internal/
├── observability/
│   └── metrics.go            # ADD: JobProcessedTotal, JobDurationSeconds
└── worker/
    ├── middleware.go         # Existing: LoggingMiddleware, TracingMiddleware
    ├── metrics_middleware.go # NEW: MetricsMiddleware
    └── metrics_middleware_test.go
```

---

### Metrics Definition Pattern

Follow existing HTTP metrics pattern in `internal/observability/metrics.go`:

```go
// Job metrics for monitoring worker performance.
var (
    // JobProcessedTotal counts total job executions by task_type, queue, status.
    JobProcessedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "job_processed_total",
            Help: "Total jobs processed",
        },
        []string{"task_type", "queue", "status"},
    )

    // JobDurationSeconds measures job execution duration in seconds.
    JobDurationSeconds = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "job_duration_seconds",
            Help:    "Job duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"task_type", "queue"},
    )
)
```

---

### MetricsMiddleware Pattern

```go
// internal/worker/metrics_middleware.go
package worker

import (
    "context"
    "time"

    "github.com/hibiken/asynq"
    "github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// MetricsMiddleware records Prometheus metrics for task execution.
func MetricsMiddleware() asynq.MiddlewareFunc {
    return func(next asynq.Handler) asynq.Handler {
        return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
            start := time.Now()
            queue, _ := asynq.GetQueue(ctx)

            err := next.ProcessTask(ctx, t)

            duration := time.Since(start).Seconds()
            taskType := t.Type()
            status := "success"
            if err != nil {
                status = "failed"
            }

            observability.JobProcessedTotal.WithLabelValues(taskType, queue, status).Inc()
            observability.JobDurationSeconds.WithLabelValues(taskType, queue).Observe(duration)

            return err
        })
    }
}
```

---

### Middleware Order in Worker

```go
// cmd/worker/main.go
srv.Use(
    worker.RecoveryMiddleware(logger),  // First: catch panics
    worker.TracingMiddleware(),          // Second: create spans
    worker.MetricsMiddleware(),          // Third: record metrics
    worker.LoggingMiddleware(logger),    // Fourth: log with context
)
```

---

### Previous Story Learnings (8-3)

From Story 8-3 Sample Async Job:
- Use `zap.String("task_type", ...)` for logging
- Use `asynq.GetQueue(ctx)` to get queue name
- Handler struct pattern with dependency injection
- Code review fixes: task_type in logs, proper error wrapping

---

### Prometheus Query Examples

```promql
# Job success rate
sum(rate(job_processed_total{status="success"}[5m])) / sum(rate(job_processed_total[5m]))

# Job latency p95
histogram_quantile(0.95, sum(rate(job_duration_seconds_bucket[5m])) by (le, task_type))

# Failed jobs by type
sum(rate(job_processed_total{status="failed"}[5m])) by (task_type)
```

---

### File List

**Create:**
- `internal/worker/metrics_middleware.go`
- `internal/worker/metrics_middleware_test.go`

**Modify:**
- `internal/observability/metrics.go` - Add job metrics
- `cmd/worker/main.go` - Add MetricsMiddleware to chain

---

## Dev Agent Record

### Agent Model Used
{{agent_model_name_version}}

### Completion Notes
- AC2 (structured logging) already done in Story 8-2
- Only need to implement Prometheus metrics (AC1)
- Follow existing HTTP metrics pattern for consistency

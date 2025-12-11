# Story 5.8: Add Context Fields to Logs

Status: done

## Story

As a SRE,
I want trace_id, request_id in logs,
So that I can correlate logs with traces.

## Acceptance Criteria

### AC1: Context fields in logs
**Given** HTTP request with trace context
**When** log is written
**Then** `trace_id`, `request_id`, `path`, `method` are included

---

## Tasks / Subtasks

- [x] **Task 1: Add trace_id to logging middleware** (AC: #1)
  - [x] Extract trace_id from OTEL span context
  - [x] Add trace_id field to logger.Info()

- [x] **Task 2: Verify all required fields** (AC: #1)
  - [x] Confirm request_id is present (already exists)
  - [x] Confirm path is present (already exists)
  - [x] Confirm method is present (already exists)
  - [x] Confirm trace_id is present (new)

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Current Implementation (logging.go)

Already logs: `method`, `path`, `status`, `latency`, `request_id`
Missing: `trace_id`

### Add trace_id from OTEL

```go
import "go.opentelemetry.io/otel/trace"

func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... existing code ...
            
            // Extract trace_id from span context
            spanCtx := trace.SpanContextFromContext(r.Context())
            traceID := ""
            if spanCtx.HasTraceID() {
                traceID = spanCtx.TraceID().String()
            }
            
            logger.Info("request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", ww.statusCode),
                zap.Duration("latency", latency),
                zap.String("request_id", GetRequestID(r.Context())),
                zap.String("trace_id", traceID),  // NEW
            )
        })
    }
}
```

### Expected Log Output

```json
{
  "level": "info",
  "msg": "request",
  "method": "GET",
  "path": "/api/v1/health",
  "status": 200,
  "latency": "1.234ms",
  "request_id": "abc-123",
  "trace_id": "0af7651916cd43dd8448eb211c80319c"
}
```

### Architecture Compliance

**Layer:** `internal/interface/http/middleware/`
**Pattern:** Extends existing logging middleware
**Benefit:** Correlate logs with distributed traces

### References

- [Source: docs/epics.md#Story-5.8]
- [Story 3.3 - Logging Middleware](file:///docs/sprint-artifacts/3-3-implement-logging-middleware.md)
- [Story 3.5 - OpenTelemetry](file:///docs/sprint-artifacts/3-5-add-opentelemetry-trace-propagation.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Eighth story in Epic 5: Observability Suite.
Extends Story 3.3 logging with trace_id.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to modify:
- `internal/interface/http/middleware/logging.go` - Add trace_id field
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

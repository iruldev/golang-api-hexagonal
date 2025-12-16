# Story 2.3b: Implement Trace Correlation in Logs

Status: done

## Story

**As a** developer,
**I want** trace_id and span_id in log entries,
**So that** I can correlate logs with traces.

## Acceptance Criteria

1. **Given** tracing is enabled and request has an active span
   **When** a log entry is written within that request context
   **Then** log entry includes `traceId` field (32 hex chars)
   **And** log entry includes `spanId` field (16 hex chars)

2. **Given** tracing is disabled
   **When** a log entry is written
   **Then** `traceId` and `spanId` fields are absent (not empty string)
   **And** logging functions normally without errors

*Covers: FR14*

## Tasks / Subtasks

- [x] Task 1: Create GetSpanID helper (AC: #1, #2)
  - [x] Add `GetSpanID(ctx context.Context) string` to tracing package
  - [x] Return span ID as 16 hex char string
  - [x] Return empty string if no span in context

- [x] Task 2: Update logging middleware to include trace IDs (AC: #1, #2)
  - [x] Modify `logging.go` to call `GetTraceID(ctx)` and `GetSpanID(ctx)`
  - [x] Add `traceId` and `spanId` fields to log entry
  - [x] Only add fields if values are non-empty (not absent if disabled)

- [x] Task 3: Create trace-aware slog handler (optional enhancement)
  - [x] Using simpler approach of adding fields in middleware only

- [x] Task 4: Write unit tests (AC: #1, #2)
  - [x] Test log entry includes traceId when tracing enabled
  - [x] Test log entry includes spanId when tracing enabled
  - [x] Test log entry excludes trace fields when tracing disabled
  - [x] Test trace IDs have correct format (32 hex / 16 hex)

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

This story primarily modifies **Transport layer** middleware:
- ✅ ALLOWED: domain, chi, otel, stdlib
- ❌ FORBIDDEN: pgx, direct infra imports

The helpers (`GetTraceID`, `GetSpanID`) are in middleware package within transport layer.

### GetTraceID Implementation [Source: Story 2.3a]

Story 2.3a already created `GetTraceID(ctx)` in `tracing.go`:

```go
// internal/transport/http/middleware/tracing.go (already exists)
func GetTraceID(ctx context.Context) string {
    spanCtx := trace.SpanContextFromContext(ctx)
    if !spanCtx.HasTraceID() {
        return ""
    }
    return spanCtx.TraceID().String()
}
```

### GetSpanID Implementation Pattern

```go
// Add to internal/transport/http/middleware/tracing.go
func GetSpanID(ctx context.Context) string {
    spanCtx := trace.SpanContextFromContext(ctx)
    if !spanCtx.HasSpanID() {
        return ""
    }
    return spanCtx.SpanID().String()
}
```

### Logging Middleware Update Pattern

```go
// internal/transport/http/middleware/logging.go
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... existing code ...
            
            // Build log args
            args := []any{
                "method", r.Method,
                "route", routePattern,
                "status", ww.Status(),
                "duration_ms", time.Since(start).Milliseconds(),
                "bytes", ww.Bytes(),
                "requestId", GetRequestID(ctx),
            }
            
            // Add trace context only if present (AC: fields absent if disabled)
            if traceID := GetTraceID(ctx); traceID != "" {
                args = append(args, "traceId", traceID)
            }
            if spanID := GetSpanID(ctx); spanID != "" {
                args = append(args, "spanId", spanID)
            }
            
            logger.InfoContext(ctx, "request completed", args...)
        })
    }
}
```

### Trace ID Formats [Source: OpenTelemetry spec]

| Field | Format | Length |
|-------|--------|--------|
| traceId | Hex string | 32 chars (128 bits) |
| spanId | Hex string | 16 chars (64 bits) |

Example:
```json
{
  "time": "2025-12-17T00:50:27Z",
  "level": "INFO",
  "msg": "request completed",
  "service": "golang-api-hexagonal",
  "env": "development",
  "method": "GET",
  "route": "/health",
  "status": 200,
  "duration_ms": 1,
  "bytes": 27,
  "requestId": "abc123...",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "00f067aa0ba902b7"
}
```

### Previous Story Learnings [Source: Story 2.1, 2.2, 2.3a]

**Files relevant to this story:**
- `internal/transport/http/middleware/tracing.go` - Already has `GetTraceID()`, need to add `GetSpanID()`
- `internal/transport/http/middleware/logging.go` - Need to add trace fields

**Key patterns:**
- `trace.SpanContextFromContext(ctx)` gets OpenTelemetry span context
- `spanCtx.HasTraceID()` / `spanCtx.HasSpanID()` check if valid
- `spanCtx.TraceID().String()` / `spanCtx.SpanID().String()` get hex string

**Middleware order (from Story 2.3a):**
RequestID → Tracing → Logging → RealIP → Recoverer

This order ensures span context is available in Logging middleware.

## Technical Requirements

- **Go version:** 1.24+ [Aligned with go.mod and updated docs]
- **Existing deps:** OpenTelemetry already added in Story 2.3a
- **traceId format:** 32 hex characters (128-bit)
- **spanId format:** 16 hex characters (64-bit)

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Fields should be ABSENT (not empty) when tracing disabled
- Use conditional append to log args
- Maintain middleware order: Tracing before Logging

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- ✅ Added `GetSpanID(ctx context.Context) string` helper to `tracing.go` that extracts span ID from OpenTelemetry span context
- ✅ Updated `logging.go` to conditionally include `traceId` and `spanId` fields only when present (fields are ABSENT when tracing disabled, per AC#2)
- ✅ Added `logKeySpanID` constant for consistent field naming
- ✅ Added 3 new tests to `tracing_test.go` for GetSpanID functionality
- ✅ Added 2 new tests to `logging_test.go` for trace correlation: enabled and disabled scenarios
- ✅ All 37 middleware tests pass, no regressions

### File List

- internal/transport/http/middleware/tracing.go (modified)
- internal/transport/http/middleware/logging.go (modified)
- internal/transport/http/middleware/tracing_test.go (modified)
- internal/transport/http/middleware/logging_test.go (modified)
- docs/sprint-artifacts/sprint-status.yaml (modified)
- go.mod (modified)
- docs/architecture.md (modified)
- docs/prd.md (modified)
- docs/epics.md (modified)

### Change Log

- 2025-12-17: Implemented trace correlation in logs (Story 2.3b)

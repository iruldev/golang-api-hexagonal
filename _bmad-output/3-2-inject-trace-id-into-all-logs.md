# Story 3.2: Inject trace_id into All Logs

Status: done

## Story

**As a** SRE,
**I want** trace_id injected into all log entries,
**So that** I can correlate logs with distributed traces.

**FR:** FR15

## Acceptance Criteria

1. ✅ **Given** tracing is enabled, **When** logs are written, **Then** all log entries include `trace_id` field
2. ✅ **Given** trace_id comes from OTel span context, **When** any downstream component logs, **Then** the log includes the same `trace_id`
3. ✅ **Given** the implementation, **When** unit tests are run, **Then** trace_id injection is verified

## Implementation Summary

### Task 0: Refactor trace context to ctxutil ✅
- Created `ctxutil/trace.go` with `GetTraceID`, `SetTraceID`, `GetSpanID`, `SetSpanID`
- Updated `tracing.go` to use `ctxutil.SetTraceID/SetSpanID`
- Updated `GetTraceID/GetSpanID` in `tracing.go` to check ctxutil first (with OTel fallback)

### Task 1: Extend LoggerFromContext ✅
- Extended `LoggerFromContext` in `observability/logger.go` to include:
  - `request_id` (from Story 3.1)
  - `trace_id` (new)
  - `span_id` (new)
- Zero IDs (`00000000...`) are filtered out

### Task 3: Unit Tests ✅
- `TestLoggerFromContext_WithTraceID` - verifies traceId in log
- `TestLoggerFromContext_WithAllIDs` - verifies all IDs together
- `TestLoggerFromContext_ZeroTraceIDFiltered` - verifies filtering

## Dev Notes

All tests pass:
- `internal/infra/observability/...` - 18 tests PASS
- `internal/transport/http/middleware/...` - 60+ tests PASS

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/ctxutil/trace.go` | **[NEW]** Trace/span ID context utilities |
| `internal/transport/http/middleware/tracing.go` | Updated to use ctxutil |
| `internal/infra/observability/logger.go` | Extended LoggerFromContext |
| `internal/infra/observability/logger_test.go` | Added 4 new tests |
| `internal/transport/http/ctxutil/trace_test.go` | **[NEW]** Unit tests for trace utils |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/ctxutil/trace.go` - NEW
- `internal/transport/http/middleware/tracing.go` - MODIFIED
- `internal/infra/observability/logger.go` - MODIFIED
- `internal/infra/observability/logger_test.go` - MODIFIED
- `internal/transport/http/ctxutil/trace_test.go` - NEW

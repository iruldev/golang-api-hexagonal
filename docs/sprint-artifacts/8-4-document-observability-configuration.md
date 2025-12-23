# Story 8.4: Document Observability Configuration

Status: Done

## Story

As a **developer**,
I want **observability documentation**,
so that **I can configure logging, tracing, and metrics**.

## Acceptance Criteria

1. **Given** I want to configure observability, **When** I read `docs/observability.md`, **Then** I see:
   - Logging configuration (`LOG_LEVEL`, JSON format)
   - Tracing setup (`OTEL_ENABLED`, `OTEL_EXPORTER_OTLP_ENDPOINT`)
   - Metrics endpoint (`/metrics`)
   - Request correlation explanation

2. **And** example log output showing correlation:
   ```json
   {"time":"...","level":"info","msg":"user created","service":"golang-api-hexagonal","env":"development","requestId":"abc123","traceId":"def456","spanId":"..."}
   ```

3. **And** optional docker-compose snippet with Jaeger/Grafana

*Covers: FR67*

## Tasks / Subtasks

- [x] Task 1: Create `docs/observability.md` Document Structure (AC: #1)
  - [x] 1.1 Create document with clear section headers
  - [x] 1.2 Add table of contents for easy navigation
  - [x] 1.3 Add overview explaining the observability stack (slog, OpenTelemetry, Prometheus)

- [x] Task 2: Document Logging Configuration (AC: #1)
  - [x] 2.1 Document `LOG_LEVEL` environment variable (debug, info, warn, error)
  - [x] 2.2 Explain JSON structured logging format using slog with `NewJSONHandler`
  - [x] 2.3 Document required log fields: `time`, `level`, `msg`, `service`, `env`
  - [x] 2.4 Document request-scoped fields: `requestId`, `traceId`, `spanId`
  - [x] 2.5 Provide example log output for different log levels
  - [x] 2.6 Explain log level filtering behavior

- [x] Task 3: Document Tracing Setup (AC: #1)
  - [x] 3.1 Document `OTEL_ENABLED` (true/false, default: false)
  - [x] 3.2 Document `OTEL_EXPORTER_OTLP_ENDPOINT` (e.g., localhost:4317)
  - [x] 3.3 Document `OTEL_EXPORTER_OTLP_INSECURE` for local development
  - [x] 3.4 Explain W3C Trace Context propagation (`traceparent` header)
  - [x] 3.5 Explain span attributes: `http.method`, `http.route`, `http.status_code`
  - [x] 3.6 Document tracer initialization and graceful shutdown

- [x] Task 4: Document Metrics Endpoint (AC: #1)
  - [x] 4.1 Document `/metrics` endpoint with Prometheus format
  - [x] 4.2 Document built-in HTTP metrics:
    - `http_requests_total{method, route, status}` (counter)
    - `http_request_duration_seconds{method, route}` (histogram)
  - [x] 4.3 Document Go runtime metrics (`go_goroutines`, `go_memstats_*`)
  - [x] 4.4 Explain how to add custom metrics using observability package

- [x] Task 5: Document Request Correlation (AC: #1, #2)
  - [x] 5.1 Explain request_id generation and propagation
  - [x] 5.2 Explain X-Request-ID header passthrough
  - [x] 5.3 Explain trace_id/span_id correlation with logs
  - [x] 5.4 Provide example correlated log output (JSON)
  - [x] 5.5 Explain how to trace a request through logs

- [x] Task 6: Add Docker Compose Snippet for Observability Stack (AC: #3)
  - [x] 6.1 Add Jaeger docker-compose.observability.yml snippet
  - [x] 6.2 Add Prometheus scrape configuration example
  - [x] 6.3 Add Grafana configuration (optional)
  - [x] 6.4 Document commands to start observability stack

- [x] Task 7: Add Custom Metrics Guide
  - [x] 7.1 Document `NewCounter`, `NewHistogram`, `NewGauge` functions
  - [x] 7.2 Document `MustNew*` variants for initialization
  - [x] 7.3 Provide code examples for registering custom metrics
  - [x] 7.4 Reference the `internal/shared/metrics.HTTPMetrics` interface

- [x] Task 8: Review and Verify (AC: #1-3)
  - [x] 8.1 Verify all environment variables exist in `.env.example`
  - [x] 8.2 Verify `/metrics` endpoint returns Prometheus format
  - [x] 8.3 Ensure document is scannable with clear headers
  - [x] 8.4 Test example commands and output accuracy

## Dependencies & Blockers

- **Depends on:** Epic 2 (Observability Stack) - Completed
  - Story 2.1: Implement Structured JSON Logging ✅
  - Story 2.2: Implement Request ID Middleware ✅
  - Story 2.3a: Implement OpenTelemetry Tracing Setup ✅
  - Story 2.3b: Implement Trace Correlation in Logs ✅
  - Story 2.4: Implement Prometheus Metrics Endpoint ✅
  - Story 2.5: Add Custom Metrics Utilities ✅
- **Uses:** Existing observability implementation in `internal/infra/observability/`
- **Uses:** Existing config in `.env.example`

## Assumptions & Open Questions

- Assumes observability stack is fully implemented (Epic 2 complete)
- Jaeger/Grafana docker-compose snippets are optional recommendations, not required infrastructure
- Target audience: developers who want to configure and understand the observability features

## Definition of Done

- [x] `docs/observability.md` created with all required sections
- [x] Logging configuration documented with LOG_LEVEL examples
- [x] Tracing setup documented with all OTEL_* environment variables
- [x] Metrics endpoint documented with built-in metrics list
- [x] Request correlation explained with correlated log example
- [x] Optional docker-compose snippets provided for Jaeger/Prometheus
- [x] Custom metrics registration guide included
- [x] All documented configurations verified to work

## Non-Functional Requirements

- Documentation should be scannable with clear headers
- Include actual environment variable values and examples
- Add tips/notes for common configurations
- Keep document practical and action-oriented
- Use GitHub-style alerts for warnings/tips (NOTE, TIP, WARNING)
- Include copy-paste ready configurations

## Testing & Verification

### Manual Verification Steps

1. **Environment Variables:** Verify all documented env vars exist in `.env.example`
2. **Metrics Endpoint:** Curl `/metrics` and verify Prometheus format output
3. **Log Format:** Start service and verify JSON log output structure
4. **Correlation:** Make request and verify requestId/traceId in logs

### Example Verification Commands

```bash
# Verify metrics endpoint
curl http://localhost:8080/metrics | head -50

# Verify log output (start service and check stdout)
make run

# Check available observability config
grep -E "OTEL|LOG_LEVEL" .env.example
```

## Dev Notes

### Observability Implementation Files

| File | Purpose |
|------|---------|
| `internal/infra/observability/logger.go` | Structured JSON logger using slog |
| `internal/infra/observability/tracer.go` | OpenTelemetry tracer initialization |
| `internal/infra/observability/metrics.go` | Prometheus metrics registry and utilities |
| `internal/shared/metrics/http_metrics.go` | HTTPMetrics interface definition |

### Environment Variables Summary (from .env.example)

| Variable | Default | Purpose |
|----------|---------|---------|
| `LOG_LEVEL` | `info` | Logging level: debug, info, warn, error |
| `OTEL_ENABLED` | `false` | Enable OpenTelemetry tracing |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTLP gRPC endpoint |
| `OTEL_EXPORTER_OTLP_INSECURE` | `false` | Use plaintext for local dev |

### Logger Key Constants (from logger.go)

```go
LogKeyService   = "service"
LogKeyEnv       = "env"
LogKeyRequestID = "requestId"
LogKeyTraceID   = "traceId"
LogKeyMethod    = "method"
LogKeyRoute     = "route"
LogKeyStatus    = "status"
LogKeyDuration  = "duration_ms"
LogKeyBytes     = "bytes"
```

### Built-in HTTP Metrics (from metrics.go)

| Metric Name | Type | Labels | Description |
|-------------|------|--------|-------------|
| `http_requests_total` | Counter | method, route, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, route | Request duration distribution |

### Custom Metrics API Example

```go
// Create and register a counter
counter, err := observability.NewCounter(registry, 
    "myapp_orders_total", 
    "Total number of orders processed",
    []string{"status"})

// Use the counter
counter.WithLabelValues("completed").Inc()

// Or use Must* variant for initialization (panics on error)
counter := observability.MustNewCounter(registry, 
    "myapp_orders_total", 
    "Total number of orders processed",
    []string{"status"})
```

### Recommended Document Structure

```markdown
# Observability Guide

## Overview
Brief intro to the observability stack (slog, OpenTelemetry, Prometheus)

## Logging
### Configuration
### Log Format
### Log Levels
### Example Output

## Tracing (OpenTelemetry)
### Configuration
### W3C Trace Context
### Span Attributes
### Viewing Traces

## Metrics
### Configuration
### Built-in Metrics
### Custom Metrics
### Prometheus Scraping

## Request Correlation
### Request ID
### Trace Correlation
### End-to-End Tracing

## Optional: Local Observability Stack
### Jaeger (Tracing UI)
### Prometheus (Metrics)
### Grafana (Dashboards)
```

### References

- [Source: docs/epics.md#Story 8.4] Lines 1711-1735
- [Source: docs/architecture.md#Observability] Lines 498-503
- [Source: docs/project-context.md#Logging Rules] Lines 244-263
- [Source: .env.example] Lines 37-48 (OTEL config), Lines 18-27 (LOG_LEVEL, SERVICE_NAME)
- [Source: internal/infra/observability/] Logger, tracer, metrics implementation
- [Source: FR67] Observability documentation explains logging, tracing, and metrics configuration

### Epic 8 Context

Epic 8 implements Documentation & Developer Guides:
- **8.1:** README Quick Start ✅ (done)
- **8.2:** Architecture and Layer Responsibilities ✅ (done)
- **8.3:** Local Development Workflow ✅ (done)
- **8.4 (this story):** Observability Configuration ← current
- **8.5:** Guide for Adding New Modules (backlog)
- **8.6:** Guide for Adding New Adapters (backlog)

### Previous Story Learnings (8.3)

From Story 8.3 implementation:
- Use GitHub-style alerts (NOTE, TIP, WARNING, CAUTION, IMPORTANT) throughout
- Include copy-paste ready commands and configurations
- Verify all documented commands work before completing
- Use tables for quick reference
- Document written in Indonesia per config setting (communication_language)

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-23)

Files analyzed:
- `docs/epics.md` - Story 8.4 acceptance criteria (lines 1711-1735)
- `.env.example` - Environment variables for observability (88 lines)
- `internal/infra/observability/logger.go` - Logger implementation (65 lines)
- `internal/infra/observability/tracer.go` - Tracer implementation (78 lines)
- `internal/infra/observability/metrics.go` - Metrics implementation (316 lines)
- `docs/project-context.md` - Layer rules and logging conventions
- `docs/sprint-artifacts/8-3-document-local-development-workflow.md` - Previous story learnings

### Agent Model Used

Google Gemini (Antigravity)

### Debug Log References

N/A

### Completion Notes List

- ✅ Created comprehensive `docs/observability.md` documentation (400+ lines)
- ✅ Documented logging configuration: LOG_LEVEL, JSON format, slog fields
- ✅ Documented tracing setup: OTEL_ENABLED, OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE
- ✅ Documented metrics endpoint: /metrics with built-in HTTP metrics and Go runtime metrics
- ✅ Explained request correlation: request_id, trace_id, span_id with example JSON output
- ✅ Added Docker Compose snippets for Jaeger, Prometheus, Grafana
- ✅ Documented custom metrics API: NewCounter, NewHistogram, NewGauge, MustNew* variants
- ✅ Verified environment variables exist in `.env.example`
- ✅ Used GitHub-style alerts (TIP, IMPORTANT, NOTE, CAUTION) throughout

### File List

**Created:**
- `docs/observability.md` - Main observability documentation

### Change Log

| Date | Change |
|------|--------|
| 2025-12-23 | Story 8.4 drafted by create-story workflow |
| 2025-12-23 | Created `docs/observability.md` with complete observability documentation covering logging, tracing, metrics, request correlation, custom metrics, and Docker Compose snippets |


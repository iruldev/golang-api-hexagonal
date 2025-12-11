# Story 3.5: Add OpenTelemetry Trace Propagation

Status: done

## Story

As a SRE,
I want trace context propagated via OTEL,
So that I can trace requests across services.

## Acceptance Criteria

### AC1: Span created with request details
**Given** OTEL exporter is configured
**When** a request is processed
**Then** span is created with request details

### AC2: Trace ID available in context
**Given** OTEL exporter is configured
**When** a request is processed
**Then** trace_id is available in context

### AC3: Child spans can be created
**Given** OTEL exporter is configured
**When** a handler needs to trace operations
**Then** child spans can be created in handlers

---

## Tasks / Subtasks

- [x] **Task 1: Add OTEL dependencies** (AC: #1, #2, #3)
  - [x] Run `go get go.opentelemetry.io/otel`
  - [x] Run `go get go.opentelemetry.io/otel/sdk`
  - [x] Run `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
  - [x] Run `go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
  - [x] Verify dependencies in go.mod

- [x] **Task 2: Create tracer initialization** (AC: #1, #2)
  - [x] Create `internal/observability/tracer.go`
  - [x] Implement `NewTracerProvider(cfg *ObservabilityConfig) (*sdktrace.TracerProvider, error)`
  - [x] Configure OTLP HTTP exporter
  - [x] Set service name from config
  - [x] Return shutdown function for cleanup

- [x] **Task 3: Create OTEL middleware** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/middleware/otel.go`
  - [x] Use otelhttp to wrap handler
  - [x] Extract trace_id from span context
  - [x] Store trace_id in request context

- [x] **Task 4: Create context helper** (AC: #2, #3)
  - [x] Create `GetTraceID(ctx context.Context) string`
  - [x] Create `GetSpan(ctx context.Context) trace.Span`
  - [x] Allow handlers to create child spans

- [x] **Task 5: Wire middleware into router** (AC: #1, #2, #3)
  - [x] Update `router.go` to initialize tracer
  - [x] Add OTEL middleware to chain
  - [x] Wire cfg.Observability into router

- [x] **Task 6: Create middleware tests** (AC: #1, #2, #3)
  - [x] Create `internal/interface/http/middleware/otel_test.go`
  - [x] Test: span is created for request
  - [x] Test: trace_id is in context
  - [x] Test: child span can be created

- [x] **Task 7: Create tracer tests** (AC: #1)
  - [x] Create `internal/observability/tracer_test.go`
  - [x] Test: provider is created successfully
  - [x] Test: shutdown closes provider cleanly

- [x] **Task 8: Verify implementation** (AC: #1, #2, #3)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Tracer Initialization Pattern

```go
// internal/observability/tracer.go
package observability

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

    "github.com/iruldev/golang-api-hexagonal/internal/config"
)

// NewTracerProvider creates a new OpenTelemetry tracer provider.
func NewTracerProvider(ctx context.Context, cfg *config.ObservabilityConfig) (*sdktrace.TracerProvider, func(context.Context) error, error) {
    // Create OTLP exporter
    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(cfg.ExporterEndpoint),
        otlptracehttp.WithInsecure(), // For local dev
    )
    if err != nil {
        return nil, nil, err
    }

    // Create resource with service name
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
        ),
    )
    if err != nil {
        return nil, nil, err
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )

    // Set global provider and propagator
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, tp.Shutdown, nil
}
```

### OTEL Middleware Pattern

```go
// internal/interface/http/middleware/otel.go
package middleware

import (
    "net/http"

    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Otel middleware wraps the handler with OpenTelemetry tracing.
func Otel(operation string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return otelhttp.NewHandler(next, operation)
    }
}
```

### Context Helpers

```go
// GetTraceID extracts trace ID from context.
func GetTraceID(ctx context.Context) string {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().HasTraceID() {
        return span.SpanContext().TraceID().String()
    }
    return ""
}

// GetSpan returns the current span from context.
func GetSpan(ctx context.Context) trace.Span {
    return trace.SpanFromContext(ctx)
}
```

### Router Integration

```go
// internal/interface/http/router.go
func NewRouter(cfg *config.Config) chi.Router {
    logger, _ := observability.NewLogger(&cfg.Log, cfg.App.Env)

    // Initialize tracer if configured
    if cfg.Observability.ExporterEndpoint != "" {
        _, shutdown, err := observability.NewTracerProvider(context.Background(), &cfg.Observability)
        if err != nil {
            log.Printf("Failed to initialize tracer: %v", err)
        } else {
            // Register shutdown in application lifecycle
            // (handled by main.go graceful shutdown)
            _ = shutdown
        }
    }

    r := chi.NewRouter()

    // Global middleware (order matters!)
    r.Use(middleware.Recovery(logger))
    r.Use(middleware.RequestID)
    r.Use(middleware.Otel("api"))  // Story 3.5
    r.Use(middleware.Logging(logger))

    // ...
}
```

### Architecture Compliance

**Layer:** `internal/observability` and `internal/interface/http/middleware`
**Pattern:** OpenTelemetry standard instrumentation
**Dependency:** Official OTEL Go SDK

### Configuration Reference

From `internal/config/config.go`:
```go
type ObservabilityConfig struct {
    ExporterEndpoint string `koanf:"exporter_otlp_endpoint"`
    ServiceName      string `koanf:"service_name"`
}
```

From `.env.example`:
```
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=golang-api-hexagonal
```

### Dependencies

**New:**
- `go.opentelemetry.io/otel`
- `go.opentelemetry.io/otel/sdk`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`

### References

- [Source: docs/epics.md#Story-3.5]
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Story 3.4 - Recovery Middleware](file:///docs/sprint-artifacts/3-4-implement-panic-recovery-middleware.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fifth story in Epic 3: HTTP API Core.
Adds distributed tracing to HTTP requests.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/observability/tracer.go` - Tracer initialization
- `internal/observability/tracer_test.go` - Tracer tests
- `internal/interface/http/middleware/otel.go` - OTEL middleware
- `internal/interface/http/middleware/otel_test.go` - Middleware tests

Files to modify:
- `go.mod` - Add OTEL dependencies
- `internal/interface/http/router.go` - Wire OTEL middleware
- `.env.example` - Add OTEL config hints
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

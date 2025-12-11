# Story 5.9: Create OpenTelemetry Spans

Status: done

## Story

As a SRE,
I want OTEL spans for HTTP requests,
So that I can visualize request flow in Jaeger.

## Acceptance Criteria

### AC1: OTEL spans created and exported
**Given** OTEL exporter is configured
**When** HTTP request is processed
**Then** span is created with request attributes
**And** span is exported to configured backend

---

## Tasks / Subtasks

- [x] **All tasks already completed in Story 3.5!**

---

## Dev Notes

> **NOTE:** This story was already completed as part of **Story 3.5: Add OpenTelemetry Trace Propagation**!

### Current Implementation

**OTEL Middleware (middleware/otel.go):**
```go
func Otel(operation string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return otelhttp.NewHandler(next, operation)
    }
}
```

**Tracer Provider (observability/tracer.go):**
```go
func NewTracerProvider(ctx context.Context, cfg *config.ObservabilityConfig) (*sdktrace.TracerProvider, func(context.Context) error, error) {
    // Creates OTLP exporter
    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(cfg.ExporterEndpoint),
    )
    // Creates resource with service name
    // Sets global provider and propagator (TraceContext, Baggage)
    return tp, tp.Shutdown, nil
}
```

**Span Includes:**
- HTTP method
- URL path
- Status code
- Request duration
- Service name

### Configuration

```env
OBSERVABILITY_EXPORTER_ENDPOINT=localhost:4318
OBSERVABILITY_SERVICE_NAME=golang-api-hexagonal
```

### References

- [Story 3.5 - OpenTelemetry](file:///docs/sprint-artifacts/3-5-add-opentelemetry-trace-propagation.md)

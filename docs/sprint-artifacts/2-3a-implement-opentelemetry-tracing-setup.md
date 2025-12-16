# Story 2.3a: Implement OpenTelemetry Tracing Setup

Status: done

## Story

**As a** developer,
**I want** distributed tracing with OpenTelemetry,
**So that** I can trace requests across service boundaries.

## Acceptance Criteria

1. **Given** `OTEL_ENABLED=true` and `OTEL_EXPORTER_OTLP_ENDPOINT` is set
   **When** the service starts
   **Then** OpenTelemetry tracer provider is initialized
   **And** traces are exported to the configured endpoint

2. **Given** an HTTP request is received (tracing enabled)
   **When** the request is processed
   **Then** a span is created with attributes:
   - `http.method`
   - `http.route` (Chi route pattern, e.g., `/api/v1/users/{id}`)
   - `http.status_code`
   **And** span is properly ended when request completes

3. **Given** request contains `traceparent` header (W3C Trace Context)
   **When** the request is processed
   **Then** the span continues the existing trace (context extracted)

4. **Given** `OTEL_ENABLED=false` (default)
   **When** the service starts
   **Then** no tracer provider is initialized
   **And** application functions normally without exporting spans

*Covers: FR16-18*

## Tasks / Subtasks

- [x] Task 1: Extend Config for OpenTelemetry (AC: #1, #4)
  - [x] Add `OTELEnabled` bool field (default: false)
  - [x] Add `OTELExporterEndpoint` string field (optional)
  - [x] Add `OTELServiceName` to use ServiceName from config
  - [x] Update `.env.example` with OTEL config options

- [x] Task 2: Create tracer provider initialization (AC: #1, #4)
  - [x] Create `internal/infra/observability/tracer.go`
  - [x] Implement `InitTracer(cfg *config.Config) (*sdktrace.TracerProvider, error)`
  - [x] Configure OTLP exporter when enabled
  - [x] Return noop tracer provider when disabled
  - [x] Set service name from config

- [x] Task 3: Create HTTP tracing middleware (AC: #2, #3)
  - [x] Create `internal/transport/http/middleware/tracing.go`
  - [x] Use `otelhttp` handler wrapper or custom middleware
  - [x] Extract W3C trace context from `traceparent` header
  - [x] Create span with http.method, http.route, http.status_code
  - [x] End span when request completes

- [x] Task 4: Wire tracer into main.go (AC: #1, #4)
  - [x] Initialize tracer provider based on config
  - [x] Register tracer provider globally: `otel.SetTracerProvider(tp)`
  - [x] Handle graceful shutdown: `tp.Shutdown(ctx)`
  - [x] Add tracing middleware to router

- [x] Task 5: Write tests (AC: #1, #2, #3, #4)
  - [x] Test tracer disabled by default (noop)
  - [x] Test tracer enabled with endpoint
  - [x] Test span creation with correct attributes
  - [x] Test trace context propagation

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Tracer is in **Infra layer** (`internal/infra/observability/`):
- ✅ ALLOWED: domain, otel, slog, external packages
- ❌ FORBIDDEN: app, transport

Tracing middleware is in **Transport layer** (`internal/transport/http/middleware/`):
- ✅ ALLOWED: domain, chi, otel, stdlib
- ❌ FORBIDDEN: pgx, direct infra imports

### Technology Stack [Source: docs/project-context.md]

| Component | Package | Version |
|-----------|---------|---------|
| Tracing | go.opentelemetry.io/otel | latest |
| OTLP Exporter | go.opentelemetry.io/otel/exporters/otlp/otlptrace | latest |
| HTTP Instrumentation | go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp | latest |

### Tracer Initialization Pattern [Source: docs/architecture.md]

```go
// internal/infra/observability/tracer.go
package observability

import (
    "context"
    "fmt"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    
    "github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

func InitTracer(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
    if !cfg.OTELEnabled {
        // Return noop provider
        return sdktrace.NewTracerProvider(), nil
    }
    
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.OTELExporterEndpoint),
        otlptracegrpc.WithInsecure(), // For local dev
    )
    if err != nil {
        return nil, fmt.Errorf("observability.InitTracer: %w", err)
    }
    
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(cfg.ServiceName),
            semconv.DeploymentEnvironment(cfg.Env),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("observability.InitTracer: resource: %w", err)
    }
    
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )
    
    // Set global propagator for W3C Trace Context
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    
    return tp, nil
}
```

### Tracing Middleware Pattern [Source: docs/architecture.md]

```go
// internal/transport/http/middleware/tracing.go
package middleware

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/iruldev/golang-api-hexagonal/transport/http"

func Tracing(next http.Handler) http.Handler {
    tracer := otel.Tracer(tracerName)
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract trace context from headers (W3C Trace Context)
        ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
        
        // Get route pattern for span name
        routePattern := chi.RouteContext(r.Context()).RoutePattern()
        if routePattern == "" {
            routePattern = r.URL.Path
        }
        
        ctx, span := tracer.Start(ctx, routePattern,
            trace.WithSpanKind(trace.SpanKindServer),
            trace.WithAttributes(
                attribute.String("http.method", r.Method),
                attribute.String("http.route", routePattern),
            ),
        )
        defer span.End()
        
        // Use response wrapper to capture status
        ww := NewResponseWrapper(w)
        
        next.ServeHTTP(ww, r.WithContext(ctx))
        
        // Add status code after request completes
        span.SetAttributes(attribute.Int("http.status_code", ww.Status()))
    })
}
```

### Config Extensions [Source: Story 1.2]

```go
// Add to internal/infra/config/config.go
type Config struct {
    // ... existing fields ...
    
    // OpenTelemetry
    OTELEnabled          bool   `envconfig:"OTEL_ENABLED" default:"false"`
    OTELExporterEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}
```

### .env.example additions

```bash
# OpenTelemetry (optional)
OTEL_ENABLED=false
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

### Middleware Order [Source: Story 2.2]

Current order from Story 2.2: RequestID → Logging → RealIP → Recoverer

New order with tracing: RequestID → Tracing → Logging → RealIP → Recoverer

### Previous Story Learnings [Source: Story 2.1, 2.2]

**Files created:**
- `internal/infra/observability/logger.go` - Logger
- `internal/transport/http/middleware/logging.go` - Request logging
- `internal/transport/http/middleware/requestid.go` - Request ID
- `internal/transport/http/middleware/response_wrapper.go` - Status/bytes capture

**Key patterns:**
- Middleware uses `http.HandlerFunc` pattern
- Response wrapper can be reused for status capture
- `GetRequestID(ctx)` helper for context values
- Route pattern via `chi.RouteContext(r.Context()).RoutePattern()`

## Technical Requirements

- **Go version:** 1.23+ [Source: docs/project-context.md]
- **OTEL packages:**
  - `go.opentelemetry.io/otel`
  - `go.opentelemetry.io/otel/sdk/trace`
  - `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`
  - `go.opentelemetry.io/otel/propagation`
- **W3C Trace Context:** Extract from `traceparent` header

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- OTEL disabled by default (no external deps unless configured)
- Use W3C Trace Context for propagation
- Graceful shutdown required for tracer provider
- Use Chi route pattern for span names (not raw paths)

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

N/A

### Completion Notes List

- ✅ Added `OTELEnabled` and `OTELExporterEndpoint` config fields
- ✅ Created `InitTracer()` function with noop provider when disabled, OTLP gRPC exporter when enabled
- ✅ Created `Tracing` middleware with W3C Trace Context extraction, span creation, and `GetTraceID()` helper
- ✅ Wired tracer into `main.go` with global provider registration and graceful shutdown
- ✅ Updated router middleware order: RequestID → Tracing → Logging → RealIP → Recoverer
- ✅ Added comprehensive tests: 3 tracer tests, 6 tracing middleware tests
- ✅ All tests pass with no regressions

### File List

**New Files:**
- `internal/infra/observability/tracer.go`
- `internal/infra/observability/tracer_test.go`
- `internal/transport/http/middleware/tracing.go`
- `internal/transport/http/middleware/tracing_test.go`

**Modified Files:**
- `internal/infra/config/config.go` - Added OTEL config fields
- `.env.example` - Added OTEL configuration section
- `cmd/api/main.go` - Added tracer initialization and shutdown
- `internal/transport/http/router.go` - Added Tracing middleware
- `go.mod` - Added OpenTelemetry dependencies
- `go.sum` - Updated module checksums
- `docs/sprint-artifacts/sprint-status.yaml` - Synced story status

### Change Log

- 2025-12-17: Implemented OpenTelemetry tracing setup (Story 2.3a)

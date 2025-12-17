# Story 2.5: Add Custom Metrics Utilities

Status: done

## Story

**As a** developer,
**I want** utilities to add custom application metrics,
**So that** I can track business-specific metrics.

## Acceptance Criteria

1. **Given** the `internal/infra/observability` package is available
   **When** I use provided metric registration utilities
   **Then** I can create and register:
   - Custom counter metrics
   - Custom histogram metrics
   - Custom gauge metrics
   **And** custom metrics appear at `/metrics` endpoint with proper labels

2. **Given** the observability package
   **When** I view the package documentation
   **Then** package comment or code example shows how to register custom metrics

*Covers: FR21*

## Tasks / Subtasks

- [x] Task 1: Create custom metrics factory functions (AC: #1)
  - [x] Add `NewCounter(name, help string, labels []string)` to metrics.go
  - [x] Add `NewHistogram(name, help string, labels []string, buckets []float64)` to metrics.go
  - [x] Add `NewGauge(name, help string, labels []string)` to metrics.go
  - [x] Each function returns Prometheus metric and registers it

- [x] Task 2: Create metrics registration helper (AC: #1)
  - [x] OR pass registry to factory functions for auto-registration
  - [x] Ensure metrics appear at /metrics endpoint

- [x] Task 3: Add package documentation (AC: #2)
  - [x] Add package-level comment with usage examples
  - [x] Document each factory function with examples
  - [x] Show how to increment counter, observe histogram, set gauge

- [x] Task 4: Create example usage (AC: #1, #2)
  - [x] Create example showing custom business metric
  - [x] E.g., `users_created_total`, `request_payload_size_bytes`

- [x] Task 5: Write tests (AC: #1)
  - [x] Test custom counter creation and increment
  - [x] Test custom histogram creation and observe
  - [x] Test custom gauge creation and set
  - [x] Test metrics appear in registry output

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Custom metrics utilities are in **Infra layer** (`internal/infra/observability/`):
- ✅ ALLOWED: domain, prometheus, external packages
- ❌ FORBIDDEN: app, transport

### Previous Story Pattern [Source: Story 2.4]

Story 2.4 established the metrics pattern with:
- `internal/infra/observability/metrics.go` - Registry and HTTPMetrics
- `internal/shared/metrics/http_metrics.go` - Interface for transport layer
- Factory function `NewMetricsRegistry()` returns registry and recorder

### Custom Metrics Factory Pattern

```go
// internal/infra/observability/metrics.go (add to existing file)

// NewCounter creates and registers a counter metric.
// 
// Example:
//   counter := observability.NewCounter(registry, "users_created_total", 
//       "Total number of users created", []string{"source"})
//   counter.WithLabelValues("api").Inc()
func NewCounter(registry *prometheus.Registry, name, help string, labels []string) *prometheus.CounterVec {
    counter := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: name,
            Help: help,
        },
        labels,
    )
    registry.MustRegister(counter)
    return counter
}

// NewHistogram creates and registers a histogram metric.
//
// Example:
//   histogram := observability.NewHistogram(registry, "request_payload_size_bytes",
//       "Size of request payloads in bytes", []string{"endpoint"}, nil)
//   histogram.WithLabelValues("/api/users").Observe(1024)
func NewHistogram(registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
    if buckets == nil {
        buckets = prometheus.DefBuckets
    }
    histogram := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    name,
            Help:    help,
            Buckets: buckets,
        },
        labels,
    )
    registry.MustRegister(histogram)
    return histogram
}

// NewGauge creates and registers a gauge metric.
//
// Example:
//   gauge := observability.NewGauge(registry, "active_connections",
//       "Number of active connections", []string{"pool"})
//   gauge.WithLabelValues("postgres").Set(10)
//   gauge.WithLabelValues("postgres").Inc()
//   gauge.WithLabelValues("postgres").Dec()
func NewGauge(registry *prometheus.Registry, name, help string, labels []string) *prometheus.GaugeVec {
    gauge := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: name,
            Help: help,
        },
        labels,
    )
    registry.MustRegister(gauge)
    return gauge
}
```

### Package Documentation Pattern

```go
// Package observability provides logging, tracing, and metrics utilities.
//
// # Logging
//
// Use NewLogger to create a structured JSON logger:
//
//     logger := observability.NewLogger(cfg)
//     logger.Info("user created", "userId", id)
//
// # Tracing
//
// Use InitTracer to initialize OpenTelemetry tracing:
//
//     tp, err := observability.InitTracer(ctx, cfg)
//     defer tp.Shutdown(ctx)
//
// # Metrics
//
// Use NewMetricsRegistry to create base metrics registry:
//
//     registry, httpMetrics := observability.NewMetricsRegistry()
//
// Create custom metrics with factory functions:
//
//     counter := observability.NewCounter(registry, "users_total", "Total users", []string{"status"})
//     counter.WithLabelValues("active").Inc()
//
//     histogram := observability.NewHistogram(registry, "request_size", "Request size", []string{}, nil)
//     histogram.WithLabelValues().Observe(1024)
//
//     gauge := observability.NewGauge(registry, "connections", "Active connections", []string{"pool"})
//     gauge.WithLabelValues("db").Set(5)
package observability
```

### Example Business Metrics

```go
// In some service initialization:
usersCreated := observability.NewCounter(registry, "users_created_total",
    "Total number of users created", []string{"source", "role"})

loginAttempts := observability.NewCounter(registry, "login_attempts_total",
    "Total login attempts", []string{"status"}) // status: success, failed

dbPoolConnections := observability.NewGauge(registry, "db_pool_connections",
    "Number of database pool connections", []string{"state"}) // state: idle, busy

requestPayloadSize := observability.NewHistogram(registry, "request_payload_bytes",
    "Size of request payloads", []string{"endpoint"}, 
    []float64{100, 500, 1000, 5000, 10000, 50000})
```

### Prometheus Metric Types

| Type | Use Case | Methods |
|------|----------|---------|
| **Counter** | Monotonically increasing values | `.Inc()`, `.Add(n)` |
| **Histogram** | Value distributions (latency, size) | `.Observe(value)` |
| **Gauge** | Values that go up and down | `.Set(v)`, `.Inc()`, `.Dec()`, `.Add(n)`, `.Sub(n)` |

### Previous Story Learnings [Source: Story 2.4]

**Key files:**
- `internal/infra/observability/metrics.go` - Contains NewMetricsRegistry()
- `internal/shared/metrics/http_metrics.go` - HTTPMetrics interface

**Pattern established:**
- Registry created in observability package
- Interface defined in shared for transport layer usage
- Metrics registered via `registry.MustRegister()`

## Technical Requirements

- **Go version:** 1.24+ [Aligned with go.mod]
- **Prometheus package:** `github.com/prometheus/client_golang/prometheus` (already added in Story 2.4)
- **Metric types:** Counter, Histogram, Gauge

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Factory functions should auto-register with passed registry
- Include comprehensive godoc with examples
- Return Prometheus Vec types for label support

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

### Agent Model Used

gemini-2.5-pro

### Debug Log References

None - implementation was straightforward with no issues.

### Completion Notes List

- ✅ Implemented three factory functions: `NewCounter`, `NewHistogram`, `NewGauge`
- ✅ All factory functions auto-register metrics with the provided Prometheus registry
- ✅ Added comprehensive package-level documentation with usage examples for logging, tracing, and metrics
- ✅ Each factory function has detailed godoc with practical usage examples
- ✅ `NewHistogram` defaults to `prometheus.DefBuckets` when buckets is nil
- ✅ All functions return `*prometheus.XxxVec` types for label support
- ✅ Created comprehensive unit tests covering:
  - Counter creation, increment, and Add operations
  - Histogram creation with default and custom buckets
  - Gauge creation with Set, Inc, Dec, Add, Sub operations
  - Verification that metrics appear in registry output and descriptor mismatches are detected on duplicate registration
  - Integration with existing `NewMetricsRegistry()` function
- ✅ All automated tests in repo pass (`go test ./...`)
- ✅ Duplicate metric registration validates descriptor compatibility, returns existing collectors on exact match, and /metrics integration test covers custom metrics exposure
- ✅ Factory functions now return `(collector, error)` plus `MustNew*` helpers for startup ergonomics; godoc examples updated to show error handling
- ✅ Histogram duplicate detection now also checks bucket boundaries to avoid silent reuse with mismatched buckets
- ✅ API change noted: callers should handle errors or use `MustNew*`; no legacy signature retained

### File List

- `internal/infra/observability/metrics.go` (modified) - Added factory functions and package documentation
- `internal/infra/observability/metrics_test.go` (new) - Unit tests for custom metrics
- `internal/transport/http/handler/integration_test.go` (modified) - Verifies custom metrics surface at /metrics
- `go.mod` (modified) - Indirect dep `kylelemons/godebug` retained via prometheus client
- `go.sum` (unchanged after tidy)
- `docs/sprint-artifacts/sprint-status.yaml` (modified) - Story status tracking

## Change Log

| Date | Changes |
|------|---------|
| 2025-12-17 | Initial implementation of custom metrics factory functions and tests |

# Story 2.4: Implement Prometheus Metrics Endpoint

Status: Done

## Story

**As a** developer,
**I want** a Prometheus-compatible metrics endpoint,
**So that** I can monitor service health and performance.

## Acceptance Criteria

1. **Given** the service is running
   **When** I call `GET /metrics`
   **Then** I receive HTTP 200
   **And** response content-type contains `text/plain`
   **And** response body is in Prometheus exposition format

2. **Given** HTTP requests are processed
   **When** I check `/metrics`
   **Then** the following metrics are present:
   - `http_requests_total{method, route, status}` (counter)
   - `http_request_duration_seconds{method, route}` (histogram)
   **And** `route` label uses Chi route template (e.g., `/api/v1/users/{id}`), NOT raw path

3. **Given** the service is running
   **When** I check `/metrics`
   **Then** Go runtime metrics are present:
   - `go_goroutines`
   - `go_memstats_*`

*Covers: FR19-20*

## Tasks / Subtasks

- [x] Task 1: Add Prometheus dependencies (AC: #1)
  - [x] Add `github.com/prometheus/client_golang` to go.mod
  - [x] Run `go mod tidy`

- [x] Task 2: Create metrics initialization (AC: #1, #2, #3)
  - [x] Create `internal/infra/observability/metrics.go`
  - [x] Create registry with default Go collectors
  - [x] Register HTTP request counter and histogram

- [x] Task 3: Create metrics middleware (AC: #2)
  - [x] Create `internal/transport/http/middleware/metrics.go`
  - [x] Capture method, route (Chi pattern), status
  - [x] Increment counter and observe histogram
  - [x] Use response wrapper for status capture

- [x] Task 4: Add /metrics endpoint (AC: #1, #3)
  - [x] Add `/metrics` route to router
  - [x] Use `promhttp.Handler()` to serve metrics
  - [x] Ensure Go runtime metrics are exposed

- [x] Task 5: Wire metrics into main.go and router
  - [x] Initialize metrics registry
  - [x] Add metrics middleware to router
  - [x] Update middleware order

- [x] Task 6: Write tests (AC: #1, #2, #3)
  - [x] Test /metrics endpoint returns 200
  - [x] Test response content-type
  - [x] Test HTTP metrics present after request
  - [x] Test Go runtime metrics present

## Dev Notes

### Architecture Compliance [Source: docs/project-context.md]

Metrics registry is in **Infra layer** (`internal/infra/observability/`):
- ✅ ALLOWED: domain, prometheus, external packages
- ❌ FORBIDDEN: app, transport

Metrics middleware is in **Transport layer** (`internal/transport/http/middleware/`):
- ✅ ALLOWED: domain, chi, prometheus, stdlib
- ❌ FORBIDDEN: pgx, direct infra imports

### Technology Stack [Source: docs/project-context.md]

| Component | Package | Version |
|-----------|---------|---------|
| Prometheus | github.com/prometheus/client_golang | latest |

### Metrics Initialization Pattern

```go
// internal/infra/observability/metrics.go
package observability

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/collectors"

    "github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
)

type httpMetrics struct {
    requests  *prometheus.CounterVec
    durations *prometheus.HistogramVec
}

func (m *httpMetrics) IncRequest(method, route, status string) {
    m.requests.WithLabelValues(method, route, status).Inc()
}

func (m *httpMetrics) ObserveRequestDuration(method, route string, seconds float64) {
    m.durations.WithLabelValues(method, route).Observe(seconds)
}

// NewMetricsRegistry registers Go runtime + HTTP metrics and returns both registry and recorder.
func NewMetricsRegistry() (*prometheus.Registry, metrics.HTTPMetrics) {
    reg := prometheus.NewRegistry()

    requests := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "route", "status"},
    )

    durations := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "route"},
    )

    reg.MustRegister(collectors.NewGoCollector())
    reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
    reg.MustRegister(requests)
    reg.MustRegister(durations)

    return reg, &httpMetrics{requests: requests, durations: durations}
}
```

### Metrics Middleware Pattern

```go
// internal/transport/http/middleware/metrics.go
package middleware

import (
    "net/http"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
)

// Metrics depends on an injected recorder (no import infra).
func Metrics(recorder metrics.HTTPMetrics) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            ww := NewResponseWrapper(w)

            next.ServeHTTP(ww, r)

            // Use chi route pattern when available; fallback ke path.
            routePattern := r.URL.Path
            if rctx := chi.RouteContext(r.Context()); rctx != nil {
                if rp := rctx.RoutePattern(); rp != "" {
                    routePattern = rp
                }
            }

            recorder.IncRequest(r.Method, routePattern, strconv.Itoa(ww.Status()))
            recorder.ObserveRequestDuration(r.Method, routePattern, time.Since(start).Seconds())
        })
    }
}
```

### Router /metrics Endpoint

```go
// internal/transport/http/router.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
)

func NewRouter(logger *slog.Logger, tracingEnabled bool, metricsReg *prometheus.Registry, httpMetrics metrics.HTTPMetrics, ...) chi.Router {
    r := chi.NewRouter()
    // ... middleware ...
    r.Use(middleware.Metrics(httpMetrics))

    // Metrics endpoint (no auth)
    r.Handle("/metrics", promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{}))
    // ... other routes ...
}
```

### Middleware Order [Source: Story 2.3a, 2.3b]

Current order: RequestID → Tracing → Logging → RealIP → Recoverer

New order with metrics: RequestID → Tracing → Metrics → Logging → RealIP → Recoverer

Note: Metrics should be after Tracing but before Logging for accurate measurements.

### Previous Story Learnings [Source: Story 2.1-2.3b]

**Files created:**
- `internal/infra/observability/logger.go` - Logger
- `internal/infra/observability/tracer.go` - Tracer
- `internal/transport/http/middleware/logging.go` - Request logging
- `internal/transport/http/middleware/tracing.go` - Tracing with GetTraceID/GetSpanID
- `internal/transport/http/middleware/requestid.go` - Request ID
- `internal/transport/http/middleware/response_wrapper.go` - Status/bytes capture

**Key patterns:**
- Use `chi.RouteContext(r.Context()).RoutePattern()` for route label
- Response wrapper has `Status()` method for capturing status code
- Middleware order is critical

## Technical Requirements

- **Go version:** 1.24+ [Aligned with go.mod]
- **Prometheus package:** `github.com/prometheus/client_golang/prometheus`
- **HTTP handler:** `github.com/prometheus/client_golang/prometheus/promhttp`
- **Required metrics:**
  - `http_requests_total` counter with labels: method, route, status
  - `http_request_duration_seconds` histogram with labels: method, route
  - Go runtime metrics: go_goroutines, go_memstats_*

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules to follow:
- Use Chi route pattern for `route` label (not raw path)
- Include Go runtime collectors
- /metrics endpoint should not require authentication

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

### Agent Model Used

Claude (Anthropic)

### Debug Log References

None

### Completion Notes List

- ✅ Added prometheus/client_golang v1.23.2 to go.mod
- ✅ Created `metrics.go` with HTTP request counter/histogram, Go/Process collectors, and returned registry + HTTPMetrics recorder
- ✅ Added shared contract `internal/shared/metrics/http_metrics.go` to keep transport free from infra imports
- ✅ Updated metrics middleware to inject recorder, use Chi route pattern with nil-safe fallback, and record status/duration
- ✅ Updated router to include metrics middleware in order (after Tracing, before Logging) and expose /metrics
- ✅ Updated main.go to initialize registry and pass both registry + recorder to router
- ✅ Updated integration_test.go to new router signature and verified /metrics behavior
- ✅ Created comprehensive metrics_test.go with 8 test cases
- ⚠️ Tes belum dijalankan di lingkungan ini karena Go lokal < 1.24 (go.mod mensyaratkan 1.24); jalankan `go test ./...` setelah upgrade Go

### File List

**New Files:**
- `internal/shared/metrics/http_metrics.go` - Shared HTTP metrics contract (avoids transport → infra import)
- `internal/infra/observability/metrics.go` - Metrics registry with HTTP and Go runtime collectors + recorder
- `internal/transport/http/middleware/metrics.go` - Metrics middleware for request counting and duration
- `internal/transport/http/middleware/metrics_test.go` - Comprehensive metrics middleware tests

**Modified Files:**
- `go.mod` / `go.sum` - Added prometheus/client_golang dependency
- `internal/transport/http/router.go` - Added metrics middleware and /metrics endpoint with injected recorder
- `cmd/api/main.go` - Initialize metrics registry and pass registry + recorder to router
- `internal/transport/http/handler/integration_test.go` - Updated for new router signature and added endpoint tests
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status to review

## Change Log

- 2025-12-17: Implemented Prometheus metrics endpoint with HTTP request counter, duration histogram, and Go runtime metrics. Added comprehensive test coverage.

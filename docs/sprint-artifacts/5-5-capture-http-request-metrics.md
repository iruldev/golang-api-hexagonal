# Story 5.5: Capture HTTP Request Metrics

Status: done

## Story

As a SRE,
I want HTTP request count, latency, and errors captured,
So that I can monitor API performance.

## Acceptance Criteria

### AC1: HTTP metrics captured and exposed
**Given** HTTP requests are made
**When** I check `/metrics`
**Then** `http_requests_total{method, path, status}` is present
**And** `http_request_duration_seconds{method, path}` histogram exists

---

## Tasks / Subtasks

- [x] **Task 1: Create metrics collector** (AC: #1)
  - [x] Define http_requests_total counter
  - [x] Define http_request_duration_seconds histogram
  - [x] Register with Prometheus registry (via promauto)

- [x] **Task 2: Create metrics middleware** (AC: #1)
  - [x] Capture method, path, status code
  - [x] Record request duration
  - [x] Increment request counter

- [x] **Task 3: Register middleware in router** (AC: #1)
  - [x] Add metrics middleware to chain
  - [x] Applied after Recovery, RequestID, before Otel, Logging

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Metrics Definitions

```go
// internal/observability/metrics.go
var (
    HttpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HttpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration)
}
```

### Metrics Middleware

```go
// internal/interface/http/middleware/metrics.go
func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        ww := NewResponseWriter(w) // Captures status code
        next.ServeHTTP(ww, r)
        
        duration := time.Since(start).Seconds()
        path := r.URL.Path
        method := r.Method
        status := strconv.Itoa(ww.StatusCode())
        
        observability.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
        observability.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
    })
}
```

### Expected /metrics Output

```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/healthz",status="200"} 5
http_requests_total{method="GET",path="/api/v1/health",status="200"} 10

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/healthz",le="0.005"} 5
```

### Architecture Compliance

**Layer:** `internal/observability/` + `internal/interface/http/middleware/`
**Pattern:** Chi middleware with Prometheus collectors
**Benefit:** RED metrics (Rate, Errors, Duration) for SRE monitoring

### References

- [Source: docs/epics.md#Story-5.5]
- [Story 5.4 - Prometheus Metrics](file:///docs/sprint-artifacts/5-4-expose-prometheus-metrics-endpoint.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fifth story in Epic 5: Observability Suite.
Builds on Story 5.4's /metrics endpoint.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/observability/metrics.go` - HTTP metrics collectors
- `internal/interface/http/middleware/metrics.go` - Metrics middleware
- `internal/interface/http/httpx/response_writer.go` - Wrap ResponseWriter

Files to modify:
- `internal/interface/http/router.go` - Add metrics middleware
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

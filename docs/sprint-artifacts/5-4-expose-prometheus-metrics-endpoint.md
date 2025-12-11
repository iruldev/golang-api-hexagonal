# Story 5.4: Expose Prometheus Metrics Endpoint

Status: done

## Story

As a SRE,
I want `/metrics` endpoint for Prometheus,
So that I can scrape application metrics.

## Acceptance Criteria

### AC1: Prometheus metrics endpoint returns metrics
**Given** the HTTP server is running
**When** I request `GET /metrics`
**Then** response is Prometheus text format
**And** standard Go metrics are included

---

## Tasks / Subtasks

- [x] **Task 1: Add Prometheus dependencies** (AC: #1)
  - [x] Add `prometheus/client_golang` to go.mod
  - [x] Import required packages

- [x] **Task 2: Register /metrics handler** (AC: #1)
  - [x] Use `promhttp.Handler()` for /metrics endpoint
  - [x] Register at root level (not under /api/v1)

- [x] **Task 3: Add Go runtime metrics** (AC: #1)
  - [x] Default Go collectors enabled by promhttp
  - [x] Memory, GC, goroutines metrics included

- [x] **Task 4: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Dependencies

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Implementation

**File:** `internal/interface/http/router.go`
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// In NewRouter:
r.Handle("/metrics", promhttp.Handler())
```

### Expected Output

```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
...
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 8
```

### Architecture Compliance

**Layer:** `internal/interface/http/`
**Pattern:** Standard Prometheus handler
**Benefit:** Native Prometheus scraping support

### References

- [Source: docs/epics.md#Story-5.4]
- [prometheus/client_golang](https://github.com/prometheus/client_golang)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fourth story in Epic 5: Observability Suite.
First new implementation work in Epic 5.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to modify:
- `go.mod` - Add Prometheus dependency
- `internal/interface/http/router.go` - Add /metrics handler
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

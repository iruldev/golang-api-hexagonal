# Story 2.5b: Create Internal Router

Status: done

## Story

As a **security engineer**,
I want /metrics to be served on the internal router only,
so that it is not exposed on the public port.

**Dependencies:** Story 2.5a (INTERNAL_PORT config) ✅

## Acceptance Criteria

1. **Given** `NewInternalRouter()` function exists
   **When** called with metrics registry
   **Then** router serves /metrics endpoint

2. **Given** public router (NewRouter)
   **When** /metrics is accessed
   **Then** returns 404 (not routed)

3. **And** /health and /ready remain on public router

## Tasks / Subtasks

- [x] Task 1: Create NewInternalRouter function
  - [x] Added `NewInternalRouter()` in router.go
  - [x] Accepts metrics registry + logger
  - [x] Mounts /metrics handler

- [x] Task 2: Remove /metrics from NewRouter
  - [x] Removed line 95 `/metrics` handler
  - [x] /health, /ready remain on public router

- [x] Task 3: Add Unit Tests
  - [x] `TestNewRouter_MetricsNotExposed` - public router /metrics → 404
  - [x] `TestNewInternalRouter_MetricsAvailable` - internal router /metrics → 200

- [x] Task 4: Update Integration Tests
  - [x] Updated `TestMetricsEndpoint` to use `NewInternalRouter()`

- [x] Task 5: Review Follow-ups (AI)
  - [x] [HIGH] Fixed regression: Wired internal server in main.go to prevent metrics outage

## Dev Notes

### NewInternalRouter Implementation

```go
func NewInternalRouter(
    logger *slog.Logger,
    metricsReg *prometheus.Registry,
) *chi.Mux {
    r := chi.NewRouter()
    r.Use(chiMiddleware.Recoverer)
    r.Handle("/metrics", promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{}))
    return r
}
```

### Key Changes

| File | Change |
|------|--------|
| `router.go` | Added `NewInternalRouter()`, removed /metrics from `NewRouter()` |
| `router_test.go` | Added 2 tests for router separation |
| `integration_test.go` | Updated to use `NewInternalRouter()` for /metrics tests |
| `config_test.go` | Fixed stale `TestLoad_InvalidPortRange` (PORT=0 now valid) |
| `main.go` | Wired up internal server on internal port |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- Build: SUCCESS
- New Tests: 2 PASS
- Integration Tests: All PASS
- Regression: 16 packages ALL PASS

### Completion Notes List

- Created `NewInternalRouter()` function in router.go
- Removed /metrics from public router (now returns 404)
- Added 2 unit tests for router separation
- Updated integration tests to use internal router
- Wired up internal server in main.go (Dual Server Startup implemented)

### File List

- `internal/transport/http/router.go` - MODIFIED (added NewInternalRouter, removed /metrics from NewRouter)
- `internal/transport/http/router_test.go` - MODIFIED (added 2 tests)
- `internal/transport/http/handler/integration_test.go` - MODIFIED (use NewInternalRouter)
- `internal/infra/config/config_test.go` - MODIFIED (fixed PORT test)
- `internal/infra/config/config.go` - MODIFIED (added internal port config)
- `.env.example` - MODIFIED (added internal port docs)
- `cmd/api/main.go` - MODIFIED (wired internal server)
- `README.md` - MODIFIED (updated config and quickstart)

### Change Log

- 2024-12-24: Created NewInternalRouter() for internal /metrics endpoint
- 2024-12-24: Removed /metrics from public router
- 2024-12-24: Added router separation tests
- 2024-12-24: Updated integration tests to use internal router
- 2024-12-24: [Code Review] Wired internal server in main.go (Fixes Critical Regression)

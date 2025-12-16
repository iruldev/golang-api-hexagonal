# System-Level Test Design

**Project:** golang-api-hexagonal
**Date:** 2025-12-16
**Author:** Chat
**Status:** Approved
**Phase:** Solutioning (Phase 3) - Testability Review

---

## Executive Summary

**Scope:** System-level testability review for hexagonal Go API boilerplate

**Assessment Results:**
- ✅ **Controllability:** PASS — Interface-based design, mockable dependencies
- ✅ **Observability:** PASS — Structured logging, metrics, health endpoints
- ✅ **Reliability:** PASS — Layer isolation, transaction rollback, depguard enforcement

**Test Pyramid:**
- Unit: 70% (domain, app, handlers)
- Integration: 25% (postgres repos, config)
- API/E2E: 5% (critical HTTP paths)

**Critical Blockers:** None

---

## Testability Assessment

### Controllability ✅ PASS

| Criterion | Assessment | Evidence |
|-----------|------------|----------|
| Control system state | ✅ | Repository pattern allows direct DB manipulation via test fixtures |
| Mockable dependencies | ✅ | All external deps behind interfaces (`UserRepository`, `Querier`, `TxManager`) |
| Trigger error conditions | ✅ | Domain sentinel errors (`ErrUserNotFound`) enable controlled failure injection |
| Test data seeding | ✅ | goose migrations + pgx allow isolated test database seeding |
| Configuration override | ✅ | envconfig enables test-specific env vars |
| Deterministic IDs | ✅ | `IDGenerator` interface enables fake generator in tests |
| Deterministic time | ✅ | `Clock` interface enables fixed timestamps in tests |

**Strengths:**
- Hexagonal architecture inherently supports mocking at layer boundaries
- `Querier` interface allows swapping real DB with mock/test DB
- `TxManager.WithTx()` enables transaction rollback for test isolation
- Domain purity (`type ID string`) means no external deps to mock in domain tests

---

### Observability ✅ PASS

| Criterion | Assessment | Evidence |
|-----------|------------|----------|
| Inspect system state | ✅ | slog JSON structured logging with `requestId`, `traceId` |
| Deterministic results | ✅ | UUID v7 generated at boundaries via `IDGenerator` |
| Validate NFRs | ✅ | Prometheus metrics endpoint, OpenTelemetry tracing |
| Health endpoints | ✅ | `/health` (liveness), `/ready` (readiness) per architecture |
| Audit trail | ✅ | `AuditEvent` entity with DB persistence |

---

### Reliability ✅ PASS

| Criterion | Assessment | Evidence |
|-----------|------------|----------|
| Tests isolated | ✅ | Hexagonal layers enforce separation; depguard prevents cross-layer pollution |
| Parallel-safe | ✅ | Repository pattern + TxManager enables per-test transactions |
| Reproduce failures | ✅ | Deterministic UUID v7, structured logs with context |
| Loosely coupled | ✅ | All infra behind interfaces; no concrete deps in domain/app |
| Cleanup discipline | ✅ | Transaction rollback or explicit cleanup via fixtures |

---

## Architecturally Significant Requirements (ASRs)

| ASR ID | Requirement | Source | Prob | Impact | Score | Testability Challenge |
|--------|-------------|--------|------|--------|-------|----------------------|
| **ASR-1** | Domain + App coverage ≥ 80% | NFR1 | 2 | 3 | 6 | Need isolated unit tests without external deps |
| **ASR-2** | Health endpoint < 10ms p95 | NFR8 | 1 | 2 | 2 | Performance test infrastructure |
| **ASR-3** | JWT token validation | NFR12 | 2 | 3 | 6 | Auth middleware integration tests |
| **ASR-4** | Rate limiting works | NFR15 | 2 | 2 | 4 | Load test to verify limits |
| **ASR-5** | RFC 7807 error format | NFR17 | 2 | 2 | 4 | API contract tests |
| **ASR-6** | Graceful shutdown | NFR23 | 2 | 3 | 6 | Integration test with signal handling |
| **ASR-7** | depguard boundaries enforced | NFR6 | 3 | 3 | 9 | CI-enforced, not runtime testable |
| **ASR-8** | Structured logging | NFR36 | 1 | 2 | 2 | Log output validation |

**High-Priority (Score ≥6):** ASR-1, ASR-3, ASR-6, ASR-7

---

## Test Levels Strategy

### Test Pyramid

```
      ┌─────────┐
      │  API/E2E │  5% - Critical paths only (health, auth, CRUD)
      │   (5%)   │
      ├─────────┤
      │  Integ  │  25% - Postgres repos, migrations, config
      │  (25%)  │
      ├─────────┤
      │  Unit   │  70% - Domain, App, Handlers, Middleware
      │  (70%)  │
      └─────────┘
```

### Level Mapping by Layer

| Layer | Primary Test Level | Coverage Target | Tools |
|-------|-------------------|-----------------|-------|
| `internal/domain/` | Unit | 100% | stdlib `testing` |
| `internal/app/` | Unit | 90% | testify + mocks |
| `internal/transport/http/handler/` | Unit + API | 80% | testify + httptest |
| `internal/transport/http/middleware/` | Unit + API | 80% | httptest |
| `internal/infra/postgres/` | Integration | 80% | testcontainers-go + pgx |
| `internal/infra/config/` | Unit | 100% | testify |

### Rationale

- Backend API with hexagonal architecture = heavy unit testing
- Domain layer is pure Go (stdlib only) = fast unit tests
- App layer depends only on domain interfaces = easy mocking
- Infra layer has external deps (pgx) = integration tests with testcontainers
- No UI = minimal E2E, focus on API contract validation

---

## NFR Testing Approach

### Security Testing

| NFR | Test Type | Tool | Approach |
|-----|-----------|------|----------|
| JWT Validation | Integration | httptest + testify | Valid/invalid/expired token scenarios |
| Authorization | Unit | testify + mocks | App layer permission checks |
| Rate Limiting | Integration | httptest | Burst requests, verify 429 response |
| Input Validation | Unit | go-playground/validator | Invalid input returns RFC 7807 error |
| OWASP Headers | Integration | httptest | Verify security headers present |

### Performance Testing

| NFR | Test Type | Tool | Approach |
|-----|-----------|------|----------|
| Health < 10ms p95 | Benchmark | `go test -bench` | Benchmark health handler |
| Graceful Shutdown | Integration | Custom harness | Send SIGTERM, verify cleanup |
| DB Connection Pool | Load (v2) | k6 or vegeta | Concurrent requests under load |

### Reliability Testing

| NFR | Test Type | Tool | Approach |
|-----|-----------|------|----------|
| Error Recovery | Unit | testify | App layer error handling paths |
| Transaction Rollback | Integration | pgx + testcontainers | Simulate failure mid-transaction |
| Health/Ready Probes | Integration | httptest | Verify correct status codes |
| Structured Logging | Unit | slog test handler | Verify log fields present |

### Maintainability Testing

| NFR | Test Type | Tool | Approach |
|-----|-----------|------|----------|
| 80% Coverage | CI | `go test -coverprofile` | Coverage gate in CI |
| depguard Boundaries | CI | golangci-lint | Import violation = build failure |
| govulncheck | CI | govulncheck | Vulnerability scanning |
| Lint Errors = 0 | CI | golangci-lint | Zero tolerance |

---

## Test Environment Requirements

| Environment | Purpose | Infrastructure |
|-------------|---------|----------------|
| **Local** | Development + Unit tests | Go toolchain only |
| **CI** | Full test suite | GitHub Actions + testcontainers |
| **Integration** | Postgres integration tests | testcontainers-go (ephemeral) |
| **Load (v2)** | Performance validation | k6 + docker-compose |

---

## Sprint 0 Test Infrastructure Tasks

### P0 (Must Have)

| Task | Description | Owner |
|------|-------------|-------|
| **testcontainers-go setup** | Configure ephemeral Postgres for integration tests | Dev |
| **Test package structure** | Create `test/integration/`, fixtures, helpers | Dev |
| **Coverage CI gate** | `go test -coverprofile` with 80% threshold for domain+app | Dev |
| **IDGenerator interface** | Domain interface + fake generator for deterministic tests | Dev |
| **Clock interface** | Domain interface + fake clock for timestamp assertions | Dev |

### P1 (Should Have)

| Task | Description | Owner |
|------|-------------|-------|
| **slog test handler** | Capture and assert log output in tests | Dev |
| **RFC 7807 contract tests** | 2-3 httptest assertions for error response shape | Dev |
| **Benchmark tests** | Health endpoint performance baseline | Dev |
| **Flaky test controls** | Retry for container startup, pool.Ping() | Dev |

### P2 (Nice to Have)

| Task | Description | Owner |
|------|-------------|-------|
| **Load test framework** | k6 scripts for performance validation (v2) | Dev |
| **Chaos testing** | Fault injection for resilience validation | Dev |

---

## Deterministic Testing Patterns

### IDGenerator Interface (Domain)

```go
// internal/domain/id.go
package domain

// IDGenerator creates unique identifiers
// Real implementation uses UUID v7, test implementation returns predictable values
type IDGenerator interface {
    NewID() ID
}

// internal/infra/idgen/uuid.go
type UUIDGenerator struct{}

func (g *UUIDGenerator) NewID() domain.ID {
    id, _ := uuid.NewV7()
    return domain.ID(id.String())
}

// test/mocks/fake_id_generator.go
type FakeIDGenerator struct {
    NextID string
}

func (g *FakeIDGenerator) NewID() domain.ID {
    return domain.ID(g.NextID)
}
```

### Clock Interface (Domain)

```go
// internal/domain/clock.go
package domain

import "time"

// Clock provides current time
// Real implementation uses time.Now(), test implementation returns fixed time
type Clock interface {
    Now() time.Time
}

// internal/infra/clock/real.go
type RealClock struct{}

func (c *RealClock) Now() time.Time {
    return time.Now().UTC()
}

// test/mocks/fake_clock.go
type FakeClock struct {
    FixedTime time.Time
}

func (c *FakeClock) Now() time.Time {
    return c.FixedTime
}
```

---

## Integration Test Setup

### Testcontainers Configuration

```go
// test/integration/postgres_container.go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
    t.Helper()
    ctx := context.Background()

    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(30*time.Second),
        ),
    )
    require.NoError(t, err)

    connString, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    // Retry pool connection (flaky test control)
    var pool *pgxpool.Pool
    for i := 0; i < 3; i++ {
        pool, err = pgxpool.New(ctx, connString)
        if err == nil {
            if pingErr := pool.Ping(ctx); pingErr == nil {
                break
            }
        }
        time.Sleep(time.Second)
    }
    require.NoError(t, err)
    require.NoError(t, pool.Ping(ctx))

    // Run migrations
    // goose.Up(pool, "migrations")

    cleanup := func() {
        pool.Close()
        _ = container.Terminate(ctx)
    }

    return pool, cleanup
}
```

### Integration Test Rules

```go
// test/integration/integration_test.go
//go:build integration

package integration

// IMPORTANT: Integration tests sharing container should NOT use t.Parallel()
// Each test gets isolated transaction that rolls back

func TestUserRepository_Create(t *testing.T) {
    // NO t.Parallel() - shares container
    pool, cleanup := SetupPostgresContainer(t)
    defer cleanup()

    repo := postgres.NewUserRepository()

    // Use transaction for isolation
    tx, err := pool.Begin(context.Background())
    require.NoError(t, err)
    defer tx.Rollback(context.Background()) // Always rollback

    // Test logic using tx as Querier...
}
```

---

## RFC 7807 Contract Tests

```go
// internal/transport/http/handler/error_contract_test.go
package handler_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRFC7807_ValidationError(t *testing.T) {
    // Setup router with handler that returns validation error
    router := setupTestRouter()

    req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"email": "invalid"}`))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    // Assert Content-Type
    assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

    // Assert status
    assert.Equal(t, http.StatusBadRequest, rec.Code)

    // Assert RFC 7807 required fields
    var problem map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &problem)
    require.NoError(t, err)

    assert.Contains(t, problem, "type")
    assert.Contains(t, problem, "title")
    assert.Contains(t, problem, "status")
    assert.Contains(t, problem, "detail")
    assert.Contains(t, problem, "instance")
    assert.Contains(t, problem, "code")

    // Assert validationErrors shape (for validation errors)
    if errors, ok := problem["validationErrors"].([]interface{}); ok {
        require.NotEmpty(t, errors)
        firstError := errors[0].(map[string]interface{})
        assert.Contains(t, firstError, "field")
        assert.Contains(t, firstError, "message")
    }
}

func TestRFC7807_NotFound(t *testing.T) {
    router := setupTestRouter()

    req := httptest.NewRequest("GET", "/api/v1/users/nonexistent-id", nil)
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
    assert.Equal(t, http.StatusNotFound, rec.Code)

    var problem map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &problem)
    require.NoError(t, err)

    assert.Equal(t, "USER_NOT_FOUND", problem["code"])
    assert.Equal(t, float64(404), problem["status"])
}

func TestRFC7807_InternalError(t *testing.T) {
    router := setupTestRouter()

    // Trigger internal error (e.g., via mock that fails)
    req := httptest.NewRequest("GET", "/api/v1/users/trigger-error", nil)
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
    assert.Equal(t, http.StatusInternalServerError, rec.Code)

    var problem map[string]interface{}
    err := json.Unmarshal(rec.Body.Bytes(), &problem)
    require.NoError(t, err)

    // Internal errors should NOT expose details
    assert.Equal(t, "INTERNAL_ERROR", problem["code"])
    assert.NotContains(t, problem["detail"], "sql")
    assert.NotContains(t, problem["detail"], "panic")
}
```

---

## Quality Gate Criteria

### CI Pipeline Gates

| Gate | Threshold | Action on Failure |
|------|-----------|-------------------|
| **Unit Tests** | 100% pass | Block merge |
| **Integration Tests** | 100% pass | Block merge |
| **Coverage (domain+app)** | ≥ 80% | Block merge |
| **golangci-lint** | 0 errors | Block merge |
| **govulncheck** | 0 critical/high | Block merge |
| **depguard boundaries** | 0 violations | Block merge |

### Test Execution Order (CI)

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    steps:
      - name: Unit Tests
        run: go test -race -coverprofile=coverage.out ./internal/domain/... ./internal/app/...

      - name: Check Coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% below 80% threshold"
            exit 1
          fi

      - name: Integration Tests
        run: go test -race -tags=integration ./test/integration/...

      - name: Lint
        run: golangci-lint run

      - name: Vulnerability Scan
        run: govulncheck ./...
```

---

## Testability Concerns

### ⚠️ Minor Concerns (Non-Blocking)

| Concern | Impact | Resolution |
|---------|--------|------------|
| No E2E framework yet | Low | httptest-based API tests sufficient for MVP |
| testcontainers not set up | Medium | Sprint 0 task (P0) |
| No structured log capture | Low | Add slog test handler (P1) |

### ✅ No Critical Blockers

Architecture is highly testable:
1. Hexagonal layer separation enables isolated unit tests
2. Interface-based design supports easy mocking
3. depguard enforces boundaries at CI level
4. Transaction rollback pattern for integration test cleanup
5. Pure domain layer = zero external deps to mock

---

## Implementation Handoff

**For Sprint 0:**

1. Implement `IDGenerator` and `Clock` interfaces in domain
2. Set up testcontainers-go for Postgres integration tests
3. Create test helper packages (`test/integration/`, `test/mocks/`)
4. Add RFC 7807 contract tests (3 scenarios)
5. Configure coverage gate in CI (80% for domain+app)
6. Add slog test handler for log assertions

**Test Infrastructure Files:**

```
test/
├── integration/
│   ├── postgres_container.go  # testcontainers setup
│   └── user_repo_test.go      # integration tests
├── mocks/
│   ├── fake_id_generator.go
│   ├── fake_clock.go
│   └── fake_user_repo.go
└── helpers/
    └── slog_handler.go        # log capture for assertions
```

---

**Test Design Status:** ✅ APPROVED FOR SPRINT 0

**Next Workflows:**
- `testarch-framework` — Set up test framework infrastructure
- `testarch-ci` — Configure CI pipeline with quality gates

---

**Generated by:** BMad TEA Agent - Test Architect Module
**Workflow:** `.bmad/bmm/testarch/test-design`
**Version:** 4.0 (BMad v6)

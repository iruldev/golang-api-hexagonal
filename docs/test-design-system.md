# System-Level Test Design

**Project:** Backend Service Golang Boilerplate
**Date:** 2025-12-10
**Author:** Gan
**Mode:** System-Level (Phase 3 - Solutioning)
**Status:** Draft

---

## Testability Assessment

### Controllability ✅ PASS

| Criterion | Status | Evidence |
|-----------|--------|----------|
| State control | ✅ | sqlc + golang-migrate, factory patterns |
| External deps mockable | ✅ | Interface-based (TxManager, Logger, Cache, etc.) |
| Error injection | ✅ | AppError model, sentinel errors per domain |
| Dependency injection | ✅ | Composition root in `internal/app/app.go` |

**Details:**
- Database state controllable via migrations and test fixtures
- All external dependencies defined as interfaces in domain layer
- Error conditions can be simulated via mock implementations
- No global state, all dependencies injected at composition root

### Observability ✅ PASS

| Criterion | Status | Evidence |
|-----------|--------|----------|
| State inspection | ✅ | zap structured logging, OTEL tracing |
| Deterministic results | ✅ | Table-driven tests, AAA pattern |
| NFR validation | ✅ | /metrics, trace_id in logs, healthchecks |
| Test debugging | ✅ | Structured logs with context fields |

**Details:**
- All operations logged with trace_id and request_id
- Metrics exposed at /metrics endpoint (Prometheus)
- Health checks at /healthz and /readyz
- OpenTelemetry spans for request tracing

### Reliability ✅ PASS

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Test isolation | ✅ | t.Parallel(), no shared state |
| Failure reproduction | ✅ | Deterministic patterns, runtimeutil.Clock |
| Loose coupling | ✅ | Hexagonal boundaries, layer import rules |
| Clean boundaries | ✅ | domain→usecase→infra→interface |

**Details:**
- Tests can run in parallel without interference
- Clock abstraction enables deterministic time testing
- Layer boundaries prevent tight coupling
- Import rules documented and enforceable

---

## Architecturally Significant Requirements (ASRs)

| ID | Requirement | Category | Prob | Impact | Score | Testing Approach |
|----|-------------|----------|------|--------|-------|------------------|
| ASR-001 | Setup <30min | PERF | 2 | 2 | 4 | Integration: clone→run timer |
| ASR-002 | p95 <100ms | PERF | 2 | 3 | **6** | Load test with k6 |
| ASR-003 | Panic recovery 100% | TECH | 2 | 3 | **6** | Integration: panic handler test |
| ASR-004 | Graceful shutdown | TECH | 2 | 3 | **6** | Integration: SIGTERM handling |
| ASR-005 | Unit coverage ≥70% | TECH | 2 | 2 | 4 | CI coverage gate |
| ASR-006 | No hardcoded secrets | SEC | 2 | 3 | **6** | Static analysis + scanning |

**High-Priority ASRs (Score ≥6):** 4 items require immediate attention

---

## Test Levels Strategy

Based on **backend service boilerplate** architecture:

| Level | Target % | Rationale |
|-------|----------|-----------|
| **Unit** | 60% | Business logic, error handling, mappers, validators |
| **Integration** | 30% | DB queries (sqlc), HTTP handlers, middleware chain |
| **E2E** | 10% | Full request flow through example module |

### Unit Tests (60%)
- Domain entity validations
- Usecase business logic
- Error mapping functions
- JSON struct serialization
- Config validation

### Integration Tests (30%)
- Repository implementations (pgx + sqlc)
- HTTP handler request/response cycle
- Middleware chain behavior
- Database migrations
- Health check endpoints

### E2E Tests (10%)
- Example module full flow (create note → list notes)
- Error response verification
- Observability validation (logs, traces)

---

## NFR Testing Approach

### Security (SEC)
| Test | Tool | Frequency |
|------|------|-----------|
| No hardcoded secrets | gitleaks | Every commit |
| Dependency CVE scan | govulncheck | Weekly |
| Error sanitization | Integration test | PR |
| Input validation | Unit test | PR |

### Performance (PERF)
| Test | Tool | Threshold |
|------|------|-----------|
| Response time p95 | k6 | <100ms |
| Startup time | Integration | <5s |
| Memory usage | pprof | Baseline +10% |

### Reliability (TECH)
| Test | Type | Validation |
|------|------|------------|
| Panic recovery | Integration | No 5xx leak |
| Graceful shutdown | Integration | In-flight complete |
| DB connection retry | Integration | Reconnect success |

### Maintainability
| Metric | Tool | Threshold |
|--------|------|-----------|
| Unit coverage | go test -cover | ≥70% |
| Lint pass | golangci-lint | 100% |
| Cyclomatic complexity | golangci-lint | ≤15 |

---

## Test Environment Requirements

### Local Development
```
docker-compose:
  - postgres (primary)
  - jaeger (optional, tracing)
```

### CI/CD Pipeline
```
GitHub Actions:
  - go test ./...
  - golangci-lint run
  - govulncheck
  - coverage report
```

### Staging (Optional)
- Full observability stack
- Performance testing with k6

---

## Testability Concerns

### No Blockers Found ✅

The architecture is well-designed for testability:
- Interface-based dependencies
- Composition root pattern
- Hexagonal layer separation
- Deterministic time via Clock abstraction

### Minor Recommendations

1. **Add testcontainers-go** for real Postgres in integration tests
2. **Consider goleak** for goroutine leak detection
3. **Add httptest** utilities in `internal/interface/http/httpx/testing.go`

---

## Recommendations for Sprint 0

### Test Infrastructure Setup

| Task | Priority | Owner |
|------|----------|-------|
| Configure go test with coverage | P0 | Dev |
| Setup golangci-lint config | P0 | Dev |
| Create test fixtures for note module | P0 | Dev |
| Add Makefile targets (test, lint, coverage) | P0 | Dev |
| Configure CI workflow | P1 | DevOps |

### Framework Recommendations

| Framework | Purpose |
|-----------|---------|
| testify | Assertions (require/assert) |
| gomock or mockery | Mock generation |
| testcontainers-go | Real DB in tests |
| httptest | HTTP handler testing |
| k6 | Performance testing |

---

## Quality Gate Criteria

### Sprint 0 Gate
- [ ] All unit tests pass
- [ ] golangci-lint passes
- [ ] Coverage ≥70% on example module
- [ ] No CVE vulnerabilities (govulncheck)

### Pre-Release Gate
- [ ] All P0 tests pass (100%)
- [ ] All P1 tests pass (≥95%)
- [ ] Performance targets met
- [ ] Security scan clean

---

## Related Documents

- PRD: `docs/prd.md`
- Architecture: `docs/architecture.md`
- Project Context: `docs/project_context.md`

---

**Generated by:** TEA Agent - Test Architect Module
**Workflow:** `.bmad/bmm/testarch/test-design`
**Mode:** System-Level (Phase 3)

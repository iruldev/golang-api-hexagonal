# System-Level Test Design

**Date:** 2025-12-15
**Author:** Gan
**Project:** Backend Service Golang Boilerplate (Go Golden Template)
**Status:** Approved

---

## Executive Summary

**Mode:** System-Level Testability Review (Phase 3 - Solutioning)
**Architecture:** Hexagonal (Clean Architecture) Brownfield Upgrade
**Purpose:** Assess architecture testability before sprint planning

**Overall Assessment:** ✅ PASS - Architecture is testable with minor recommendations

---

## Testability Assessment

### Controllability ✅ PASS

| Aspect | Status | Evidence |
|--------|--------|----------|
| State Control | ✅ | `make reset` clears DB, Redis, containers |
| API Seeding | ✅ | sqlc + pgx enables programmatic data setup |
| Dependency Injection | ✅ | Constructor injection in all layers |
| Mockable Interfaces | ✅ | Repository pattern with domain interfaces |
| Error Injection | ⚠️ | Wrapper pattern in Epic 2 enables this |
| External Mocks | ✅ | Interface-based design supports mocking |

**Recommendations:**
- Implement fault injection via `internal/infra/wrapper/` context wrappers
- Create test factories in `tests/testutil/factories.go`

### Observability ✅ PASS

| Aspect | Status | Evidence |
|--------|--------|----------|
| Structured Logging | ✅ | Zap JSON logs with consistent field names |
| Distributed Tracing | ✅ | OpenTelemetry spans (Epic 4) |
| Metrics | ✅ | Prometheus `/metrics` endpoint |
| Audit Logging | ✅ | Story 4.4 implements audit trail |
| Test Results | ✅ | `make verify` with clear exit codes |

**Recommendations:**
- Ensure trace_id propagates to all async workers
- Add test-specific log levels

### Reliability ✅ PASS

| Aspect | Status | Evidence |
|--------|--------|----------|
| Test Isolation | ✅ | Hybrid: unit collocated, integration separate |
| Parallel Safety | ⚠️ | Needs transaction isolation for DB tests |
| Reproducibility | ✅ | sqlc queries versioned, deterministic |
| Cleanup | ✅ | `make reset`, test transaction rollback |
| Loose Coupling | ✅ | Hexagonal layers with depguard enforcement |

**Recommendations:**
- Use transaction wrapping for integration tests
- Build tag: `//go:build integration`

---

## Architecturally Significant Requirements (ASRs)

| ASR ID | Requirement | Category | Prob | Impact | Score | Mitigation |
|--------|-------------|----------|------|--------|-------|------------|
| ASR-1 | Context propagation mandatory | TECH | 2 | 2 | 4 | contextcheck linter + wrappers |
| ASR-2 | Coverage ≥80% domain/usecase | TECH | 2 | 2 | 4 | CI coverage gate in `make verify` |
| ASR-3 | CI p50 ≤8min, p95 ≤15min | PERF | 2 | 3 | **6** | Parallel jobs, caching, selective tests |
| ASR-4 | 0 Critical vulnerabilities | SEC | 1 | 3 | 3 | govulncheck in CI pipeline |
| ASR-5 | make lint ≤60sec | PERF | 2 | 2 | 4 | golangci-lint caching |
| ASR-6 | Test flake rate <1% | TECH | 2 | 2 | 4 | Deterministic tests, isolation |
| ASR-7 | Rate limiting works under load | PERF | 2 | 2 | 4 | API integration tests with Redis |

**High Priority (Score ≥6):** ASR-3 (CI performance) requires immediate attention

---

## Test Levels Strategy

| Level | Percentage | Framework | Target |
|-------|------------|-----------|--------|
| **Unit** | 70% | `go test` + testify | Domain logic, usecase |
| **Integration** | 25% | `go test` + dockertest | DB, Redis, HTTP clients |
| **E2E** | 5% | httptest | Critical API paths |

**Rationale (API-Heavy Platform):**
- High unit percentage for fast feedback on domain logic
- Integration tests for infrastructure boundaries (DB, cache)
- Minimal E2E for critical user journeys only

### Test Framework Decisions

| Purpose | Tool | Notes |
|---------|------|-------|
| Unit Tests | `testing` + `testify/assert` | Standard Go, familiar |
| Mocking | `testify/mock` or mockery | Interface-based |
| DB Tests | `dockertest` or `testcontainers-go` | Real PostgreSQL in container |
| HTTP Tests | `httptest` | Built-in, chi-compatible |
| Load Tests | `k6` | Performance NFRs validation |
| Coverage | `go test -cover` | CI gate enforcement |

---

## NFR Testing Approach

### Security (NFR-S1 to NFR-S3)

| NFR | Test Approach | Tool |
|-----|---------------|------|
| 0 Critical vulns | CI vulnerability scan | govulncheck |
| Secrets via env only | Config validation tests | Unit test + gitleaks CI |
| Auth/RBAC | Middleware unit tests | Go testing + httptest |

### Performance (NFR-P1 to NFR-P5)

| NFR | Test Approach | Tool |
|-----|---------------|------|
| CI p50 ≤8min | CI duration tracking | GitHub Actions timing |
| make lint ≤60sec | Lint benchmark | `time make lint` in CI |
| make verify ≤3min | Test suite timing | CI step timing |

### Reliability (NFR-R1 to NFR-R4)

| NFR | Test Approach | Tool |
|-----|---------------|------|
| Flake rate <1% | Test stability tracking | CI flake detection |
| Pass rate >95% | CI success metrics | GitHub Actions dashboard |
| Reproducibility | Deterministic seed data | Test fixtures with factories |

### Maintainability (NFR-M1 to NFR-M5)

| NFR | Test Approach | Tool |
|-----|---------------|------|
| Coverage ≥80% | Coverage gate | `go test -cover` |
| Complexity ≤15 | Static analysis | gocyclo in golangci-lint |

---

## Test Environment Requirements

| Environment | Purpose | Infrastructure |
|-------------|---------|----------------|
| **Local** | Dev testing | docker-compose (postgres, redis) |
| **CI** | Automated tests | GitHub Actions + Docker services |
| **Staging** | Integration validation | Kubernetes (optional) |

**Required Services:**
- PostgreSQL 16+ (via docker-compose or testcontainers)
- Redis 7+ (via docker-compose or testcontainers)

---

## Testability Concerns (Minor)

| Concern | Severity | Status | Recommendation |
|---------|----------|--------|----------------|
| No fault injection | Medium | Planned | Story 2.3 wrappers enable this |
| Integration test isolation | Low | Planned | Transaction rollback pattern |
| Test fixtures missing | Medium | Planned | Create `tests/testutil/factories.go` |
| Async worker testing | Low | Future | Add test hooks for worker queues |

**No blockers identified.** All concerns are addressable in MVP stories.

---

## Recommendations for Sprint 0 / Epic 1

### Test Framework Setup (Story 1.5 CI Pipeline)

1. **Configure coverage gate** in CI to block PRs below 80%
2. **Add test caching** to improve CI performance
3. **Setup parallel test execution** for faster feedback

### Test Infrastructure (New Stories Optional)

1. **Create test factories** in `tests/testutil/factories.go`
   - User, Note, API Key factories
   - Auto-cleanup with t.Cleanup()

2. **Setup dockertest** for integration tests
   - PostgreSQL container per test suite
   - Redis container for rate limit tests

3. **Add build tags** for test organization
   - `//go:build unit` (default)
   - `//go:build integration`

---

## Quality Gate Criteria

### Pass/Fail Thresholds

| Metric | Threshold | Enforcement |
|--------|-----------|-------------|
| P0 tests | 100% pass | CI blocks merge |
| P1 tests | ≥95% pass | CI warning |
| Coverage (domain/usecase) | ≥80% | CI blocks merge |
| Lint | 0 errors | CI blocks merge |
| Security scan | 0 critical vulns | CI blocks merge |

### Non-Negotiable Requirements

- [ ] All unit tests pass
- [ ] Coverage ≥80% for `internal/domain/`, `internal/usecase/`
- [ ] No golangci-lint errors
- [ ] No critical security vulnerabilities
- [ ] CI completes in <15min (p95)

---

## Summary

| Dimension | Assessment |
|-----------|------------|
| **Controllability** | ✅ PASS |
| **Observability** | ✅ PASS |
| **Reliability** | ✅ PASS |
| **Overall** | ✅ **READY FOR IMPLEMENTATION** |

**High Priority ASRs:** ASR-3 (CI performance ≤8min p50) requires attention during Epic 1.

**Next Steps:**
1. Run `/bmad-bmm-workflows-sprint-planning` to create sprint status
2. Start Epic 1: Foundation & Quality Gates
3. Implement test infrastructure in Story 1.5

---

**Generated by:** BMad TEA Agent - Test Architect Module
**Workflow:** `.bmad/bmm/testarch/test-design` (System-Level Mode)
**Version:** 4.0 (BMad v6)

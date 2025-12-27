# Brainstorming Session: golang-api-hexagonal vNext

> **Session Date:** 2025-12-27  
> **Topic:** International-grade Go backend boilerplate vNext  
> **Facilitator:** BMad Analyst  
> **Technique:** First Principles + SCAMPER + Risk Radar

---

## Session Overview

**Topic:** "International-grade Go backend boilerplate vNext" yang ambitious tapi maintainable, dengan fokus kuat di testing architecture cleanup + konsistensi struktur.

**Goals:**
1. Prioritized improvement backlog
2. Target repository structure
3. Risk & bug radar
4. Dependency recommendations
5. Roadmap (quick wins → sprint → long term)

---

## 1. Prioritized Improvement Backlog

### 🔴 HIGH Priority (Blocking International-Grade Status)

| # | Improvement | Why | Acceptance Criteria |
|---|-------------|-----|---------------------|
| H1 | **Split jumbo test files** | 22-23KB files = hard to navigate, review, maintain | Each test file ≤500 LOC, categorized by concern |
| H2 | **Consolidate mocks to single location** | Scattered mocks = duplication, inconsistency, drift | All mocks in `internal/testutil/mocks/` |
| H3 | **Delete/implement empty test_helpers_test.go** | Dead file = tech debt, confusing, false reference | File either deleted or has >1 meaningful test helper |
| H4 | **Add tests for shared/metrics** | Package without tests = coverage gap, regression risk | Interface compliance + behavior contract tests |
| H5 | **Build tags for integration tests** | No separation = CI confusion, slow feedback | `// +build integration` + `make test-unit` vs `make test-integration` |
| H6 | **Add testcontainers/dockertest** | Flaky/missing DB tests = unreliable integration | All DB tests use containerized PostgreSQL |
| H7 | **Go 1.25 upgrade assessment** | Stay current = security + perf + features | go.mod updated, CI green, benchmarks compared |

### 🟡 MEDIUM Priority (Enterprise-Grade Enhancement)

| # | Improvement | Why | Acceptance Criteria |
|---|-------------|-----|---------------------|
| M1 | **Granular error codes** | Generic errors = poor DX, hard debugging | Domain→App→Transport error code mapping documented |
| M2 | **Server timeout audit** | Missing timeouts = tail latency, resource leaks | ReadTimeout, WriteTimeout, IdleTimeout configured |
| M3 | **Context deadline propagation** | DB queries without deadline = potential leak | All repo methods respect ctx deadline |
| M4 | **Rate limiter IP keying audit** | Proxy misconfig = bypass risk | X-Forwarded-For trusted proxy tests |
| M5 | **goleak goroutine detection** | Goroutine leaks = slow memory creep | `go.uber.org/goleak` in test teardown |
| M6 | **Race detector in CI** | Data races = subtle bugs | `go test -race` as CI gate (at least nightly) |
| M7 | **Config validation strict mode** | Silent defaults = prod misconfig | Fail-fast on invalid config, prod auth required |

### 🟢 LOW Priority (Polish & Future-Proofing)

| # | Improvement | Why | Acceptance Criteria |
|---|-------------|-----|---------------------|
| L1 | **Cleanup internal_review_tmp** | Temporary folder = clutter | Folder deleted or documented purpose |
| L2 | **Coverage HTML generation** | Text coverage hard to navigate | `make coverage-html` target |
| L3 | **Idempotency key pattern** | POST retry safety | Middleware skeleton + documentation |
| L4 | **Feature flags interface** | Gradual rollout capability | Interface + mock provider + sample |
| L5 | **Background jobs skeleton** | Common need, hard to retrofit | Worker pattern + graceful shutdown |
| L6 | **OpenAPI codegen exploration** | Contract-first DX | Evaluate oapi-codegen vs current approach |

---

## 2. Target Repository Structure (Testing & Mocks)

### Current State (Problematic)
```
internal/
├── transport/http/handler/
│   ├── user_test.go          # 22KB - too large!
│   └── helpers_test.go       # scattered helpers
├── transport/http/middleware/
│   ├── auth_test.go          # 23KB - too large!
│   └── auth_test_helper_test.go  # scattered
├── infra/postgres/
│   └── test_helpers_test.go  # EMPTY (22 bytes)
└── (mocks scattered throughout test files)
```

### Target State (International Standard)
```
internal/
├── testutil/                          # 🆕 Centralized test infrastructure
│   ├── mocks/                         # 🆕 All mocks consolidated
│   │   ├── user_repo.go               # implements domain.UserRepository
│   │   ├── audit_repo.go              # implements domain.AuditEventRepository
│   │   ├── tx_manager.go              # implements domain.TxManager
│   │   ├── querier.go                 # implements domain.Querier
│   │   ├── id_generator.go            # implements domain.IDGenerator
│   │   └── redactor.go                # implements domain.Redactor
│   ├── fixtures/                      # 🆕 Test data factories
│   │   ├── user.go                    # UserFixture factory
│   │   └── audit.go                   # AuditEventFixture factory
│   ├── containers/                    # 🆕 Testcontainers setup
│   │   └── postgres.go                # PostgreSQL container helper
│   └── assertions/                    # 🆕 Custom test assertions
│       └── http.go                    # HTTP response assertions
│
├── transport/http/handler/
│   ├── user_test.go                   # Core happy path tests only (~200 LOC)
│   ├── user_validation_test.go        # 🆕 Validation edge cases
│   ├── user_error_test.go             # 🆕 Error handling scenarios
│   └── user_integration_test.go       # 🆕 Build tag: integration
│
├── transport/http/middleware/
│   ├── auth_test.go                   # Core auth tests (~300 LOC)
│   ├── auth_jwt_test.go               # 🆕 JWT-specific scenarios
│   ├── auth_claims_test.go            # 🆕 Claims extraction
│   └── auth_error_test.go             # 🆕 Error/edge cases
│
├── shared/metrics/
│   └── http_metrics_test.go           # 🆕 Interface compliance tests
│
└── infra/postgres/
    └── (test_helpers_test.go DELETED)
```

### Makefile Additions
```makefile
## test-unit: Run unit tests only (fast feedback)
test-unit:
	go test -v -short ./...

## test-integration: Run integration tests (requires containers)
test-integration:
	go test -v -tags=integration -race ./...

## test-race: Run tests with race detector
test-race:
	go test -race ./...

## coverage-html: Generate HTML coverage report
coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Open coverage.html in browser"
```

---

## 3. Risk & Bug Radar

### 🔴 Critical Risks (Production Impact)

| Risk | Scenario | Likelihood | Impact | Mitigation |
|------|----------|------------|--------|------------|
| **Context leak** | Handler doesn't propagate ctx deadline to DB query | Medium | High (resource exhaustion) | Add integration test for timeout behavior |
| **Auth drift** | Prod runs with JWT_ENABLED=false due to config mistake | Low | Critical (security breach) | Config validation: prod MUST enable auth |
| **Rate limit bypass** | Attacker uses different X-Forwarded-For values | Medium | Medium (DoS) | Trusted proxy whitelist + tests |

### 🟡 Medium Risks (Quality Degradation)

| Risk | Scenario | Likelihood | Impact | Mitigation |
|------|----------|------------|--------|------------|
| **Goroutine leak** | Background operation not cancelled on shutdown | Medium | Medium (memory creep) | Add goleak to tests |
| **Test drift** | Mocks diverge from real implementations | High | Medium (false confidence) | Centralized mocks + interface compliance tests |
| **Coverage regression** | New code skips testing | Medium | Medium | Coverage gate in CI (already have) |

### 🟢 Low Risks (Tech Debt)

| Risk | Scenario | Likelihood | Impact | Mitigation |
|------|----------|------------|--------|------------|
| **Dependency staleness** | Go version or deps fall behind | High | Low | Dependabot/Renovate + quarterly audit |
| **Documentation drift** | Code changes, docs don't | High | Low | Doc validation in PR checklist |

---

## 4. Dependency Recommendations

### 🔧 Core (Strongly Recommended)

| Package | Purpose | Justification |
|---------|---------|---------------|
| `github.com/testcontainers/testcontainers-go` | Container-based integration tests | Industry standard, reproducible, CI-friendly |
| `go.uber.org/goleak` | Goroutine leak detection | Catches subtle bugs before production |

### ⚡ Quick Wins (Minimal Effort)

| Package | Purpose | Justification |
|---------|---------|---------------|
| `github.com/stretchr/testify` (already have) | Assertions + mocks | Continue using, consolidate mock usage |

### 🚀 Ambitious (Evaluate for vNext)

| Package | Purpose | Trade-off |
|---------|---------|-----------|
| `github.com/cockroachdb/errors` | Rich error stack traces | Extra dependency vs better debugging |
| `github.com/oapi-codegen/oapi-codegen` | OpenAPI server/client codegen | Learning curve vs contract guarantee |
| `github.com/samber/lo` | Go generics utilities | Convenience vs stdlib purity |
| `github.com/riverqueue/river` | Background jobs with PostgreSQL | Full-featured vs custom simple worker |

### ❌ Avoid

| Package | Reason |
|---------|--------|
| Heavy ORM (GORM, etc.) | Conflicts with sqlc + hexagonal purity |
| Vendor-specific SDKs in domain | Layer violation |

---

## 5. Roadmap

### ⚡ Quick Wins (1-3 Days)

| Day | Task | Deliverable |
|-----|------|-------------|
| 1 | Delete empty test_helpers_test.go | ✅ Clean git history |
| 1 | Delete internal_review_tmp folder | ✅ Clean repository |
| 1 | Add `make test-race` target | ✅ Race detection ready |
| 1 | Add `make coverage-html` target | ✅ Visual coverage |
| 2 | Create `internal/testutil/mocks/` structure | ✅ Mock consolidation started |
| 2 | Add tests for `shared/metrics/` | ✅ Coverage gap closed |
| 3 | Split `handler/user_test.go` → 3 files | ✅ First jumbo file fixed |

### 📦 Sprint-Scale (1-2 Weeks)

| Week | Epic | Stories |
|------|------|---------|
| Week 1 | **Testing Architecture Overhaul** | Split all jumbo tests, consolidate all mocks, add testcontainers, implement build tags |
| Week 2 | **Reliability Hardening** | Timeout audit, context propagation, goleak integration, config validation strict mode |

### 🎯 Long Term (1-2 Months)

| Month | Epic | Outcome |
|-------|------|---------|
| Month 1 | **Enterprise Error Model** | Granular error codes, RFC 7807 full compliance, client-friendly diagnostics |
| Month 1 | **Go 1.25 Migration** | Updated toolchain, benchmarks, leverage new features |
| Month 2 | **Ambitious Features** | Idempotency keys, feature flags interface, background jobs skeleton |

---

## 6. Session Synthesis

### Key Insights 💡

1. **Testing adalah bottleneck #1** - Cleanup di sini akan unlock velocity untuk semua improvement lain
2. **Stack sudah kuat** - Ini bukan rewrite, ini polish + standardization
3. **Quick wins tersedia banyak** - 3 hari pertama bisa deliver value nyata
4. **Risk profile manageable** - Tidak ada showstopper, tapi ada beberapa "ticking bombs" yang perlu defused

### Recommended Next Actions

1. **Approve prioritization** → Move to PRD dengan backlog ini
2. **Start with Quick Wins** → Build momentum, reduce debt
3. **Defer Ambitious features** → Fokus core quality dulu

---

*Session facilitated by BMad Brainstorming Workflow*
*Techniques used: First Principles Thinking, SCAMPER, Risk Radar, Roadmap Mapping*

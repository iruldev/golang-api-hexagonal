---
stepsCompleted: [1, 2, 3, 4, 7, 8, 9, 10, 11]
inputDocuments:
  - docs/analysis/brainstorming-session-2025-12-15.md
  - docs/analysis/research/technical-go-golden-template-2025-12-15.md
  - docs/index.md
documentCounts:
  briefs: 0
  research: 1
  brainstorming: 1
  projectDocs: 1
workflowType: 'prd'
lastStep: 11
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-15'
---

# Product Requirements Document - Backend Service Golang Boilerplate

**Author:** Gan
**Date:** 2025-12-15

---

## Executive Summary

**Backend Service Golang Boilerplate** adalah inisiatif upgrade besar untuk mentransformasi repository Go backend yang sudah ada menjadi **enterprise-grade "golden template"** yang siap diadopsi seluruh tim.

Repository ini sudah memiliki arsitektur **Hexagonal (Ports & Adapters)** yang solid dengan tech stack modern (Go 1.24, chi v5, PostgreSQL, Redis, Kafka, OpenTelemetry). Upgrade ini bukan tentang menambah fitur bisnis baru, melainkan **menaikkan standar infrastruktur dan tooling** ke level production-grade yang enforceable.

**Core Problem:**
- Standar saat ini *documented* tapi belum *enforced* → inkonsistensi
- Developer experience belum optimal → onboarding lambat
- Safety defaults belum uniform → technical debt tersembunyi

**Solution Approach:**
Incremental, phase-based adoption dengan prioritas:
1. **Quality Gates** (CI, lint, test, coverage, boundary enforcement)
2. **DX Enhancement** (1-command workflow, container-first)
3. **Safety Defaults** (timeouts, typed errors, idempotency)
4. **Observability Polish** (dashboards, alerts, runbooks)
5. **Golden Path Documentation** (living reference implementation)

### What Makes This Special

Upgrade ini berbeda karena fokus pada **developer adoption through excellent DX**:

- **Enforced Standards** - Bukan hanya policy document, tapi CI gates yang memaksa konsistensi
- **"1 Command" Philosophy** - `make up` untuk start, `make verify` untuk validate, `make reset` untuk clean slate
- **Golden Path as Code** - Reference implementation yang selalu build & test, bukan dokumentasi statis
- **Safety by Default** - Timeout, context propagation, dan idempotency pattern built-in

**The Moment of Success:** Developer baru membuka repo, run `make up`, dan dalam 15 menit sudah submit PR pertama mengikuti golden path yang established.

### Project Classification

| Attribute | Value |
|-----------|-------|
| **Technical Type** | api_backend (monolith multi-entrypoint) |
| **Domain** | Developer Tooling / Internal Platform |
| **Complexity** | Medium |
| **Project Context** | Brownfield - extending existing system |
| **Upgrade Scope** | Platform/Architecture/Tooling (non-business features) |
| **Adoption Strategy** | Incremental, phase-based |

---

## Success Criteria

### User Success (Developer Experience)

| Metric | Target | Stretch |
|--------|--------|---------|
| **Time-to-First-PR (TTFP)** | ≤ 4 jam | ≤ 2 jam |
| **Onboarding to `make up` success** | Median ≤ 30 min | p95 ≤ 60 min |
| **One-command parity** | ≥ 95% success Mac/Linux | tanpa debugging manual |
| **Golden-path compliance** | ≥ 80% PR ikut template | tanpa structure review |

**Success Moment:** Developer baru run `make up`, service jalan, submit PR pertama dalam hitungan jam - bukan hari.

### Business Success (Platform Outcomes)

**3-Month Targets:**
- 100% PR lewat CI gates (lint/test/coverage/boundary) sebelum merge
- Style debate di PR review berkurang drastis
- Coverage stabil ≥ 80% untuk usecase/domain layer
- Minimal 1 golden path module sebagai living reference
- Onboarding median ≤ 30 menit

**12-Month Targets:**
- >90% repos/modul internal mengadopsi blueprint ini
- MTTR turun ≥ 30-50% (trace/log/metrics konsisten)
- Tech debt dari inconsistent patterns turun signifikan
- Zero "style/structure" comments di PR reviews

### Technical Success

| Metric | Target |
|--------|--------|
| **CI pass rate on main** | >95% tanpa "fix CI" loop |
| **Test coverage** | ≥80% untuk domain/usecase |
| **Boundary violations** | 0 (blocked by depguard) |
| **Build time** | <5 min untuk full CI |

### Measurable Outcomes

**Key Metric (Headline):** Time-to-First-PR (TTFP) Median
- Measures adoption through DX
- Measures quality of foundation
- Leading indicator of template success

**Priority Metrics:**
1. TTFP median & p95
2. Onboarding time to `make up` success
3. CI pass rate on main
4. PR review cycle time (style/structure debates)
5. MTTR (post-observability polish)

---

## Product Scope

### MVP - Phase 1 & 2 (Foundation)

**Must-Have untuk Template Usable:**
- [ ] golangci-lint v2 strict mode dengan depguard boundaries
- [ ] CI pipeline dengan coverage enforcement (≥80%)
- [ ] Makefile consolidation (`up`, `down`, `reset`, `verify`)
- [ ] .env.example dengan validation on startup
- [ ] Basic golden path documentation

**Done When:** CI gates blocking, new dev productive ≤4 jam

### Growth Features - Phase 3 & 4 (Enhancement)

**Competitive Advantages:**
- [ ] Safety defaults (timeout wrappers, typed errors)
- [ ] Idempotency middleware template
- [ ] Observability polish (dashboards, alerts, runbooks)
- [ ] Context propagation mandatory (strict mode)
- [ ] bplat generator enhancements

**Done When:** MTTR turun, incidents "boring"

### Vision - Phase 5 (Polish)

**Dream Version:**
- [ ] examples/goldenpath/ sebagai living reference yang selalu tested
- [ ] ADR template + existing decisions documented
- [ ] Policy pack (dep allowlist, error codes, log fields)
- [ ] Golden PR bot untuk boundary violation comments
- [ ] Chaos mode lokal untuk resilience testing

**Done When:** Template siap untuk adoption seluruh tim

---

## User Journeys

### User Types

| User Type | Focus | Why Included |
|-----------|-------|-------------|
| **New Developer** | Onboarding, first PR | Measures TTFP, DX quality |
| **Senior Developer** | Adding features | Validates golden path efficiency |
| **Tech Lead / Reviewer** | PR reviews, standards | Validates enforcement works |
| **Platform Engineer** | CI/CD, infra | Maintains toolchain |
| **QA Engineer** | Test reliability | Quality gates, testing pyramid |
| **Security Engineer** | Audit, compliance | Secure-by-default validation |
| **SRE / On-call** | Incident response | "Incidents boring" validation |

---

### Journey 1: Andi — New Developer Onboarding

**Persona:** Andi, backend developer, baru join minggu ini. Background Spring Boot, belum familiar Go hexagonal.

**Before Golden Template:**
Andi clone repo, bingung mau mulai dari mana. Ada banyak folder dan pattern yang asing. Setup local env memakan waktu 1-2 hari dengan banyak debugging. PR pertama kena banyak review comments soal "structure" dan "pattern".

**After Golden Template:**
Andi run `make up` dan dalam 30 menit semua services jalan lokal. Docker-compose handles semua dependencies. Dia buka `examples/goldenpath/` dan langsung paham pattern yang diharapkan - handler, usecase, repository, tests. Dalam 4 jam, Andi submit PR pertama untuk fitur kecil. CI langsung pass - depguard catch kalau structure salah, linter enforce style. Review cuma bahas logic, bukan "kenapa folder namanya gitu".

**Success Indicator:** TTFP ≤ 4 jam, zero "structure" comments di PR review.

---

### Journey 2: Dina — Senior Developer Adding Feature

**Persona:** Dina, senior backend 2 tahun di tim, sangat familiar dengan codebase.

**Before Golden Template:**
Dina dapat task bikin module baru. Setup folder structure manual, copy-paste dari module lain, sering ada inkonsistensi kecil. PR review masih ada style debates.

**After Golden Template:**
Dina run `bplat generate module payment` - scaffold handler/usecase/repo langsung ready dengan test stubs. Dia fokus 100% ke business logic. PR langsung pass semua gates. Review cuma bahas edge cases dan business logic. Modul baru otomatis punya tracing, logging, dan metrics yang konsisten.

**Success Indicator:** Golden-path compliance ≥80%, PR review fokus ke logic.

---

### Journey 3: Budi — Tech Lead Reviewing PRs

**Persona:** Budi, tech lead, review 5-10 PRs per hari.

**Before Golden Template:**
Banyak waktu habis untuk "please move ini ke layer yang benar" atau "pakai error pattern yang standard". Inconsistency antar developer bikin review exhausting.

**After Golden Template:**
Depguard dan linter sudah catch structure issues sebelum Budi lihat. CI red = PR belum ready. Budi fokus review business logic, edge cases, dan potential bugs. Style/structure debate hilang karena tooling yang enforce, bukan manusia.

**Success Indicator:** Zero "style/structure" comments, PR review cycle time turun.

---

### Journey 4: Rina — Platform Engineer Maintaining CI

**Persona:** Rina, platform/devops eng, maintain CI pipeline dan infra.

**Before Golden Template:**
Sering debugging "works on my machine" issues. Dev env dan CI env beda behavior. Config scattered di berbagai tempat.

**After Golden Template:**
Container-first dev berarti local env = CI env = staging config. Makefile targets (`up`, `down`, `reset`, `verify`) konsisten. `.env.example` always up-to-date dengan validation on startup. Rina setup observability dashboards yang semua modul otomatis ikuti karena pattern konsisten.

**Success Indicator:** One-command parity ≥95%, CI pass rate >95%.

---

### Journey 5: Santi — QA Ensuring Test Stability

**Persona:** Santi, QA engineer, fokus ke test quality dan reliability.

**Before Golden Template:**
Flaky tests sering bikin CI red random. Integration test setup complicated. Debugging test failures susah karena no clear isolation.

**After Golden Template:**
Santi run `make verify` - semua unit + integration tests jalan. `make reset` untuk clean slate sebelum integration test. Test deterministik karena container isolation. Flaky test policy jelas: quarantine + create ticket. Testing pyramid enforced oleh coverage requirements (≥80% usecase/domain).

**Success Indicator:** CI flake rate <1%, test coverage ≥80%.

---

### Journey 6: Arif — Security Review Cepat

**Persona:** Arif, security/appsec engineer, review repos untuk compliance.

**Before Golden Template:**
Manual scan untuk hardcoded secrets. Dependency vulnerabilities sering missed. Auth patterns inconsistent antar module. Logging kadang bocorin sensitive data.

**After Golden Template:**
Arif cek repo: gitleaks di pre-commit, govulncheck + gosec di CI. Error responses mengikuti standard (no stack traces to client). Auth middleware template dengan RBAC patterns. Structured logging dengan predefined safe fields. Dependency allowlist via depguard.

**Success Indicator:** Zero hardcoded secrets, govulncheck green, auth patterns consistent.

---

### Journey 7: Nita — On-call Incident Jadi Boring

**Persona:** Nita, SRE, on-call rotation untuk production incidents.

**Before Golden Template:**
Incidents sering chaotic. Logs unstructured dan susah correlate. Tracing incomplete. Runbooks outdated atau tidak ada.

**After Golden Template:**
Nita dapat alert → klik dashboard → lihat metrics anomaly → get trace-id → log correlation automatic → runbook link di alert → langkah-langkah jelas. Root cause identified cepat karena observability konsisten di semua module. Incidents jadi "boring" dalam arti positif.

**Success Indicator:** MTTR turun ≥30-50%, incidents predictable.

---

### Journey Requirements Summary

| Journey | Reveals Requirements For |
|---------|------------------------|
| Andi (New Dev) | Onboarding docs, make up, examples/goldenpath |
| Dina (Senior Dev) | bplat generator, golden path patterns |
| Budi (Tech Lead) | CI gates, depguard, linter config |
| Rina (Platform) | Container-first, Makefile, env validation |
| Santi (QA) | Test pyramid, make verify/reset, flaky policy |
| Arif (Security) | Secret scanning, dep scanning, auth template |
| Nita (On-call) | Dashboards, tracing, runbooks, alerts |

---

## API Backend Specific Requirements

### Project-Type Overview

Golden template ini mengstandarkan **api_backend** patterns untuk Go monolith dengan multi-entrypoint (server, worker, scheduler, CLI). Fokus pada consistency, enforceability, dan developer adoption.

### Endpoint Standards

**Versioning Policy:**

| Protocol | Versioning Strategy | Example |
|----------|-------------------|---------|
| **REST** | Path-based `/api/v1` | `/api/v1/notes`, `/api/v2/notes` |
| **gRPC** | Package name | `service.v1.NoteService` |
| **GraphQL** | Schema evolution | Deprecate fields, no path versioning |

**Rules:**
- Default semua REST endpoints di `/api/v1`
- `/api/v2` hanya untuk breaking changes besar
- Semua handlers follow pattern: `handler → usecase → repository`

### Authentication Model

**Standard Stack:**

| Component | Implementation | Scope |
|-----------|---------------|-------|
| **JWT** | Primary auth method | User authentication |
| **RBAC Middleware** | `RequireRole()` | Route protection |
| **API Keys** | Rotatable, multiple active | Service-to-service, CLI |

**API Key Requirements:**
- Multiple active keys per service
- Expiry date mandatory
- Revocation support
- Rotation without downtime

**Future Extension (not MVP):**
- OAuth2 flows (enterprise extension)
- mTLS for service mesh

### Response Envelope Contract

```go
type Envelope struct {
    Data  interface{}    `json:"data,omitempty"`
    Error *ErrorResponse `json:"error,omitempty"`
    Meta  *Meta          `json:"meta,omitempty"`
}

type ErrorResponse struct {
    Code    string            `json:"code"`              // Public, stable (e.g., "NOTE_NOT_FOUND")
    Message string            `json:"message"`           // Safe untuk user
    Hint    string            `json:"hint,omitempty"`    // Self-documenting
    Details []ValidationError `json:"details,omitempty"` // Field validation errors
}

type Meta struct {
    TraceID    string      `json:"trace_id"`              // Wajib
    RequestID  string      `json:"request_id,omitempty"`  // Recommended
    Pagination *Pagination `json:"pagination,omitempty"`  // Optional
    Warnings   []string    `json:"warnings,omitempty"`    // Non-blocking warnings
}
```

**Mandatory Fields:**
- `meta.trace_id` - Always present for debugging
- `error.code` - Public, stable error identifier

### Rate Limiting

**Policy:** Default ON (template available), opt-in per route untuk enforcement

| Route Type | Rate Limit |
|------------|------------|
| Public endpoints | Mandatory enable |
| Internal/admin | Recommended |
| Health checks | Exempt |

**Implementation:**
- Middleware template di golden path
- Config-driven thresholds
- Redis-backed untuk distributed rate limiting
- Future: upgrade ke "default for all" setelah baseline jelas

### API Documentation

| Protocol | Documentation Source |
|----------|---------------------|
| **REST** | OpenAPI/Swagger generated (source of truth) |
| **gRPC** | Proto files + grpcurl/buf docs |
| **GraphQL** | gqlgen schema introspection |
| **Cookbook** | Markdown golden path examples |

**Generation Requirements:**
- OpenAPI spec auto-generated dari code annotations
- Swagger UI available di `/docs` (dev only)
- Proto → documentation pipeline

### Implementation Considerations

**Template Files to Create:**
- `internal/interface/http/response/envelope.go`
- `internal/interface/http/middleware/ratelimit.go`
- `internal/interface/http/middleware/apikey.go`
- `api/openapi.yaml` (generated)

**CI Integration:**
- OpenAPI spec diff check on PR
- Proto breaking change detection
- Response envelope compliance linting

---

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Platform MVP - Build strong foundation for all future development
**Resource Requirements:** 1-2 senior backend engineers + platform/devops support
**Total Timeline:** 8-14 weeks (5 phases)

### Phase 1-2: Foundation (Week 1-4)

**Core Deliverables:**
- [ ] golangci-lint v2 strict mode + depguard boundaries
- [ ] CI pipeline dengan coverage enforcement (≥80%)
- [ ] Makefile consolidation (`up`, `down`, `reset`, `verify`)
- [ ] .env.example dengan validation on startup
- [ ] Basic golden path documentation

**Added to MVP (high-leverage):**
- [ ] Standard HTTP Envelope minimal (`meta.trace_id` + `error.code`)
- [ ] Pre-commit hooks: `make hooks` untuk auto-install (recommended)
- [ ] Context propagation mandatory (lint rule / wrapper baseline)

**Done When:**
- CI gates blocking all PRs
- New dev productive ≤4 jam
- All endpoints return trace_id on error

### Phase 3-4: Enhancement (Week 5-10)

**Safety & Observability:**
- [ ] Timeout wrappers untuk all IO
- [ ] Typed errors dengan public codes + hints
- [ ] Idempotency middleware template
- [ ] Observability polish (dashboards, alerts)
- [ ] Runbook templates
- [ ] bplat generator enhancements

**Done When:**
- MTTR turun ≥30%
- Incidents "boring" dengan clear trace

### Phase 5: Polish (Week 11-14)

**Living Reference & Automation:**
- [ ] examples/goldenpath/ sebagai living reference (always tested)
- [ ] ADR template + existing decisions documented
- [ ] Policy pack (dep allowlist, error codes, log fields)
- [ ] Golden PR bot untuk boundary violation comments
- [ ] Feature flag framework (optional)

**Done When:**
- Template siap untuk adoption seluruh tim
- >90% internal repos can adopt

---

### Risk Mitigation Strategy

| Risk Type | Risk | Mitigation |
|-----------|------|------------|
| **Technical** | Breaking existing CI | Incremental rollout, opt-in first |
| **Adoption** | Developer resistance | DX-first approach, 1-command workflow |
| **Resource** | Limited bandwidth | Phase-based approach, prioritize high-leverage items |
| **Scope** | Feature creep | Clear phase boundaries, defer non-essential |

### Timeline Summary

| Phase | Weeks | Focus |
|-------|-------|-------|
| Phase 1 | 1-2 | CI Gates + Linting |
| Phase 2 | 3-4 | DX + Container + Envelope |
| Phase 3 | 5-8 | Safety Defaults |
| Phase 4 | 9-10 | Observability |
| Phase 5 | 11-14 | Polish + Adoption |

---

## Functional Requirements

### Quality Gates & CI/CD

- **FR1:** Platform Engineer dapat configure golangci-lint v2 dengan strict mode dan custom rules
- **FR2:** CI Pipeline dapat block PR yang tidak pass linting, testing, atau coverage threshold
- **FR3:** Platform Engineer dapat configure depguard rules untuk enforce layer boundaries
- **FR4:** Developer dapat melihat lint violations dengan link ke rule + contoh fix
- **FR5:** CI Pipeline dapat generate dan display coverage report per package/layer (domain/usecase/interface) dengan baseline file untuk mencegah drop

### Developer Experience

- **FR6:** Developer baru dapat run `make up` untuk start semua services secara container-first
- **FR7:** Developer dapat run `make verify` untuk execute full test suite locally
- **FR8:** Developer dapat run `make reset` untuk clean slate environment
- **FR9:** Developer dapat install pre-commit hooks via `make hooks`
- **FR10:** Developer dapat generate new module scaffold via `bplat generate`
- **FR11:** System dapat validate environment configuration on startup (.env)

### API Standards & Response Contract

- **FR12:** All HTTP endpoints harus return response dalam Envelope format (data/error/meta)
- **FR13:** All error responses harus include `error.code` yang public dan stable
- **FR14:** All responses harus include `meta.trace_id` untuk debugging
- **FR15:** System dapat return validation errors dengan field-level detail
- **FR16:** REST endpoints harus mengikuti path-based versioning (/api/v1)

### Authentication & Authorization

- **FR17:** System dapat authenticate users via JWT token
- **FR18:** System dapat authorize access via RBAC middleware (`RequireRole()`)
- **FR19:** System dapat authenticate service-to-service via API keys
- **FR20:** Admin dapat create, rotate, dan revoke API keys
- **FR21:** API keys harus support multiple active keys dan expiry

### Rate Limiting

- **FR22:** Platform Engineer dapat configure rate limiting per route
- **FR23:** Rate limit middleware harus tersedia sebagai template
- **FR24:** System dapat enforce rate limits dengan Redis-backed storage
- **FR25:** Health check endpoints harus exempt dari rate limiting

### Context Propagation & Tracing

- **FR26:** All IO operations harus menerima context sebagai parameter pertama
- **FR27:** System dapat propagate trace context across service boundaries
- **FR28:** Logs harus correlate dengan trace_id dan request_id
- **FR29:** Lint rules harus enforce context propagation patterns

### Documentation

- **FR30:** System dapat generate OpenAPI spec dari code annotations
- **FR31:** Developer dapat access Swagger UI di /docs (dev environment)
- **FR32:** gRPC services harus have proto files sebagai documentation source
- **FR33:** Golden path examples harus selalu up-to-date dan tested

### Observability

- **FR34:** System dapat emit structured logs dalam JSON format (Zap)
- **FR35:** System dapat emit metrics ke Prometheus endpoint
- **FR36:** System dapat emit traces ke OpenTelemetry collector
- **FR37:** Platform Engineer dapat configure dashboards templates
- **FR38:** Platform Engineer dapat configure alert rules templates
- **FR39:** Runbook templates harus tersedia untuk common incidents

### Security & Config Hygiene

- **FR40:** System tidak boleh start jika config required missing/invalid (fail-fast) dengan error output yang aman (tanpa leak secret)
- **FR41:** CI pipeline dapat menjalankan secret scanning dan block PR yang contain hardcoded credentials

### Golden Path & Policy Enforcement

- **FR42:** System menyediakan policy pack configuration (lint + depguard + error codes registry) sebagai single source of truth untuk CI dan tooling lokal

### Migration Readiness (Brownfield)

- **FR43:** System menyediakan compatibility mode untuk response/route pada golden path module dengan contract tests untuk prevent breaking changes

---

## Non-Functional Requirements

### Performance

| NFR | Requirement | Target |
|-----|-------------|--------|
| **NFR-P1** | Full CI pipeline (lint+unit+integration) | p50 ≤ 8 min, p95 ≤ 15 min |
| **NFR-P1b** | PR quick checks (lint+unit only) | ≤ 5 min |
| **NFR-P2** | `make up` to all services running | ≤ 2 min |
| **NFR-P3** | `make verify` (lint+unit) | ≤ 3 min |
| **NFR-P3b** | `make verify-full` (with integration) | ≤ 10-15 min |
| **NFR-P4** | Lint execution time (full project) | ≤ 60 sec |

### Security

| NFR | Requirement | Target |
|-----|-------------|--------|
| **NFR-S1** | Zero hardcoded secrets in codebase | 100% (gitleaks enforced) |
| **NFR-S2** | Dependency vulnerability scan | 0 Critical; High requires waiver + expiry |
| **NFR-S3** | Secret storage policy | Secrets via env/secret manager only; non-secret config allowed |
| **NFR-S4** | Error responses production mode | No stack traces exposed |
| **NFR-S5** | Audit log retention | Immutable, 90 days |

### Reliability

| NFR | Requirement | Target |
|-----|-------------|--------|
| **NFR-R1** | CI flake rate (test fail without code change, per week) | < 1% |
| **NFR-R2** | CI pass rate on main | > 95% |
| **NFR-R3** | Container startup reliability | > 99% |
| **NFR-R4** | Test reproducibility | ≥ 99% (quarantine policy for outliers) |
| **NFR-R5** | Health check response after startup | Within 5 sec |

### Maintainability

| NFR | Requirement | Target |
|-----|-------------|--------|
| **NFR-M1** | Test coverage (domain/usecase) | ≥ 80% |
| **NFR-M2** | Layer boundary violations | 0 (depguard enforced) |
| **NFR-M3** | Cyclomatic complexity | ≤ 15 (gocyclo, threshold enforced gradually) |
| **NFR-M4** | Code duplication | < 3% (dupl, threshold enforced gradually) |
| **NFR-M5** | API documentation coverage | 100% public APIs (OpenAPI + proto) + golden path cookbook |
| **NFR-M6** | New dev to first PR | ≤ 4 jam |

### Developer Experience

| NFR | Requirement | Target |
|-----|-------------|--------|
| **NFR-DX1** | One-command setup success | ≥ 95% Mac/Linux |
| **NFR-DX2** | Golden path compliance | ≥ 80% new modules |
| **NFR-DX3** | PR style issue turnaround | 0 (automated by linter) |
| **NFR-DX4** | Lint output clarity | Includes rule link + fix suggestion |
| **NFR-DX5** | First-time setup failure rate | < 5% with troubleshooting guide |

---
stepsCompleted: [1, 2, 3, 4, 6, 7, 8, 9, 10, 11]
inputDocuments:
  - 'docs/analysis/product-brief-backend-service-golang-boilerplate-2025-12-10.md'
  - 'docs/analysis/research/technical-golang-enterprise-boilerplate-research-2025-12-10.md'
  - 'docs/analysis/brainstorming-session-2025-12-10.md'
documentCounts:
  briefs: 1
  research: 1
  brainstorming: 1
  projectDocs: 0
workflowType: 'prd'
lastStep: 11
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-10'
---

# Product Requirements Document - Backend Service Golang Boilerplate

**Author:** Gan
**Date:** 2025-12-10

---

## Executive Summary

**Backend Service Golang Boilerplate** adalah enterprise-grade "golden template" yang menyediakan fondasi production-ready untuk membangun backend services di Go. Template ini mengintegrasikan **observability-first architecture**, **opinionated tech stack**, dan **AI-native developer experience** dalam satu paket siap pakai.

### Vision Statement

Menyediakan "golden template" yang memungkinkan tim engineering:
- Memulai service baru dalam **< 30 menit** (bukan hari)
- Menghasilkan service yang **sehat sejak lahir** (observability, testing, struktur konsisten)
- Berkolaborasi dengan **AI assistant secara aman** melalui AGENTS.md
- Fokus pada **domain logic**, bukan repetitive infrastructure setup

### What Makes This Special

1. **Three Pillars Architecture** (Berjalan – Diamati – Dipercaya)
   - Foundation filosofi arsitektur yang memastikan setiap service memenuhi standar production-readiness

2. **AI-Native by Design**
   - AGENTS.md sebagai first-class citizen
   - Struktur project memungkinkan AI assistant berkolaborasi dengan safety guardrails

3. **Regulated-Ready**
   - Designed dengan standar fintech/regulated environment in mind
   - Audit trail hooks, observability ketat, time safety (UTC-first)

4. **Opinionated but Not Locking**
   - Stack tegas (chi, sqlc+pgx, zap, koanf, asynq, OTEL)
   - Hook patterns untuk adaptasi ke environment berbeda

5. **Strong DX + Quality Bundle**
   - Bukan skeleton, tapi paket lengkap: observability, testing, documentation
   - One-command dev experience (`make dev`)

---

## Project Classification

| Attribute | Value |
|-----------|-------|
| **Technical Type** | Developer Tool + API Backend |
| **Domain** | General Backend (Regulated/Fintech-Ready) |
| **Complexity** | Medium-High (Rich Platform Capability) |
| **Project Context** | Greenfield - New Project |

### Classification Rationale

**Complexity Medium-High karena:**
- Scope mencakup observability suite, job system hooks, testing setup, AI integration
- Bukan rumit dari sisi fitur bisnis, tapi **kaya dari sisi platform capability**
- Target output adalah enterprise-grade foundation, bukan simple starter

**Regulated-Ready karena:**
- Audit trail hooks built-in
- Time/timezone safety (UTC-first)
- Strict observability (healthcheck jujur, metrics standar)
- Designed untuk bisa dipakai di fintech/regulated environment

**AI-Native karena:**
- AGENTS.md bukan afterthought
- Struktur project memungkinkan AI + developer kolab dengan aman
- Documentation sebagai contract untuk AI behavior

---

## Success Criteria

### User Success

#### Functional Metrics

| Persona | Metric | Target |
|---------|--------|--------|
| **Andi (Backend Engineer)** | Time-to-first-service | < 30 menit |
| | Time-to-first-feature | ≤ 2 hari kerja |
| **Rudi (Tech Lead)** | Service adoption rate | ≥ 80% dalam 6-12 bulan |
| **Maya (SRE)** | Observability coverage | 100% service baru |
| **Dina (New Joiner)** | Onboarding to first PR | ≤ 1 minggu |

#### Emotional Success & Aha Moments

| Persona | Aha Moment | Delight Criteria |
|---------|------------|------------------|
| **Andi** | Clone → `make dev` → 30min: endpoint hit, logs, metrics visible | "Langsung kerja di domain logic" |
| **Rudi** | First PR review: no comments on struktur/error handling | "PR 80% tentang bisnis" |
| **Maya** | Add service to same dashboard template | "Semua service terlihat sama" |
| **Dina** | Follow example + AGENTS.md → create endpoint independently | "Bisa kontribusi tanpa takut" |

---

### Business Success

| Metric | Target | Timeline |
|--------|--------|----------|
| Pilot Adoption | 2-3 services end-to-end | 3 months |
| Engineer Feedback | 3+ engineers dari berbeda squad | 3 months |
| Adoption Rate | ≥ 70-80% service baru | 12 months |
| Squad Usage | ≥ 3 squads | 12 months |
| Setup Time Reduction | ≥ 50-70% lebih rendah | 12 months |
| Incident Reduction | ≥ 30% (missing basics) | 12 months |

---

### Technical Success

| Category | Metric | Success Criteria |
|----------|--------|------------------|
| **Architectural Consistency** | Structure compliance | ≥ 90% follow internal/interface/usecase/domain/infra |
| **Testability** | Example module coverage | ≥ 70% unit + ≥ 1 integration |
| **Static Quality** | PR first-run success | ≥ 80% pass lint+test |
| **Observability Quality** | Log fields | level, timestamp, trace_id, request_id, path, method, status |

---

### Measurable Outcomes Summary

| Outcome | Before | After | Improvement |
|---------|--------|-------|-------------|
| Time-to-first-service | Hours-Days | < 30 min | 10-100x faster |
| Setup overhead | 50-70% time | Near zero | ≥ 50% saved |
| Pattern consistency | Variable | Standardized | 100% consistent |
| Observability baseline | Often missing | Always present | 100% coverage |

---

## Product Scope

### MVP - Minimum Viable Product (v1)

| # | Category | Components |
|---|----------|------------|
| 1 | Foundation | main.go, wiring, chi router, middleware |
| 2 | Config | koanf package, struct validation, .env.example |
| 3 | Database | pgx + sqlc + golang-migrate |
| 4 | Observability | zap, healthz/readyz, /metrics, OTEL hooks |
| 5 | DX Tooling | Makefile, docker-compose |
| 6 | Quality | golangci-lint, CI example |
| 7 | Sample Module | Entity → Repo → Usecase → Handler → Tests |
| 8 | Documentation | README, ARCHITECTURE, AGENTS.md |

---

### Growth Features (v1.1)

| Feature | Rationale |
|---------|-----------|
| Redis + asynq jobs | Structure ready, implementation v1.1 |
| Full testcontainers | docker-compose cukup untuk MVP |
| Prometheus + Grafana | Jaeger only di MVP |

---

### Out of Scope (Explicit)

| Item | Status |
|------|--------|
| gRPC & GraphQL | v2+ or never |
| Multiple database | Postgres only |
| Full Auth/RBAC | Interface stub only |
| Rate limiting impl | Interface + docs only |
| Feature flags | Future pluggable |
| Admin UI | Backend only |

---

## User Journeys

### Journey 1: Andi Pratama - From Repetitive Setup to Domain Focus

**Who:** Andi, 28 tahun, Backend Engineer Mid-level, 3 tahun di fintech.

**The Pain:** Setiap service baru, habiskan 2-3 hari untuk setup infrastructure. Pattern antar service tidak konsisten. Code review penuh komentar soal struktur, bukan logic bisnis.

**The Journey:**

Senin pagi, Andi dapat tiket: "Bikin service notification, deadline Jumat."

```bash
git clone git@github.com:company/golang-backend-boilerplate.git notification-service
cd notification-service && cp .env.example .env && make dev
```

**25 menit kemudian**: Service running. Jaeger tracing aktif. Prometheus metrics jalan. Healthcheck respond benar.

```bash
cp -r internal/module/example internal/module/notification
```

**Rabu sore**: PR pertama. Rudi review: "LGTM, logic bisnis jelas." Zero komentar tentang struktur.

**Aha:** "Langsung fokus ke apa yang mau dibangun."

---

### Journey 2: Rudi - Standardizing Without Micromanaging

**Who:** Rudi, Tech Lead/Architect, responsible untuk 5 squad.

**The Pain:** Setiap squad punya style berbeda. Code review berat. Onboarding lambat.

**The Journey:**

Rudi review ARCHITECTURE.md: "Three Pillars ini exactly what I've been preaching!"
Review AGENTS.md: "AI assistant bisa kerja sesuai standard kita?"
Present ke Architecture Guild: "Ini mandatory starting point."

**1 bulan kemudian**: 3 service baru pakai boilerplate. PR review time turun 40%.

**Aha:** "PR 80% tentang bisnis, bukan plumbing."

---

### Journey 3: Maya - One Template Dashboard for All

**Who:** Maya, SRE/Platform Engineer, responsible untuk 20+ microservices.

**The Pain:** Setiap service beda. Dashboard custom per service. Incident investigation lambat.

**The Journey:**

notification-service naik ke staging. Maya notice:
- `/readyz` return 503 ketika Redis down (honest!)
- `/metrics` expose standard HTTP metrics
- Log JSON dengan trace_id, request_id, path, method, status
- Tracing connected ke Jaeger

Maya import "Go Service Template Dashboard", point ke metrics. **5 menit**: dashboard live.

**2 bulan kemudian**: 5 service, 1 dashboard template, incident investigation 30% lebih cepat.

**Aha:** "Semua service terlihat sehat/sakit dengan cara yang sama."

---

### Journey 4: Dina - From Confused to Contributing

**Who:** Dina, Junior Backend Engineer, 3 bulan join, first time Go production.

**The Pain:** Setiap project beda. Takut salah. Imposter syndrome tinggi.

**The Journey:**

Task: Tambah endpoint di notification-service.
Andi: "Baca AGENTS.md, lihat contoh di example module."

Dina ikuti AGENTS.md step-by-step, referensi example module.

**3 jam kemudian**: PR dengan endpoint baru, unit test 75% coverage, passing CI.
Andi review: "Clean! LGTM."

**Aha:** "Bisa contribute tanpa takut merusak pattern."

---

### Journey Requirements Summary

| Journey | Key Capabilities |
|---------|-----------------|
| **Andi** | One-command setup, sample module, consistent structure |
| **Rudi** | ARCHITECTURE.md, AGENTS.md, enforceable patterns |
| **Maya** | Honest healthcheck, standard metrics, consistent logging |
| **Dina** | Clear guidelines, example module, test templates |

### Cross-Cutting Requirements

1. **Documentation First-Class** - README, ARCHITECTURE, AGENTS.md
2. **Example Module** - Living documentation yang bisa di-copy
3. **Observability Built-In** - Metrics, logging, tracing, healthcheck
4. **DX Tooling** - Makefile, docker-compose, .env.example
5. **Quality Gates** - Linting, testing, CI pipeline

---

## Innovation & Novel Patterns

### 1. AI-Native Developer Experience (Key Innovation)

**The Paradigm Shift:**

| Traditional Boilerplate | This Boilerplate |
|------------------------|------------------|
| Human-first, AI afterthought | Human + AI collaborative peers |
| AI produces inconsistent results | AI guided by AGENTS.md contract |
| Pattern violations common | Pattern compliance by design |

**AGENTS.md as AI Contract:**
- Explicit agreement between project and AI coding assistant
- Contains: structure, stack, test strategy, standards, anti-patterns
- Instructions for safe refactoring and feature addition

**Validation Plan:**

| Test | Success Criteria |
|------|------------------|
| Engineer + AI with AGENTS.md | Code follows layer structure |
| Compare: with vs without | Fewer pattern violation comments |
| AI-assisted PRs | "Pattern salah" comments ↓ |

---

### 2. Three Pillars Architecture

| Pillar | Indonesian | Meaning |
|--------|------------|---------|
| **Berjalan** | Running | Service functions correctly |
| **Diamati** | Observable | Visible via logs, metrics, tracing |
| **Dipercaya** | Trustworthy | Graceful shutdown, time safety, testable |

**Innovation:** Made explicit design framework, not implicit. "Diamati" and "Dipercaya" are minimum definition, not nice-to-have.

---

### 3. Opinionated but Not Locking

**Philosophy:** Strong opinions AND clear escape hatches

| Capability | Hook Interface | Possible Implementations |
|------------|----------------|--------------------------|
| Auth | ✅ | JWT, OAuth2, RBAC |
| Caching | ✅ | Redis, Memcached, in-memory |
| Rate Limiting | ✅ | Redis-based, external |
| Event Bus | ✅ | Kafka, RabbitMQ, NATS |
| Secrets | ✅ | Vault, AWS SM, GCP SM |

**Innovation:** Clear "happy path" for 80% + adaptable for enterprise requirements.

---

### Innovation Risk Mitigation

| Innovation | Risk | Mitigation |
|------------|------|------------|
| AI-Native | AGENTS.md stale | PR review checklist |
| Three Pillars | Over-engineering | Clear minimum examples |
| Hook Patterns | Abstraction overhead | Default implementations |

---

## Developer Tool Specific Requirements

### Language & Runtime Support

| Go Version | Status | Notes |
|------------|--------|-------|
| **Go 1.24.x** | ✅ Supported + CI | Primary target |
| Go 1.23.x | ⚠️ May work | Not guaranteed |
| < 1.22 | ❌ Not supported | - |

---

### Installation & Setup

| Component | Approach |
|-----------|----------|
| **Primary Method** | GitHub "Use this template" or `git clone` |
| Dependencies | Go Modules (`go.mod`) |
| External Tools | Makefile + optional `tools.go` |
| CLI Generator | Out of scope (future v2) |

---

### Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/observability` | Logger, metrics, tracer |
| `internal/config` | Config loading |
| `internal/interface/http/httpx` | Response helpers |
| `internal/usecase/*` | Business logic |
| `internal/infra/postgres/*` | Repositories |
| `internal/runtimeutil` | Clock, ID generator |

### Extension Hooks

| Interface | Purpose |
|-----------|---------|
| Logger | Logging abstraction |
| Cache | Caching abstraction |
| RateLimiter | Rate limiting |
| EventPublisher | Event publishing |
| SecretProvider | Secret management |

---

### Documentation Matrix

| Document | Audience | Content |
|----------|----------|---------|
| README.md | All engineers | Quickstart, links |
| ARCHITECTURE.md | Tech leads | Three Pillars, layering |
| AGENTS.md | AI + humans | Patterns, DO/DON'T |
| TESTING.md | Test writers | AAA, examples |

---

### Examples & Migration

**Example Module:** Complete pipeline (Entity → Repo → Usecase → Handler → Tests)

**Migration Paths:**
- New Services: Use template → add domain
- Existing Services: Copy patterns → gradual refactor

---

## Project Scoping & Phased Development

### MVP Strategy

| Dimension | Value |
|-----------|-------|
| **Approach** | Platform MVP |
| **Team Size** | 1-2 engineers |
| **Timeline** | 2-4 weeks |

---

### MVP Feature Boundaries

#### ✅ Must Have (Non-Negotiable)

| Category | Scope |
|----------|-------|
| Foundation | main.go, wiring, chi, middleware |
| Config | koanf + validation + .env.example |
| Database | pgx + sqlc + migrate |
| Observability | zap, healthz/readyz, metrics, OTEL |
| DX Tooling | Makefile, docker-compose |
| Quality | golangci-lint + CI |
| Sample Module | Full E2E example |
| Documentation | README, ARCHITECTURE, AGENTS |

#### 🟡 v1.1 Candidates (If Time Tight)

| Feature | MVP | v1.1 |
|---------|-----|------|
| asynq | Interface only | Full worker |
| testcontainers | docker-compose | Programmatic |
| Obs stack | Jaeger only | Prometheus + Grafana |

---

### Risk Analysis

#### Technical Risks

| Risk | Mitigation |
|------|------------|
| sqlc learning curve | Example + docs |
| OTEL complexity | Minimal MVP |
| Overcomplexity | Simple entry UX |
| Maintenance | Version policy |

#### Adoption Risks

| Risk | Mitigation |
|------|------------|
| Developer resistance | DX-first, success stories |
| Tech lead preferences | Involve in design |
| "Kaku" perception | Show hook patterns |

---

### Solo Engineer 2-Week Plan

**Week 1:** Foundation + Config + DB + Example + Makefile
**Week 2:** Observability + CI + Tests + Docs

**Trim if needed:** asynq, testcontainers, full obs stack

---

## Functional Requirements

### 1. Project Setup & Initialization

- FR1: Developer can clone boilerplate and run service locally within 30 minutes
- FR2: Developer can configure service via environment variables or config file
- FR3: Developer can start all dependencies with single command (`make dev`)
- FR4: Developer can view example environment configuration via `.env.example`
- FR5: System initializes with graceful shutdown handling on OS signals

### 2. Configuration Management

- FR6: System can load configuration from environment variables
- FR7: System can load configuration from config file (optional)
- FR8: System validates configuration at startup and fails fast on invalid config
- FR9: System binds configuration to typed struct for type safety
- FR10: Developer can see clear error messages when configuration is invalid

### 3. HTTP API Foundation

- FR11: System exposes versioned HTTP API endpoints (`/api/v1/...`)
- FR12: System generates unique request ID for each incoming request
- FR13: System logs all HTTP requests with structured fields
- FR14: System recovers from panics in handlers and returns 500 response
- FR15: System propagates trace context via OpenTelemetry middleware
- FR16: Developer can add new HTTP endpoints following documented patterns
- FR17: System returns consistent response envelope for success/error
- FR18: System maps application errors to appropriate HTTP status codes

### 4. Database & Persistence

- FR19: System connects to PostgreSQL database with connection pooling
- FR20: System handles database connection timeouts gracefully
- FR21: Developer can write type-safe SQL queries using sqlc
- FR22: Developer can run database migrations via `make migrate-up`
- FR23: Developer can rollback database migrations via `make migrate-down`
- FR24: Developer can generate repository code via `make gen`
- FR25: System checks database connectivity as part of readiness check

### 5. Observability & Monitoring

- FR26: System exposes liveness endpoint (`/healthz`) returning 200
- FR27: System exposes readiness endpoint (`/readyz`) with dependency check
- FR28: System returns 503 on readiness when database fails
- FR29: System exposes Prometheus metrics endpoint (`/metrics`)
- FR30: System captures HTTP request count, latency, error count
- FR31: System produces structured JSON logs in production
- FR32: System produces human-readable logs in development
- FR33: System includes trace_id, request_id, path, method, status in logs
- FR34: System creates OpenTelemetry spans for HTTP requests

### 6. Developer Experience & Tooling

- FR35: Developer can run all tests with `make test`
- FR36: Developer can run linter with `make lint`
- FR37: System provides golangci-lint configuration
- FR38: Developer can view CI pipeline example
- FR39: Developer can follow example module as pattern
- FR40: Developer can understand architecture via ARCHITECTURE.md
- FR41: AI assistants can follow AGENTS.md for consistent code
- FR42: Developer can copy example module for new domain

### 7. Extension & Hooks

- FR43: System defines Logger interface
- FR44: System defines Cache interface
- FR45: System defines RateLimiter interface
- FR46: System defines EventPublisher interface
- FR47: System defines SecretProvider interface
- FR48: Developer can implement custom providers

### 8. Sample Module

- FR49: System includes complete example module
- FR50: Example includes entity with validation
- FR51: Example includes SQL migration
- FR52: Example includes sqlc repository
- FR53: Example includes usecase logic
- FR54: Example includes HTTP handler
- FR55: Example includes unit tests
- FR56: Example includes integration test

### FR Summary

| Area | Count |
|------|-------|
| Project Setup | 5 |
| Configuration | 5 |
| HTTP API | 8 |
| Database | 7 |
| Observability | 9 |
| DX & Tooling | 8 |
| Extensions | 6 |
| Sample Module | 8 |
| **Total** | **56** |

---

## Non-Functional Requirements

### Performance

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR1: Setup time | < 30 minutes | Clone → running |
| NFR2: make dev startup | < 60 seconds | Quick iteration |
| NFR3: HTTP response | < 100ms p95 | Responsive |
| NFR4: DB query latency | < 50ms p95 | Efficient |
| NFR5: Graceful shutdown | < 30 seconds | Clean exit |

### Reliability

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR6: Panic recovery | 100% | No server crash |
| NFR7: Shutdown | All requests complete | No drops |
| NFR8: Healthcheck | Honest /readyz | 503 when DB down |
| NFR9: Config validation | Fail-fast | No silent errors |

### Maintainability

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR10: Lint pass | 100% | Clean baseline |
| NFR11: No circular | Zero | Clean architecture |
| NFR12: Layer separation | Hexagonal | Compliance |
| NFR13: Complexity | cyclomatic ≤ 15 | Readable |
| NFR14: File size | ≤ 500 LOC | Manageable |

### Testability

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR15: Unit coverage | ≥ 70% example | Pattern demo |
| NFR16: Test execution | make test works | CI parity |
| NFR17: Isolation | Independent tests | Parallel safe |
| NFR18: Integration | 1+ HTTP+DB | E2E validation |

### Observability

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR19: Log format | JSON (prod) | Parseable |
| NFR20: Log context | trace_id, request_id | Correlation |
| NFR21: Metrics | Prometheus | Standard |
| NFR22: Tracing | OTEL ready | Future-proof |
| NFR23: Health latency | ≤ 10ms | Probe-friendly |

### Security Baseline

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR24: Secrets | Env/file only | No hardcode |
| NFR25: Errors | No stack traces | Leakage prevention |
| NFR26: Validation | All inputs | Basic protection |
| NFR27: Dependencies | No critical CVE | Security posture |

### Developer Experience

| NFR | Target | Rationale |
|-----|--------|-----------|
| NFR28: Doc accuracy | 100% works | Trust |
| NFR29: Example clarity | Follow → works | Efficiency |
| NFR30: Error messages | Clear, actionable | UX |
| NFR31: AI compat | AGENTS.md = compliance | AI-native |

### NFR Summary

| Category | Count |
|----------|-------|
| Performance | 5 |
| Reliability | 4 |
| Maintainability | 5 |
| Testability | 4 |
| Observability | 5 |
| Security | 4 |
| Developer Experience | 4 |
| **Total** | **31** |

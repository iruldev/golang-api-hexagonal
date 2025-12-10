---
stepsCompleted: [1, 2, 3]
inputDocuments: []
session_topic: 'Enterprise Golang Backend Boilerplate - Golden Template Design'
session_goals: 'Core capabilities list, architecture principles, module ideas, testing patterns, anti-patterns identification'
selected_approach: 'ai-recommended'
techniques_used: ['First Principles Thinking', 'SCAMPER Method', 'Reverse Brainstorming']
ideas_generated: ['Three Pillars Framework', '6 Hook Patterns', '6 Golden Template Dimensions', '9 Tech Substitutions', '8 Unified Packages', '5 Domain Adaptations', '25+ Anti-Patterns']
context_file: ''
---

# Brainstorming Session Results

**Facilitator:** Gan
**Date:** 2025-12-10

## Session Overview

**Topic:** Enterprise Golang Backend Boilerplate - Golden Template Design

**Goals:**
1. 📋 Daftar kemampuan inti (HTTP API, DB layer, job queue, config, observability, security)
2. 🧱 Set prinsip arsitektur & quality bar (regulated-ready, observability-first, testable by design)
3. 🧰 Daftar modul & fitur konkret (must-have vs nice-to-have)
4. 🧪 Ide pola testing & AI integration (AGENTS.md, test structure)
5. ⚠️ List risiko & anti-pattern yang harus dihindari

### Focus Areas
- **Architecture & Tech Stack:** Layering, observability-first, background jobs, CI/CD hooks
- **Developer Experience:** Makefile/CLI, documentation structure, extensibility patterns
- **Standards & Guardrails:** Opinionated principles, boundaries for flexibility

## Technique Selection

**Approach:** AI-Recommended Techniques
**Analysis Context:** Enterprise Golang Backend Boilerplate with focus on architecture, DX, and standards

**Recommended Techniques:**
1. **First Principles Thinking:** Strip assumptions, rebuild from fundamental truths
2. **SCAMPER Method:** Systematic exploration of all component aspects
3. **Reverse Brainstorming:** Uncover anti-patterns by asking "how to fail"

---

## Technique Execution Results

### 🧱 Phase 1: First Principles Thinking

**Core Question:** "Apa yang PASTI HARUS ada agar service production-ready?"

#### The Three Pillars Framework

##### Pillar 1: BERJALAN (Basic Functionality)

| Component | Minimum Requirements |
|-----------|---------------------|
| **Entry Point** | `main.go` dengan config load, logger init, dependency init, HTTP server, graceful shutdown |
| **HTTP Server & Routing** | Stable server, middleware support (logging, recovery, request ID), path/query params |
| **Config Management** | ENV-based (wajib), environment separation (local/dev/staging/prod), defaults & required validation |
| **Dependency Connections** | DB/broker driver dengan connection pool & timeout |
| **Error Handling** | Global recover middleware, consistent JSON error response format |

##### Pillar 2: DIAMATI (Observability)

| Component | Minimum Requirements |
|-----------|---------------------|
| **Structured Logging** | JSON format untuk production, log levels, timestamp + request_id + path/method/status |
| **Health Endpoints** | `GET /healthz` (liveness), `GET /readyz` (readiness) |
| **Metrics** | Prometheus-style: request count, latency histogram, error count via `/metrics` |
| **Tracing Hook** | Middleware integration untuk trace/span per request, OTLP-ready |

##### Pillar 3: DIPERCAYA (Production-Worthy)

| Component | Minimum Requirements |
|-----------|---------------------|
| **Reliability** | Graceful shutdown, timeouts (HTTP client, DB), sensible retry policy |
| **Security Baseline** | No secrets in code, no PII in logs, auth middleware hook (JWT/API key abstraction) |
| **Input Validation** | Required fields & format validation sebelum processing |
| **Environment Separation** | `APP_ENV` dengan behavior per-environment (log format, strictness) |
| **Testing** | Unit tests (domain/usecase), smoke/integration tests (HTTP + DB) |
| **Linting & CI** | golangci-lint, basic CI pipeline (`go test`, `lint run`) |
| **Time Safety** | UTC internal, timezone conversion only at boundaries |

#### 💡 Key Insight
> "Kalau salah satu cluster besar ini nggak ada, itu bukan enterprise-grade boilerplate, paling tinggi cuma level example project."

---

#### Hook Patterns: "Contract Ready, Implementation Flexible"

| Hook Area | Interface/Contract | Implementations |
|-----------|-------------------|-----------------|
| **Caching** | `Cache` (Get/Set/Del + TTL) | In-memory, Redis, Memcached |
| **Rate Limiting** | `RateLimiter.Allow(ctx, key)` | In-memory token bucket, Redis, API Gateway |
| **Feature Flags** | `FeatureFlagClient.IsEnabled(ctx, flag, actor)` | File/ENV, LaunchDarkly, Unleash |
| **Outbound Integration** | `NotificationSender`, `HttpClient` wrapper | Mock/no-op, SendGrid, Twilio |
| **Event Bus** | `EventPublisher`, `EventSubscriber` | In-memory, Kafka, RabbitMQ, NATS, PubSub |
| **Audit Trail** | `AuditLogger`, `DomainEventSink` | Log file, DB table, Event stream |

> **Principle:** Semua hal yang lintas domain, banyak variasi vendor, dan jadi infra decision → disediakan hook + contoh, bukan di-hardcode.

---

#### Golden Template Differentiators (Beyond Minimum)

##### A. Developer Experience (DX)

| Aspect | Implementation |
|--------|----------------|
| **One-command dev** | `make dev` → docker-compose + app auto-reload |
| **End-to-end scaffolding** | Sample entity dengan migration, repo, usecase, handler, job, tests |
| **Makefile komplit** | `test`, `lint`, `gen`, `migrate-up`, `seed`, dll |
| **Config samples** | `.env.local`, `.env.staging`, `.env.prod.example` |

##### B. Documentation

| Doc | Purpose |
|-----|---------|
| **README.md** | "5 min to run", tech stack, common pitfalls |
| **ARCHITECTURE.md** | Layering, data flow, dependency direction, diagrams |
| **AGENTS.md** | AI instructions: structure, test patterns, DO/DON'T |
| **CONTRIBUTING.md** | PR/commit/style standards |

##### C. Observability & Operability

- Metrics & tracing nyala sejak awal (route labels, latency histogram)
- Contoh business metric (e.g., `notes_created_total`)
- Healthcheck yang jujur (`/healthz` vs `/readyz` berbeda)
- Sample alerting rules (HighErrorRate, HighLatency, AppDown)

##### D. Quality & Governance

- Lint + format + security check (golangci-lint, gosec)
- Test structure yang diajarkan (unit mock, integration testcontainers)
- Opinionated error model (wrap, categorize, HTTP mapping)

##### E. Time & Timezone Safety

- DB: `TIMESTAMPTZ` in UTC
- Go: `time.Time` always UTC
- Package `timeutil` dengan `NowUTC()`, consistent Parse/Format

##### F. Extensibility

- Modular folder layout per domain
- Configurable on/off (`ENABLE_ASYNC_JOBS`, `ENABLE_METRICS`)

---

#### 🏆 The Distinction

| Type | Characteristic |
|------|----------------|
| **Working Boilerplate** | Bisa dipakai kalau udah ngerti banyak hal, masih mikir setup |
| **Golden Template** | Three Pillars solid + DX/obs/docs/hooks → engineer baru productive dalam jam |

---

### 📐 Phase 2: SCAMPER Method

#### S - SUBSTITUTE: Enterprise-Grade Tech Stack Decisions

| Area | Common Default | Golden Template Choice | Rationale |
|------|---------------|----------------------|-----------|
| **HTTP Framework** | Gin/Echo | `chi` + `otelhttp` | Idiomatic, composable, less magic, trace-ready |
| **Database** | GORM | `sqlc` + `pgx` | Type-safe SQL, explicit queries, reviewable |
| **Logging** | logrus/printf | `zap` | High-perf structured logging, enterprise standard |
| **Config** | raw ENV/Viper | `koanf` + struct validation | Composable, testable, fail-fast |
| **Testing** | go test minimal | `testify` + `testcontainers` + AAA | Readable, real DBs in test, clear patterns |
| **Background Jobs** | DIY goroutine/cron | `asynq` | Retry, DLQ, dashboard, observability |
| **HTTP Client** | raw `http.Client` | Own wrapper + otel + retry | One correct way for outbound calls |
| **Build/Dev** | manual commands | `Makefile` as DX interface | Unified UX, `make help` for discovery |
| **API Docs** | swaggo comments | `oapi-codegen` (OpenAPI-first) | Contract-first, generate handlers/clients |

> **Philosophy:** "Substitute Go defaults with enterprise-proven choices that prioritize explicitness, testability, and observability"

---

#### C - COMBINE: Unified Packages for Simplicity

| Combined Concern | Package | Components Merged |
|-----------------|---------|-------------------|
| **Observability Super-Paket** | `internal/observability` | Logger + TracerProvider + MeterProvider, single `Init(cfg)` |
| **Error + HTTP Response** | `internal/interface/http/httpx` | AppError model + Success/Fail response helpers |
| **Config + Validation + Secrets** | `internal/infra/config` | koanf load + struct bind + validate + secret provider interface |
| **HTTP Client + Retry + Observability** | `internal/infra/httpclient` | Timeout + otelhttp + retry + logging in one constructor |
| **Sample Feature Bundle** | `internal/modules/notes` (example) | Entity + migration + repo + usecase + handler + tests |
| **Jobs + Events + Audit** | `activities` / `events` layer | EventDispatcher + EnqueueJob + RecordAuditTrail unified |
| **Runtime Utilities** | `runtimeutil` / `foundation` | Clock interface + ID generator + idempotency keys |
| **Quality + AI Alignment** | `AGENTS.md` + `TESTING.md` + `.golangci.yml` | AI instructions + test standards + lint rules in sync |

> **Philosophy:** "Combine concerns that are mentally linked, so engineers think of them as ONE thing, not three libraries"

---

#### A - ADAPT: Patterns from Other Domains

| Source Domain | Pattern | Adaptation |
|--------------|---------|------------|
| **Cloud-native/DevOps** | 12-Factor App, K8s-friendly | Config via env, stateless, healthcheck first-class |
| **Frontend (React)** | One clear structure + data flow | `interface → usecase → domain → infra` pattern |
| **Game Dev** | ECS, determinism | Pure domain logic, Clock/ID as dependencies |
| **Fintech/Regulated** | Audit trail, traceability | Hook for audit logger & domain events |
| **SRE** | Golden signals (latency, traffic, error, saturation) | Auto-expose 4 signals from day one |

---

#### M - MODIFY: Enlarge vs Shrink

| Perbesar (More than typical) | Perkecil (Less than typical) |
|------------------------------|------------------------------|
| Observability = wajib, bukan bonus | Folder nesting berlebihan |
| Testing dengan real example | Generic interface premature |
| Docs (README, ARCH, AGENTS) as MVP | Config modes (cukup 3-4 env) |

---

#### P - PUT TO OTHER USES: Multi-Purpose Components

| Component | Primary Purpose | Additional Uses |
|-----------|----------------|-----------------|
| **Healthcheck** | Liveness/readiness | Dependency graph, debug ops |
| **Logger** | Debug | Audit events, simple analytics |
| **Metrics** | Monitoring | SLO tracking, capacity planning |
| **AGENTS.md** | AI guide | Local engineering standard |
| **Sample module** | Example | Copy-paste template, training material |

---

#### E - ELIMINATE: What to Remove

| Remove | Reason |
|--------|--------|
| Over-abstraction | Interface dengan 1 impl, mulai concrete |
| Wrapper kosong stdlib | Tanpa value tambah |
| Support semua protokol v1 | Fokus REST, gRPC/GraphQL documented future |
| Framework mengunci | Golden = kuat tapi tidak mengurung |
| Config hardcode | Semua ke env/config |

---

#### R - REVERSE/REARRANGE: Inverted Flows

| Traditional Order | Golden Template Order |
|-------------------|----------------------|
| Feature → observability → test | Quality-first: obs + error → feature |
| Code → test → docs | Skeleton docs + test pattern first |
| "We have Postgres, what features?" | Domain/usecase first → adapter later |
| Global vars everywhere | Wiring di composition root only |
| Read docs → then try | `make dev` → see it work → then read docs |

---

### ⚠️ Phase 3: Reverse Brainstorming - Anti-Patterns & Risks

**Core Question:** "Bagaimana cara membuat boilerplate ini GAGAL TOTAL?"

#### 1. DX Nightmare 🤬

| Anti-Pattern | Description |
|-------------|-------------|
| Setup super ribet | 7+ services manual, 10+ shell steps, no one-command start |
| Dokumentasi bohong | Outdated README, missing env vars, broken examples |
| Tooling inconsistent | Mixed test runners, lint config ignored by CI |
| Folder structure random | Mixed naming (controllers/handlers/services), no clear pattern |
| Too much magic | Autogen everything, hard to debug, ritual-based startup |

#### 2. Maintenance Hell 🔥

| Anti-Pattern | Description |
|-------------|-------------|
| God package/function | `pkg/common`, `pkg/utils`, 2000+ line files |
| Over-abstraction | Interface for every single implementation |
| Circular dependency | usecase ↔ infra imports |
| No boundaries | Handler does DB + cache + HTTP + events + file in one func |
| No standard error model | Mixed error types, random HTTP status mapping |

#### 3. Performance Disaster 💸

| Anti-Pattern | Description |
|-------------|-------------|
| No timeout | `context.Background()` everywhere, default http.Client |
| Goroutine liar | `go func()` without context, error handling, limits |
| N+1 queries | ORM loops without understanding, no index guidance |
| Heavy sync logging | Verbose plaintext, no sampling, no level control |
| Debug mode always ON | Verbose stacks in production |

#### 4. Security Holes 🔓

| Anti-Pattern | Description |
|-------------|-------------|
| Log sensitive data | Passwords, tokens, PII in plaintext logs |
| Hardcode secrets | API keys, DB passwords in source code |
| No auth pattern | All endpoints open, no middleware hooks |
| No input validation | Direct JSON → DB, string concat SQL |
| CORS/exposure ngawur | `*` everywhere, /metrics /debug public |

#### 5. Adoption Failure 🚫

| Anti-Pattern | Description |
|-------------|-------------|
| Too complex from day 1 | REST + gRPC + GraphQL + WebSocket + CLI + multi-DB |
| No real examples | Only Hello World, no best practice demo |
| Opinionated without docs | Strong decisions but no WHY explained |
| Unstable/breaking often | Frequent changes without migration path |
| Hard to customize | Locked choices, no extension points |

#### 6. Quality Killers 💀

| Anti-Pattern | Description |
|-------------|-------------|
| No tests | "Test when stable" mentality |
| No linting | Style and bugs left unchecked |
| Time/timezone chaos | Mixed local/UTC, various DB formats |

---

## Session Summary

### 🎯 Techniques Used
1. **First Principles Thinking** → Three Pillars Framework + 6 Hook Patterns + 6 Golden Template Dimensions
2. **SCAMPER Method** → 7 elements with 30+ actionable insights
3. **Reverse Brainstorming** → 6 failure categories as anti-pattern checklist

### 📊 Total Ideas Generated
- **Core capabilities:** 15+ components identified
- **Architecture principles:** 10+ principles defined
- **Tech stack decisions:** 9 substitutions
- **Package designs:** 8 unified packages
- **Anti-patterns:** 25+ explicit DON'Ts

### 🏆 Key Frameworks Created
1. **Three Pillars:** Berjalan → Diamati → Dipercaya
2. **Hook Pattern:** Contract ready, implementation flexible
3. **Golden Template Distinction:** Working boilerplate vs Golden Template
4. **Quality-First Flow:** Observability → Error model → Features

### ➡️ Next Steps
This brainstorming output is ready to be transformed into:
- PRD (Product Requirements Document)
- Architecture Document
- AGENTS.md (AI Instructions)
- Implementation Backlog

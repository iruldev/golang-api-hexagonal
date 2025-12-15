---
stepsCompleted: [1]
inputDocuments: []
session_topic: 'Mendesain Golden Template Production-Grade untuk Go Backend'
session_goals: 'Decision pack, Blueprint repo, Golden path spec, Toolchain & CI spec, Migration plan'
selected_approach: 'progressive-flow'
techniques_used: []
ideas_generated: []
context_file: 'docs/project-overview.md'
---

# Brainstorming Session: Golden Template Production-Grade

**Date:** 2025-12-15
**Facilitator:** AI Assistant
**Participant:** Gan

---

## Session Overview

**Topic:** Mendesain "Golden Template" Production-Grade untuk Go Backend

**Goals:**
1. Decision Pack - Keputusan final + rationale (router, error, config, auth, logging, DB)
2. Blueprint Repo - Struktur folder + "rules of the road" per layer
3. Golden Path Spec - 1 fitur end-to-end sebagai reference implementation
4. Toolchain & CI Spec - Linting, testing pyramid, coverage, release
5. Migration Plan - Phase-based dengan definisi "done"

### Context Guidance

Berdasarkan dokumentasi proyek yang sudah ada:
- Proyek: Backend Service Golang Boilerplate
- Arsitektur: Hexagonal (Ports & Adapters)
- Tech Stack: Go 1.24, chi v5, PostgreSQL, Redis, Kafka, gRPC, GraphQL
- Existing: Domain, Usecase, Interface, Infra layers sudah terdefinisi

### Session Setup

**Approach:** Progressive Flow - Mulai dari eksplorasi luas, lalu narrow down secara sistematis untuk menghasilkan keputusan konkret.

---

## Phase 1: Divergent Exploration ✅

### Fundamental Truths (Non-Negotiable)

1. **Build once, run anywhere** → local/dev/staging/prod parity tinggi (config-driven)
2. **Standar enforced, bukan documented** → CI gate + tooling, bukan hanya doc
3. **Semua footguns akan kepakai** → default aman: timeouts, context, retries, rate-limit
4. **Dependency graph tanpa aturan = spaghetti** → import boundary + depguard wajib
5. **Observability kontrak, bukan fitur** → trace-id, structured log, metrics minimal
6. **Data contract raja** → schema/migrations/versioning disiplin, backward compat
7. **Concurrency tanpa guard = bug** → context, idempotency, dedup, locking template
8. **Timeout = requirement** → semua IO wajib timeout + deadline
9. **Testing non-deterministic = flaky = dibenci** → repeatable, seed jelas, isolasi
10. **Local dev ribet = tim berhenti pakai** → "1 command" + "clean reset"
11. **Runbook bagian sistem** → tanpa runbook, incident chaos
12. **Versioning & release discipline** → semantic versioning + changelog

### What-If Scenarios

| Scenario | Exploration |
|----------|-------------|
| **Dev baru produktif 15 menit** | `make up` = compose+migrate+seed+run; `make verify` = lint+unit+integration; first-task guide |
| **Zero configuration** | Default config siap (dev), override env, `.env.example` auto-generated + validation startup |
| **Every error self-documents** | Typed errors + public error code + hint; error include trace_id + request_id |
| **Tests write themselves** | Contract tests dari OpenAPI/Proto/GraphQL; golden test cases untuk error mapping |
| **Migration zero-downtime** | Feature flags built-in; expand/contract pattern; background backfill template |
| **Module bisa di-generate** | `bplat new module payments` → handler/usecase/repo/migrations/tests/docs |
| **Incidents "boring"** | Default dashboards + SLO + alerts + runbook links di alert description |
| **Security audit tinggal "centang"** | Secret scanning, dependency scanning, SBOM, minimal privileges, TLS-default |

### Wild Ideas (Unfiltered)

1. **Golden path as code** - `examples/goldenpath/` yang selalu build & test, referensi hidup
2. **Policy pack** - `policy/` untuk dep allowlist, error codes, log fields mandatory
3. **One command to reproduce incident** - Script replay request dari log ke staging
4. **Chaos mode lokal** - Toggle inject latency/error ke DB/queue
5. **Auto-ADR generator** - Perubahan besar butuh ADR, template auto dari CLI
6. **Strict mode** - Build fail kalau handler tanpa tracing span, query tanpa context
7. **Golden PR bot** - Komentar otomatis kalau boundary dilanggar
8. **Unified contract layer** - Satu spec (OpenAPI/Proto) source-of-truth → generate stubs
9. **Migration simulator** - CI job migrate up/down + verify no destructive
10. **Idempotency kit** - Middleware + storage pattern template

**Total Ideas Generated:** 30+

---

## Phase 2: Pattern Recognition ✅

### Six Thinking Hats Analysis

| Hat | Perspective | Key Insights |
|-----|-------------|--------------|
| ⚪ White | Facts | Hexagonal arch exists, basic OTel, chi+sqlc ready; missing: enforcement, policy pack |
| ❤️ Red | Emotions | Exciting: golden path as code, strict mode. Scary: migration, too many rules |
| 💛 Yellow | Benefits | 3x productivity, 50% faster MTTR, zero CVE, consistency, tech debt prevention |
| ⬛ Black | Risks | Over-engineering, dev resistance, maintenance, breaking changes, perf overhead |
| 💚 Green | Alternatives | Tiered adoption, opt-in strict mode, progressive enhancement, escape hatches |
| 🔵 Blue | Summary | 5 layers identified: Foundation, Safety, Quality, Documentation, Migration |

### Identified Pattern Clusters

1. **🏗️ Foundation Layer** - Config, logging, tracing, error handling
2. **🛡️ Safety Layer** - Timeouts, retries, circuit breakers, idempotency
3. **✅ Quality Layer** - Linting, testing, coverage, CI gates
4. **📚 Documentation Layer** - Runbooks, ADRs, golden path
5. **🚀 Migration Layer** - Feature flags, phased rollout, backward compat

### Priority Ranking (User Decision)

| Priority | Cluster | Rationale |
|----------|---------|-----------|
| **1** | B. Quality Gates | Foundation for consistency dan safety |
| **2** | A. DX | Adoption enabler, reduce friction |
| **3** | D. Safety Defaults | Reliability-by-default |
| **4** | C. Observability | Meaningful signals after safety |
| **5** | E. Migration Tools | Bertahap, plan first |

**Final Order: B → A → D → C → E**

---

## Phase 3: Convergent Focus ✅

### SCAMPER Analysis - Top 3 Clusters

#### Cluster B: Quality Gates
| Lens | Application |
|------|-------------|
| Substitute | golangci-lint → custom analyzer bundle |
| Combine | lint + test + coverage dalam 1 gate |
| Adapt | pre-commit + CI dual enforcement |
| Modify | strict mode (fail) vs warning mode |
| Eliminate | Manual code review untuk style |

#### Cluster A: DX
| Lens | Application |
|------|-------------|
| Combine | `make up` = compose + migrate + seed + server |
| Adapt | .env.example auto-validation on startup |
| Modify | `make reset` untuk clean slate |

#### Cluster D: Safety Defaults
| Lens | Application |
|------|-------------|
| Combine | Timeout + context + retry dalam 1 wrapper |
| Modify | Default timeouts (30s HTTP, 5s DB) |
| Eliminate | Panic in normal flow |

### Key Decisions Made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Pre-commit hooks | Recommended | Flexibility, tidak blocking dev |
| Unit test coverage | 80% | Achievable, meaningful |
| Strict mode | Default ON | Enforce dari awal |
| Dev environment | Container-first | Parity tinggi |
| HTTP timeout | 30s | Reasonable default |

---

## Phase 4: Decision Crystallization ✅

_All brainstorming consolidated into actionable deliverables._

---

## Final Outputs

### 1. Decision Pack

| Category | Decision | Rationale |
|----------|----------|-----------|
| **Router** | chi v5 | Already adopted, stable, middleware-friendly |
| **Protocol Boundaries** | HTTP→Domain clean, no chi types in usecase | Testability, portability |
| **Error Standard** | Typed domain errors + public codes + hint | Self-documenting, trace-friendly |
| **Response Standard** | `response.Envelope{Data, Error, Meta}` | Consistent client contract |
| **Config** | koanf + .env + validation on startup | Flexible, fail-fast |
| **Auth Strategy** | JWT default + API key for service-to-service | Already implemented |
| **Logging** | Zap structured JSON + mandatory fields (trace_id, request_id) | Queryable, correlated |
| **Tracing** | OpenTelemetry auto-instrumentation + manual spans for business ops | Full observability |
| **DB Pattern** | sqlc + repository interface + transaction wrapper | Type-safe, testable |
| **Queue Pattern** | Asynq + idempotency middleware | Reliable, deduped |

### 2. Blueprint Repo

```
golang-api-hexagonal/
├── cmd/                          # Entry points ONLY
│   ├── server/main.go           # ✅ HTTP server bootstrap
│   ├── worker/main.go           # ✅ Job processor
│   ├── scheduler/main.go        # ✅ Cron jobs
│   └── bplat/                   # ✅ CLI scaffolding
│
├── internal/                     # Private code
│   ├── domain/                  # 🔒 NO external imports
│   │   └── {module}/
│   │       ├── entity.go        # Entities + Validate()
│   │       ├── errors.go        # Typed domain errors
│   │       └── repository.go    # Interface (port)
│   │
│   ├── usecase/                 # 🔒 Only imports domain
│   │   └── {module}/
│   │       └── usecase.go       # Business logic
│   │
│   ├── interface/               # Adapters IN
│   │   ├── http/
│   │   │   ├── middleware/      # Auth, rate-limit, tracing
│   │   │   ├── {module}/        # Handlers + DTOs
│   │   │   ├── response/        # Envelope pattern
│   │   │   └── router.go        # Route registration
│   │   ├── grpc/
│   │   └── graphql/
│   │
│   ├── infra/                   # Adapters OUT
│   │   ├── postgres/            # sqlc implementations
│   │   ├── redis/               # Cache + rate limiter
│   │   └── kafka/               # Event publisher
│   │
│   ├── worker/                  # Job handlers
│   │   ├── tasks/               # Task definitions
│   │   └── patterns/            # Fanout, idempotency
│   │
│   ├── observability/           # Cross-cutting
│   │   ├── logger.go            # Zap setup
│   │   ├── tracer.go            # OTel setup
│   │   ├── metrics.go           # Prometheus
│   │   └── audit.go             # Audit logging
│   │
│   └── config/                  # Configuration
│
├── db/
│   ├── migrations/              # golang-migrate files
│   └── queries/                 # sqlc queries
│
├── examples/
│   └── goldenpath/              # 🌟 LIVING REFERENCE
│
├── policy/                      # 🛡️ Enforcement rules
│   ├── depguard.yaml            # Import boundaries
│   └── error-codes.yaml         # Registered error codes
│
└── deploy/
    ├── prometheus/alerts.yaml
    └── grafana/dashboards/
```

**Rules of the Road:**

| Layer | CAN Import | CANNOT Import |
|-------|------------|---------------|
| domain | stdlib only | usecase, interface, infra |
| usecase | domain | interface, infra |
| interface | domain, usecase | - |
| infra | domain | usecase, interface |

### 3. Golden Path Spec

**Example Module: `note` (CRUD Reference)**

```
Feature: Create Note (End-to-End)

1. HTTP Layer (interface/http/note/)
   ├── handler.go          # POST /notes
   ├── dto.go              # CreateNoteRequest/Response
   └── handler_test.go     # Table-driven tests

2. Validation
   ├── Request validation  # DTO level
   └── Domain validation   # entity.Validate()

3. Use Case Layer (usecase/note/)
   ├── usecase.go          # Create() with audit logging
   └── usecase_test.go     # Mock repo tests

4. Repository Layer (domain + infra)
   ├── domain/note/repository.go  # Interface
   └── infra/postgres/note.go     # Implementation

5. Database
   ├── migrations/*.sql    # Up/down migrations
   └── queries/note.sql    # sqlc CRUD

6. Observability
   ├── Tracing span        # Automatic via middleware
   ├── Metrics             # request_count, latency
   └── Audit log           # LogAudit on create

7. Tests
   ├── Unit tests          # handler, usecase
   ├── Integration test    # testcontainers
   └── Golden response     # Expected JSON snapshots
```

### 4. Toolchain & CI Spec

**Makefile Targets:**

```makefile
# DX Commands
up:          docker-compose up -d && migrate-up && run-dev
down:        docker-compose down
reset:       down && docker volume prune && up
verify:      lint test integration

# Quality Gates
lint:        golangci-lint run ./... --config .golangci.yml
lint-fix:    golangci-lint run ./... --fix
test:        go test -race -cover ./... -coverprofile=coverage.out
coverage:    go tool cover -func=coverage.out | grep total
integration: go test -tags=integration ./...

# Code Generation
sqlc:        sqlc generate
proto:       buf generate
gql:         go generate ./internal/interface/graphql/...
generate:    sqlc proto gql
```

**CI Pipeline (.github/workflows/ci.yml):**

```yaml
jobs:
  quality-gate:
    steps:
      - lint           # golangci-lint (strict mode)
      - test           # unit tests + coverage
      - coverage-check # fail if < 80%
      - integration    # testcontainers

  security:
    steps:
      - secret-scan    # gitleaks
      - dependency-scan # govulncheck
      - sbom           # generate SBOM

  build:
    needs: [quality-gate, security]
    steps:
      - build-binaries
      - build-docker
```

**Coverage Target:** 80% (enforced in CI)
**Pre-commit:** Recommended (not blocking)
**Strict Mode:** Default ON

### 5. Migration Plan

**Phase 1: Foundation (Week 1-2)**
- [ ] Lock .golangci.yml config
- [ ] Add depguard rules for import boundaries
- [ ] Setup CI quality gate (lint + test)
- [ ] Add coverage enforcement (80%)

**Done when:** CI green, all PRs must pass gates

---

**Phase 2: DX Enhancement (Week 3-4)**
- [ ] Consolidate Makefile (up/down/reset/verify)
- [ ] Auto-generate .env.example with validation
- [ ] Create first-task onboarding guide
- [ ] Document quickstart < 5 steps

**Done when:** New dev productive in <30 min

---

**Phase 3: Safety Defaults (Week 5-6)**
- [ ] Add timeout wrappers (HTTP 30s, DB 5s)
- [ ] Implement typed errors with codes
- [ ] Add idempotency middleware template
- [ ] Graceful shutdown for server/worker

**Done when:** All IO has timeouts, errors self-documenting

---

**Phase 4: Observability (Week 7-8)**
- [ ] Mandatory trace_id in all responses
- [ ] Structured logging contract
- [ ] Default Prometheus alerts
- [ ] Runbook for each alert

**Done when:** Incidents can be traced end-to-end

---

**Phase 5: Polish (Week 9-10)**
- [ ] examples/goldenpath/ as living reference
- [ ] ADR template + existing decisions
- [ ] bplat generate enhancements
- [ ] Feature flag framework if needed

**Done when:** Golden template ready for team adoption

---

*Brainstorming Session Complete - 2025-12-15*

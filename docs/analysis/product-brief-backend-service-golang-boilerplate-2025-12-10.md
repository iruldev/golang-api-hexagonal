---
stepsCompleted: [1, 2, 3, 4, 5, 6]
inputDocuments:
  - 'docs/analysis/research/technical-golang-enterprise-boilerplate-research-2025-12-10.md'
  - 'docs/analysis/brainstorming-session-2025-12-10.md'
workflowType: 'product-brief'
lastStep: 6
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-10'
---

# Product Brief: Backend Service Golang Boilerplate

**Date:** 2025-12-10
**Author:** Gan

---

## Executive Summary

**Backend Service Golang Boilerplate** adalah enterprise-grade "golden template" yang menyediakan fondasi production-ready untuk membangun backend services di Go. Berbeda dari boilerplate sederhana yang hanya menyediakan skeleton kode, template ini mengintegrasikan **observability-first architecture**, **opinionated tech stack**, dan **AI-native developer experience** dalam satu paket yang siap pakai.

**Target Impact:**
- Time-to-first-feature: dari minggu → hari
- Onboarding engineer baru: dari hari → jam
- Konsistensi arsitektur: dari fragmentasi → standardisasi

**Core Philosophy:** Opinionated but not locking – keputusan tegas untuk akselerasi, dengan hook untuk fleksibilitas.

---

## Core Vision

### Problem Statement

Tim engineering backend menghadapi **empat masalah utama** saat membangun service baru:

1. **Repetitive Setup** – Setiap service baru dimulai hampir dari nol: routing, config, logging, DB, jobs, observability, testing, CI
2. **Inconsistent Patterns** – Struktur folder, error handling, logging, dan testing berbeda antar service
3. **Missing Production Basics** – Healthcheck, metrics, tracing, security hygiene sering ketinggalan sampai menjelang go-live
4. **Slow Onboarding** – Engineer baru butuh waktu lama memahami "cara yang benar" karena knowledge tersebar di kepala senior

### Problem Impact

| Stakeholder | Pain Point |
|-------------|------------|
| **Backend Engineers** | Capek mengulang hal yang sama, ingin fokus ke domain logic |
| **Tech Leads/Architects** | Sulit menjaga konsistensi & kualitas antar squad |
| **SRE/Platform Team** | Setup alert dan dashboard sulit karena variasi pattern |
| **New Joiners** | Bingung karena setiap project berbeda, belajar lambat |

### Why Existing Solutions Fall Short

| Current Solution | Gap |
|-----------------|-----|
| **Copy-paste project lama** | Membawa legacy decisions & technical debt |
| **Internal boilerplate lama** | Tidak ter-update dengan best practices 2024-2025, minim observability |
| **Open-source starters** | Jarang enterprise-ready, terlalu framework-centric |
| **Squad-own patterns** | Fragmentasi style, sulit sharing knowledge |

### Proposed Solution

**Golden Template** yang menyediakan:

**Foundation:**
- Three Pillars Architecture: Berjalan (functionality) → Diamati (observability) → Dipercaya (production-worthy)
- Opinionated Stack: chi, sqlc+pgx, zap, koanf, asynq, OpenTelemetry
- Hook Pattern: Auth, cache, rate limit, event bus, secret manager – contract ready, implementation flexible

**Developer Experience:**
- One-command dev: `make dev` → service running dengan DB, metrics, tracing
- End-to-end sample: Example module dengan migration, repo, usecase, handler, job, tests
- Makefile lengkap: test, lint, gen, migrate, seed

**Documentation & AI:**
- AGENTS.md sebagai first-class citizen – panduan untuk AI dan manusia
- ARCHITECTURE.md, README.md, TESTING.md terintegrasi

### Key Differentiators

1. **Enterprise-grade dari hari pertama** – Bukan CRUD demo, tapi fondasi untuk fintech/regulated environment
2. **Opinionated but not locking** – Stack tegas dengan hook untuk adaptasi
3. **AI-native by design** – AGENTS.md sebagai standard, AI assistant selaras dengan project
4. **Strong DX + Quality bundle** – Bukan skeleton, tapi paket lengkap (observability, testing, docs)
5. **Battle-tested decisions** – Keputusan dari pengalaman nyata, bukan teoretis
6. **Perfect timing** – Go mature, OTEL standard, AI assistant mainstream – semua konvergen sekarang

---

## Target Users

### Primary User: Backend Engineer

**Persona: Andi Pratama**
- **Profile:** 28 tahun, Backend Engineer Mid-level, 3 tahun pengalaman di fintech
- **Context:** Sering diminta bikin/maintain microservice baru (collections, risk, disbursement)
- **Mindset:** Sudah cukup senior untuk tahu "harusnya gimana", tapi sering terjebak reality: deadline mepet, banyak utang teknis

**Current Pain (Without Golden Template):**

| Time | Activity | Frustration |
|------|----------|-------------|
| Pagi | Copy project lama, bersih-bersih hal tidak relevan | Wasting time on repetitive setup |
| Siang | Setup config, DB, logging, routing dari scratch | Bingung ikut pattern service mana |
| Sore | Fitur jalan tapi healthcheck/metrics/test belum ready | On-call horror karena kurang observability |

**Transformed Experience (With Golden Template):**

| Time | Activity | Outcome |
|------|----------|---------|
| Pagi | Clone boilerplate → copy example module → `make dev` | Service struktur sudah ready |
| Siang | Fokus modif entity, usecase, repo untuk domain logic | HTTP, DB, metrics, tracing sudah jalan |
| Sore | Tambah test berdasar template → push → CI auto lint+test | Code review fokus business logic |

**"Aha Moment":** "Saya tidak perlu lagi mikir routing, logging, DB wiring — tinggal fokus ke *apa* yang mau dibangun."

---

### Secondary Users

| Persona | Role | Pain Point | Value | Influence |
|---------|------|------------|-------|-----------|
| **Rudi** | Tech Lead/Architect | Pattern beda tiap service, code review berat | "Service baru WAJIB start dari boilerplate" | Decision maker |
| **Maya** | SRE/Platform | Alert & dashboard config berbeda tiap service | Metrics, tracing, healthcheck standar | Strong influencer |
| **Dina** | New Joiner/Junior | Tiap project beda, susah paham "cara benar" | Satu repo untuk belajar struktur ideal | Beneficiary terbesar |

---

### User Journey

| Phase | Timeline | Activity | Outcome |
|-------|----------|----------|---------|
| **Discovery** | Day 0 | Tech Lead introduce atau find di GitHub | Trigger: butuh service baru |
| **First Experience** | Day 0-1 | Clone → `make dev` → service running <30min | "Ini lengkap banget dari awal" |
| **Realize Value** | Week 1-2 | Create module, write tests, PR review | "Service baru sehat sejak lahir" |

---

## Success Metrics

### User Success Metrics (Per Persona)

| Persona | Metric | Target |
|---------|--------|--------|
| **Andi (Backend Engineer)** | Time-to-first-service | < 30 menit |
| | Time-to-first-feature | ≤ 2 hari kerja |
| | Setup rework reduction | ≥ 70% |
| | Repeat usage | Yes untuk service berikutnya |
| **Rudi (Tech Lead)** | Service adoption | ≥ 80% service baru dalam 6-12 bulan |
| | Code review time untuk "basic" | ↓ ≥ 50% |
| | Default standards | Healthcheck, metrics, logging, testing by default |
| **Maya (SRE)** | Observability coverage | 100% service baru |
| | Incident reduction | ↓ X% dari missing metrics/trace/log |
| | Template dashboard/alert | 1 set untuk semua service |
| **Dina (New Joiner)** | Onboarding to first PR | ≤ 1 minggu |
| | Self-sufficiency | Endpoint + test tanpa blocking senior |

---

### Business Objectives

| Timeline | Objective | Target |
|----------|-----------|--------|
| **3 Months** | Pilot adoption | 2-3 service end-to-end |
| | Feedback collection | 3+ engineers feedback |
| | Baseline metrics | Establish time-to-first-service baseline |
| **12 Months** | Adoption rate | ≥ 70-80% service baru |
| | Standardization | 100% dengan observability + testing + struktur konsisten |
| | Reliability impact | ≥ 30% incident reduction (missing basics) |
| | Setup time reduction | N hari → N jam |
| | Onboarding efficiency | PR meaningful ≤ 2 minggu |

---

### Key Performance Indicators

#### Adoption & Usage

| KPI | Definition | Target |
|-----|------------|--------|
| Boilerplate Adoption Rate | % service baru dari boilerplate | 3mo: ≥40%, 12mo: ≥80% |
| Repeat Usage | # engineer pakai untuk ≥2 service | Mayoritas mid-senior dalam 12mo |

#### Delivery & DX

| KPI | Definition | Target |
|-----|------------|--------|
| Time-to-First-Working-Service | Clone → service jalan (healthz + metrics) | Median < 30 menit |
| Time-to-First-Feature | Project start → fitur siap di dev | Median ≤ 2 hari kerja |
| Setup Overhead Reduction | Jam setup vs sebelum boilerplate | ≥ 50-70% lebih kecil |

#### Reliability & Observability

| KPI | Definition | Target |
|-----|------------|--------|
| Observability Baseline Coverage | % service dengan healthz + readyz + metrics + OTEL | 100% |
| Incident Investigation Time | Rata-rata waktu root cause | ↓ ≥ 20-30% |
| Incidents from Missing Basics | Count incidents karena missing logs/metrics | → 0 |

#### Quality & Maintainability

| KPI | Definition | Target |
|-----|------------|--------|
| Test Coverage Minimal | Coverage module core | ≥ 60% + unit + integration |
| Lint & CI Compliance | % PR lulus CI first run | ≥ 80% |

---

## MVP Scope

### Core Features (18 Items)

#### A. Foundation & Runtime

| # | Feature | Description |
|---|---------|-------------|
| 1 | Entrypoint & Wiring | main.go dengan config, logger, DB init, HTTP server, graceful shutdown |
| 2 | HTTP Server (chi) | Router + middleware (request ID, logging, panic recovery, otelhttp) |
| 3 | Config System (koanf) | Load env/file → struct → validate → fail-fast |
| 4 | Database Layer | pgx + sqlc + golang-migrate |

#### B. Observability & Operability

| # | Feature | Description |
|---|---------|-------------|
| 5 | Structured Logging (zap) | JSON prod, pretty local, trace_id, request_id |
| 6 | Healthcheck Endpoints | /healthz (liveness), /readyz (readiness dengan dependency check) |
| 7 | Metrics (Prometheus) | /metrics endpoint, HTTP request count, latency, errors |
| 8 | Tracing Hooks (OTEL) | otelhttp middleware, tracer provider stub |

#### C. Developer Experience

| # | Feature | Description |
|---|---------|-------------|
| 9 | Makefile | dev, test, lint, migrate-up/down, gen |
| 10 | docker-compose | Postgres + optional Redis/Jaeger |
| 11 | Environment Template | .env.example dengan semua config keys |

#### D. Quality & Testing

| # | Feature | Description |
|---|---------|-------------|
| 12 | Linting Config | golangci-lint + CI sample |
| 13 | Testing Setup | testify + struktur unit/integration |

#### E. Sample Module (E2E Example)

| # | Feature | Description |
|---|---------|-------------|
| 14 | Example Domain | Note/Task entity → repo → usecase → handler → migration → tests |

#### F. Documentation & AI

| # | Feature | Description |
|---|---------|-------------|
| 15 | README.md | Getting started, stack overview, feature list |
| 16 | ARCHITECTURE.md | Three Pillars, layering, data flow |
| 17 | AGENTS.md | Stack, struktur, test strategy, DO/DON'T untuk AI |
| 18 | TESTING.md | Guidelines per layer, AAA pattern |

---

### Out of Scope for MVP

| Item | Reason |
|------|--------|
| gRPC & GraphQL | Future extension, dokumentasi saja |
| Multiple Database | Fokus Postgres only |
| Advanced Auth/RBAC | Hook/interface saja, tanpa implementation |
| Rate Limiting/Circuit Breaker | Interface stub, integrasi later |
| Feature Flags | Future pluggable |
| Admin UI/Dashboard | Backend only |
| Multi-tenant/Sharding | v2+ patterns |

---

### MVP Success Criteria

| Criteria | Target |
|----------|--------|
| Time-to-first-service | < 30 menit via `make dev` |
| Sample module works | CRUD endpoint + test passing |
| Observability active | healthz + readyz + metrics + trace hooks |
| Documentation complete | README + ARCHITECTURE + AGENTS readable |
| CI passing | lint + test dalam sample workflow |

---

### Future Vision (2-3 Years)

| Phase | Capability |
|-------|------------|
| **Ecosystem Mode** | CLI generator, optional modules (gRPC, GraphQL, multi-tenant) |
| **Platform Integration** | Scaffolder, CI/CD templates, dev portal integration |
| **Security Hardening** | Audit trail, compliance logging, Vault/Secret Manager |
| **Observability Suite** | Bundled Grafana dashboards, alerting rules |
| **Open-Source Community** | External adoption, Go service best practice baseline |

---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments: []
workflowType: 'product-brief'
lastStep: 5
project_name: 'golang-api-hexagonal'
user_name: 'Chat'
date: '2025-12-16'
---

# Product Brief: golang-api-hexagonal

**Date:** 2025-12-16
**Author:** Chat

---

## Executive Summary

**golang-api-hexagonal** adalah production-ready Go service boilerplate yang mengimplementasikan hexagonal architecture secara benar—bukan sekadar struktur folder, tetapi dengan boundary enforcement yang nyata melalui linting rules dan layering tests.

Boilerplate ini dirancang untuk menyelesaikan masalah klasik tim engineering: **setup repetitif**, **inkonsistensi antar service**, dan **production concerns yang selalu jadi afterthought**. Dengan philosophy "opinionated but reasonable defaults", developer bisa bootstrap service baru dalam hitungan menit dengan semua production concerns (observability, security, audit) sudah terpasang dari hari pertama.

Target utama adalah **tim engineering (startup hingga enterprise)** yang membangun multiple microservices dan membutuhkan standardisasi arsitektur, serta **developer individu** yang membutuhkan baseline production-ready tanpa debat ulang di setiap proyek baru.

---

## Core Vision

### Problem Statement

Tim engineering dan developer Go menghadapi masalah berulang setiap memulai service baru:

1. **Setup repetitif dan memakan waktu**: config, logging, database, migration, healthcheck, graceful shutdown, docker compose, dan CI harus di-setup ulang setiap kali.

2. **Inkonsistensi antar service**: struktur folder, error style, naming convention, dan cara handling context/timeout berbeda-beda—membuat onboarding developer baru dan maintenance menjadi mahal.

3. **Production concerns yang selalu tertunda**: observability (tracing, metrics, logging), security headers, rate limiting, dan audit trail sering ditambahkan belakangan sebagai technical debt.

4. **Debat arsitektur yang berulang**: setiap service baru memicu diskusi "pakai apa" alih-alih fokus pada delivery.

### Problem Impact

Tanpa penyelesaian, masalah ini menyebabkan:

- **Production incidents**: missing timeouts, poor graceful shutdown, tidak ada readiness/liveness probes, log yang tidak bisa di-trace
- **Security hygiene lemah**: input validation, secrets handling, dan authorization checks tersebar inkonsisten
- **Onboarding lambat**: setiap service berbeda strukturnya, developer baru butuh waktu lama memahami "cara kerja" tiap repository
- **Velocity menurun**: waktu terbuang untuk setup dan debugging masalah yang seharusnya sudah solved

### Why Existing Solutions Fall Short

Solusi yang ada di ekosistem Go memiliki gap signifikan:

| Solusi | Kekurangan |
|--------|------------|
| **Copy-paste service lama** | Membawa technical debt, inkonsisten, memerlukan cleanup |
| **Mulai dari scratch** | Repetitif, hasil variatif antar developer |
| **Template internal** | Sering outdated, tidak mengikuti praktik terbaru |
| **go-kit, Uber Fx** | Terlalu heavy, learning curve tinggi, magic DI |
| **Gin/Echo/Fiber scaffolds** | Terlalu minimal, hanya router—kosong untuk production |
| **"Clean architecture" boilerplates** | Sekadar folder structure, boundary sering dilanggar (domain import infra) |
| **Berbagai GitHub boilerplates** | DX & dokumentasi kurang jelas, observability/security tidak lengkap |

### Proposed Solution

**golang-api-hexagonal** menyediakan:

**1. True Hexagonal Architecture**
- `domain`: entity, value object, port interfaces (repo, clock, idgen, audit sink)
- `app`: usecase orchestration, authorization policy, transaction management
- `transport`: HTTP/gRPC mapping, request/response, middleware
- `infra`: port implementations (DB, Redis, MQ, OpenTelemetry, logger)

**2. Production-Ready Day 1**
- **Observability**: OpenTelemetry (tracing + metrics), Prometheus endpoint, structured JSON logging
- **Security**: request limits, timeouts, secure headers, CORS, rate limiting, input validation, JWT/OIDC auth middleware
- **Audit Trail**: structured audit events dengan PII redaction, DB sink + optional queue publish

**3. Developer Experience**
- `git clone` → `make setup` → `make run`: service jalan lokal dalam menit
- `make test` / `make lint` / `make ci`: satu pintu quality gates
- Dokumentasi per concern: config, DB, auth, observability, deployment
- Contoh modul end-to-end (`users`) sebagai reference implementation

**4. Boundary Enforcement**
- Import rules dan linting untuk menjaga hexagonal boundaries
- Layering tests yang memvalidasi arsitektur tetap konsisten
- CI checks sebagai guardrail saat tim berkembang

**5. Testing Strategy Built-in**
- Unit tests (domain/app) yang cepat
- Integration tests dengan container/compose
- Contract/E2E minimal (health + satu flow API)

### Key Differentiators

| Differentiator | Penjelasan |
|----------------|------------|
| **Real Hexagonal** | Bukan sekadar folder structure—ada enforcement via linting + layering tests |
| **Opinionated but Reasonable** | Minimal pilihan, semua production concerns sudah ada dengan sensible defaults |
| **DX-First Documentation** | Preskriptif: "jalan lokal dalam menit", "cara tambah modul", "cara buat adapter baru" |
| **Guardrails for Scale** | Import rules + CI checks menjaga konsistensi saat tim dan service bertambah |
| **International Standards** | 12-factor app, cloud-native ready, OWASP hygiene, CI quality gates |

**Timing**: Go sudah mainstream untuk high-performance backend di enterprise. Cloud-native/Kubernetes adoption menuntut baseline konsisten. Tim engineering makin butuh standardisasi karena jumlah service dan engineer bertumbuh cepat.

---

## Target Users

### Primary Users

#### 1. Tech Lead / Senior Engineer di Startup/Scale-up — "Andi"

**Profil:**
- Tech Lead atau Senior Backend Engineer di startup/scale-up
- Mengelola 5–30 microservices dengan tim 5–30 engineer
- Growth cepat, sering membuat service baru
- Infrastruktur standar: PostgreSQL, Redis, Kubernetes/ECS

**Frustrasi Saat Ini:**
- PR review berantakan karena coding style berbeda-beda antar service
- Observability bolong—incident sulit di-trace
- Timeout dan context handling inconsistent
- Onboarding developer baru memakan waktu lama
- Copy-paste template lama membawa technical debt

**Trigger Adopsi:**
- `make run` langsung jalan tanpa setup panjang
- Contoh modul end-to-end sebagai reference
- Aturan boundary yang membuat code review lebih mudah
- OpenTelemetry, logging, dan metrics sudah terpasang

**User Journey:**
1. **Discovery**: Googling "go clean architecture boilerplate" atau "hexagonal go template", menemukan repo dengan README kuat
2. **Evaluation**: Cek README (how-to-run), struktur folder, contoh modul, CI badge, release tags, license, issue activity
3. **Onboarding (30 menit)**:
   - `git clone` → `make setup`
   - `make run` (compose + migrate)
   - Hit `GET /health` + `POST /users`
   - Baca "How to add module" (5–10 menit)
4. **Aha! Moment**: Tracing/logging/metrics dengan correlation ID tanpa setup ribet + boundary checks bikin PR review lebih ringan
5. **Adoption**: Buat pilot service kecil, tulis "service standard" doc 1 halaman, enforce via PR checklist

---

#### 2. Platform/Architecture Engineer di Enterprise — "Budi"

**Profil:**
- Platform Engineer atau Principal Engineer di enterprise
- Bertanggung jawab atas standardisasi arsitektur lintas tim
- Harus navigate governance: Architecture Board, Security review, Compliance

**Concern Tambahan:**
- Audit trail formal untuk compliance
- Access control dan authorization yang konsisten
- Secrets management yang proper
- Logging policy dan retention sesuai regulasi
- Standar deployment dan threat modeling baseline
- SBOM dan vulnerability scanning untuk supply chain security

**Decision Maker:**
- Platform/Architecture Team + Security approval
- Final decision oleh Principal Engineer, Head of Engineering, atau EM

**Trigger Adopsi:**
- Guardrails di CI (lint rules, import rules, layering tests)
- Security baseline yang sudah default
- Audit events dengan format standar
- Dokumentasi runbook + compliance checklist

**User Journey:**
1. **Discovery**: Internal tech radar, benchmarking template, rekomendasi komunitas OSS
2. **Evaluation Checklist**:
   - Security baseline (authn/authz, validation, rate limit, headers)
   - Observability standar (OTel, metrics endpoint)
   - Audit trail model
   - Supply chain: license, pinned deps, govulncheck/SBOM, CI quality gates
   - Maintainability: boundaries, test strategy, upgrade path
3. **Internal Selling**:
   - Ke Security: Tunjukkan default controls, policy di usecase layer, audit events, secret handling
   - Ke Stakeholder: Hitung penghematan bootstrap time + onboarding time + konsistensi
4. **Rollout**:
   - Fork menjadi "company template"
   - Buat "golden path" dengan 1–2 reference services
   - Migrasi bertahap: service baru wajib pakai template, service lama adopt komponen dulu (logging/OTel/health)

---

#### 3. Solo Developer / Freelancer — "Citra"

**Profil:**
- Freelancer, side-project builder, atau OSS contributor
- Sering memulai repository baru
- Waktu terbatas, butuh cepat dari demo ke production
- Deploy sederhana: VM atau lightweight Kubernetes

**Frustrasi Saat Ini:**
- Setup panjang menghabiskan waktu yang bisa untuk coding
- Boilerplate yang ada terlalu minimal atau terlalu kompleks
- Harus rakit sendiri: router, config, DB, migrate, logging, health

**Yang Paling Valuable:**
- Scaffold yang clean dengan "baterai sudah terpasang"
- Dokumentasi singkat dan to-the-point
- Contoh test yang bisa di-copy

**User Journey:**
1. **Discovery**: Cari "production ready golang api template", lihat rekomendasi di GitHub/YouTube/blog
2. **Decision**: Pilih kalau "jalan langsung" + docs singkat + tidak terlalu heavy + ada contoh CRUD yang rapih
3. **First Success**: Dalam 15–30 menit sudah deploy MVP (health + auth + DB + migrate), saat incident kecil log + trace bikin debug cepat—"hemat waktu gue banget!"

---

### Secondary Users

| User | Benefit |
|------|---------|
| **DevOps/SRE** | Health/readiness probes, metrics, tracing, structured logs yang konsisten → monitoring dan incident response lebih mudah |
| **Engineering Manager** | Konsistensi lintas tim, onboarding lebih cepat, kualitas delivery meningkat |
| **Security Team** | Baseline security + audit trail standar → security review lebih cepat |
| **New Hire** | Belajar satu pola yang berlaku di semua services → produktif lebih cepat |

---

### User Journey Summary

```
Discovery → Evaluation → Onboarding → Aha! Moment → Adoption/Advocacy

Andi (Startup):     Google → README check → 30min setup → Tracing works! → Pilot service → Team standard
Budi (Enterprise):  Tech radar → Checklist eval → Board approval → Golden path → Gradual rollout
Citra (Solo):       Search → Quick eval → 15min MVP → Debug fast → Keep using
```

---

## Success Metrics

### User Success Metrics

#### Andi (Tech Lead Startup)

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Bootstrap Time** | < 30 menit | Service baru jalan lokal + CI + observability + security baseline |
| **First Deploy Ready** | < 1 hari kerja | Service siap untuk production deployment |
| **PR Review Efficiency** | Berkurang | Lebih sedikit komentar "kok beda?" karena struktur & error/log standar |
| **Basic Incidents** | Menurun | Incident terkait timeout, shutdown, missing health berkurang |

**Behavior Signals:**
- Repo dijadikan template internal tim
- Tim membuat service kedua/ketiga menggunakan template
- Muncul guideline "service standard" dari Tech Lead

---

#### Budi (Platform Engineer Enterprise)

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Compliance Baseline** | 100% checklist | Authn/authz placement jelas, audit events, secrets, vuln scanning |
| **Cross-team Adoption** | Multi-team | Template bisa dipakai konsisten lintas tim |
| **Bootstrap/Onboarding Time** | 2-3 hari → < 1 hari | Pengurangan waktu setup untuk developer baru |
| **Implementation Variance** | Menurun | Lebih sedikit defect pattern berulang |
| **Review Cycle Time** | Lebih cepat | Audit/security review menjadi lebih singkat |

**Architecture Board Approval Criteria:**
- CI gates + boundary enforcement guardrails aktif
- Dokumentasi governance (how to extend, upgrade policy)
- Reference service yang lulus security review

---

#### Citra (Solo Dev)

| Metric | Target | Measurement |
|--------|--------|-------------|
| **MVP Deploy Time** | < 1-2 jam | DB + migrate + auth + health ready untuk deploy |
| **Debug Efficiency** | Lebih cepat | Log/trace siap pakai untuk troubleshooting |

**Reuse Signals:**
- Menggunakan lagi untuk proyek berikutnya
- Membuat starter repo pribadi berbasis template
- Merekomendasikan ke teman/tim lain

---

### Business Objectives

#### Primary Objective
**Menjadi "Golden Path" Go Microservice yang Kredibel**

Membangun reputasi sebagai standard de-facto untuk Go microservice architecture yang:
- Trusted oleh engineering community
- Diakui sebagai best practice implementation
- Menjadi reference untuk platform engineering

#### Secondary Objective (Opsional)
**Foundation untuk Enablement**

Potensi sebagai foundation untuk:
- Training dan workshop
- Consulting engagement
- Internal enablement di organisasi
- *(Bukan fokus monetisasi pada tahap awal)*

---

### Key Performance Indicators

#### Adoption KPIs

| KPI | Description | Target |
|-----|-------------|--------|
| **GitHub Stars** | Community interest indicator | Growth trend positif |
| **Forks** | Active derivation count | Menunjukkan real usage |
| **Template Usage** | "Use this template" clicks | Direct adoption metric |
| **Derivative Repos** | Aktif repos yang berbasis template | Real-world implementation |

#### Engagement KPIs

| KPI | Description | Target |
|-----|-------------|--------|
| **Quality Issues/PRs** | Kontribusi berkualitas | Active community participation |
| **Architecture Discussions** | Diskusi mendalam tentang design | Engaged expert users |
| **Docs/Examples Contributions** | Community-driven improvements | Self-sustaining ecosystem |

#### Real-world Usage KPIs

| KPI | Description | Target |
|-----|-------------|--------|
| **Companies/Teams Using** | Daftar adopter yang visible | Growing adopter list |
| **Testimonials** | Feedback dari real users | Credibility proof |
| **Case Studies** | Documented success stories | Adoption evidence |

#### Quality KPIs

| KPI | Description | Target |
|-----|-------------|--------|
| **CI Status** | Build pipeline health | Always green |
| **Test Coverage** | Domain/app layer coverage | High coverage (>80%) |
| **Security Scanning** | govulncheck results | Clean/no critical |
| **Dependency Updates** | Dependabot/Renovate | Regular & timely |

#### Documentation KPIs

| KPI | Description | Target |
|-----|-------------|--------|
| **Time-to-First-Success** | Waktu user pertama kali berhasil run | < 30 menit tanpa bantuan |
| **Repeated Questions** | FAQ yang sama muncul berulang | Menurun (docs improvement indicator) |
| **Self-service Rate** | User bisa jalan tanpa tanya | Tinggi |

---

## MVP Scope

### Core Features

#### A. Hexagonal Architecture Foundation

| Feature | Description | Priority |
|---------|-------------|----------|
| **Folder Structure** | `domain`, `app`, `transport`, `infra` dengan boundary rules | Must Have |
| **Port Interfaces** | Repository, clock, idgen, audit sink interfaces di domain | Must Have |
| **Reference Module** | `users` module end-to-end (create + get + list) sebagai contoh implementasi lengkap | Must Have |
| **Import Rules** | Linting rules untuk enforce hexagonal boundaries | Must Have |

#### B. Production Concerns

| Feature | Description | Priority |
|---------|-------------|----------|
| **Config Management** | Environment-based configuration dengan sensible defaults | Must Have |
| **Structured Logging** | JSON format, leveled, dengan context propagation | Must Have |
| **OpenTelemetry Tracing** | Distributed tracing dengan trace/span propagation | Must Have |
| **Metrics Endpoint** | Prometheus-compatible `/metrics` endpoint dengan contoh metrics | Must Have |
| **Health Endpoints** | `/health` (liveness) + `/ready` (readiness) | Must Have |
| **Graceful Shutdown** | Proper signal handling dan connection draining | Must Have |

#### C. Security Baseline

| Feature | Description | Priority |
|---------|-------------|----------|
| **Request Validation** | Schema-based input validation | Must Have |
| **Secure Headers** | Security headers (HSTS, X-Frame-Options, etc.) | Must Have |
| **Rate Limiting** | Configurable rate limit middleware | Must Have |
| **Auth Middleware** | JWT verification + context user extraction (OIDC-ready) | Must Have |
| **Audit Trail** | Structured audit events dengan DB sink | Must Have |

#### D. Developer Experience

| Feature | Description | Priority |
|---------|-------------|----------|
| **Make Targets** | `make setup`, `make run`, `make test`, `make lint`, `make ci` | Must Have |
| **Docker Compose** | PostgreSQL + service untuk local development | Must Have |
| **Migrations** | Database migration setup dengan tooling | Must Have |
| **CI Pipeline** | Lint + test + build + import rules check | Must Have |
| **Documentation** | README + "How to add module" + "Local dev" + "Observability" guides | Must Have |

#### E. Cross-cutting (Tambahan)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Standard Error Response** | Consistent error format across all endpoints | Must Have |
| **Request ID Middleware** | Correlation ID generation dan propagation | Must Have |

---

### Out of Scope for MVP

| Feature | Rationale | Target Version |
|---------|-----------|----------------|
| **CLI Generator** | Fokus template dulu, generator setelah pola stabil | v2+ |
| **gRPC Transport** | HTTP solid dulu, gRPC setelah modul contoh matang | v2 |
| **Layering Tests** | Import rules cukup untuk MVP, layering tests setelah struktur stabil | v2 |
| **Redis/Cache Adapter** | Fokus DB dulu, cache setelah core solid | v2 |
| **Advanced Audit** | Event sourcing terlalu kompleks untuk MVP | v3+ |
| **Multi-tenancy** | Enterprise feature, bukan core value | v3+ |
| **Message Queue Adapter** | Kafka/RabbitMQ untuk advanced use cases | v3+ |
| **WebSocket Support** | Real-time bukan core use case | v3+ |
| **GraphQL** | Alternative transport, bukan priority | v3+ |
| **Complex API Versioning** | Simple versioning cukup untuk MVP | v3+ |

---

### MVP Success Criteria

| Criteria | Measurement | Target |
|----------|-------------|--------|
| **Bootstrap Time** | `git clone` → service running lokal | < 15 menit |
| **First API Call** | Health + users endpoint working | < 30 menit dari clone |
| **CI Green** | Lint + test + build passing | 100% |
| **Boundary Enforcement** | Import rules catching violations | Aktif di CI |
| **Documentation Complete** | User bisa run tanpa tanya | Self-service rate tinggi |
| **Observability Working** | Tracing + logging + metrics visible | Functional out-of-box |

**Go/No-Go for v2:**
- ✅ MVP features complete dan documented
- ✅ 5+ external users/stars menunjukkan interest
- ✅ Feedback dari real usage (issues, discussions)
- ✅ Core patterns validated dan stable

---

### Future Vision

#### Version 2.0 — "Production Hardened"

| Feature | Description |
|---------|-------------|
| **Layering Tests** | Automated architecture validation tests |
| **Redis Adapter** | Cache adapter dengan contoh implementation |
| **gRPC Transport** | Full gRPC support dengan protobuf |
| **Enhanced Metrics** | More comprehensive OTel metrics |
| **Performance Benchmarks** | Documented performance characteristics |

#### Version 3.0 — "Enterprise Ready"

| Feature | Description |
|---------|-------------|
| **Message Queue Adapter** | Kafka/RabbitMQ/NATS support |
| **Template Variants** | Cron/worker template, event-driven template |
| **Enterprise Hardening** | Policy packs, SBOM automation, compliance docs |
| **Advanced Audit** | Event sourcing option, SIEM integration |
| **Multi-tenancy** | Tenant isolation patterns |

#### Long-term Vision

```
MVP (v1)           →  v2 (Hardened)      →  v3 (Enterprise)       →  Ecosystem
─────────────────────────────────────────────────────────────────────────────
HTTP + Postgres       + gRPC + Redis        + MQ + Variants          + CLI
Basic OTel            + Layering Tests      + Policy Packs           + Templates
Import Rules          + Benchmarks          + Compliance             + Community
```

---

<!-- Content will be appended sequentially through collaborative workflow steps -->

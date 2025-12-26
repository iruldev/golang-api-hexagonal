---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
status: 'approved'
inputDocuments:
  - "_bmad-output/research/technical-production-boilerplate-research-2024-12-24.md"
  - "_bmad-output/index.md"
  - "_bmad-output/architecture.md"
  - "_bmad-output/api-contracts.md"
  - "_bmad-output/data-models.md"
  - "_bmad-output/source-tree-analysis.md"
  - "_bmad-output/project-overview.md"
  - "_bmad-output/development-guide.md"
  - "_bmad-output/project-scan-report.json"
documentCounts:
  briefs: 0
  research: 1
  brainstorming: 0
  projectDocs: 8
workflowType: 'prd'
lastStep: 11
project_name: 'golang-api-hexagonal'
user_name: 'Gan'
date: '2024-12-24'
---


---

# Product Requirements Document - golang-api-hexagonal

**Author:** Gan  
**Date:** 2024-12-24  
**Status:** Draft  

---

## Executive Summary

**golang-api-hexagonal** adalah production-ready Go API boilerplate berbasis **hexagonal architecture** yang akan di-upgrade menjadi **international production-grade backend boilerplate** untuk tim internal.

### Vision

Menyediakan standardized, secure-by-default, observable, dan maintainable backend boilerplate yang dapat digunakan tim internal untuk membangun service baru dengan standar internasional — tanpa harus memikirkan ulang security, observability, atau DI patterns dari nol.

### Problem Statement

- Existing codebase memiliki beberapa **correctness issues** (P0 bugs) yang perlu diperbaiki
- Security patterns tidak konsisten (authz scattered, JWT validation belum lengkap)
- Manual dependency wiring di main.go yang semakin kompleks
- Belum ada standardized secret management pattern
- Developer onboarding memerlukan banyak tribal knowledge

### Target Users

- **Primary:** Tim internal sebagai standardized boilerplate untuk service baru
- **Secondary:** Potensial internal enterprise template sharing

### What Makes This Special

1. **Secure-by-Default:** No-auth guard, konsisten AuthZ di app layer, PII redaction built-in
2. **Platform-Agnostic Secrets:** `*_FILE` pattern untuk Vault/Cloud SM tanpa vendor lock-in
3. **Developer Experience:** Wire DI (compile-time), sqlc (typed SQL), tight CI gating
4. **Brownfield Ready:** Complete docs untuk AI-assisted development, ADRs, runbooks

---

## Project Classification

| Field | Value |
|-------|-------|
| **Technical Type** | API Backend (Hexagonal Architecture) |
| **Domain** | Developer Tooling / Infrastructure |
| **Complexity** | Medium-High |
| **Project Context** | Brownfield - extending existing production-ready boilerplate |

### Technology Stack (Existing)

| Component | Technology |
|-----------|------------|
| Language | Go 1.24.11 |
| HTTP Framework | Chi v5.2.3 |
| Database | PostgreSQL (pgx v5.7.6) |
| Migrations | Goose v3.26.0 |
| Observability | OpenTelemetry + Prometheus |
| Authentication | JWT (HS256, optional) |

### Upgrade Scope

| Priority | Area | Key Items |
|----------|------|-----------|
| **P0** | Bug Fixes | Config usage, pool lifecycle, audit metadata, UUID v7 errors |
| **P1** | Security | No-auth guard, AuthZ consistency, JWT validation, internal endpoints |
| **P1** | API Contract | OpenAPI spec, JSON decoding hardening, response headers |
| **P1** | Reliability | Health check semantics, HTTP hardening, graceful shutdown |
| **P2** | Secrets | `*_FILE` pattern, Vault/Cloud SM readiness |
| **P2** | DI | Google Wire migration, lifecycle management |
| **P2** | Data Layer | sqlc adoption, email CITEXT, pool tuning |
| **P3** | DX/Governance | Toolchain pinning, lint expansion, SECURITY.md, ADRs |

### Non-Goals

- Multi-tenancy (design note only, tidak implementasi)
- New business modules (tetap Users + Audit)
- Major breaking changes (prefer versioned approach)

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **DI Framework** | Google Wire | Compile-time, deterministic, minimal magic |
| **Secret Management** | Platform injection + `*_FILE` | Portable, Vault-unaware app |
| **SQL Layer** | sqlc (infra-only) | Typed SQL, domain isolation maintained |
| **Breaking Changes** | Avoided | Versioned approach (v2) if needed |

---

## Success Criteria

### User Success (Developer-focused)

| Criterion | Target | Source |
|-----------|--------|--------|
| Setup Time | < 30 menit | Initial |
| First Endpoint | < 2 jam dengan guide | Initial |
| Local Dev Parity | `make up && make migrate && make test` sukses tanpa manual steps | User |
| Generator Hygiene | `make generate` → `git diff --exit-code` | User |
| Zero-friction Setup | Fresh clone `make setup && make test` passes | Pre-mortem |
| Quick Start | README ≤ 5 commands to running server | Pre-mortem |
| Breaking Changes | Documented in CHANGELOG with migration guide | Pre-mortem |

### Business Success

| Criterion | Target | Source |
|-----------|--------|--------|
| Adoption | ≥ 1 production service dalam 3 bulan | Initial |
| Reuse | ≥ 2 service template usage dalam 6 bulan | User (stretch) |
| Incident Readiness | Runbook + oncall checklist (DB, OTEL, auth, rate-limit) | User |

### Technical Success

| Criterion | Target | Source |
|-----------|--------|--------|
| Coverage | Domain + App ≥ 80% | Initial |
| Integration Tests | CI job `go test -tags=integration` dengan DB `_test` | User |
| CI Gates | lint, test, vuln, secret scan + generate check | User |
| CI Performance | Full CI ≤ 10 minutes | Pre-mortem |
| Local Parity | `make ci` identical to GitHub Actions | Pre-mortem |
| Toolchain Pinning | `go.mod` toolchain directive, tools versioned | Pre-mortem |

### Security Success

| Criterion | Target | Source |
|-----------|--------|--------|
| No-auth Guard | `ENV=production` memaksa JWT enabled | User |
| No-auth CI Gate | CI fails if prod + JWT disabled | Pre-mortem |
| JWT Validation | require `exp` + validate `iss`/`aud` + clock skew config | User |
| Algorithm Whitelist | JWT parser only accepts configured algorithm | Red Team |
| Secret Non-Exposure | JWT_SECRET never in logs/errors | Red Team |
| Min Key Length | JWT_SECRET ≥ 32 bytes enforced | Red Team |
| /metrics Protection | Not publicly accessible (internal port/allowlist/auth) | User |
| Metrics Content Audit | No sensitive labels in /metrics | Red Team |
| TRUST_PROXY | RealIP only active if `TRUST_PROXY=true` | User |
| Rate Limit Test | Integration test verifies IP extraction correctness | Red Team |
| Error Audit | RFC7807 errors no raw traces/PII | Pre-mortem |
| Generic Error 500 | Internal errors → generic message | Red Team |
| Constant-Time Auth | Auth failures consistent timing (prevent enumeration) | Red Team |
| IDOR Test Coverage | Integration tests verify users cannot access other users' data | Red Team |

### Observability Success

| Criterion | Target | Source |
|-----------|--------|--------|
| Correlation | `request_id` + `trace_id` di logs, errors, audit | User |
| Cardinality Safe | Route label pakai pattern (no `{id}`) | User |
| Correlation Audit | Manual test: verify requestId matches across layers | Pre-mortem |

### Reliability Success

| Criterion | Target | Source |
|-----------|--------|--------|
| Readiness Safe | Tidak mutate runtime state (no close/reset DB pool) | User |
| HTTP Hardening | `ReadHeaderTimeout` + `MaxHeaderBytes` configured | User |
| Generate Idempotency | CI verifies no diff after `make generate` | Pre-mortem |

### Architecture Success

| Criterion | Target | Source |
|-----------|--------|--------|
| depguard | violations = 0 | Initial |
| No Infra Leakage | domain/app tidak import sqlc/wire/fx/vault SDK | User |

---

## Product Scope

### MVP - Phase 1 (P0/P1)

**Focus:** Correctness + Security + Reliability

- All P0 bugs fixed (config, pool, metadata, UUID)
- Security baseline complete (AuthZ, JWT validation, no-auth guard)
- API contract discipline (JSON decoding, response headers)
- Reliability hardening (HTTP server, graceful shutdown, health checks)
- CI gates expanded (generate check, vuln scan)

### Growth - Phase 2 (P2)

**Focus:** Platform Readiness + Developer Velocity

- Secret management (`*_FILE` pattern, documented Vault integration)
- DI migration (Google Wire, lifecycle cleanup)
- Data layer modernization (sqlc, CITEXT, pool tuning)
- Integration test suite with DB `_test`

### Vision - Phase 3 (P3)

**Focus:** Governance + Long-term Maintainability

- DX & Governance (SECURITY.md, CONTRIBUTING.md, ADRs)
- OpenAPI spec generation
- Expanded lint suite (gosec, revive, errorlint)
- Runbook and oncall documentation

---

## User Journeys

### Journey 1: Andi Pratama - Building a New Service in Record Time

**Persona:** Backend Developer, 3 years experience, new to the team

Andi baru bergabung dengan tim dan ditugaskan untuk membangun Order Service baru. Sebelumnya di tim lama, setiap service baru membutuhkan 2-3 minggu hanya untuk setup boilerplate.

Pagi pertama, tech lead menunjukkan golang-api-hexagonal repo. Ketika Andi menjalankan `make setup && make up && make migrate && make test`, semuanya hijau dalam 20 menit. Tidak ada error aneh, tidak ada manual steps tersembunyi.

Mengikuti `docs/guides/adding-module.md`, Andi berhasil menambahkan Order entity, repository, dan use case dalam 2 jam. Ketika dia push PR, CI gates menangkap semua issues early. Seminggu kemudian, Order Service sudah di production.

**Requirements Revealed:** Zero-friction setup, comprehensive guides, pre-configured observability, CI gates

---

### Journey 2: Dewi Lestari - Debugging a Production Incident at 3 AM

**Persona:** Senior DevOps/SRE, 5 years experience, on-call

Dewi terbangun oleh PagerDuty alert: "Order Service - High Error Rate." Di log explorer, dia filter by `service=order-service` dan langsung menemukan pattern error dengan `request_id: abc-123-xyz`.

Dengan request_id itu, dia trace di Jaeger dan melihat span yang gagal di database connection. Memeriksa runbook di `docs/runbook.md`, ada section "Database Connection Issues" dengan langkah-langkah: check pool metrics, verify DATABASE_URL, restart pods jika perlu.

Karena boilerplate sudah menggunakan proper health check semantics (readiness yang tidak mutate state), Kubernetes sudah otomatis menghentikan traffic ke pods yang unhealthy. Sistem recover dalam 15 menit.

**Requirements Revealed:** Request/trace correlation, runbook documentation, health check semantics, consistent observability

---

### Journey 3: Budi Santoso - Security Review Before Major Release

**Persona:** Security Engineer, responsible for security reviews

Budi menerima request untuk security review Order Service. Karena dibangun dari golang-api-hexagonal, Budi sudah tahu baseline security-nya:
- Security headers (CSP, HSTS, X-Frame-Options) ✓
- JWT validation dengan algorithm whitelist ✓
- No-auth guard memastikan auth tidak bisa disabled di production ✓
- /metrics di internal port ✓

Budi hanya perlu fokus review business logic security. Review approved dalam 1 hari instead of 3.

**Requirements Revealed:** Documented security baseline, SECURITY.md, consistent security patterns

---

### Journey 4: Rina Maharani - External Penetration Testing

**Persona:** External Penetration Tester, annual security assessment

Rina mulai scanning dan menemukan pattern yang konsisten:
- Semua service menolak JWT dengan `alg: none` ✓
- Error messages generic, tidak bocorkan stack traces ✓
- Rate limiting mencegah brute force ✓

Attack vectors yang dicoba:
- JWT forgery → blocked by algorithm whitelist
- X-Forwarded-For spoofing → TRUST_PROXY=false by default
- IDOR attempts → consistent authz middleware blocking

Final report: "No critical or high findings. Baseline security posture excellent."

**Requirements Revealed:** Attack vector mitigations, consistent security posture, security by default

---

### Journey 5: Fajar Nugroho - New Team Member Onboarding

**Persona:** Fresh Graduate, first job as backend developer

Fajar baru lulus kuliah. Di hari pertama, mengikuti README:
1. `make setup` - tools terinstall otomatis ✓
2. `make up && make migrate` - database ready ✓
3. `make test` - semua hijau ✓
4. `make run` dan hit `/health` - working! ✓

Dokumentasi membantu: `docs/architecture.md` menjelaskan layer structure, `docs/guides/adding-module.md` menunjukkan step-by-step. Minggu kedua, Fajar sudah submit PR pertama.

**Requirements Revealed:** Self-documenting codebase, clear onboarding path, example modules, no tribal knowledge

---

### Journey 6: Sari Wijayanti - Platform Engineer Maintaining Standards

**Persona:** Platform/Infrastructure Engineer, maintainer of internal templates

Sari bertanggung jawab memastikan semua services mengikuti company standards. Dengan golang-api-hexagonal sebagai single source of truth, Sari bisa:
1. Update security baseline di satu tempat
2. Add new lint rules yang otomatis propagate
3. Track adoption across teams

Sari maintain ADRs di `docs/adr/`. Ketika ada pertanyaan "kenapa pakai Wire bukan Fx?", jawabannya terdokumentasi dengan rationale.

**Requirements Revealed:** Single source of truth, versioned boilerplate, ADRs, easy upgrade path

---

### Journey 7: Hendra - Compliance Audit Evidence Collection

**Persona:** Internal Audit/Compliance Officer

Hendra mengumpulkan evidence untuk SOC 2 audit. Dengan golang-api-hexagonal sebagai standard:
- **Access Controls**: JWT auth dengan documented claims validation ✓
- **Audit Trail**: Semua write operations logged di audit_events table ✓
- **Change Management**: CI/CD gates, PR required ✓
- **Monitoring**: Prometheus metrics, OTEL tracing ✓

Evidence dari single source: `docs/SECURITY.md`, `docs/architecture.md`, `docs/runbook.md`, CI/CD logs.

**Requirements Revealed:** SECURITY.md dengan control descriptions, audit trail, runbooks, evidence-ready documentation

---

### Journey Requirements Summary

| Capability Area | Required By |
|-----------------|-------------|
| Zero-friction Setup | Developer, New Member |
| Comprehensive Guides | Developer, New Member |
| Pre-configured Observability | Developer, DevOps |
| Request/Trace Correlation | DevOps |
| Runbook Documentation | DevOps, Compliance |
| Security Baseline (SECURITY.md) | Security Reviewer, Pen Tester, Compliance |
| Consistent Patterns | All users |
| ADRs | Platform Engineer |
| Audit Trail | Compliance |
| Self-documenting Code | New Member |

---

## API Backend Specific Requirements

### Endpoint Specifications

| Endpoint | Method | Auth | Purpose | Upgrade Notes |
|----------|--------|------|---------|---------------|
| `/health` | GET | No | Liveness probe | No changes |
| `/ready` | GET | No | Readiness probe | Fix: no pool mutation |
| `/metrics` | GET | Internal | Prometheus | Move to internal port |
| `/api/v1/users` | GET | JWT | List users | Add IDOR tests |
| `/api/v1/users` | POST | JWT | Create user | Consistent AuthZ |
| `/api/v1/users/{id}` | GET | JWT | Get user | Add IDOR tests |

**Future Additions:**
- OpenAPI 3.1 spec auto-generation
- Contract tests based on OpenAPI spec

### Authentication Model

| Aspect | Current | Upgrade Target |
|--------|---------|----------------|
| Algorithm | HS256 only | Whitelist configured algorithm |
| Claims | Basic | Require `exp`, validate `iss`/`aud` |
| Guard | Optional | Fail-closed in production (`ENV=production` → JWT required) |
| Key Length | Any | Minimum 32 bytes enforced |
| Rotation | Manual | Document RS256/JWKS migration path |

### Data & Error Handling

| Aspect | Current | Upgrade Target |
|--------|---------|----------------|
| Request Parsing | Lenient JSON | Strict: reject unknown fields, no trailing data |
| Error Format | RFC 7807 | Add `request_id`, `trace_id` extensions |
| Internal Errors | May leak details | Generic message only, details in logs |
| Validation | Mixed | Consistent validator with structured errors |

### Rate Limiting

| Aspect | Current | Upgrade Target |
|--------|---------|----------------|
| Strategy | IP-based | Fix TRUST_PROXY extraction |
| Config | `RATE_LIMIT_RPS` | Add integration test for correctness |
| Bypass Prevention | Partial | RealIP only if `TRUST_PROXY=true` |
| Headers | Not exposed | Add `X-RateLimit-*` headers |

### API Documentation

| Item | Status | Priority |
|------|--------|----------|
| README API section | ✅ Exists | Maintain |
| OpenAPI 3.1 spec | 🆕 Add | P3 |
| Postman/Insomnia collection | Optional | Nice-to-have |
| Contract tests | 🆕 Add with OpenAPI | P3 |

---

## Functional Requirements

### 1. Correctness & Bug Fixes

- **FR1:** System uses `AUDIT_REDACT_EMAIL` config correctly for PII redaction
- **FR2:** Database pool lifecycle manages reconnection properly
- **FR3:** Audit events include `ActorID` and `RequestID` metadata
- **FR4:** UUID v7 parsing errors handled gracefully (no 500)
- **FR5:** Schema migration (`schema_info`) is idempotent

### 2. Security & Authentication

- **FR6:** System validates JWT with configured algorithm whitelist only
- **FR7:** System requires `exp` claim and validates `iss`/`aud` claims
- **FR8:** System enforces JWT when `ENV=production` (no-auth guard)
- **FR9:** System rejects JWT secrets shorter than 32 bytes
- **FR10:** System protects `/metrics` from public access (internal port/auth)
- **FR11:** System extracts real IP only when `TRUST_PROXY=true`
- **FR12:** Users can only access their own data (IDOR prevention)
- **FR13:** System applies consistent authorization at application layer
- **FR41:** System never logs or returns JWT_SECRET / Authorization header / sensitive tokens
- **FR42:** System returns generic 500 errors (details in logs only)
- **FR43:** System validates/normalizes inbound `X-Request-ID` (length/charset)
- **FR44:** Auth checks are constant-time to prevent user enumeration

### 3. Observability & Correlation

- **FR14:** System injects `request_id` into all log entries
- **FR15:** System injects `trace_id` into all log entries
- **FR16:** Error responses include `request_id` extension (RFC7807)
- **FR17:** Audit events include `request_id` for correlation
- **FR18:** Metrics use route patterns (no `{id}` in labels for safe cardinality)
- **FR45:** Error responses include BOTH `request_id` and `trace_id` extensions
- **FR46:** `/metrics` content is audited (no sensitive labels/values)

### 4. API Contract & Reliability

- **FR19:** API rejects JSON with unknown fields
- **FR20:** API rejects JSON with trailing data
- **FR21:** Health endpoint (`/ready`) does not mutate runtime state
- **FR22:** System configures HTTP `ReadHeaderTimeout`
- **FR23:** System performs graceful shutdown with cleanup chain
- **FR24:** Server responds with `Location` header on 201 Created
- **FR47:** System produces OpenAPI 3.1 spec for all endpoints
- **FR48:** CI runs OpenAPI-based contract tests
- **FR51:** System sets `MaxHeaderBytes` and `ReadHeaderTimeout` via config

### 5. Rate Limiting & Networking

- **FR49:** Integration test verifies IP extraction logic (TRUST_PROXY scenarios)
- **FR50:** RealIP middleware is enabled only when `TRUST_PROXY=true`

### 6. Developer Experience

- **FR25:** Developer can setup local environment with `make setup`
- **FR26:** Developer can run all unit tests with `make test`
- **FR27:** Developer can run integration tests with test database
- **FR28:** Developer can generate code with `make generate`
- **FR29:** CI verifies generated code is up-to-date (no diff)
- **FR30:** CI runs lint, test, vuln scan, secret scan gates
- **FR52:** CI enforces wire/sqlc generate idempotency (`git diff --exit-code`)

### 7. Governance & Documentation

- **FR31:** Project has `SECURITY.md` with threat model
- **FR32:** Project has `CONTRIBUTING.md` guidelines
- **FR33:** Project has runbook for common operations
- **FR34:** Project has ADRs for key decisions
- **FR35:** README documents setup in ≤5 commands

### 8. Data Layer & Infrastructure

- **FR36:** Database pool is configurable via environment variables
- **FR37:** Email uniqueness is case-insensitive (CITEXT)
- **FR38:** sqlc generates typed queries for Users module
- **FR39:** sqlc generates typed queries for Audit module
- **FR40:** Wire generates dependency injection code

---

## Non-Functional Requirements

### Performance

- **NFR1:** API p99 response time < 100ms for non-DB endpoints; p99 < 300ms for DB-bound endpoints at baseline load
- **NFR2:** Health check `/health` responds < 10ms
- **NFR3:** CI full pipeline completes ≤ 10 minutes with parallel jobs (unit vs integration vs scans)
- **NFR20:** Service startup time < 10s on standard environment (without migration)
- **NFR21:** Memory footprint baseline < 200MB (or defined budget per infra)

### Security

- **NFR4:** TLS 1.2+ required if service terminates TLS; document ingress termination + mTLS internal option
- **NFR5:** No secrets in repository; support env + `*_FILE` pattern; secret source audited (Vault/Cloud/K8s)
- **NFR6:** No sensitive data in logs (PII redaction active)
- **NFR22:** JWT rotation without downtime (dual keys / JWKS caching policy)
- **NFR23:** `/metrics` restricted (internal port / network policy / minimal auth)
- **NFR24:** `Request-ID` bounded (max length/charset) to prevent log injection

### Reliability

- **NFR7:** Graceful shutdown completes within 30 seconds
- **NFR8:** Health checks distinguish liveness vs readiness
- **NFR9:** DB outage does not cause runtime mutation; readiness can fail but must not close/reset pool used by handlers
- **NFR25:** All dependencies have timeouts (DB/HTTP clients); no unbounded waits

### Maintainability

- **NFR10:** Domain + App test coverage ≥ 80%
- **NFR11:** golangci-lint clean with required linters (gosec, revive, errorlint, bodyclose) in CI
- **NFR12:** All dependencies scanned for known vulnerabilities
- **NFR13:** Generated code (wire/sqlc) reproducible and idempotent
- **NFR26:** No infra leakage: domain/app do not import wire/sqlc/vault SDK
- **NFR27:** ADR coverage for major decisions (DI, secrets, sqlc, auth model)

### Observability

- **NFR14:** All requests correlated via `request_id` + `trace_id` in logs + RFC7807 + audit events
- **NFR15:** Structured logging with consistent field names
- **NFR16:** OTel semantic conventions for traces and metrics
- **NFR28:** Metrics cardinality budget enforced (route label uses pattern; no user-provided labels)
- **NFR29:** Audit redaction policy configurable and tested (`AUDIT_REDACT_EMAIL` verified effective)

### Integration & Delivery

- **NFR17:** OpenTelemetry exporter configurable via environment
- **NFR18:** Prometheus metrics endpoint available
- **NFR19:** OIDC/JWKS rotation supported without restart (future)
- **NFR30:** `make ci` replicates GitHub Actions steps (local parity)
- **NFR31:** Supply-chain hygiene: SBOM generated + govulncheck/OSV scan in CI


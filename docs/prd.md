---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
inputDocuments:
  - "docs/analysis/product-brief-golang-api-hexagonal-2025-12-16.md"
documentCounts:
  briefs: 1
  research: 0
  brainstorming: 0
  projectDocs: 0
workflowType: 'prd'
lastStep: 11
project_name: 'golang-api-hexagonal'
user_name: 'Chat'
date: '2025-12-16'
---

# Product Requirements Document - golang-api-hexagonal

**Author:** Chat
**Date:** 2025-12-16

---

## Executive Summary

**golang-api-hexagonal** adalah production-ready Go service boilerplate yang mengimplementasikan hexagonal architecture secara benar—bukan sekadar struktur folder, tetapi dengan boundary enforcement melalui linting rules dan CI checks.

### Vision

Menyediakan "golden path" untuk tim engineering yang membangun Go microservices, dengan philosophy **"opinionated but reasonable defaults"**. Developer bisa bootstrap service baru dalam hitungan menit dengan semua production concerns (observability, security, audit) sudah terpasang dari hari pertama.

### Problem Space

Tim engineering dan developer Go menghadapi masalah berulang setiap memulai service baru:

1. **Setup repetitif dan memakan waktu** — config, logging, database, migration, healthcheck, graceful shutdown, docker compose, dan CI harus di-setup ulang setiap kali
2. **Inkonsistensi antar service** — struktur folder, error style, naming convention, dan cara handling context/timeout berbeda-beda, membuat onboarding dan maintenance mahal
3. **Production concerns yang selalu tertunda** — observability, security headers, rate limiting, dan audit trail sering ditambahkan belakangan sebagai technical debt
4. **Debat arsitektur yang berulang** — setiap service baru memicu diskusi "pakai apa" alih-alih fokus pada delivery

### Target Users

**Primary Users:**
- **Tech Lead / Senior Engineer di Startup/Scale-up** — Mengelola 5-30 services, butuh konsistensi dan cepat bootstrap
- **Platform/Architecture Engineer di Enterprise** — Butuh compliance checklist, security baseline, dan governance
- **Solo Developer / Freelancer** — Waktu terbatas, butuh scaffold dengan "baterai sudah terpasang"

**Secondary Users:**
- **SRE/DevOps** — Benefit dari health probes, metrics, tracing, structured logs yang konsisten
- **Security Reviewer** — Benefit dari baseline security + audit trail standar

### What Makes This Special

| Differentiator | Description |
|----------------|-------------|
| **Real Hexagonal** | Bukan sekadar folder structure—ada enforcement via linting + layering tests |
| **Opinionated but Reasonable** | Minimal pilihan, semua production concerns sudah ada dengan sensible defaults |
| **DX-First Documentation** | Preskriptif: "jalan lokal dalam menit", "cara tambah modul", "cara buat adapter baru" |
| **Guardrails for Scale** | Import rules + CI checks menjaga konsistensi saat tim dan service bertambah |
| **International Standards** | 12-factor app, cloud-native ready, OWASP hygiene, CI quality gates |

---

## Project Classification

| Aspect | Value |
|--------|-------|
| **Technical Type** | Developer Tool (Go service boilerplate/template) |
| **Domain** | General — Cloud-native Go service template |
| **Complexity** | Medium (security baseline, audit trail, OpenTelemetry, boundary enforcement) |
| **Project Context** | Greenfield - new project |

### Complexity Rationale

Complexity dinaikkan ke **medium** karena MVP scope mencakup:
- Security baseline (authn/authz, validation, rate limit, secure headers)
- Audit trail dengan structured events dan DB sink
- OpenTelemetry integration (tracing + metrics)
- Hexagonal boundary enforcement via linting dan CI checks

Ini menaikkan quality bar di atas "standard requirements" dan memerlukan careful design untuk semua concerns tersebut.

---

## Success Criteria

### User Success

#### Time-to-First-Success

| User | Target | Measurement |
|------|--------|-------------|
| **Andi (Tech Lead)** | < 30 menit | Clone → `POST /users` berhasil |
| **Budi (Enterprise)** | < 1 hari | Compliance checklist terpenuhi + Architecture Board review pass |
| **Citra (Solo Dev)** | < 30 menit | Clone → `POST /users` berhasil |

#### "Aha!" Moments

- **Observability Connected**: Trace + log + request-id nyambung (satu request kelihatan di log & trace) **tanpa konfigurasi manual**
- **Boundaries Work**: Import violation terdeteksi di CI sebelum merge
- **Production Ready**: Health/ready/metrics endpoints working out-of-box

#### Adoption Proof

| User | Adoption Signal |
|------|-----------------|
| **Andi** | Minimal 1 service pilot dipakai tim |
| **Budi** | Minimal 1 internal fork sebagai company template |
| **Citra** | Reuse untuk proyek kedua |

---

### Business Success

#### Primary Objective

**Menjadi "Golden Path" Go Microservice yang Kredibel**

| Timeframe | Success Indicator |
|-----------|-------------------|
| **3 months** | MVP complete, 5+ GitHub stars, 1+ external user feedback |
| **6 months** | 50+ stars, 5+ forks, 2+ real-world adoption testimonials |
| **12 months** | 200+ stars, recognized in Go community, company adoptions documented |

#### Secondary Objective

**Foundation untuk Enablement**
- Training/workshop materials derivable from docs
- Consulting inquiries based on template quality

---

### Technical Success

#### Quality Gates (Target v1.0)

| Gate | Requirement |
|------|-------------|
| **Tests** | `go test ./... -race` pass (minimal di CI untuk package core) |
| **Coverage** | `domain` + `app` ≥ 80% (overall boleh lebih rendah) |
| **Lint** | `golangci-lint` 0 error |
| **Security** | `govulncheck` clean (atau documented exception) |
| **Build** | `go build ./...` pass dan reproducible via Makefile |

#### Maintainability

| Aspect | Requirement |
|--------|-------------|
| **Boundary Enforcement** | Import rules aktif + minimal 1 test membuktikan pelanggaran terdeteksi |
| **Dependency Footprint** | Reasonable — tidak pakai DI framework berat (wire/fx opsional v2) |
| **Code Organization** | Clear separation: domain → app → transport → infra |

#### Performance Baseline (MVP)

| Aspect | Requirement |
|--------|-------------|
| **Timeouts** | Handler punya timeouts, request size limit |
| **Goroutine Safety** | Tidak ada goroutine leak (graceful shutdown + ctx cancel) |
| **Indicative Benchmark** | `GET /health` p95 < 10ms lokal (bukan SLA produksi) |

---

### Measurable Outcomes

#### Adoption KPIs

| KPI | Target (v1.0) | Target (6mo) |
|-----|---------------|--------------|
| **GitHub Stars** | 5+ | 50+ |
| **Forks** | 1+ | 5+ |
| **Template Usage** | Tracked | Growing |
| **Real Adoptions** | 1 testimonial | 2+ testimonials |

#### Quality KPIs

| KPI | Target |
|-----|--------|
| **CI Status** | Always green |
| **Test Coverage (domain/app)** | ≥ 80% |
| **Security Scanning** | Clean / documented exceptions |
| **Dependency Updates** | Regular via Dependabot/Renovate |

#### Documentation KPIs

| KPI | Target |
|-----|--------|
| **Time-to-First-Success** | < 30 menit tanpa bantuan |
| **Self-service Rate** | User bisa run tanpa tanya |
| **Docs Accuracy** | Steps match reality (no missing steps) |

---

## Product Scope

### MVP - Minimum Viable Product

#### Definition of Done (v1.0 Release Criteria)

**Happy Path End-to-End (Local):**
- [ ] `make setup && make run` jalan tanpa edit kode
- [ ] `GET /health` = 200
- [ ] `GET /ready` = 200 (setelah DB up)
- [ ] `POST /users` create → 201
- [ ] `GET /users/{id}` → 200
- [ ] `GET /users` → 200
- [ ] Migrations otomatis jalan (atau `make migrate up` yang jelas)

**Observability Works:**
- [ ] Log JSON keluar dengan `request_id` (dan `trace_id` jika tracing enabled)
- [ ] `/metrics` reachable dengan minimal 1 metric custom + default HTTP metrics
- [ ] Tracing menghasilkan span untuk request (export ke collector atau stdout exporter untuk dev)

**Security Baseline Works:**
- [ ] Request invalid → 400 dengan standard error response
- [ ] Rate limit trigger → 429
- [ ] Endpoint protected butuh JWT → 401/403 sesuai kasus
- [ ] Secure headers muncul di response

**Audit Works:**
- [ ] Create user menghasilkan 1 audit event tersimpan (DB sink)
- [ ] Field sensitif ter-redact

**CI Gates:**
- [ ] Pipeline lint + test + build + boundary-check **green** di default branch

**Docs Self-Service:**
- [ ] README punya "run local", "add module", "config", "observability"
- [ ] Langkah-langkah benar-benar match (tidak missing steps)

---

### Growth Features (Post-MVP / v2)

| Feature | Description |
|---------|-------------|
| **Layering Tests** | Automated architecture validation tests |
| **Redis Adapter** | Cache adapter dengan contoh implementation |
| **gRPC Transport** | Full gRPC support dengan protobuf |
| **Enhanced Metrics** | More comprehensive OTel metrics |
| **Performance Benchmarks** | Documented performance characteristics |
| **Wire/Fx Option** | Optional DI framework integration |

---

### Vision (Future / v3+)

| Feature | Description |
|---------|-------------|
| **Message Queue Adapter** | Kafka/RabbitMQ/NATS support |
| **Template Variants** | Cron/worker template, event-driven template |
| **Enterprise Hardening** | Policy packs, SBOM automation, compliance docs |
| **Advanced Audit** | Event sourcing option, SIEM integration |
| **Multi-tenancy** | Tenant isolation patterns |
| **CLI Generator** | `go-hex new module users` style tooling |

---

## User Journeys

### Journey 1: Andi Santos — From Chaos to Team Standard

**Character Profile:**
- **Role:** Tech Lead di fintech startup, 8 tahun pengalaman Go
- **Team:** 12 engineers, 15+ microservices
- **Pain:** Setiap service berbeda struktur, debug susah karena log tidak ada request-id/trace-id, incident berulang karena missing production concerns

**The Story:**

Andi baru selesai post-mortem incident ketiga bulan ini. Kali ini service payment gagal graceful shutdown saat deploy, menyebabkan beberapa transaksi orphaned. Yang bikin frustrasi: butuh 2 jam untuk trace masalah karena log antar service tidak ada correlation ID. "Kenapa kita harus reinvent wheel setiap bikin service baru?" pikirnya.

Malam itu, sambil scrolling GitHub, Andi menemukan **golang-api-hexagonal**. README-nya langsung menarik: "production-ready dari hari pertama, observability built-in, boundary enforcement." Skeptis tapi penasaran, dia clone dan jalankan `make setup && make run`.

**15 menit kemudian**, service sudah jalan. Andi hit `POST /users`, lalu buka log — dan terkejut melihat JSON structured log dengan `request_id` dan `trace_id` yang sama persis dengan yang muncul di response header. "Ini yang kita butuhkan," gumamnya.

Keesokan harinya, Andi pitch ke tim: "Kita bikin pilot dengan template ini untuk service notification yang akan dibangun minggu depan." Dua sprint kemudian, service notification live tanpa incident. PR review jadi lebih cepat karena semua orang sudah familiar dengan struktur. Import rules menangkap junior engineer yang tidak sengaja import infra package dari domain layer.

**Tiga bulan kemudian**, semua service baru wajib pakai template ini. Andi menulis "Service Standard" doc satu halaman yang merujuk ke golang-api-hexagonal. Incident "basic" (timeout, shutdown, missing health) turun 80%. Onboarding engineer baru dari 2 minggu jadi 3 hari.

**Andi's "Aha!" Moment:** "Trace + log + request-id nyambung tanpa config manual. Ini game changer untuk debugging."

**Capabilities Revealed:**
- Structured logging dengan request-id/trace-id correlation
- OpenTelemetry tracing out-of-box
- Graceful shutdown yang proper
- Import rules untuk boundary enforcement
- Standard error response format
- Health/readiness endpoints

---

### Journey 2: Citra Dewi — Weekend Project to Production in Hours

**Character Profile:**
- **Role:** Freelance backend developer, 4 tahun pengalaman
- **Situation:** Dapat project inventory API, deadline 2 minggu
- **Constraint:** Deploy ke VPS dengan Docker atau k8s sederhana, budget terbatas

**The Story:**

Jumat malam, Citra baru selesai call dengan klien baru — toko online yang butuh API inventory management. Deadline ketat: 2 minggu untuk MVP. Citra ingat project terakhir: 3 hari habis hanya untuk setup boilerplate, config logging, dan health check. Kali ini harus lebih cepat.

Citra googling "production ready golang api template" dan menemukan **golang-api-hexagonal**. "Hexagonal architecture, observability, security baseline... sounds too good to be true," pikirnya. Tapi star count dan README yang detail meyakinkannya untuk coba.

Sabtu pagi, 9 AM. `git clone`, `make setup`, `make run`. **10 menit kemudian**, Citra sudah bisa hit `POST /users` dan melihat response dengan proper error format. Health endpoint working. Log JSON rapi. "Okay, ini beneran works," katanya sambil tersenyum.

Citra mulai coding fitur inventory. Dengan melihat `users` module sebagai reference, dia tahu persis di mana taruh domain logic, di mana repository interface, di mana handler HTTP. **Sabtu sore**, CRUD inventory sudah jalan dengan validasi dan audit trail.

**Minggu**, Citra setup deployment. `docker-compose.yml` sudah ada, tinggal adjust untuk production. Environment config jelas — semua via env vars sesuai 12-factor. Health endpoint siap untuk load balancer. `make ci` jalan lokal untuk pastikan semua green sebelum push.

**Senin pagi**, MVP sudah live di VPS klien. Citra kirim invoice lebih cepat dari estimasi. Klien impressed dengan "professional setup" — ada `/health`, `/ready`, `/metrics`. Saat ada bug kecil di production, Citra debug dalam 5 menit berkat structured log dengan request ID.

**Dua bulan kemudian**, Citra sudah pakai template ini untuk 3 project lain. Dia bahkan bikin fork pribadi dengan beberapa customization untuk recurring client needs.

**Citra's "Aha!" Moment:** "Dari clone sampai production dalam weekend. Debug cepat karena log dan trace sudah siap. Ini hemat waktu gue banget!"

**Capabilities Revealed:**
- Quick bootstrap: clone → run dalam menit
- Reference module (`users`) sebagai learning template
- Docker Compose untuk local dan production
- Environment-based config (12-factor)
- Health/ready endpoints untuk deployment
- Structured logging untuk debugging
- Standard error response

---

### Journey 3: Dimas Pratama — First Day to First PR

**Character Profile:**
- **Role:** Backend Engineer, baru join tim Andi
- **Experience:** 2 tahun Go, tapi di perusahaan lama setiap service berbeda struktur
- **First Day Target:** Run lokal + pahami flow `users` module + bikin PR kecil

**The Story:**

Senin pagi, hari pertama Dimas di startup fintech. Di perusahaan sebelumnya, onboarding butuh 2 minggu — setiap service punya struktur berbeda, dokumentasi tersebar, dan senior engineer selalu sibuk untuk pair programming.

Andi, Tech Lead barunya, kirim Slack: "Welcome! Clone repo service-notification, ikuti README, harusnya bisa run dalam 30 menit. Kalau stuck, ping aku."

Dimas buka README. Langkah jelas: `make setup` untuk install tools, `make run` untuk start service dengan dependencies. **20 menit kemudian**, service jalan lokal. Dimas hit beberapa endpoint, semua response konsisten dengan format yang sama.

"Sekarang baca folder structure di README," lanjut Andi. Dimas buka section "Project Structure" — `domain/`, `app/`, `transport/`, `infra/` dengan penjelasan singkat masing-masing. Konsep hexagonal yang dia baca di artikel jadi konkret: domain tidak import apapun dari luar, usecase di app layer, HTTP handler di transport.

**Siang hari**, Dimas explore `users` module sebagai reference. Dia trace flow dari HTTP handler → usecase → repository → database. Pattern-nya konsisten. Validation di handler, business logic di usecase, database query di repository. "Kalau mau tambah field, tinggal ikutin pattern ini," pikirnya.

**Sore hari**, Andi kasih task kecil: "Tambah endpoint `GET /users/{id}/activity` yang return audit events untuk user tersebut. Audit events sudah tersimpan di DB." Dimas buka audit module, lihat pattern-nya, dan mulai coding.

**Sebelum pulang**, Dimas push PR pertamanya. CI green — lint, test, boundary check semua pass. Andi review dan approve dengan satu minor comment. "Good job, PR pertama di hari pertama!" kata Andi.

**Minggu kedua**, Dimas sudah bisa independently handle feature request tanpa banyak bertanya. Pattern sudah familiar, docs menjawab pertanyaan umum, dan structure konsisten di semua service.

**Dimas's "Aha!" Moment:** "Semua service sama strukturnya. Belajar satu, ngerti semua. Onboarding tercepat yang pernah gue alami."

**Capabilities Revealed:**
- Clear README dengan step-by-step
- Consistent project structure
- Reference module sebagai learning path
- CI dengan boundary check
- Audit trail yang queryable
- Self-service documentation

---

### Journey Requirements Summary

| Journey | Key Capabilities Revealed |
|---------|---------------------------|
| **Andi (Adoption)** | Observability (log + trace correlation), graceful shutdown, boundary enforcement, standard error format, health endpoints |
| **Citra (Quick Start)** | Fast bootstrap, reference module, Docker Compose, env config, health/ready for deployment, debugging with structured logs |
| **Dimas (Onboarding)** | Clear docs, consistent structure, reference module for learning, CI quality gates, audit queryable |

#### Consolidated Capability Areas for MVP

1. **Developer Experience**
   - `make setup/run/test/lint` commands
   - Clear README dengan getting started
   - Reference module (`users`) sebagai template

2. **Observability**
   - Structured JSON logging dengan request-id
   - OpenTelemetry tracing dengan trace-id correlation
   - Prometheus metrics endpoint
   - Health/readiness endpoints

3. **Security Baseline**
   - Request validation dengan standard error
   - JWT auth middleware ready
   - Secure headers
   - Rate limiting

4. **Architecture Enforcement**
   - Hexagonal folder structure
   - Import rules / linting
   - CI boundary checks

5. **Audit**
   - Audit events dengan DB sink
   - PII redaction
   - Queryable audit trail

6. **Deployment Ready**
   - Docker Compose
   - Environment-based config
   - Graceful shutdown

---

## Developer Tool Specific Requirements

### Project-Type Overview

**golang-api-hexagonal** adalah Go service boilerplate yang dikategorikan sebagai **Developer Tool** — menyediakan foundation code yang bisa di-clone dan dikustomisasi untuk membangun production-ready microservices.

**Value Proposition sebagai Developer Tool:**
- Zero-to-running dalam menit (bukan jam/hari)
- Opinionated structure yang konsisten
- Production concerns sudah built-in
- Learning path via reference module

---

### Technical Environment Matrix

#### Go Version Requirements

| Version | Support Level | Notes |
|---------|---------------|-------|
| **Go 1.23+** | Primary | Latest stable, fully tested |
| **Go 1.22** | CI Tested | Backward compatibility verified |
| **Go 1.21 and below** | Unsupported | May work, no guarantees |

> **Note:** Aligned with Architecture and Project Context documents (Go 1.23+ as primary target).

#### Operating System Support

| OS | Support Level | Notes |
|----|---------------|-------|
| **Linux** | Full | Primary development & production target |
| **macOS** | Full | Intel & Apple Silicon |
| **Windows** | WSL2 | Native Windows not supported; WSL2 required |

#### Container Runtime

| Runtime | Support Level | Notes |
|---------|---------------|-------|
| **Docker** | Required (MVP) | `docker compose` plugin (v2 syntax) |
| **Podman** | Best-effort (v2) | Community contribution welcome |

---

### Installation & Setup Methods

#### Prerequisites (MVP)

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.25+ | Language runtime |
| **Docker** | 24+ | Container runtime |
| **docker compose** | v2 plugin | Local infrastructure |
| **make** | any | Task runner |
| **golangci-lint** | latest | Code quality |
| **goose** | latest | Database migrations |

#### Optional Tools (v2)

| Tool | Purpose |
|------|---------|
| **sqlc** | Type-safe SQL generation |
| **Podman** | Alternative container runtime |

#### Setup Flow

```bash
# 1. Clone template
git clone https://github.com/iruldev/golang-api-hexagonal.git my-service
cd my-service

# 2. Install tools (if not present)
make setup

# 3. Start infrastructure
make infra-up  # or: docker compose up -d

# 4. Run migrations
make migrate-up

# 5. Run service
make run

# 6. Verify
curl http://localhost:8080/health
```

---

### IDE & Tooling Integration

#### VS Code (Primary)

**Included in `.vscode/`:**

| File | Purpose |
|------|---------|
| `settings.json` | Go extension settings, format on save, lint on save |
| `launch.json` | Debug configurations (local, with docker, attach) |
| `extensions.json` | Recommended extensions list |

**Recommended Extensions:**
- `golang.go` — Official Go extension
- `ms-azuretools.vscode-docker` — Docker integration
- `redhat.vscode-yaml` — YAML support

#### GoLand (Compatible)

- Project structure compatible with GoLand conventions
- Run configurations exportable
- No GoLand-specific files in repo (user preference)

#### Linting Configuration

**`.golangci.yml` included with:**
- Import ordering (goimports)
- Error handling (errcheck)
- Security checks (gosec)
- Complexity limits (gocyclo)
- **Custom: boundary enforcement rules**

---

### Documentation Structure

#### MVP Documentation Set

```
README.md                    # Quick start, badges, overview
docs/
├── ARCHITECTURE.md          # Hexagonal structure, layers, boundaries
├── LOCAL_DEV.md             # Detailed local development guide
└── OBSERVABILITY.md         # Logging, tracing, metrics setup
```

#### README.md Sections

| Section | Content |
|---------|---------|
| **Badges** | CI status, Go version, License |
| **Overview** | What & why |
| **Quick Start** | 5-step setup |
| **Project Structure** | Folder overview |
| **Configuration** | Environment variables |
| **Development** | Common commands |
| **Contributing** | (placeholder for v2) |

#### docs/ARCHITECTURE.md

| Section | Content |
|---------|---------|
| **Hexagonal Overview** | Diagram, layer responsibilities |
| **Directory Structure** | Detailed folder explanation |
| **Dependency Rules** | What imports what (enforcement) |
| **Adding a Module** | Step-by-step guide |
| **Adding an Adapter** | Step-by-step guide |

#### docs/LOCAL_DEV.md

| Section | Content |
|---------|---------|
| **Prerequisites** | Tools & versions |
| **First-time Setup** | Detailed walkthrough |
| **Daily Workflow** | Common tasks |
| **Troubleshooting** | Common issues & fixes |
| **IDE Setup** | VS Code & GoLand tips |

#### docs/OBSERVABILITY.md

| Section | Content |
|---------|---------|
| **Logging** | JSON format, levels, correlation IDs |
| **Tracing** | OpenTelemetry setup, span examples |
| **Metrics** | Prometheus endpoint, custom metrics |
| **Local Debugging** | Viewing traces/logs locally |

---

### Code Examples & Reference Patterns

#### Reference Module: `users`

**Purpose:** Complete working example covering all layers

| Layer | Example |
|-------|---------|
| **Domain** | `User` entity, `UserRepository` interface, domain errors |
| **App** | `CreateUserUseCase`, `GetUserUseCase`, validation |
| **Transport** | HTTP handlers, request/response DTOs, middleware chain |
| **Infra** | PostgreSQL repository implementation, migrations |

**Demonstrates:**
- CRUD operations across all layers
- Request validation with standard errors
- Audit event emission
- Transaction handling
- Context propagation (request-id, trace-id)

#### Future Examples (v2)

| Example | Description |
|---------|-------------|
| **Background Worker** | Message consumer pattern |
| **Scheduled Job** | Cron-style recurring task |
| **gRPC Transport** | Alternative transport layer |

---

### Implementation Considerations

#### Migration Strategy (goose)

| Aspect | Approach |
|--------|----------|
| **Tool** | goose (preferred over go-migrate) |
| **Location** | `infra/migrations/` |
| **Naming** | `YYYYMMDDHHMMSS_description.sql` |
| **Rollback** | Down migrations required |
| **CI** | Migrations run before tests |

#### Build & CI Commands

| Command | Action |
|---------|--------|
| `make setup` | Install required tools |
| `make run` | Run service locally |
| `make test` | Run all tests with race detector |
| `make lint` | Run golangci-lint |
| `make ci` | Full CI pipeline locally |
| `make infra-up` | Start Docker infrastructure |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback last migration |

#### Version Constraints

| Dependency | Constraint | Rationale |
|------------|------------|-----------|
| **Go** | 1.25+ | Latest features, security |
| **PostgreSQL** | 15+ | JSON, performance |
| **Redis** | 7+ (v2) | Streams support |

---

## Functional Requirements

### 1. Project Setup & Bootstrap

- **FR1:** Developer can clone the repository and have a working service running locally within 30 minutes
- **FR2:** Developer can run a single setup command to install all required development tools
- **FR3:** Developer can start all infrastructure dependencies with a single command
- **FR4:** Developer can run database migrations with a single command
- **FR5:** Developer can start the service locally with a single command
- **FR6:** Developer can verify the service is running correctly via health endpoint

### 2. Reference Implementation (Users Module)

- **FR7:** Developer can create a new user via HTTP API
- **FR8:** Developer can retrieve a single user by ID via HTTP API
- **FR9:** Developer can retrieve a list of users via HTTP API
- **FR10:** Developer can use the users module as a reference pattern for creating new modules
- **FR11:** Developer can trace the complete request flow from HTTP handler through all layers to database

### 3. Observability

#### Logging

- **FR12:** System emits structured JSON logs for all requests
- **FR13:** System includes request_id in all log entries for a given request
- **FR14:** System includes trace_id in log entries when tracing is enabled
- **FR15:** System returns request_id in response headers for client correlation

#### Tracing

- **FR16:** System generates OpenTelemetry traces for all HTTP requests
- **FR17:** System propagates trace context across service boundaries
- **FR18:** Developer can configure trace export destination via environment variables

#### Metrics

- **FR19:** System exposes Prometheus-compatible metrics endpoint
- **FR20:** System emits default HTTP metrics (request count, latency, status codes)
- **FR21:** Developer can add custom application metrics using provided utilities

#### Health & Readiness

- **FR22:** System exposes health endpoint for liveness checks
- **FR23:** System exposes readiness endpoint that validates database connectivity
- **FR24:** Health and readiness endpoints respond appropriately based on system state

### 4. Security Baseline

#### Request Validation

- **FR25:** System validates incoming request payloads against defined schemas
- **FR26:** System returns standardized error responses for validation failures
- **FR27:** System enforces request size limits

#### Authentication & Authorization

- **FR28:** Developer can protect endpoints with JWT authentication middleware
- **FR29:** System returns 401 for requests without valid authentication
- **FR30:** System returns 403 for requests without required authorization
- **FR31:** Developer can implement authorization logic at the use case layer

#### Security Headers & Rate Limiting

- **FR32:** System includes security headers in all HTTP responses
- **FR33:** System implements rate limiting per client
- **FR34:** System returns 429 when rate limit is exceeded

### 5. Audit Trail

- **FR35:** System records audit events for significant business operations
- **FR36:** System stores audit events in the database
- **FR37:** System redacts PII fields in audit event payloads
- **FR38:** Developer can query audit events for a specific entity
- **FR39:** Developer can extend audit event types for new modules

### 6. Architecture Enforcement

#### Hexagonal Structure

- **FR40:** Project follows hexagonal architecture with domain, app, transport, and infra layers
- **FR41:** Domain layer contains entities and repository interfaces with no external dependencies
- **FR42:** App layer contains use cases that orchestrate domain logic
- **FR43:** Transport layer handles HTTP concerns and DTO transformations
- **FR44:** Infra layer contains implementations of domain interfaces

#### Boundary Enforcement

- **FR45:** Linting rules detect import violations between layers
- **FR46:** CI pipeline fails when boundary violations are detected
- **FR47:** Developer receives clear error messages indicating which boundary was violated

### 7. Development Workflow

#### Local Development

- **FR48:** Developer can run full test suite locally with race detection
- **FR49:** Developer can run linting checks locally
- **FR50:** Developer can run full CI pipeline locally before pushing
- **FR51:** Developer can start/stop infrastructure containers independently

#### Database Migrations

- **FR52:** Developer can create new migration files following naming convention
- **FR53:** Developer can apply pending migrations
- **FR54:** Developer can rollback last migration
- **FR55:** Migrations run automatically in CI before tests

### 8. Configuration Management

- **FR56:** Developer can configure all settings via environment variables
- **FR57:** System provides sensible defaults for all configuration options
- **FR58:** System validates required configuration on startup
- **FR59:** System fails fast with clear error messages for invalid configuration

### 9. Error Handling

- **FR60:** System returns standardized error response format for all errors
- **FR61:** System distinguishes between client errors (4xx) and server errors (5xx)
- **FR62:** System includes error codes for programmatic error handling
- **FR63:** System logs appropriate detail for errors without exposing sensitive information

### 10. Documentation

- **FR64:** README provides quick start instructions that work without modification
- **FR65:** Architecture documentation explains hexagonal structure and layer responsibilities
- **FR66:** Local development documentation covers setup, daily workflow, and troubleshooting
- **FR67:** Observability documentation explains logging, tracing, and metrics configuration
- **FR68:** Documentation includes step-by-step guide for adding new modules
- **FR69:** Documentation includes step-by-step guide for adding new adapters

---

## Non-Functional Requirements

### Code Quality

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR1** | Test coverage for domain and app layers | ≥ 80% coverage |
| **NFR2** | All tests pass with race detector enabled | `go test -race` zero failures |
| **NFR3** | Zero linting errors on default branch | `golangci-lint` returns 0 errors |
| **NFR4** | No known security vulnerabilities in dependencies | `govulncheck` clean or documented exceptions |
| **NFR5** | Build is reproducible | Same source produces identical binary |
| **NFR6** | Code follows consistent formatting | `gofmt` produces no changes |

### Performance Baseline

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR7** | Health endpoint responds quickly | `GET /health` p95 < 10ms (local) |
| **NFR8** | No goroutine leaks under normal operation | Goroutine count stable over time |
| **NFR9** | Graceful shutdown completes in-flight requests | Shutdown within configured timeout |
| **NFR10** | Request handlers have timeouts | All handlers respect context cancellation |
| **NFR11** | Request size limits enforced | Requests exceeding limit rejected with 413 |

### Security

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR12** | All sensitive configuration via environment variables | No secrets in code or config files |
| **NFR13** | Security headers present in all responses | OWASP recommended headers included |
| **NFR14** | Rate limiting active on all public endpoints | Rate limit triggers 429 response |
| **NFR15** | Input validation on all user-supplied data | Invalid input returns 400 with error details |
| **NFR16** | PII redaction in audit logs | Sensitive fields masked in audit events |
| **NFR17** | JWT validation for protected endpoints | Invalid/expired tokens return 401 |
| **NFR18** | Error responses do not leak internal details | Stack traces not exposed to clients |

### Reliability

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR19** | Service starts successfully with valid configuration | Startup completes without errors |
| **NFR20** | Service fails fast with invalid configuration | Clear error message on startup failure |
| **NFR21** | Health endpoint reflects actual service health | Returns 503 when unhealthy |
| **NFR22** | Readiness endpoint reflects dependency status | Returns 503 when dependencies unavailable |
| **NFR23** | Graceful shutdown handles SIGTERM | In-flight requests complete before exit |
| **NFR24** | Database connections properly pooled and managed | No connection leaks |

### Portability

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR25** | Runs on Linux without modification | `make run` works on Linux |
| **NFR26** | Runs on macOS without modification | `make run` works on macOS (Intel & Apple Silicon) |
| **NFR27** | Runs on Windows via WSL2 | `make run` works in WSL2 environment |
| **NFR28** | Containerized with Docker | `docker build` produces working image |
| **NFR29** | Local development via docker compose | `docker compose up` starts all dependencies |

### Developer Experience

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR30** | Setup time for new developer | Clone to running service < 30 minutes |
| **NFR31** | Documentation accuracy | All documented steps work without modification |
| **NFR32** | CI feedback time | Full CI pipeline < 5 minutes |
| **NFR33** | Clear error messages | Errors include actionable guidance |
| **NFR34** | Consistent patterns across modules | All modules follow same structure |

### Observability Quality

| NFR | Requirement | Measurement |
|-----|-------------|-------------|
| **NFR35** | Log format consistency | All logs are valid JSON |
| **NFR36** | Request traceability | Every request has unique request_id |
| **NFR37** | Trace correlation | trace_id appears in logs when tracing enabled |
| **NFR38** | Metrics endpoint availability | `/metrics` returns Prometheus format |
| **NFR39** | Log levels configurable | LOG_LEVEL env var controls verbosity |

---

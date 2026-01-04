# Dokumentasi Proyek: golang-api-hexagonal

> **Dokumentasi Brownfield** - Dihasilkan oleh BMad Document Project
> **Tanggal:** 2025-12-29
> **Scan Level:** Exhaustive
> **Total Files:** 140 Go files

---

## Ringkasan Proyek

| Aspek | Detail |
|-------|--------|
| **Nama** | golang-api-hexagonal |
| **Tipe** | Backend API (Monolith) |
| **Arsitektur** | Hexagonal (Ports & Adapters) |
| **Bahasa** | Go 1.25.5 |
| **Framework** | Chi v5.2.3 |
| **Database** | PostgreSQL 15 (pgx v5.7.6) |
| **DI Framework** | Uber Fx v1.24.0 |
| **Test Coverage** | 80% threshold (domain+app) |

---

## Quick Reference

| Komponen | Teknologi | Versi |
|----------|-----------|-------|
| Web Framework | go-chi/chi | v5.2.3 |
| Database Driver | jackc/pgx | v5.7.6 |
| DI Container | uber/fx | v1.24.0 |
| Code Generation | sqlc | v1.28.0 |
| Migrations | goose | v3.26.0 |
| Tracing | OpenTelemetry | v1.39.0 |
| Metrics | Prometheus | v1.23.2 |
| JWT | golang-jwt | v5.3.0 |
| Validation | go-playground/validator | v10.22.0 |
| Linter | golangci-lint | v1.64.2 |
| Mocking | uber/mock | v0.5.2 |
| Integration Tests | testcontainers-go | v0.37.0 |

---

## Dokumentasi yang Tersedia

### Core Documentation
| File | Deskripsi | Status |
|------|-----------|--------|
| [architecture.md](./architecture.md) | Arsitektur hexagonal, layer rules, gap analysis | Updated |
| [patterns.md](./patterns.md) | Copy-paste code patterns untuk semua layers | New |
| [adr/](./adr/index.md) | Architecture Decision Records (ADRs) | New |
| [runbooks/](./runbooks/index.md) | Operational runbooks untuk incident response | New |
| [source-tree-analysis.md](./source-tree-analysis.md) | Struktur direktori dengan anotasi lengkap | Updated |
| [local-development.md](./local-development.md) | Setup development environment | Complete |
| [observability.md](./observability.md) | Metrics, tracing, logging setup | Complete |
| [testing-guide.md](./testing-guide.md) | Strategi testing dan coverage | Complete |
| [testing-patterns.md](./testing-patterns.md) | Testing conventions dan best practices | Complete |
| [adoption-guide.md](./adoption-guide.md) | Panduan adopsi untuk tim baru | Complete |
| [copy-paste-kit/](./copy-paste-kit/README.md) | Template files siap pakai | New |

### API & OpenAPI
| File | Deskripsi |
|------|-----------|
| [openapi.yaml](./openapi.yaml) | OpenAPI 3.1 specification |

### Checklists
| File | Deskripsi |
|------|-----------|
| [metrics-audit-checklist.md](./metrics-audit-checklist.md) | Checklist audit metrics |
| [observability-security-checklist.md](./observability-security-checklist.md) | Checklist keamanan observability |

### Scan Reports
| File | Deskripsi |
|------|-----------|
| [project-scan-report.json](./project-scan-report.json) | Machine-readable scan state (for workflow resume) |

---

## Entry Points

| Entry Point | Path | Deskripsi |
|-------------|------|-----------|
| Main Application | `cmd/api/main.go` | Entry point aplikasi dengan Uber Fx |
| Router | `internal/transport/http/router.go` | HTTP routing + middleware chain |
| Fx Module | `internal/infra/fx/module.go` | Dependency injection wiring |
| Config | `internal/infra/config/config.go` | Configuration loading |

---

## Getting Started

```bash
# 1. Clone dan setup
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal
make setup

# 2. Start infrastructure
make infra-up

# 3. Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
make migrate-up

# 4. Run service
make run

# 5. Verify (in another terminal)
curl http://localhost:8080/health
```

---

## Statistik Kode

| Metrik | Nilai |
|--------|-------|
| **Total Go Files** | 140 |
| **Production Files** | ~70 |
| **Test Files** | ~70 |
| **Domain Layer** | 12 files |
| **Application Layer** | 11 files |
| **Transport Layer** | 36 files |
| **Infrastructure Layer** | 24 files |
| **Test Utilities** | 13 files |
| **Coverage Threshold** | 80% (domain+app) |
| **CI Quality Gates** | 8 gates |
| **Migrations** | 4 |
| **API Endpoints** | 6 (4 public, 2 internal) |

---

## Arsitektur Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/api/main.go                          │
│                 (Application Entry Point)                   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  internal/infra/fx/                         │
│              (Uber Fx DI Wiring Layer)                      │
└─────────────────────────────────────────────────────────────┘
        │               │               │               │
        ▼               ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  transport/  │ │    app/      │ │   domain/    │ │   infra/     │
│     http/    │ │ (Use Cases)  │ │ (Entities)   │ │ (Adapters)   │
│  (Handlers)  │ │              │ │              │ │              │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

### Layer Rules (Enforced by depguard)

| Layer | Dapat Import | Tidak Boleh Import |
|-------|--------------|-------------------|
| domain/ | stdlib only | app, transport, infra |
| app/ | domain | transport, infra, slog, http |
| transport/ | domain, app, shared | infra |
| infra/ | domain, shared, external | app, transport |

---

## International Grade Gap Analysis

### What's Already Excellent
- Architecture enforcement via linting
- Comprehensive test coverage (80%)
- Security middleware (JWT, rate limiting, OWASP headers)
- Full observability stack (OTEL + Prometheus + slog)
- Integration tests with Testcontainers
- Security-focused tests (IDOR, auth edge cases)

### Areas for Improvement
See [architecture.md](./architecture.md) Section 6.2 for detailed gap analysis.

**Priority 1 (High)**:
- Enhance OpenAPI spec completeness
- Implement full RFC 7807 error codes taxonomy

**Priority 2 (Medium)**:
- Add application-level retry/circuit breaker
- Split large test files

**Priority 3 (Low)**:
- Add X-RateLimit-* headers
- Enhanced health check with dependency details

---

## Referensi Penting Untuk AI Agents

> **Untuk Brownfield PRD:** Gunakan file ini sebagai entry point untuk memahami context proyek.

**Critical Files:**
1. `internal/domain/*.go` - Domain entities dan interfaces
2. `internal/app/*/*.go` - Use cases dan application services
3. `internal/infra/fx/module.go` - Dependency injection graph
4. `.golangci.yml` - Layer rules enforcement
5. `Makefile` - Development commands reference
6. `internal/testutil/` - Test utilities, mocks, fixtures, containers

**Key Patterns:**
- All domain entities use stdlib only
- Use cases accept interfaces, not concrete implementations
- Handlers validate input, use cases validate business logic
- Transaction management via TxManager interface
- PII redaction via Redactor interface

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow (Exhaustive Scan)*

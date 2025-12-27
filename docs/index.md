# Dokumentasi Proyek: golang-api-hexagonal

> **Dokumentasi Brownfield** - Dihasilkan oleh BMad Document Project  
> **Tanggal:** 2025-12-27  
> **Scan Level:** Deep

---

## Ringkasan Proyek

| Aspek | Detail |
|-------|--------|
| **Nama** | golang-api-hexagonal |
| **Tipe** | Backend API (Monolith) |
| **Arsitektur** | Hexagonal (Ports & Adapters) |
| **Bahasa** | Go 1.24.11 |
| **Framework** | Chi v5.2.3 |
| **Database** | PostgreSQL (pgx v5.7.6) |
| **DI Framework** | Uber Fx v1.24.0 |

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

---

## Dokumentasi yang Dihasilkan

### Arsitektur & Struktur
- [Arsitektur](./architecture.md) - Arsitektur hexagonal dan layer rules
- [Source Tree Analysis](./source-tree-analysis.md) - Struktur direktori dengan anotasi
- [Data Models](./data-models.md) - Schema database dan migrasi

### Pengembangan
- [Development Guide](./development-guide.md) - Setup, build, test commands
- [API Contracts](./api-contracts.md) - REST API endpoints dan contracts
- [Testing Guide](./testing-guide.md) - Strategi testing dan coverage

### Operasional
- [Observability](./observability.md) - Metrics, tracing, logging
- [Local Development](./local-development.md) - Setup development environment

---

## Dokumentasi Existing

| File | Deskripsi |
|------|-----------|
| [README.md](../README.md) | Overview proyek dan quick start |
| [openapi.yaml](./openapi.yaml) | OpenAPI 3.1 specification |
| [observability.md](./observability.md) | Panduan observability stack |
| [local-development.md](./local-development.md) | Panduan development lokal |
| [metrics-audit-checklist.md](./metrics-audit-checklist.md) | Checklist audit metrics |
| [observability-security-checklist.md](./observability-security-checklist.md) | Checklist keamanan |
| [guides/](./guides/) | Panduan tambahan |

---

## Entry Points

| Entry Point | Path | Deskripsi |
|-------------|------|-----------|
| Main Application | `cmd/api/main.go` | Entry point aplikasi |
| Router | `internal/transport/http/router.go` | HTTP routing setup |
| Fx Module | `internal/infra/fx/module.go` | Dependency injection wiring |

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
| Test Files | 49 |
| Coverage Threshold | ≥80% (domain+app) |
| CI Gates | 8 quality gates |
| Migrations | 4 |
| API Endpoints | 6 (4 public, 2 internal) |

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

---

## Referensi Penting Untuk AI Agents

> **Untuk Brownfield PRD:** Gunakan file ini sebagai entry point untuk memahami context proyek.

**Critical Files:**
1. `internal/domain/*.go` - Domain entities dan interfaces
2. `internal/app/*/*.go` - Use cases dan application services
3. `internal/infra/fx/module.go` - Dependency injection graph
4. `.golangci.yml` - Layer rules enforcement
5. `Makefile` - Development commands reference

---

*Dokumentasi ini dihasilkan oleh BMad Method - Document Project Workflow*

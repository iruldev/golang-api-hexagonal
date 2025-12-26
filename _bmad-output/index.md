# Project Documentation Index

**golang-api-hexagonal** - Production-Ready Go API Boilerplate

---

## Project Overview

- **Type:** Backend API (Monolith)
- **Primary Language:** Go 1.24.11
- **Architecture:** Hexagonal (Ports & Adapters)
- **Database:** PostgreSQL

---

## Quick Reference

| Property | Value |
|----------|-------|
| HTTP Framework | Chi v5.2.3 |
| Database Driver | pgx v5.7.6 |
| Migration Tool | Goose v3.26.0 |
| Observability | OpenTelemetry + Prometheus |
| Entry Point | `cmd/api/main.go` |

---

## Generated Documentation

### Core Documentation

- [Project Overview](./project-overview.md) - Executive summary and key features
- [Architecture](./architecture.md) - Detailed architecture documentation
- [Source Tree Analysis](./source-tree-analysis.md) - Annotated directory structure

### API & Data

- [API Contracts](./api-contracts.md) - HTTP API documentation
- [Data Models](./data-models.md) - Database schema and entities

### Development

- [Development Guide](./development-guide.md) - Local development setup
- [Local Development](./local-development.md) - Detailed local dev workflow
- [Observability](./observability.md) - Logging, tracing, and metrics configuration

### How-To Guides

- [Adding Module](./guides/adding-module.md) - Guide for adding new modules
- [Adding Adapter](./guides/adding-adapter.md) - Guide for adding new adapters
- [Adding Audit Events](./guides/adding-audit-events.md) - Guide for adding audit event types

---

## Getting Started

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

### Quick Start

```bash
# 1. Setup
make setup

# 2. Start PostgreSQL
make infra-up

# 3. Run migrations
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
make migrate-up

# 4. Run the service
make run

# 5. Verify
curl http://localhost:8080/health
```

---

## Key Modules

### Users Module
- CRUD operations for user management
- Repository pattern with transaction support

### Audit Module
- Audit trail for business operations
- PII redaction for compliance

---

## CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`) includes:
- Linting (golangci-lint)
- Testing with coverage
- Security scanning (govulncheck)
- Migration validation
- Docker build verification

---

## Configuration

All configuration via environment variables. See:
- [.env.example](../.env.example) - Environment variable template
- [Configuration Section](./architecture.md#configuration) - Detailed config docs

---

## For AI-Assisted Development

When creating brownfield PRD, provide this index as context:

```
Reference documentation at: docs/index.md
```

Key documents for AI agents:
1. `docs/architecture.md` - Architecture patterns and rules
2. `docs/source-tree-analysis.md` - Project structure
3. `docs/api-contracts.md` - Existing API endpoints
4. `docs/data-models.md` - Database schema

---

## Document Versioning

| Document | Last Updated | Status |
|----------|--------------|--------|
| project-overview.md | 2024-12-24 | ✅ Complete |
| architecture.md | 2024-12-24 | ✅ Complete |
| source-tree-analysis.md | 2024-12-24 | ✅ Complete |
| api-contracts.md | 2024-12-24 | ✅ Complete |
| data-models.md | 2024-12-24 | ✅ Complete |
| development-guide.md | 2024-12-24 | ✅ Complete |
| local-development.md | Existing | ✅ Complete |
| observability.md | Existing | ✅ Complete |
| guides/adding-module.md | Existing | ✅ Complete |
| guides/adding-adapter.md | Existing | ✅ Complete |
| guides/adding-audit-events.md | Existing | ✅ Complete |

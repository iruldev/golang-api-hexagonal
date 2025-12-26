# Project Overview

**Project:** golang-api-hexagonal  
**Type:** Production-Ready Backend API Boilerplate  
**Architecture:** Hexagonal (Ports & Adapters)  
**Language:** Go 1.24.11  

## Executive Summary

A production-ready Go API built with **hexagonal architecture**, featuring comprehensive observability, security hardening, and developer-friendly tooling. Designed as a **boilerplate template** for creating new backend services with international production standards.

## Quick Reference

| Property | Value |
|----------|-------|
| **Primary Language** | Go 1.24.11 |
| **HTTP Framework** | Chi v5.2.3 |
| **Database** | PostgreSQL (pgx v5.7.6) |
| **Migration Tool** | Goose v3.26.0 |
| **Observability** | OpenTelemetry + Prometheus |
| **Authentication** | JWT (HS256, optional) |
| **Configuration** | Environment variables (envconfig) |
| **Entry Point** | `cmd/api/main.go` |

## Current Modules

### Users Module
- CRUD operations for user management
- Entities: `User` (ID, Email, FirstName, LastName, timestamps)
- Repository pattern with transaction support
- PII redaction for logs/audit

### Audit Module
- Audit trail for business operations
- Event types: `user.created`, `user.updated`, `user.deleted`
- JSON payload with automatic PII redaction
- Request correlation via RequestID

## API Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/health` | GET | No | Liveness probe |
| `/ready` | GET | No | Readiness probe (DB check) |
| `/metrics` | GET | No | Prometheus metrics |
| `/api/v1/users` | GET | JWT* | List users (paginated) |
| `/api/v1/users` | POST | JWT* | Create user |
| `/api/v1/users/{id}` | GET | JWT* | Get user by ID |

*JWT authentication is optional (controlled by `JWT_ENABLED` env var)

## Key Features

### Security
- OWASP security headers (CSP, HSTS, X-Frame-Options)
- JWT authentication middleware
- Rate limiting (IP-based or per-user)
- Request body size limiting
- PII redaction in logs and audit trails

### Observability
- Structured JSON logging (slog)
- OpenTelemetry distributed tracing
- Prometheus metrics (requests, latency)
- Request ID correlation across logs and traces

### Developer Experience
- Comprehensive Makefile (build, test, lint, migrate)
- Docker Compose for local PostgreSQL
- Goose migrations with validation
- 80%+ test coverage for domain/app layers
- CI/CD pipeline with GitHub Actions

## Getting Started

See [Development Guide](./development-guide.md) for setup instructions.

## Related Documentation

- [Architecture](./architecture.md) - Detailed architecture documentation
- [Source Tree Analysis](./source-tree-analysis.md) - Annotated directory structure
- [API Contracts](./api-contracts.md) - API documentation
- [Data Models](./data-models.md) - Database schema
- [Development Guide](./development-guide.md) - Local development setup

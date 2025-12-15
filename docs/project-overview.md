# Project Overview

## Backend Service Golang Boilerplate

**Enterprise-grade "golden template"** untuk membangun backend services di Go dengan **observability-first architecture**.

---

## Executive Summary

| Attribute | Value |
|-----------|-------|
| **Repository Type** | Monolith |
| **Primary Language** | Go 1.24 |
| **Architecture Pattern** | Hexagonal (Ports & Adapters) |
| **API Protocols** | HTTP (chi), gRPC, GraphQL |
| **Database** | PostgreSQL + sqlc |
| **Cache/Queue** | Redis + Asynq |
| **Message Brokers** | Kafka, RabbitMQ |
| **Observability** | OpenTelemetry, Prometheus, Grafana, Zap logging |

---

## Technology Stack

| Category | Technology | Version | Purpose |
|----------|-----------|---------|----------|
| **Language** | Go | 1.24.x | Core language |
| **HTTP Router** | chi | v5.2.3 | HTTP routing and middleware |
| **Database** | PostgreSQL | - | Primary data store |
| **DB Driver** | pgx | v5.7.6 | PostgreSQL driver |
| **Query Builder** | sqlc | - | Type-safe SQL |
| **Job Queue** | asynq | v0.25.1 | Background job processing |
| **Cache** | Redis | v9.17.2 | Caching, rate limiting, job queue backend |
| **GraphQL** | gqlgen | v0.17.84 | GraphQL server |
| **gRPC** | grpc-go | v1.77.0 | gRPC server |
| **Message Broker** | Sarama (Kafka) | v1.46.3 | Event publishing |
| **Message Broker** | amqp091-go | v1.10.0 | RabbitMQ integration |
| **Logging** | zap | v1.27.1 | Structured logging |
| **Tracing** | OpenTelemetry | v1.39.0 | Distributed tracing |
| **Metrics** | Prometheus client | v1.23.2 | Metrics collection |
| **Config** | koanf | v2.3.0 | Configuration management |
| **Auth** | JWT, OIDC | - | Authentication |
| **Testing** | testify, testcontainers | - | Testing framework |
| **CLI** | cobra | v1.10.2 | CLI framework (bplat tool) |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Entry Points                                │
│   ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │
│   │cmd/server │  │cmd/worker │  │cmd/scheduler│ │cmd/bplat │        │
│   │ (HTTP API)│  │(Job Queue)│  │  (Cron)    │ │  (CLI)   │        │
│   └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └───────────┘        │
└─────────┼──────────────┼──────────────┼─────────────────────────────┘
          │              │              │
┌─────────▼──────────────▼──────────────▼─────────────────────────────┐
│                     Interface Layer                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ HTTP (chi)   │  │   gRPC       │  │  GraphQL     │               │
│  │ + middleware │  │              │  │              │               │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │
└─────────┼─────────────────┼─────────────────┼───────────────────────┘
          │                 │                 │
┌─────────▼─────────────────▼─────────────────▼───────────────────────┐
│                       Use Case Layer                                 │
│              (Application Business Logic)                            │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────────┐
│                       Domain Layer                                   │
│           (Entities, Repository Interfaces, Errors)                  │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────────┐
│                   Infrastructure Layer                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐            │
│  │PostgreSQL│  │  Redis   │  │  Kafka   │  │ RabbitMQ │            │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘            │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Key Features

### Core Architecture
- **Hexagonal Architecture** - Clean separation of domain, usecase, interface, and infrastructure
- **Multiple Entry Points** - HTTP server, background worker, scheduler, CLI tool
- **Multi-Protocol Support** - HTTP REST, gRPC, GraphQL

### Security
- **JWT Authentication** - Token-based auth with claims
- **API Key Auth** - Service-to-service authentication
- **RBAC** - Role/Permission-based authorization (Admin, Service, User)
- **Security Headers Middleware** - Secure HTTP defaults

### Observability
- **OpenTelemetry Tracing** - Distributed tracing
- **Prometheus Metrics** - Application metrics
- **Grafana Dashboards** - Pre-configured dashboards
- **Zap Structured Logging** - JSON logging
- **Alerting Rules** - Pre-configured Prometheus alerts
- **Runbook Documentation** - Incident response guides

### Platform Features
- **Async Job Queue** - Fire-and-forget, scheduled, fanout patterns
- **Rate Limiting** - In-memory and Redis-backed
- **Feature Flags** - Environment-based provider
- **Event Publishing** - Kafka, RabbitMQ support
- **CLI Scaffolding** - bplat tool for code generation

---

## Quick Reference

### Entry Points

| Entry Point | Purpose | Command |
|-------------|---------|---------|
| `cmd/server` | HTTP API server | `go run ./cmd/server` |
| `cmd/worker` | Background job processor | `go run ./cmd/worker` |
| `cmd/scheduler` | Cron job scheduler | `go run ./cmd/scheduler` |
| `cmd/bplat` | CLI scaffolding tool | `./bin/bplat` |

### Key Commands

```bash
make dev          # Run development server
make build        # Build binaries
make test         # Run all tests
make lint         # Run golangci-lint
make sqlc         # Generate sqlc code
make migrate-up   # Run database migrations
docker-compose up # Start dependencies
```

---

## Links to Documentation

- [Architecture](./architecture.md) - Design decisions and patterns
- [Source Tree Analysis](./source-tree-analysis.md) - Directory structure
- [Development Guide](./development-guide.md) - Setup and development
- [API Contracts](./api-contracts.md) - API documentation
- [Data Models](./data-models.md) - Database schema
- [Async Jobs Guide](./async-jobs.md) - Background job patterns
- [Runbook Index](./runbook/README.md) - Incident response

---

*Generated by BMad Method document-project workflow on 2025-12-15*

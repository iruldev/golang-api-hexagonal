---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - "docs/prd.md"
  - "docs/architecture.md"
  - "docs/project-context.md"
  - "docs/test-design-system.md"
  - "docs/analysis/product-brief-golang-api-hexagonal-2025-12-16.md"
project_name: "golang-api-hexagonal"
date: "2025-12-16"
---

# golang-api-hexagonal - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for golang-api-hexagonal, decomposing the requirements from the PRD, Architecture, Test Design System, Project Context, and Product Brief into implementable stories.

---

## Requirements Inventory

### Functional Requirements

**Project Setup & Bootstrap**
- FR1: Developer can clone the repository and have a working service running locally within 30 minutes
- FR2: Developer can run a single setup command to install all required development tools
- FR3: Developer can start all infrastructure dependencies with a single command
- FR4: Developer can run database migrations with a single command
- FR5: Developer can start the service locally with a single command
- FR6: Developer can verify the service is running correctly via health endpoint

**Reference Implementation (Users Module)**
- FR7: Developer can create a new user via HTTP API
- FR8: Developer can retrieve a single user by ID via HTTP API
- FR9: Developer can retrieve a list of users via HTTP API
- FR10: Developer can use the users module as a reference pattern for creating new modules
- FR11: Developer can trace the complete request flow from HTTP handler through all layers to database

**Observability - Logging**
- FR12: System emits structured JSON logs for all requests
- FR13: System includes request_id in all log entries for a given request
- FR14: System includes trace_id in log entries when tracing is enabled
- FR15: System returns request_id in response headers for client correlation

**Observability - Tracing**
- FR16: System generates OpenTelemetry traces for all HTTP requests
- FR17: System propagates trace context across service boundaries
- FR18: Developer can configure trace export destination via environment variables

**Observability - Metrics**
- FR19: System exposes Prometheus-compatible metrics endpoint
- FR20: System emits default HTTP metrics (request count, latency, status codes)
- FR21: Developer can add custom application metrics using provided utilities

**Observability - Health & Readiness**
- FR22: System exposes health endpoint for liveness checks
- FR23: System exposes readiness endpoint that validates database connectivity
- FR24: Health and readiness endpoints respond appropriately based on system state

**Security Baseline - Request Validation**
- FR25: System validates incoming request payloads against defined schemas
- FR26: System returns standardized error responses for validation failures
- FR27: System enforces request size limits

**Security Baseline - Authentication & Authorization**
- FR28: Developer can protect endpoints with JWT authentication middleware
- FR29: System returns 401 for requests without valid authentication
- FR30: System returns 403 for requests without required authorization
- FR31: Developer can implement authorization logic at the use case layer

**Security Baseline - Security Headers & Rate Limiting**
- FR32: System includes security headers in all HTTP responses
- FR33: System implements rate limiting per client
- FR34: System returns 429 when rate limit is exceeded

**Audit Trail**
- FR35: System records audit events for significant business operations
- FR36: System stores audit events in the database
- FR37: System redacts PII fields in audit event payloads
- FR38: Developer can query audit events for a specific entity
- FR39: Developer can extend audit event types for new modules

**Architecture Enforcement - Hexagonal Structure**
- FR40: Project follows hexagonal architecture with domain, app, transport, and infra layers
- FR41: Domain layer contains entities and repository interfaces with no external dependencies
- FR42: App layer contains use cases that orchestrate domain logic
- FR43: Transport layer handles HTTP concerns and DTO transformations
- FR44: Infra layer contains implementations of domain interfaces

**Architecture Enforcement - Boundary Enforcement**
- FR45: Linting rules detect import violations between layers
- FR46: CI pipeline fails when boundary violations are detected
- FR47: Developer receives clear error messages indicating which boundary was violated

**Development Workflow - Local Development**
- FR48: Developer can run full test suite locally with race detection
- FR49: Developer can run linting checks locally
- FR50: Developer can run full CI pipeline locally before pushing
- FR51: Developer can start/stop infrastructure containers independently

**Development Workflow - Database Migrations**
- FR52: Developer can create new migration files following naming convention
- FR53: Developer can apply pending migrations
- FR54: Developer can rollback last migration
- FR55: Migrations run automatically in CI before tests

**Configuration Management**
- FR56: Developer can configure all settings via environment variables
- FR57: System provides sensible defaults for all configuration options
- FR58: System validates required configuration on startup
- FR59: System fails fast with clear error messages for invalid configuration

**Error Handling**
- FR60: System returns standardized error response format for all errors
- FR61: System distinguishes between client errors (4xx) and server errors (5xx)
- FR62: System includes error codes for programmatic error handling
- FR63: System logs appropriate detail for errors without exposing sensitive information

**Documentation**
- FR64: README provides quick start instructions that work without modification
- FR65: Architecture documentation explains hexagonal structure and layer responsibilities
- FR66: Local development documentation covers setup, daily workflow, and troubleshooting
- FR67: Observability documentation explains logging, tracing, and metrics configuration
- FR68: Documentation includes step-by-step guide for adding new modules
- FR69: Documentation includes step-by-step guide for adding new adapters

---

### Non-Functional Requirements

**Code Quality**
- NFR1: Test coverage for domain and app layers ≥ 80%
- NFR2: All tests pass with race detector enabled (`go test -race`)
- NFR3: Zero linting errors on default branch (`golangci-lint` returns 0 errors)
- NFR4: No known security vulnerabilities in dependencies (`govulncheck` clean)
- NFR5: Build is reproducible (same source produces identical binary)
- NFR6: Code follows consistent formatting (`gofmt` produces no changes)

**Performance Baseline**
- NFR7: Health endpoint responds quickly (`GET /health` p95 < 10ms local)
- NFR8: No goroutine leaks under normal operation (goroutine count stable)
- NFR9: Graceful shutdown completes in-flight requests (within timeout)
- NFR10: Request handlers have timeouts (respect context cancellation)
- NFR11: Request size limits enforced (exceeding limit → 413)

**Security**
- NFR12: All sensitive configuration via environment variables (no secrets in code)
- NFR13: Security headers present in all responses (OWASP recommended)
- NFR14: Rate limiting active on all public endpoints (triggers 429)
- NFR15: Input validation on all user-supplied data (invalid → 400)
- NFR16: PII redaction in audit logs (sensitive fields masked)
- NFR17: JWT validation for protected endpoints (invalid/expired → 401)
- NFR18: Error responses do not leak internal details (no stack traces)

**Reliability**
- NFR19: Service starts successfully with valid configuration
- NFR20: Service fails fast with invalid configuration (clear error)
- NFR21: Health endpoint reflects actual service health (503 when unhealthy)
- NFR22: Readiness endpoint reflects dependency status (503 when unavailable)
- NFR23: Graceful shutdown handles SIGTERM (in-flight complete before exit)
- NFR24: Database connections properly pooled and managed (no leaks)

**Portability**
- NFR25: Runs on Linux without modification
- NFR26: Runs on macOS without modification (Intel & Apple Silicon)
- NFR27: Runs on Windows via WSL2
- NFR28: Containerized with Docker (`docker build` works)
- NFR29: Local development via docker compose (`docker compose up` works)

**Developer Experience**
- NFR30: Setup time for new developer < 30 minutes (clone → running)
- NFR31: Documentation accuracy (all steps work without modification)
- NFR32: CI feedback time < 5 minutes (full pipeline)
- NFR33: Clear error messages (include actionable guidance)
- NFR34: Consistent patterns across modules (same structure)

**Observability Quality**
- NFR35: Log format consistency (all logs valid JSON)
- NFR36: Request traceability (every request has unique request_id)
- NFR37: Trace correlation (trace_id in logs when tracing enabled)
- NFR38: Metrics endpoint availability (`/metrics` returns Prometheus format)
- NFR39: Log levels configurable (`LOG_LEVEL` env var controls verbosity)

---

### Additional Requirements

**From Architecture Document:**
- AR-1: Use Chi v5 as HTTP router (stdlib-compatible, composable middleware)
- AR-2: Use pgx v5 for PostgreSQL driver (best performance, native features)
- AR-3: Use goose for database migrations (SQL-based, no ORM lock-in)
- AR-4: Use slog (stdlib) for structured logging (zero dependency)
- AR-5: Use OpenTelemetry for tracing and metrics (vendor-neutral)
- AR-6: Use envconfig for configuration (12-factor compliant, fail-fast)
- AR-7: Use go-playground/validator v10 for validation (struct tags)
- AR-8: Use golang-jwt/jwt v5 for JWT authentication
- AR-9: Use go-chi/httprate for rate limiting
- AR-10: Implement Repository pattern with Querier abstraction (works with pool or tx)
- AR-11: Implement Unit of Work pattern via TxManager.WithTx() for transactions
- AR-12: Use UUID v7 (time-ordered) generated in app/use case via IDGenerator interface, NOT database
- AR-13: Use RFC 7807 Problem Details for error responses
- AR-14: Use offset-based pagination for MVP (?page=1&pageSize=20)
- AR-15: Database naming: snake_case, plural tables (users, audit_events)
- AR-16: JSON field naming: camelCase
- AR-17: API paths: kebab-case, plural (/api/v1/users, /api/v1/audit-events)

**From Project Context (Guardrails):**
- PC-1: Domain layer: stdlib ONLY (no slog, uuid, pgx, chi, otel)
- PC-2: Domain layer: Use `type ID string` not uuid.UUID
- PC-3: Domain layer: NO logging — ever
- PC-4: App layer: domain imports only (no net/http, pgx, slog, otel)
- PC-5: App layer: Authorization checks happen HERE (not middleware)
- PC-6: App layer: NO logging — tracing context only
- PC-7: App layer: Generate UUID v7 for new entities via IDGenerator; Transport parses IDs from path params
- PC-8: Transport layer: Map AppError.Code → HTTP status + RFC 7807
- PC-9: Transport layer: ONLY place that knows HTTP status codes
- PC-10: Infra layer: Wrap errors with `op` string pattern
- PC-11: Infra layer: Convert domain.ID ↔ uuid.UUID at boundary
- PC-12: depguard enforcement in CI — violations = build failure

**From Test Design System:**
- TD-1: Test pyramid: 70% unit, 25% integration, 5% API/E2E
- TD-2: Domain layer: 100% unit test coverage target
- TD-3: App layer: 90% unit test coverage target
- TD-4: Handlers/Middleware: 80% coverage with testify + httptest
- TD-5: Postgres repos: Integration tests with testcontainers-go
- TD-6: Implement IDGenerator interface for deterministic tests
- TD-7: Implement Clock interface for deterministic timestamps
- TD-8: RFC 7807 contract tests (3 scenarios: validation, not found, internal)
- TD-9: CI gates: 100% test pass, 80% coverage (domain+app), 0 lint errors

**From Product Brief (MVP Scope):**
- PB-1: Must Have: Hexagonal architecture with boundary enforcement via import rules
- PB-2: Must Have: Reference module (users) with create + get + list
- PB-3: Must Have: Config, logging, tracing, metrics, health endpoints
- PB-4: Must Have: JWT auth middleware, rate limiting, secure headers, audit trail
- PB-5: Must Have: Make targets, docker-compose, migrations, CI pipeline, documentation
- PB-6: Out of scope for MVP: CLI generator, gRPC, layering tests, Redis, advanced audit

---

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1-6 | Epic 1 | Project setup, bootstrap, health verification |
| FR7-11 | Epic 4 | Users CRUD reference implementation |
| FR12-21 | Epic 2 | Logging, tracing, metrics |
| FR22-24 | Epic 1 | Health & readiness endpoints |
| FR25-27 | Epic 5 | Request validation, size limits |
| FR28-31 | Epic 5 | JWT auth, 401/403 responses |
| FR32-34 | Epic 5 | Secure headers, rate limiting |
| FR35-39 | Epic 6 | Audit trail, PII redaction |
| FR40-44 | Epic 1 | Hexagonal folder structure |
| FR45, FR47 | Epic 3 | Local lint/depguard boundary detection |
| FR46 | Epic 7 | CI fails on boundary violations |
| FR48-51 | Epic 3 | Local test, lint, CI, infra containers |
| FR52-55 | Epic 7 | Migration commands, CI migrations |
| FR56-59 | Epic 1 | Environment config, fail-fast |
| FR60-63 | Epic 4 | Standard error response format |
| FR64-69 | Epic 8 | Documentation guides |

---

## Epic List

### Epic 1: Project Foundation & First Run Experience
**Goal:** Developer dapat clone repo, run `make setup && make run`, dan melihat service berjalan dengan health endpoints dalam 15-30 menit.

**FRs covered:** FR1-6, FR22-24, FR40-44, FR56-59

**User Outcome:** Complete runnable service with health probes, hexagonal structure, env-based config.

---

### Epic 2: Observability Stack
**Goal:** Developer dapat melihat structured JSON logs dengan request_id/trace_id correlation, OpenTelemetry traces, dan Prometheus metrics di /metrics.

**FRs covered:** FR12-21

**User Outcome:** Full observability pipeline - logs, traces, metrics with correlation IDs.

---

### Epic 3: Local Quality Gates
**Goal:** Developer dapat run lint, tests, coverage checks lokal, dan boundary violations terdeteksi sebelum push.

**FRs covered:** FR45, FR47, FR48-51

**User Outcome:** Quality gates active locally - depguard, lint, tests with race detector.

---

### Epic 4: Reference Implementation (Users Module)
**Goal:** Developer dapat melihat complete CRUD pattern (handler → usecase → repository) dengan standard error handling RFC 7807.

**FRs covered:** FR7-11, FR60-63

**User Outcome:** Working reference pattern demonstrating all hexagonal layers end-to-end.

---

### Epic 5: Security & Authentication Foundation
**Goal:** Developer dapat protect Users endpoints dengan JWT middleware, rate limiting, secure headers, dan request validation.

**FRs covered:** FR25-34

**User Outcome:** Security baseline complete - auth, validation, rate limit, headers.

---

### Epic 6: Audit Trail System
**Goal:** Developer dapat record, store, dan query audit events dengan PII redaction untuk compliance requirements.

**FRs covered:** FR35-39

**User Outcome:** Compliance-ready audit trail with DB sink and PII protection.

---

### Epic 7: CI/CD Pipeline
**Goal:** Developer dapat push code dan CI pipeline (lint → test → build → boundary-check → migrations) berjalan otomatis.

**FRs covered:** FR46, FR52-55

**User Outcome:** Automated quality pipeline in GitHub Actions with migration support.

---

### Epic 8: Documentation & Developer Guides
**Goal:** Developer dapat self-service: quick start, architecture, local dev, observability, add module, add adapter - tanpa bantuan.

**FRs covered:** FR64-69

**User Outcome:** Complete documentation enabling developer autonomy.

---

## Epic Dependencies

```
Epic 1 (Foundation) ────────────────────────────────────────────┐
    │                                                           │
    ├── Epic 2 (Observability) ──┐                              │
    │                            │                              │
    │                            ▼                              │
    │                     Epic 3 (Local Quality Gates) ───► Epic 7 (CI/CD)
    │                            │                              │
    │                            ▼                              │
    │                     Epic 4 (Users Reference)              │
    │                            │                              │
    │                            ├── Epic 5 (Security) ──► Epic 6 (Audit)
    │                            │                              │
    └────────────────────────────┴──────────────────────────────┘
                                 │
                                 ▼
                          Epic 8 (Documentation)
```

---

## Epic 1: Project Foundation & First Run Experience

**Goal:** Developer dapat clone repo, run `make setup && make run`, dan melihat service berjalan dengan health endpoints dalam 15-30 menit.

**FRs covered:** FR1-6, FR22-24, FR40-44, FR56-59

---

### Story 1.1: Initialize Hexagonal Folder Structure

**As a** developer,
**I want** a pre-configured hexagonal folder structure with clear layer separation,
**So that** I can immediately understand where to place code.

**Acceptance Criteria:**

**Given** I clone the repository
**When** I view the project structure
**Then** the following directories exist and are preserved in git via `.keep`:
- `cmd/api/`
- `internal/domain/`
- `internal/app/`
- `internal/transport/http/handler/`
- `internal/transport/http/contract/`
- `internal/transport/http/middleware/`
- `internal/infra/config/`
- `internal/infra/postgres/`
- `internal/infra/observability/`
- `internal/shared/`
- `migrations/`
- `.github/workflows/`
- `docs/`
**And** `cmd/api/main.go` exists as the entry point
**And** `go build ./...` succeeds without errors

*Covers: FR40-44*

---

### Story 1.2: Implement Configuration Management

**As a** developer,
**I want** environment-based configuration with validation and sensible defaults,
**So that** the service fails fast with clear errors if misconfigured.

**Acceptance Criteria:**

**Given** I start the service without required configuration (e.g., DATABASE_URL missing)
**When** the service attempts to load configuration
**Then** the service exits with non-zero exit code
**And** the error message clearly indicates which configuration is missing (e.g., "required key DATABASE_URL missing value")

**Given** I start the service with valid configuration
**When** the service loads configuration
**Then** defaults are applied for optional values:
- `PORT` defaults to `8080`
- `LOG_LEVEL` defaults to `info`
- `ENV` defaults to `development`
- `SERVICE_NAME` defaults to `golang-api-hexagonal`
**And** `.env.example` exists with all configurable options documented
**And** configuration loader uses `kelseyhightower/envconfig` package

*Covers: FR56-59*

---

### Story 1.3: Setup Docker Compose Infrastructure

**As a** developer,
**I want** a single command to start all infrastructure dependencies,
**So that** I can run locally without manual setup.

**Acceptance Criteria:**

**Given** I have Docker installed
**When** I run `make infra-up` (or `docker compose up -d`)
**Then** PostgreSQL 15+ container starts
**And** container uses `pg_isready` healthcheck
**And** `docker compose ps` shows "healthy" status
**And** volume `pgdata` is created for data persistence
**And** `make infra-down` stops and removes containers (preserving volume)

*Covers: FR3*

---

### Story 1.4: Implement Database Migration System

**As a** developer,
**I want** a single command to run database migrations,
**So that** the schema is ready before the service starts.

**Acceptance Criteria:**

**Given** infrastructure is running (PostgreSQL accessible)
**When** I run `make migrate-up`
**Then** goose migrations in `migrations/` are applied successfully
**And** goose version table is updated in database
**And** migration files follow format `YYYYMMDDHHMMSS_description.sql` with `-- +goose Up` and `-- +goose Down` sections

**Given** migrations have been applied
**When** I run `make migrate-down`
**Then** the last migration is rolled back
**And** goose version table reflects the rollback

*Covers: FR4*

---

### Story 1.5: Implement Health & Readiness Endpoints

**As a** developer,
**I want** health and readiness endpoints,
**So that** I can verify the service is running and ready to accept traffic.

**Acceptance Criteria:**

**Given** the service is running
**When** I call `GET /health` (liveness probe)
**Then** I receive HTTP 200
**And** response body is `{"data":{"status":"ok"}}`
**And** this endpoint does NOT check database connectivity

**Given** the service is running and database is connected
**When** I call `GET /ready` (readiness probe)
**Then** I receive HTTP 200
**And** response body is `{"data":{"status":"ready","checks":{"database":"ok"}}}`

**Given** the service is running but database is NOT connected
**When** I call `GET /ready`
**Then** I receive HTTP 503
**And** response body is `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`

*Covers: FR6, FR22-24*

---

### Story 1.6: Create Makefile Commands & Developer Setup

**As a** developer,
**I want** intuitive make targets for common development tasks,
**So that** I can bootstrap and run everything easily.

**Acceptance Criteria:**

**Given** I clone the repository
**When** I run `make setup`
**Then** required tools are installed if not present (golangci-lint, goose)
**And** tool versions are printed to stdout
**And** `go mod download` completes successfully

**Given** setup is complete
**When** I run `make help`
**Then** I see a list of all available make targets with descriptions

**Given** infrastructure is running and migrations applied
**When** I run `make run`
**Then** the service starts and listens on configured port
**And** graceful shutdown handles SIGTERM (in-flight requests complete within timeout)

**Given** I follow the full workflow
**When** I execute `make setup && make infra-up && make migrate-up && make run`
**Then** the service is accessible at `http://localhost:8080/health`
**And** total time from clone to running < 15 minutes

*Covers: FR1, FR2, FR5*

---

### Epic 1 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 1.1 | Initialize Hexagonal Folder Structure | FR40-44 | S |
| 1.2 | Implement Configuration Management | FR56-59 | M |
| 1.3 | Setup Docker Compose Infrastructure | FR3 | S |
| 1.4 | Implement Database Migration System | FR4 | S |
| 1.5 | Implement Health & Readiness Endpoints | FR6, FR22-24 | M |
| 1.6 | Create Makefile Commands & Developer Setup | FR1, FR2, FR5 | M |

**Total: 6 stories covering 14 FRs**

---

## Epic 2: Observability Stack

**Goal:** Developer dapat melihat structured JSON logs dengan request_id/trace_id correlation, OpenTelemetry traces, dan Prometheus metrics di /metrics.

**FRs covered:** FR12-21

---

### Story 2.1: Implement Structured JSON Logging

**As a** developer,
**I want** structured JSON logs with consistent fields,
**So that** I can parse and search logs easily.

**Acceptance Criteria:**

**Given** the service is running
**When** any log entry is written
**Then** log is emitted in JSON format to stdout
**And** each log entry includes required fields:
- `time` (RFC 3339 format)
- `level` (debug/info/warn/error)
- `msg` (log message)
- `service` (from config SERVICE_NAME)
- `env` (from config ENV)
**And** request-scoped logs include `requestId`

**Given** `LOG_LEVEL` is set to `warn`
**When** info-level log is written
**Then** log is NOT emitted (filtered)

**Given** any log operation
**When** log entry is written
**Then** sensitive data is NEVER logged (Authorization header, password, token, secret fields)

**Given** any HTTP request is processed
**When** request completes (success or error)
**Then** request logging middleware emits log entry with:
- `method` (HTTP method)
- `route` (Chi route pattern)
- `status` (HTTP status code)
- `duration_ms` (request processing time)
- `bytes` (response size)

*Covers: FR12*

---

### Story 2.2: Implement Request ID Middleware

**As a** developer,
**I want** every request to have a unique request_id,
**So that** I can trace a single request through all log entries.

**Acceptance Criteria:**

**Given** the service receives an HTTP request without `X-Request-ID` header
**When** the request is processed
**Then** a unique `requestId` is generated (opaque random, 16 bytes hex)
**And** `requestId` is injected into request context
**And** `requestId` appears in all log entries for that request
**And** `requestId` is returned in response header `X-Request-ID`

**Given** the service receives an HTTP request WITH `X-Request-ID` header
**When** the request is processed
**Then** the provided `requestId` is used (passthrough)
**And** the same value is returned in response header `X-Request-ID`

*Covers: FR13, FR15*

---

### Story 2.3a: Implement OpenTelemetry Tracing Setup

**As a** developer,
**I want** distributed tracing with OpenTelemetry,
**So that** I can trace requests across service boundaries.

**Acceptance Criteria:**

**Given** `OTEL_ENABLED=true` and `OTEL_EXPORTER_OTLP_ENDPOINT` is set
**When** the service starts
**Then** OpenTelemetry tracer provider is initialized
**And** traces are exported to the configured endpoint

**Given** an HTTP request is received (tracing enabled)
**When** the request is processed
**Then** a span is created with attributes:
- `http.method`
- `http.route` (Chi route pattern, e.g., `/api/v1/users/{id}`)
- `http.status_code`
**And** span is properly ended when request completes

**Given** request contains `traceparent` header (W3C Trace Context)
**When** the request is processed
**Then** the span continues the existing trace (context extracted)

**Given** `OTEL_ENABLED=false` (default)
**When** the service starts
**Then** no tracer provider is initialized
**And** application functions normally without exporting spans

*Covers: FR16-18*

---

### Story 2.3b: Implement Trace Correlation in Logs

**As a** developer,
**I want** trace_id and span_id in log entries,
**So that** I can correlate logs with traces.

**Acceptance Criteria:**

**Given** tracing is enabled and request has an active span
**When** a log entry is written within that request context
**Then** log entry includes `traceId` field (32 hex chars)
**And** log entry includes `spanId` field (16 hex chars)

**Given** tracing is disabled
**When** a log entry is written
**Then** `traceId` and `spanId` fields are absent (not empty string)
**And** logging functions normally without errors

*Covers: FR14*

---

### Story 2.4: Implement Prometheus Metrics Endpoint

**As a** developer,
**I want** a Prometheus-compatible metrics endpoint,
**So that** I can monitor service health and performance.

**Acceptance Criteria:**

**Given** the service is running
**When** I call `GET /metrics`
**Then** I receive HTTP 200
**And** response content-type contains `text/plain`
**And** response body is in Prometheus exposition format

**Given** HTTP requests are processed
**When** I check `/metrics`
**Then** the following metrics are present:
- `http_requests_total{method, route, status}` (counter)
- `http_request_duration_seconds{method, route}` (histogram)
**And** `route` label uses Chi route template (e.g., `/api/v1/users/{id}`), NOT raw path

**Given** the service is running
**When** I check `/metrics`
**Then** Go runtime metrics are present:
- `go_goroutines`
- `go_memstats_*`

*Covers: FR19-20*

---

### Story 2.5: Add Custom Metrics Utilities

**As a** developer,
**I want** utilities to add custom application metrics,
**So that** I can track business-specific metrics.

**Acceptance Criteria:**

**Given** the `internal/infra/observability` package is available
**When** I use provided metric registration utilities
**Then** I can create and register:
- Custom counter metrics
- Custom histogram metrics
- Custom gauge metrics
**And** custom metrics appear at `/metrics` endpoint with proper labels

**Given** the observability package
**When** I view the package documentation
**Then** package comment or code example shows how to register custom metrics

*Covers: FR21*

---

### Epic 2 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 2.1 | Implement Structured JSON Logging | FR12 | M |
| 2.2 | Implement Request ID Middleware | FR13, FR15 | S |
| 2.3a | Implement OpenTelemetry Tracing Setup | FR16-18 | M |
| 2.3b | Implement Trace Correlation in Logs | FR14 | S |
| 2.4 | Implement Prometheus Metrics Endpoint | FR19-20 | M |
| 2.5 | Add Custom Metrics Utilities | FR21 | S |

**Total: 6 stories covering 10 FRs**

---

## Epic 3: Local Quality Gates

**Goal:** Developer dapat run lint, tests, coverage checks lokal, dan boundary violations terdeteksi sebelum push.

**FRs covered:** FR45, FR47, FR48-51

---

### Story 3.1: Configure golangci-lint with Boundary Rules

**As a** developer,
**I want** linting configured with hexagonal boundary rules,
**So that** import violations are detected locally before push.

**Acceptance Criteria:**

**Given** I have golangci-lint installed
**When** I run `make lint`
**Then** golangci-lint runs with `.golangci.yml` configuration
**And** `make lint` exits with code 0 when no violations

**Given** code in `internal/domain/` imports external package (e.g., `github.com/google/uuid`)
**When** I run `make lint`
**Then** depguard rule fails with exit code non-zero
**And** error message includes:
- Rule name (e.g., "domain-layer")
- Forbidden import path
- File path and line number

**Given** code in `internal/app/` imports `net/http` or `github.com/jackc/pgx`
**When** I run `make lint`
**Then** depguard rule fails with clear error identifying the boundary violation

*Covers: FR45, FR47*

---

### Story 3.2: Setup Unit Test Infrastructure with Coverage Gate

**As a** developer,
**I want** to run unit tests with race detection and coverage enforcement,
**So that** I can verify code correctness and maintain quality standards.

**Acceptance Criteria:**

**Given** I have test files in `internal/domain/` and `internal/app/`
**When** I run `make test`
**Then** tests execute with `-race` flag enabled
**And** test output shows pass/fail status for each test
**And** coverage profile is generated

**Given** I want to check coverage threshold
**When** I run `make coverage`
**Then** coverage is calculated for `./internal/domain/...` and `./internal/app/...`
**And** coverage percentage is displayed
**And** `make coverage` fails (exit non-zero) if combined coverage < 80%
**And** `make coverage` passes (exit 0) if combined coverage ≥ 80%

*Covers: FR48*

---

### Story 3.3: Configure Local CI Pipeline

**As a** developer,
**I want** to run the full CI pipeline locally,
**So that** I can verify everything passes before pushing.

**Acceptance Criteria:**

**Given** I want to verify my changes
**When** I run `make ci`
**Then** the following steps execute in order:
1. `go mod tidy` check
2. `gofmt` check
3. `make lint`
4. `make test`
**And** after `go mod tidy`, `git diff --exit-code go.mod go.sum` passes (no changes)
**And** after `gofmt`, `git diff --exit-code` passes (no formatting changes)
**And** pipeline fails fast on first error
**And** exit code is non-zero if any step fails

*Covers: FR49, FR50*

---

### Story 3.4: Infrastructure Container Management

**As a** developer,
**I want** to start/stop infrastructure containers independently,
**So that** I can manage local development resources.

**Acceptance Criteria:**

**Given** I want to start infrastructure
**When** I run `make infra-up`
**Then** only infrastructure containers start (not the app)
**And** containers run in detached mode
**And** command waits for healthcheck to pass

**Given** infrastructure is running
**When** I run `make infra-down`
**Then** containers stop and are removed
**And** data volume is preserved (not deleted)

**Given** I want fresh infrastructure
**When** I run `make infra-reset`
**Then** warning message is printed: "WARNING: removing volumes"
**And** containers and volumes are removed
**And** infrastructure can be started fresh with `make infra-up`

*Covers: FR51*

---

### Epic 3 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 3.1 | Configure golangci-lint with Boundary Rules | FR45, FR47 | M |
| 3.2 | Setup Unit Test Infrastructure with Coverage Gate | FR48 | S |
| 3.3 | Configure Local CI Pipeline | FR49, FR50 | S |
| 3.4 | Infrastructure Container Management | FR51 | S |

**Total: 4 stories covering 6 FRs**

---

## Epic 4: Reference Implementation (Users Module)

**Goal:** Developer dapat melihat complete CRUD pattern (handler → usecase → repository) dengan standard error handling RFC 7807.

**FRs covered:** FR7-11, FR60-63

---

### Story 4.1: Implement User Domain Entity and Repository Interface

**As a** developer,
**I want** User entity and repository interface in domain layer,
**So that** I have a clear contract for user data access.

**Acceptance Criteria:**

**Given** the domain layer
**When** I view `internal/domain/user.go`
**Then** User entity exists with fields:
- `ID` (type `ID` which is `type ID string`)
- `Email` (string)
- `FirstName` (string)
- `LastName` (string)
- `CreatedAt` (time.Time)
- `UpdatedAt` (time.Time)

**And** `UserRepository` interface is defined with methods:
- `Create(ctx context.Context, q Querier, user *User) error`
- `GetByID(ctx context.Context, q Querier, id ID) (*User, error)`
- `List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)`

**And** domain errors are defined:
- `var ErrUserNotFound = errors.New("user not found")`
- `var ErrEmailAlreadyExists = errors.New("email already exists")`

**And** `IDGenerator` interface is defined:
- `type IDGenerator interface { NewID() ID }`

**And** domain layer has NO external imports (stdlib only)

*Covers: FR10 (partial), FR41*

---

### Story 4.2: Implement User PostgreSQL Repository

**As a** developer,
**I want** UserRepository implementation for PostgreSQL,
**So that** I can persist and retrieve users from the database.

**Acceptance Criteria:**

**Given** the infra layer
**When** I view `internal/infra/postgres/user_repo.go`
**Then** `UserRepo` struct implements `domain.UserRepository`

**Given** migration file `migrations/YYYYMMDDHHMMSS_create_users.sql`
**When** migration is applied
**Then** table `users` is created with columns:
- `id` (uuid, primary key)
- `email` (varchar, unique)
- `first_name` (varchar)
- `last_name` (varchar)
- `created_at` (timestamptz)
- `updated_at` (timestamptz)

**Given** Create is called with valid user
**When** the operation succeeds
**Then** user is inserted into `users` table
**And** `domain.ID` is parsed to `uuid.UUID` at repository boundary

**Given** GetByID is called with non-existent ID
**When** query returns no rows
**Then** `domain.ErrUserNotFound` is returned (wrapped: `"userRepo.GetByID: %w"`)

**Given** Create is called with duplicate email
**When** unique constraint violation occurs
**Then** `domain.ErrEmailAlreadyExists` is returned

**Given** List is called
**When** query executes
**Then** results are ordered by `created_at DESC, id DESC`
**And** total count is returned for pagination

*Covers: FR11 (partial)*

---

### Story 4.3: Implement User Use Cases

**As a** developer,
**I want** use cases for creating, getting, and listing users,
**So that** business logic is properly orchestrated in the app layer.

**Acceptance Criteria:**

**Given** the app layer
**When** I view `internal/app/user/`
**Then** the following use cases exist:
- `CreateUserUseCase`
- `GetUserUseCase`
- `ListUsersUseCase`

**Given** CreateUserUseCase is instantiated
**When** use case is created
**Then** it accepts `IDGenerator` interface for generating user IDs
**And** ID is generated via `idGen.NewID()` (not in handler)

**Given** CreateUserUseCase is executed with valid request
**When** the operation completes
**Then** user is created via repository
**And** domain.User is returned

**Given** GetUserUseCase is executed with non-existent ID
**When** repository returns ErrUserNotFound
**Then** AppError with Code="USER_NOT_FOUND" is returned

**And** app layer has NO imports of net/http, pgx, slog, uuid
**And** use cases accept repository interfaces (not implementations)

*Covers: FR10 (partial), FR42*

---

### Story 4.4: Implement RFC 7807 Error Response Mapper

**As a** developer,
**I want** standardized error responses following RFC 7807,
**So that** all handlers can return consistent error format.

**Acceptance Criteria:**

**Given** the transport layer
**When** I view `internal/transport/http/contract/error.go`
**Then** `ProblemDetail` struct exists with fields:
- `Type` (string, URL)
- `Title` (string)
- `Status` (int)
- `Detail` (string)
- `Instance` (string)
- `Code` (string, extension)
- `ValidationErrors` ([]ValidationError, optional)

**Given** an AppError with Code="USER_NOT_FOUND"
**When** error mapper processes it
**Then** HTTP status 404 is mapped
**And** response Content-Type is `application/problem+json`

**Given** an AppError with Code="VALIDATION_ERROR"
**When** error mapper processes it
**Then** HTTP status 400 is mapped
**And** validationErrors array is populated

**Given** an unknown/internal error
**When** error mapper processes it
**Then** HTTP status 500 is returned
**And** Code="INTERNAL_ERROR"
**And** error details are NOT exposed (no stack trace, DB error)

**And** error code to HTTP status mapping is centralized

*Covers: FR60-63*

---

### Story 4.5: Implement Transport Contracts (DTOs)

**As a** developer,
**I want** request/response DTOs in transport layer,
**So that** domain entities don't leak HTTP concerns.

**Acceptance Criteria:**

**Given** the transport layer
**When** I view `internal/transport/http/contract/user.go`
**Then** the following DTOs exist:
- `CreateUserRequest` with validation tags
- `UserResponse`
- `ListUsersResponse`
- `PaginationResponse`

**And** all JSON tags use camelCase: `firstName`, `lastName`, `createdAt`
**And** timestamps serialize as RFC 3339 strings
**And** validation tags are present: `validate:"required,email"`, etc.

**Given** `CreateUserRequest` with invalid email
**When** validation runs
**Then** validation error is returned with field name

*Covers: FR43 (partial)*

---

### Story 4.6: Implement User HTTP Handlers

**As a** developer,
**I want** HTTP handlers for user CRUD operations,
**So that** users can interact with the API.

**Acceptance Criteria:**

**Given** the service is running
**When** I call `POST /api/v1/users` with valid JSON body:
```json
{"email": "test@example.com", "firstName": "John", "lastName": "Doe"}
```
**Then** I receive HTTP 201 Created
**And** response body is `{"data": {"id": "...", "email": "...", ...}}`
**And** `id` is UUID v7 format

**Given** the service is running
**When** I call `GET /api/v1/users/{id}` with valid ID
**Then** I receive HTTP 200 OK
**And** response body is `{"data": {...}}`

**Given** the service is running
**When** I call `GET /api/v1/users?page=1&pageSize=10`
**Then** I receive HTTP 200 OK
**And** response body includes `{"data": [...], "pagination": {...}}`

**Given** I call `GET /api/v1/users/{id}` with non-existent ID
**When** user is not found
**Then** I receive HTTP 404 with RFC 7807 error (via Story 4.4 mapper)

**And** handlers use error mapper from Story 4.4
**And** handlers use DTOs from Story 4.5
**And** handlers use Chi path parameters `{id}`

*Covers: FR7-9, FR43*

---

### Epic 4 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 4.1 | Implement User Domain Entity and Repository Interface | FR10, FR41 | S |
| 4.2 | Implement User PostgreSQL Repository | FR11 | M |
| 4.3 | Implement User Use Cases | FR10, FR42 | M |
| 4.4 | Implement RFC 7807 Error Response Mapper | FR60-63 | M |
| 4.5 | Implement Transport Contracts (DTOs) | FR43 | S |
| 4.6 | Implement User HTTP Handlers | FR7-9, FR43 | M |

**Total: 6 stories covering 9 FRs**

---

## Epic 5: Security & Authentication Foundation

**Goal:** Developer dapat protect Users endpoints dengan JWT middleware, rate limiting, secure headers, dan request validation.

**FRs covered:** FR25-34

---

### Story 5.1: Implement Request Validation Middleware

**As a** developer,
**I want** automatic request payload validation,
**So that** invalid requests are rejected before reaching handlers.

**Acceptance Criteria:**

**Given** a request with JSON body is decoded into DTO struct
**When** validation runs using `go-playground/validator` tags
**Then** if validation fails, HTTP 400 is returned with RFC 7807 error format
**And** `validationErrors` array contains field names and messages

**Given** request body exceeds size limit (configurable via `MAX_REQUEST_SIZE`, default 1MB)
**When** request is received
**Then** HTTP 413 Request Entity Too Large is returned
**And** request body is NOT fully read into memory

**And** validation happens after JSON decode, before handler logic

*Covers: FR25-27*

---

### Story 5.2: Implement JWT Authentication Middleware

**As a** developer,
**I want** JWT authentication middleware with deterministic time handling,
**So that** I can protect endpoints and test reliably.

**Acceptance Criteria:**

**Given** a request to protected endpoint without `Authorization` header
**When** middleware processes the request
**Then** HTTP 401 Unauthorized is returned
**And** RFC 7807 error with Code="UNAUTHORIZED"

**Given** a request with invalid JWT token (malformed, wrong signature)
**When** middleware validates the token
**Then** HTTP 401 Unauthorized is returned
**And** Code="UNAUTHORIZED" (no detail exposed to client)

**Given** middleware with injected `Clock` / `Now func() time.Time`
**When** validating claim `exp`
**Then** expiry decision uses injected `now()`, NOT `time.Now()` directly

**Given** token with `exp` < `now()`
**When** request is processed
**Then** HTTP 401 Unauthorized is returned
**And** Code="UNAUTHORIZED"

**Given** a request with valid JWT token
**When** middleware validates the token
**Then** claims are extracted and stored in request context
**And** request proceeds to handler
**And** claims are accessible via `ctxutil.GetClaims(ctx)`

**And** JWT secret is loaded from environment variable `JWT_SECRET`
**And** supported algorithms: HS256

*Covers: FR28-29*

---

### Story 5.3: Implement Authorization and Role Checking

**As a** developer,
**I want** authorization checks in the app layer,
**So that** users can only access resources they're allowed to.

**Acceptance Criteria:**

**Given** authenticated user attempts action requiring specific role
**When** use case checks authorization
**Then** authorization is checked in app layer (not middleware)
**And** if unauthorized, AppError with Code="FORBIDDEN" is returned
**And** HTTP 403 Forbidden is returned to client

**Given** user with sufficient permissions
**When** action is authorized
**Then** request proceeds normally

*Covers: FR30-31*

---

### Story 5.4: Implement Security Headers Middleware

**As a** developer,
**I want** security headers in all HTTP responses,
**So that** common web vulnerabilities are mitigated.

**Acceptance Criteria:**

**Given** any HTTP response from the service (success or error)
**When** response is sent
**Then** the following headers are present:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains` (when HTTPS)
- `Content-Security-Policy: default-src 'none'` (API-appropriate)
- `Referrer-Policy: strict-origin-when-cross-origin`

**And** headers are applied via global middleware (first in chain)
**And** headers are present on error responses (4xx, 5xx)

*Covers: FR32*

---

### Story 5.5: Implement Rate Limiting Middleware

**As a** developer,
**I want** rate limiting on API endpoints with per-user and per-IP support,
**So that** the service is protected from abuse.

**Acceptance Criteria:**

**Given** request without JWT (unauthenticated)
**When** rate limit is calculated
**Then** limiter key = resolved client IP

**Given** request with valid JWT
**When** rate limit is calculated
**Then** limiter key = `claims.userId` (per-user rate limiting)

**Given** service behind reverse proxy with `TRUST_PROXY=true`
**When** client IP is resolved
**Then** IP is extracted from `X-Forwarded-For` / `X-Real-IP`

**Given** `TRUST_PROXY=false` (default)
**When** client IP is resolved
**Then** IP is taken from `RemoteAddr`

**Given** client exceeds the rate limit
**When** request is processed
**Then** HTTP 429 Too Many Requests is returned
**And** `Retry-After` header indicates when to retry
**And** RFC 7807 error with Code="RATE_LIMIT_EXCEEDED"

**And** rate limiting uses `go-chi/httprate`
**And** rate limit values are configurable via `RATE_LIMIT_RPS` environment variable

*Covers: FR33-34*

---

### Epic 5 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 5.1 | Implement Request Validation Middleware | FR25-27 | M |
| 5.2 | Implement JWT Authentication Middleware | FR28-29 | M |
| 5.3 | Implement Authorization and Role Checking | FR30-31 | S |
| 5.4 | Implement Security Headers Middleware | FR32 | S |
| 5.5 | Implement Rate Limiting Middleware | FR33-34 | M |

**Total: 5 stories covering 10 FRs**

---

## Epic 6: Audit Trail System

**Goal:** Developer dapat record, store, dan query audit events dengan PII redaction untuk compliance requirements.

**FRs covered:** FR35-39

---

### Story 6.1: Implement Audit Event Domain Model

**As a** developer,
**I want** audit event entity and interfaces in domain layer,
**So that** I have a clear contract for audit logging.

**Acceptance Criteria:**

**Given** the domain layer
**When** I view `internal/domain/audit.go`
**Then** AuditEvent entity exists with fields:
- `ID` (type ID)
- `EventType` (string, pattern: "entity.action", e.g., "user.created")
- `ActorID` (ID, nullable for system/unauthenticated events)
- `EntityType` (string, e.g., "user")
- `EntityID` (ID, what was affected)
- `Payload` ([]byte, JSON, already redacted)
- `Timestamp` (time.Time)
- `RequestID` (string, for correlation)

**And** `AuditEventRepository` interface is defined:
- `Create(ctx context.Context, q Querier, event *AuditEvent) error`
- `ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)`

**And** domain layer has NO external imports (stdlib only)

*Covers: FR35 (partial)*

---

### Story 6.2: Implement Audit Event PostgreSQL Repository

**As a** developer,
**I want** AuditEventRepository implementation for PostgreSQL,
**So that** audit events are persisted to the database.

**Acceptance Criteria:**

**Given** migration file `migrations/YYYYMMDDHHMMSS_create_audit_events.sql`
**When** migration is applied
**Then** table `audit_events` is created with columns:
- `id` (uuid, primary key)
- `event_type` (varchar, indexed)
- `actor_id` (uuid, nullable for system events)
- `entity_type` (varchar, indexed)
- `entity_id` (uuid, indexed)
- `payload` (jsonb)
- `timestamp` (timestamptz, indexed)
- `request_id` (varchar)

**Given** Create is called with valid audit event
**When** the operation succeeds
**Then** event is inserted into `audit_events` table
**And** `domain.ID` is parsed to `uuid.UUID` at repository boundary
**And** payload []byte is stored as JSONB

**Given** ListByEntityID is called
**When** query executes
**Then** results are filtered by entity_type and entity_id
**And** ordered by `timestamp DESC`
**And** total count is returned for pagination

*Covers: FR36*

---

### Story 6.3: Implement PII Redaction Service

**As a** developer,
**I want** automatic PII redaction in audit payloads,
**So that** sensitive data is never stored in audit logs.

**Acceptance Criteria:**

**Given** audit payload contains PII fields
**When** redaction service processes the payload
**Then** following fields are fully redacted (replaced with `"[REDACTED]"`):
- `password`
- `token`
- `secret`
- `authorization`
- `creditCard` / `credit_card`
- `ssn`
- `email` (default: full redact)

**Given** configuration `AUDIT_REDACT_EMAIL=partial`
**When** email is redacted
**Then** email shows partial mask: `ab***@domain.com` (first 2 chars + domain)

**Given** `AUDIT_REDACT_EMAIL=full` (default)
**When** email is redacted
**Then** email is replaced with `"[REDACTED]"`

**And** redaction happens BEFORE conversion to []byte JSON
**And** original payload is NOT stored anywhere

*Covers: FR37*

---

### Story 6.4: Implement Audit Event Service

**As a** developer,
**I want** a service to record and query audit events synchronously,
**So that** business operations are tracked for compliance.

**Acceptance Criteria:**

**Given** the app layer
**When** I view `internal/app/audit/`
**Then** `AuditService` exists with methods:
- `Record(ctx context.Context, q Querier, event AuditEvent) error`
- `ListByEntity(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)`

**Given** Record is called within a transaction
**When** event is processed
**Then** PII redaction is applied to payload
**And** requestID is extracted from context
**And** ActorID is extracted from auth claims (nullable if no auth)
**And** event is persisted via repository in the SAME transaction

**Given** audit insert fails
**When** business transaction attempts to commit
**Then** transaction is rolled back
**And** API returns 500 with Code="INTERNAL_ERROR"

**Given** user creates a new user via API
**When** CreateUserUseCase completes successfully
**Then** audit event with type="user.created" is recorded in same transaction

*Covers: FR35, FR38*

---

### Story 6.5: Enable Extensible Audit Event Types

**As a** developer,
**I want** to easily add new audit event types,
**So that** I can extend auditing for new modules.

**Acceptance Criteria:**

**Given** a developer wants to add new module auditing
**When** they follow the audit pattern
**Then** they can define new event type constants: `const EventOrderCreated = "order.created"`
**And** call `auditService.Record()` from their use case within transaction
**And** new events appear in audit_events table

**Given** the audit module
**When** I view code comments or package documentation
**Then** example shows how to add audit events for new entity types
**And** pattern "entity.action" is documented

*Covers: FR39*

---

### Epic 6 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 6.1 | Implement Audit Event Domain Model | FR35 | S |
| 6.2 | Implement Audit Event PostgreSQL Repository | FR36 | M |
| 6.3 | Implement PII Redaction Service | FR37 | M |
| 6.4 | Implement Audit Event Service | FR35, FR38 | M |
| 6.5 | Enable Extensible Audit Event Types | FR39 | S |

**Total: 5 stories covering 5 FRs**

---

## Epic 7: CI/CD Pipeline

**Goal:** Developer dapat push code dan CI pipeline (lint → test → build → boundary-check → migrations) berjalan otomatis.

**FRs covered:** FR46, FR52-55

---

### Story 7.1: Setup GitHub Actions Workflow

**As a** developer,
**I want** CI pipeline that runs on every push and PR,
**So that** code quality is automatically verified.

**Acceptance Criteria:**

**Given** I push code to any branch
**When** push event triggers
**Then** GitHub Actions workflow `.github/workflows/ci.yml` runs

**Given** I open a Pull Request
**When** PR is created or updated
**Then** CI workflow runs and reports status to PR

**Given** CI workflow runs
**When** any step fails
**Then** workflow fails fast and reports failure to PR
**And** subsequent steps are skipped

**And** Go modules are cached between runs (actions/cache)
**And** Go build cache is enabled

*Covers: FR46 (partial)*

---

### Story 7.2: Implement CI Lint and Test Steps

**As a** developer,
**I want** lint and test steps in CI with coverage enforcement,
**So that** code quality and correctness are enforced.

**Acceptance Criteria:**

**Given** CI workflow runs
**When** lint step executes
**Then** `golangci-lint run` is executed with project config
**And** boundary violations (depguard) cause step to fail

**Given** CI workflow runs
**When** test step executes
**Then** `go test -race -coverprofile=coverage.out ./...` runs
**And** test failures cause step to fail
**And** coverage report is uploaded as artifact

**Given** coverage for `internal/domain/...` and `internal/app/...` is below 80%
**When** coverage check runs
**Then** step fails with clear message indicating coverage gap

*Covers: FR46*

---

### Story 7.3: Implement CI Build and Security Scan

**As a** developer,
**I want** build verification and security scanning in CI,
**So that** binaries are buildable and dependencies are secure.

**Acceptance Criteria:**

**Given** CI workflow runs
**When** build step executes
**Then** `go build -o /dev/null ./cmd/api` succeeds

**Given** CI workflow runs
**When** security scan step executes
**Then** `govulncheck ./...` runs
**And** known vulnerabilities cause step to **FAIL** (default behavior)

**Given** Dockerfile exists in repository
**When** Docker build step runs (conditional)
**Then** `docker build .` succeeds without pushing

**Given** Dockerfile does NOT exist
**When** CI runs
**Then** Docker build step is skipped

*Covers: FR46 (partial)*

---

### Story 7.4: Implement CI Migration Verification

**As a** developer,
**I want** migrations verified in CI,
**So that** database changes are validated before merge.

**Acceptance Criteria:**

**Given** CI workflow runs
**When** migration step executes
**Then** PostgreSQL service container is started (services: postgres)
**And** database is empty (clean state)

**Given** clean database
**When** migration verification runs
**Then** `goose up` applies ALL migrations successfully
**And** `goose down` rolls back ALL migrations to version 0
**And** full up/down cycle completes without error

**Given** migration has syntax error or fails
**When** migration step runs
**Then** step fails with clear error message

*Covers: FR55*

---

### Story 7.5: Create Migration Helper Commands

**As a** developer,
**I want** make commands for managing migrations,
**So that** I can create and manage migrations easily.

**Acceptance Criteria:**

**Given** I want to create a new migration
**When** I run `make migrate-create NAME=add_orders_table`
**Then** new migration file is created: `migrations/YYYYMMDDHHMMSS_add_orders_table.sql`
**And** file contains template with `-- +goose Up` and `-- +goose Down` sections

**Given** I want to check migration status
**When** I run `make migrate-status`
**Then** I see list of applied and pending migrations

*Covers: FR52-54*

---

### Epic 7 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 7.1 | Setup GitHub Actions Workflow | FR46 | S |
| 7.2 | Implement CI Lint and Test Steps | FR46 | M |
| 7.3 | Implement CI Build and Security Scan | FR46 | S |
| 7.4 | Implement CI Migration Verification | FR55 | M |
| 7.5 | Create Migration Helper Commands | FR52-54 | S |

**Total: 5 stories covering 5 FRs**

---

## Epic 8: Documentation & Developer Guides

**Goal:** Developer dapat self-service: quick start, architecture, local dev, observability, add module, add adapter - tanpa bantuan.

**FRs covered:** FR64-69

---

### Story 8.1: Create README Quick Start

**As a** developer,
**I want** a README with quick start instructions,
**So that** I can get the service running without external help.

**Acceptance Criteria:**

**Given** I clone the repository
**When** I read `README.md`
**Then** I see clear quick start section with commands:
1. Prerequisites (Go 1.23+, Docker)
2. `make setup`
3. `make infra-up`
4. `make migrate-up`
5. `make run`
6. Verify endpoints:
   - `curl http://localhost:8080/health` → 200
   - `curl http://localhost:8080/ready` → 200 (DB connected)
   - `curl http://localhost:8080/metrics` → Prometheus format

**And** optional "First API Call" section:
- `POST /api/v1/users` with sample body
- `GET /api/v1/users/{id}`
- `GET /api/v1/users`

**And** all commands work without modification
**And** estimated time is mentioned (< 15 minutes)

*Covers: FR64*

---

### Story 8.2: Document Architecture and Layer Responsibilities

**As a** developer,
**I want** architecture documentation explaining hexagonal structure,
**So that** I understand where to place code.

**Acceptance Criteria:**

**Given** I want to understand the architecture
**When** I read `docs/architecture.md`
**Then** I see explanation of:
- Hexagonal architecture principles
- Four layers: domain, app, transport, infra
- Responsibilities of each layer
- Import rules between layers
- Diagram showing layer relationships

**And** subsection "Boundary Enforcement in Practice" showing:
- depguard/golangci config location
- Example error message when boundary violated
- Reference to FR45-FR47

*Covers: FR65*

---

### Story 8.3: Document Local Development Workflow

**As a** developer,
**I want** local development guide,
**So that** I know daily workflow and troubleshooting.

**Acceptance Criteria:**

**Given** I want to develop locally
**When** I read `docs/local-development.md`
**Then** I see:
- Daily workflow commands (run, test, lint)
- Hot reload options (if any)
- Database management (migrations, reset)
- Troubleshooting common issues
- IDE setup recommendations (VS Code, GoLand)

*Covers: FR66*

---

### Story 8.4: Document Observability Configuration

**As a** developer,
**I want** observability documentation,
**So that** I can configure logging, tracing, and metrics.

**Acceptance Criteria:**

**Given** I want to configure observability
**When** I read `docs/observability.md`
**Then** I see:
- Logging configuration (`LOG_LEVEL`, JSON format)
- Tracing setup (`OTEL_ENABLED`, `OTEL_EXPORTER_OTLP_ENDPOINT`)
- Metrics endpoint (`/metrics`)
- Request correlation explanation

**And** example log output showing correlation:
```json
{"time":"...","level":"info","msg":"user created","service":"golang-api-hexagonal","env":"development","requestId":"abc123","traceId":"def456","spanId":"..."}
```

**And** optional docker-compose snippet with Jaeger/Grafana

*Covers: FR67*

---

### Story 8.5: Create Guide for Adding New Modules

**As a** developer,
**I want** step-by-step guide for adding new modules,
**So that** I can extend the application correctly.

**Acceptance Criteria:**

**Given** I want to add a new module (e.g., "orders")
**When** I read `docs/guides/adding-module.md`
**Then** I see step-by-step instructions:
1. Create domain entity and repository interface (`internal/domain/order.go`)
2. Create migration (`make migrate-create NAME=create_orders`)
3. Implement repository (`internal/infra/postgres/order_repo.go`)
4. Create use cases (`internal/app/order/`)
5. Create DTOs (`internal/transport/http/contract/order.go`)
6. Create handlers (`internal/transport/http/handler/order.go`)
7. Wire routes in router
8. Add audit events
9. Write tests

**And** each step references existing Users module as example

*Covers: FR68*

---

### Story 8.6: Create Guide for Adding New Adapters

**As a** developer,
**I want** guide for adding new infrastructure adapters,
**So that** I can integrate external services correctly.

**Acceptance Criteria:**

**Given** I want to add a new adapter (e.g., Redis cache, email service)
**When** I read `docs/guides/adding-adapter.md`
**Then** I see:
- Where adapters live (`internal/infra/`)
- Interface definition in domain or app layer
- Implementation in infra layer
- Configuration via environment variables
- Testing strategy (mocks, testcontainers)
- Wiring in main.go

*Covers: FR69*

---

### Epic 8 Summary

| Story | Title | FRs Covered | Size |
|-------|-------|-------------|------|
| 8.1 | Create README Quick Start | FR64 | S |
| 8.2 | Document Architecture and Layer Responsibilities | FR65 | M |
| 8.3 | Document Local Development Workflow | FR66 | S |
| 8.4 | Document Observability Configuration | FR67 | S |
| 8.5 | Create Guide for Adding New Modules | FR68 | M |
| 8.6 | Create Guide for Adding New Adapters | FR69 | S |

**Total: 6 stories covering 6 FRs**

---

## Final Summary

### Epic Overview

| Epic | Title | Stories | FRs Covered |
|------|-------|---------|-------------|
| 1 | Project Foundation & First Run Experience | 6 | 14 |
| 2 | Observability Stack | 6 | 10 |
| 3 | Local Quality Gates | 4 | 6 |
| 4 | Reference Implementation (Users Module) | 6 | 9 |
| 5 | Security & Authentication Foundation | 5 | 10 |
| 6 | Audit Trail System | 5 | 5 |
| 7 | CI/CD Pipeline | 5 | 5 |
| 8 | Documentation & Developer Guides | 6 | 6 |
| **Total** | | **43** | **69** |

---

## FR → Story Traceability Matrix

| FR | Description | Epic | Story |
|----|-------------|------|-------|
| FR1 | Clone → working service < 30 min | 1 | 1.6 |
| FR2 | Single setup command | 1 | 1.6 |
| FR3 | Single command for infra | 1 | 1.3 |
| FR4 | Single command for migrations | 1 | 1.4 |
| FR5 | Single command to start service | 1 | 1.6 |
| FR6 | Health endpoint verification | 1 | 1.5 |
| FR7 | Create user via HTTP API | 4 | 4.6 |
| FR8 | Get user by ID | 4 | 4.6 |
| FR9 | List users | 4 | 4.6 |
| FR10 | Users as reference pattern | 4 | 4.1, 4.3 |
| FR11 | Trace request flow | 4 | 4.2 |
| FR12 | Structured JSON logs | 2 | 2.1 |
| FR13 | request_id in logs | 2 | 2.2 |
| FR14 | trace_id in logs | 2 | 2.3b |
| FR15 | request_id in response headers | 2 | 2.2 |
| FR16 | OpenTelemetry traces | 2 | 2.3a |
| FR17 | Trace context propagation | 2 | 2.3a |
| FR18 | Trace export configuration | 2 | 2.3a |
| FR19 | Prometheus metrics endpoint | 2 | 2.4 |
| FR20 | Default HTTP metrics | 2 | 2.4 |
| FR21 | Custom metrics utilities | 2 | 2.5 |
| FR22 | Health endpoint | 1 | 1.5 |
| FR23 | Readiness endpoint | 1 | 1.5 |
| FR24 | Health/readiness responses | 1 | 1.5 |
| FR25 | Request payload validation | 5 | 5.1 |
| FR26 | Validation error responses | 5 | 5.1 |
| FR27 | Request size limits | 5 | 5.1 |
| FR28 | JWT authentication middleware | 5 | 5.2 |
| FR29 | 401 for invalid auth | 5 | 5.2 |
| FR30 | 403 for unauthorized | 5 | 5.3 |
| FR31 | Authorization in use case | 5 | 5.3 |
| FR32 | Security headers | 5 | 5.4 |
| FR33 | Rate limiting | 5 | 5.5 |
| FR34 | 429 for rate limit exceeded | 5 | 5.5 |
| FR35 | Record audit events | 6 | 6.1, 6.4 |
| FR36 | Store audit events in DB | 6 | 6.2 |
| FR37 | PII redaction | 6 | 6.3 |
| FR38 | Query audit events | 6 | 6.4 |
| FR39 | Extensible audit types | 6 | 6.5 |
| FR40 | Hexagonal architecture | 1 | 1.1 |
| FR41 | Domain layer no external deps | 1, 4 | 1.1, 4.1 |
| FR42 | App layer use cases | 4 | 4.3 |
| FR43 | Transport layer HTTP concerns | 4 | 4.5, 4.6 |
| FR44 | Infra implements interfaces | 1 | 1.1 |
| FR45 | Lint rules detect violations | 3 | 3.1 |
| FR46 | CI fails on violations | 7 | 7.1, 7.2, 7.3 |
| FR47 | Clear violation error messages | 3 | 3.1 |
| FR48 | Full test suite with race | 3 | 3.2 |
| FR49 | Local linting checks | 3 | 3.3 |
| FR50 | Local CI pipeline | 3 | 3.3 |
| FR51 | Start/stop infra containers | 3 | 3.4 |
| FR52 | Create migration files | 7 | 7.5 |
| FR53 | Apply migrations | 7 | 7.5 |
| FR54 | Rollback migrations | 7 | 7.5 |
| FR55 | Migrations in CI | 7 | 7.4 |
| FR56 | Env var configuration | 1 | 1.2 |
| FR57 | Sensible defaults | 1 | 1.2 |
| FR58 | Validate required config | 1 | 1.2 |
| FR59 | Fail fast on invalid config | 1 | 1.2 |
| FR60 | Standardized error format | 4 | 4.4 |
| FR61 | Client vs server errors | 4 | 4.4 |
| FR62 | Error codes | 4 | 4.4 |
| FR63 | Log errors without exposing details | 4 | 4.4 |
| FR64 | README quick start | 8 | 8.1 |
| FR65 | Architecture documentation | 8 | 8.2 |
| FR66 | Local development docs | 8 | 8.3 |
| FR67 | Observability docs | 8 | 8.4 |
| FR68 | Adding new modules guide | 8 | 8.5 |
| FR69 | Adding new adapters guide | 8 | 8.6 |

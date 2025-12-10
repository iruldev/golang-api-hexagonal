---
stepsCompleted: [1, 2, 3, 4]
status: 'complete'
completedAt: '2025-12-10'
inputDocuments:
  - 'docs/prd.md'
  - 'docs/architecture.md'
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-10'
---

# Backend Service Golang Boilerplate - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for Backend Service Golang Boilerplate, decomposing the requirements from the PRD and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

**1. Project Setup & Initialization (FR1-FR5)**
- FR1: Developer can clone boilerplate and run service locally within 30 minutes
- FR2: Developer can configure service via environment variables or config file
- FR3: Developer can start all dependencies with single command (`make dev`)
- FR4: Developer can view example environment configuration via `.env.example`
- FR5: System initializes with graceful shutdown handling on OS signals

**2. Configuration Management (FR6-FR10)**
- FR6: System can load configuration from environment variables
- FR7: System can load configuration from config file (optional)
- FR8: System validates configuration at startup and fails fast on invalid config
- FR9: System binds configuration to typed struct for type safety
- FR10: Developer can see clear error messages when configuration is invalid

**3. HTTP API Foundation (FR11-FR18)**
- FR11: System exposes versioned HTTP API endpoints (`/api/v1/...`)
- FR12: System generates unique request ID for each incoming request
- FR13: System logs all HTTP requests with structured fields
- FR14: System recovers from panics in handlers and returns 500 response
- FR15: System propagates trace context via OpenTelemetry middleware
- FR16: Developer can add new HTTP endpoints following documented patterns
- FR17: System returns consistent response envelope for success/error
- FR18: System maps application errors to appropriate HTTP status codes

**4. Database & Persistence (FR19-FR25)**
- FR19: System connects to PostgreSQL database with connection pooling
- FR20: System handles database connection timeouts gracefully
- FR21: Developer can write type-safe SQL queries using sqlc
- FR22: Developer can run database migrations via `make migrate-up`
- FR23: Developer can rollback database migrations via `make migrate-down`
- FR24: Developer can generate repository code via `make gen`
- FR25: System checks database connectivity as part of readiness check

**5. Observability & Monitoring (FR26-FR34)**
- FR26: System exposes liveness endpoint (`/healthz`) returning 200
- FR27: System exposes readiness endpoint (`/readyz`) with dependency check
- FR28: System returns 503 on readiness when database fails
- FR29: System exposes Prometheus metrics endpoint (`/metrics`)
- FR30: System captures HTTP request count, latency, error count
- FR31: System produces structured JSON logs in production
- FR32: System produces human-readable logs in development
- FR33: System includes trace_id, request_id, path, method, status in logs
- FR34: System creates OpenTelemetry spans for HTTP requests

**6. Developer Experience & Tooling (FR35-FR42)**
- FR35: Developer can run all tests with `make test`
- FR36: Developer can run linter with `make lint`
- FR37: System provides golangci-lint configuration
- FR38: Developer can view CI pipeline example
- FR39: Developer can follow example module as pattern
- FR40: Developer can understand architecture via ARCHITECTURE.md
- FR41: AI assistants can follow AGENTS.md for consistent code
- FR42: Developer can copy example module for new domain

**7. Extension & Hooks (FR43-FR48)**
- FR43: System defines Logger interface
- FR44: System defines Cache interface
- FR45: System defines RateLimiter interface
- FR46: System defines EventPublisher interface
- FR47: System defines SecretProvider interface
- FR48: Developer can implement custom providers

**8. Sample Module (FR49-FR56)**
- FR49: System includes complete example module
- FR50: Example includes entity with validation
- FR51: Example includes SQL migration
- FR52: Example includes sqlc repository
- FR53: Example includes usecase logic
- FR54: Example includes HTTP handler
- FR55: Example includes unit tests
- FR56: Example includes integration test

### Non-Functional Requirements

**Performance (NFR1-NFR5)**
- NFR1: Setup time < 30 minutes (clone → running)
- NFR2: `make dev` startup < 60 seconds
- NFR3: HTTP response < 100ms p95
- NFR4: DB query latency < 50ms p95
- NFR5: Graceful shutdown < 30 seconds

**Reliability (NFR6-NFR9)**
- NFR6: Panic recovery 100% (no server crash)
- NFR7: Shutdown completes all in-flight requests
- NFR8: Honest /readyz (503 when DB down)
- NFR9: Config validation fails fast

**Maintainability (NFR10-NFR14)**
- NFR10: 100% lint pass
- NFR11: Zero circular imports
- NFR12: Hexagonal layer separation
- NFR13: Cyclomatic complexity ≤ 15
- NFR14: File size ≤ 500 LOC

**Testability (NFR15-NFR18)**
- NFR15: Unit coverage ≥ 70% on example module
- NFR16: `make test` works
- NFR17: Tests are independent (parallel safe)
- NFR18: At least 1 integration test (HTTP+DB)

**Observability (NFR19-NFR23)**
- NFR19: JSON log format in production
- NFR20: trace_id, request_id in logs
- NFR21: Prometheus metrics exposed
- NFR22: OTEL tracing ready
- NFR23: Health endpoint latency ≤ 10ms

**Security (NFR24-NFR27)**
- NFR24: Secrets via env/file only (no hardcode)
- NFR25: No stack traces in error responses
- NFR26: All inputs validated
- NFR27: No critical CVE in dependencies

**Developer Experience (NFR28-NFR31)**
- NFR28: Documentation accuracy 100%
- NFR29: Example clarity (follow → works)
- NFR30: Clear, actionable error messages
- NFR31: AI compatibility via AGENTS.md

### Additional Requirements (from Architecture)

- Hexagonal/Clean Architecture layer separation
- chi router (go-chi/chi/v5)
- pgx + sqlc for database queries
- golang-migrate for migrations
- zap structured logger
- koanf configuration management
- OpenTelemetry tracing
- docker-compose for local development
- Makefile targets: dev, test, lint, migrate-up, migrate-down, gen

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1 | Epic 1 | Clone and run within 30 minutes |
| FR2 | Epic 1 | Configure via env/config |
| FR3 | Epic 1 | Single command start (`make dev`) |
| FR4 | Epic 1 | `.env.example` template |
| FR5 | Epic 1 | Graceful shutdown |
| FR6 | Epic 2 | Load config from env |
| FR7 | Epic 2 | Load config from file |
| FR8 | Epic 2 | Validate config at startup |
| FR9 | Epic 2 | Typed config struct |
| FR10 | Epic 2 | Clear config error messages |
| FR11 | Epic 3 | Versioned API endpoints |
| FR12 | Epic 3 | Request ID generation |
| FR13 | Epic 3 | HTTP request logging |
| FR14 | Epic 3 | Panic recovery |
| FR15 | Epic 3 | OTEL trace propagation |
| FR16 | Epic 3 | Documented endpoint patterns |
| FR17 | Epic 3 | Consistent response envelope |
| FR18 | Epic 3 | Error to HTTP status mapping |
| FR19 | Epic 4 | PostgreSQL connection pooling |
| FR20 | Epic 4 | Connection timeout handling |
| FR21 | Epic 4 | Type-safe SQL with sqlc |
| FR22 | Epic 4 | `make migrate-up` |
| FR23 | Epic 4 | `make migrate-down` |
| FR24 | Epic 4 | `make gen` for repository |
| FR25 | Epic 4 | DB readiness check |
| FR26 | Epic 5 | `/healthz` liveness |
| FR27 | Epic 5 | `/readyz` readiness |
| FR28 | Epic 5 | 503 when DB down |
| FR29 | Epic 5 | `/metrics` endpoint |
| FR30 | Epic 5 | HTTP metrics capture |
| FR31 | Epic 5 | JSON logs (prod) |
| FR32 | Epic 5 | Human logs (dev) |
| FR33 | Epic 5 | Log context fields |
| FR34 | Epic 5 | OTEL spans |
| FR35 | Epic 1 | `make test` |
| FR36 | Epic 1 | `make lint` |
| FR37 | Epic 1 | golangci-lint config |
| FR38 | Epic 1 | CI pipeline example |
| FR39 | Epic 7 | Example module as pattern |
| FR40 | Epic 7 | ARCHITECTURE.md |
| FR41 | Epic 7 | AGENTS.md |
| FR42 | Epic 7 | Copy example for new domain |
| FR43 | Epic 6 | Logger interface |
| FR44 | Epic 6 | Cache interface |
| FR45 | Epic 6 | RateLimiter interface |
| FR46 | Epic 6 | EventPublisher interface |
| FR47 | Epic 6 | SecretProvider interface |
| FR48 | Epic 6 | Custom provider support |
| FR49 | Epic 7 | Complete example module |
| FR50 | Epic 7 | Entity with validation |
| FR51 | Epic 7 | SQL migration |
| FR52 | Epic 7 | sqlc repository |
| FR53 | Epic 7 | Usecase logic |
| FR54 | Epic 7 | HTTP handler |
| FR55 | Epic 7 | Unit tests |
| FR56 | Epic 7 | Integration test |

## Epic List

### Epic 1: Project Foundation & DX Setup
Developer can clone, configure, and run service in <30 minutes with full development tooling.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR35, FR36, FR37, FR38

### Epic 2: Configuration & Environment
System boots with validated configuration from environment or file, fails fast on errors.
**FRs covered:** FR6, FR7, FR8, FR9, FR10

### Epic 3: HTTP API Core
System handles HTTP requests with tracing, logging, panic recovery, and consistent responses.
**FRs covered:** FR11, FR12, FR13, FR14, FR15, FR16, FR17, FR18

### Epic 4: Database & Persistence
System connects to PostgreSQL with type-safe queries, migrations, and connection management.
**FRs covered:** FR19, FR20, FR21, FR22, FR23, FR24, FR25

### Epic 5: Observability Suite
Service is fully observable with health endpoints, metrics, structured logging, and tracing.
**FRs covered:** FR26, FR27, FR28, FR29, FR30, FR31, FR32, FR33, FR34

### Epic 6: Extension Interfaces
Service provides hook points (interfaces) for enterprise adapters and custom implementations.
**FRs covered:** FR43, FR44, FR45, FR46, FR47, FR48

### Epic 7: Sample Module (Note)
Developer has complete reference implementation demonstrating all patterns to copy for new domains.
**FRs covered:** FR39, FR40, FR41, FR42, FR49, FR50, FR51, FR52, FR53, FR54, FR55, FR56

---

## Epic 1: Project Foundation & DX Setup

Developer can clone, configure, and run service in <30 minutes with full development tooling.

### Story 1.1: Initialize Go Module Structure

As a developer,
I want to clone the boilerplate and see a proper Go module structure,
So that I can start working on domain logic immediately.

**Acceptance Criteria:**

**Given** a fresh clone of the repository
**When** I run `go mod download`
**Then** all dependencies are fetched successfully
**And** the project compiles with `go build ./...`

### Story 1.2: Create Makefile & Docker Compose

As a developer,
I want to start all dependencies with a single command,
So that I don't need to manually configure external services.

**Acceptance Criteria:**

**Given** docker and docker-compose are installed
**When** I run `make dev`
**Then** PostgreSQL container starts on port 5432
**And** Jaeger container starts on port 16686 (optional)
**And** the Go application compiles and runs

**Given** I want to run tests
**When** I run `make test`
**Then** all tests execute with coverage report

**Given** I want to check code quality
**When** I run `make lint`
**Then** golangci-lint runs with project configuration

### Story 1.3: Setup Environment Configuration

As a developer,
I want to configure the service via environment variables,
So that I can adapt settings without code changes.

**Acceptance Criteria:**

**Given** `.env.example` exists with all required variables
**When** I copy `.env.example` to `.env`
**Then** the application starts with default configuration

**Given** `APP_ENV=development` is set
**When** the application starts
**Then** human-readable logs are produced

### Story 1.4: Implement Graceful Shutdown

As a SRE,
I want the service to shutdown gracefully,
So that in-flight requests complete before termination.

**Acceptance Criteria:**

**Given** the application is running with active requests
**When** SIGTERM or SIGINT is received
**Then** the server stops accepting new connections
**And** existing requests have up to 30 seconds to complete
**And** the application exits with code 0

### Story 1.5: Add Linting & CI Configuration

As a tech lead,
I want consistent linting rules and CI example,
So that all code follows the same quality standards.

**Acceptance Criteria:**

**Given** `.golangci.yml` exists in project root
**When** I run `make lint`
**Then** golangci-lint uses the project configuration
**And** no lint errors on clean codebase

**Given** `.github/workflows/ci.yml` exists
**When** I push to main branch
**Then** CI runs tests and linting

---

## Epic 2: Configuration & Environment

System boots with validated configuration from environment or file, fails fast on errors.

### Story 2.1: Implement Environment Variable Loading

As a developer,
I want the system to load configuration from environment variables,
So that I can configure the service without modifying files.

**Acceptance Criteria:**

**Given** environment variables `APP_PORT`, `DB_HOST`, `DB_PORT` are set
**When** the application starts
**Then** configuration is loaded from environment variables
**And** the Config struct is populated correctly

### Story 2.2: Add Optional Config File Support

As a developer,
I want to optionally load configuration from a YAML/JSON file,
So that I can use config files in certain deployment scenarios.

**Acceptance Criteria:**

**Given** `APP_CONFIG_FILE` environment variable is set
**When** the application starts
**Then** configuration is loaded from the specified file
**And** environment variables override file values

**Given** no config file is specified
**When** the application starts
**Then** only environment variables are used

### Story 2.3: Implement Config Validation with Fail-Fast

As a SRE,
I want the system to fail fast on invalid configuration,
So that misconfigurations are caught at startup, not runtime.

**Acceptance Criteria:**

**Given** required config `DB_HOST` is missing
**When** the application starts
**Then** startup fails with exit code 1
**And** error message indicates `DB_HOST is required`

**Given** `APP_PORT` is set to "invalid"
**When** the application starts
**Then** startup fails with validation error
**And** error message indicates type mismatch

### Story 2.4: Create Typed Config Struct

As a developer,
I want configuration bound to typed Go structs,
So that I get compile-time safety and IDE autocomplete.

**Acceptance Criteria:**

**Given** `internal/config/config.go` exists
**When** I import the config package
**Then** I can access `cfg.Server.Port` as `int`
**And** I can access `cfg.Database.Host` as `string`
**And** nested structures like `cfg.Observability.LogLevel` work

### Story 2.5: Provide Clear Config Error Messages

As a developer,
I want clear error messages when configuration is invalid,
So that I can quickly fix configuration issues.

**Acceptance Criteria:**

**Given** `DB_MAX_OPEN_CONNS` is set to -5
**When** the application starts
**Then** error message shows: `DB_MAX_OPEN_CONNS must be positive`

**Given** multiple config errors exist
**When** the application starts
**Then** all errors are listed in the output
**And** exit code is 1

---

## Epic 3: HTTP API Core

System handles HTTP requests with tracing, logging, panic recovery, and consistent responses.

### Story 3.1: Setup Chi Router with Versioned API

As a developer,
I want versioned API endpoints under `/api/v1/`,
So that I can evolve the API without breaking clients.

**Acceptance Criteria:**

**Given** the HTTP server is running
**When** I request `GET /api/v1/health`
**Then** response status is 200
**And** route is mounted under versioned prefix

### Story 3.2: Implement Request ID Middleware

As a SRE,
I want unique request IDs for each request,
So that I can trace requests across systems.

**Acceptance Criteria:**

**Given** a request without `X-Request-ID` header
**When** the request is processed
**Then** a UUID request ID is generated
**And** `X-Request-ID` header is set in response

**Given** a request with `X-Request-ID` header
**When** the request is processed
**Then** the provided ID is used
**And** same ID is returned in response

### Story 3.3: Add HTTP Request Logging Middleware

As a SRE,
I want all HTTP requests logged with structured fields,
So that I can monitor and debug traffic.

**Acceptance Criteria:**

**Given** any HTTP request is made
**When** the request completes
**Then** log entry contains: method, path, status, latency_ms, request_id

**Given** `APP_ENV=production`
**When** request is logged
**Then** output is JSON format

### Story 3.4: Implement Panic Recovery Middleware

As a SRE,
I want the system to recover from panics,
So that a single panic doesn't crash the server.

**Acceptance Criteria:**

**Given** a handler panics
**When** the panic occurs
**Then** response status is 500
**And** response body is generic error (no stack trace)
**And** panic is logged with stack trace
**And** server continues handling other requests

### Story 3.5: Add OpenTelemetry Trace Propagation

As a SRE,
I want trace context propagated via OTEL,
So that I can trace requests across services.

**Acceptance Criteria:**

**Given** OTEL exporter is configured
**When** a request is processed
**Then** span is created with request details
**And** trace_id is available in context
**And** child spans can be created in handlers

### Story 3.6: Create Handler Registration Pattern

As a developer,
I want documented patterns for adding endpoints,
So that I can add new handlers consistently.

**Acceptance Criteria:**

**Given** `internal/interface/http/routes.go` exists
**When** I follow the pattern to add a new handler
**Then** the route is registered under `/api/v1/`
**And** middleware chain is applied automatically

### Story 3.7: Implement Response Envelope Pattern

As a developer,
I want consistent response format for success/error,
So that clients can parse responses uniformly.

**Acceptance Criteria:**

**Given** a successful request
**When** response is sent
**Then** body is `{"success": true, "data": {...}}`

**Given** an error occurs
**When** response is sent
**Then** body is `{"success": false, "error": {"code": "ERR_*", "message": "..."}}`

### Story 3.8: Create Error to HTTP Status Mapping

As a developer,
I want application errors mapped to HTTP status codes,
So that clients receive appropriate status codes.

**Acceptance Criteria:**

**Given** `ErrNotFound` domain error
**When** handler returns this error
**Then** HTTP status is 404

**Given** `ErrValidation` domain error
**When** handler returns this error
**Then** HTTP status is 400

**Given** `ErrUnauthorized` domain error
**When** handler returns this error
**Then** HTTP status is 401

---

## Epic 4: Database & Persistence

System connects to PostgreSQL with type-safe queries, migrations, and connection management.

### Story 4.1: Setup PostgreSQL Connection with pgx

As a developer,
I want the system to connect to PostgreSQL with connection pooling,
So that database connections are efficiently managed.

**Acceptance Criteria:**

**Given** valid `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
**When** the application starts
**Then** connection pool is established
**And** `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS` are respected

### Story 4.2: Handle Database Connection Timeouts

As a SRE,
I want database connections to timeout gracefully,
So that slow database doesn't hang the application.

**Acceptance Criteria:**

**Given** database is unreachable
**When** connection attempt times out
**Then** error is logged with context
**And** application fails startup gracefully

**Given** query takes longer than timeout
**When** context is cancelled
**Then** query is cancelled
**And** appropriate error is returned

### Story 4.3: Configure sqlc for Type-Safe Queries

As a developer,
I want to write type-safe SQL queries using sqlc,
So that SQL errors are caught at compile time.

**Acceptance Criteria:**

**Given** `sqlc.yaml` configuration exists
**When** I run `make gen`
**Then** Go code is generated from SQL queries
**And** generated code has typed parameters and returns

**Given** `db/queries/*.sql` contains a query
**When** I run `make gen`
**Then** corresponding Go function is generated

### Story 4.4: Implement Database Migrations (Up)

As a developer,
I want to run database migrations via `make migrate-up`,
So that schema changes are applied consistently.

**Acceptance Criteria:**

**Given** migrations exist in `db/migrations/`
**When** I run `make migrate-up`
**Then** pending migrations are applied in order
**And** migration version is tracked

**Given** migrations are already applied
**When** I run `make migrate-up`
**Then** no migrations are re-applied
**And** output shows "no change"

### Story 4.5: Implement Database Migrations (Down)

As a developer,
I want to rollback migrations via `make migrate-down`,
So that I can undo schema changes.

**Acceptance Criteria:**

**Given** migrations have been applied
**When** I run `make migrate-down`
**Then** last migration is rolled back
**And** migration version is updated

**Given** I run `make migrate-down N=2`
**When** command completes
**Then** 2 migrations are rolled back

### Story 4.6: Add Repository Code Generation

As a developer,
I want to generate repository code via `make gen`,
So that boilerplate code is automated.

**Acceptance Criteria:**

**Given** `db/queries/note.sql` contains queries
**When** I run `make gen`
**Then** `internal/infra/postgres/note/queries.go` is generated
**And** generated code compiles without errors

### Story 4.7: Add Database Readiness Check

As a SRE,
I want database connectivity checked in readiness probe,
So that unhealthy instances are removed from load balancer.

**Acceptance Criteria:**

**Given** database is connected
**When** `/readyz` is requested
**Then** response is 200

**Given** database connection is lost
**When** `/readyz` is requested
**Then** response is 503
**And** response body indicates database unavailable

---

## Epic 5: Observability Suite

Service is fully observable with health endpoints, metrics, structured logging, and tracing.

### Story 5.1: Implement Liveness Endpoint

As a Kubernetes operator,
I want `/healthz` endpoint returning 200,
So that I can check if the service is alive.

**Acceptance Criteria:**

**Given** the HTTP server is running
**When** I request `GET /healthz`
**Then** response status is 200
**And** response body is `{"status": "ok"}`
**And** latency is < 10ms

### Story 5.2: Implement Readiness Endpoint

As a Kubernetes operator,
I want `/readyz` with dependency checks,
So that unhealthy pods are removed from service.

**Acceptance Criteria:**

**Given** all dependencies (DB) are healthy
**When** I request `GET /readyz`
**Then** response status is 200
**And** body shows `{"database": "ok"}`

### Story 5.3: Return 503 on Dependency Failure

As a SRE,
I want 503 when database is down,
So that load balancer routes traffic elsewhere.

**Acceptance Criteria:**

**Given** database connection is lost
**When** I request `GET /readyz`
**Then** response status is 503
**And** body shows `{"database": "unavailable"}`

### Story 5.4: Expose Prometheus Metrics Endpoint

As a SRE,
I want `/metrics` endpoint for Prometheus,
So that I can scrape application metrics.

**Acceptance Criteria:**

**Given** the HTTP server is running
**When** I request `GET /metrics`
**Then** response is Prometheus text format
**And** standard Go metrics are included

### Story 5.5: Capture HTTP Request Metrics

As a SRE,
I want HTTP request count, latency, and errors captured,
So that I can monitor API performance.

**Acceptance Criteria:**

**Given** HTTP requests are made
**When** I check `/metrics`
**Then** `http_requests_total{method, path, status}` is present
**And** `http_request_duration_seconds{method, path}` histogram exists

### Story 5.6: Implement Structured JSON Logging (Prod)

As a SRE,
I want JSON logs in production,
So that logs are parseable by log aggregators.

**Acceptance Criteria:**

**Given** `APP_ENV=production`
**When** application logs
**Then** output is valid JSON
**And** includes `level`, `timestamp`, `message` fields

### Story 5.7: Implement Human-Readable Logging (Dev)

As a developer,
I want readable logs in development,
So that I can debug locally without parsing JSON.

**Acceptance Criteria:**

**Given** `APP_ENV=development`
**When** application logs
**Then** output is human-readable format
**And** colors are used for log levels

### Story 5.8: Add Context Fields to Logs

As a SRE,
I want trace_id, request_id in logs,
So that I can correlate logs with traces.

**Acceptance Criteria:**

**Given** HTTP request with trace context
**When** log is written
**Then** `trace_id`, `request_id`, `path`, `method` are included

### Story 5.9: Create OpenTelemetry Spans

As a SRE,
I want OTEL spans for HTTP requests,
So that I can visualize request flow in Jaeger.

**Acceptance Criteria:**

**Given** OTEL exporter is configured
**When** HTTP request is processed
**Then** span is created with request attributes
**And** span is exported to configured backend

---

## Epic 6: Extension Interfaces

Service provides hook points (interfaces) for enterprise adapters and custom implementations.

### Story 6.1: Define Logger Interface

As a developer,
I want a Logger interface abstraction,
So that I can swap logging implementations.

**Acceptance Criteria:**

**Given** `internal/observability/logger.go` exists
**When** I implement Logger interface
**Then** methods include: Debug, Info, Warn, Error, With(fields)
**And** default implementation uses zap

### Story 6.2: Define Cache Interface

As a developer,
I want a Cache interface abstraction,
So that I can plug in Redis, Memcached, or in-memory cache.

**Acceptance Criteria:**

**Given** `internal/runtimeutil/cache.go` exists
**When** I review the interface
**Then** methods include: Get, Set, Delete, Exists
**And** interface documentation explains usage

### Story 6.3: Define RateLimiter Interface

As a developer,
I want a RateLimiter interface abstraction,
So that I can implement rate limiting with different backends.

**Acceptance Criteria:**

**Given** `internal/runtimeutil/ratelimiter.go` exists
**When** I review the interface
**Then** methods include: Allow(key), Limit(key, rate)
**And** interface is compatible with middleware usage

### Story 6.4: Define EventPublisher Interface

As a developer,
I want an EventPublisher interface abstraction,
So that I can publish events to Kafka, RabbitMQ, or NATS.

**Acceptance Criteria:**

**Given** `internal/runtimeutil/events.go` exists
**When** I review the interface
**Then** methods include: Publish(topic, event), PublishAsync(topic, event)
**And** event struct has ID, Type, Payload, Timestamp

### Story 6.5: Define SecretProvider Interface

As a developer,
I want a SecretProvider interface abstraction,
So that I can fetch secrets from Vault, AWS SM, or GCP SM.

**Acceptance Criteria:**

**Given** `internal/runtimeutil/secrets.go` exists
**When** I review the interface
**Then** methods include: GetSecret(key), GetSecretWithTTL(key)
**And** default implementation reads from environment

### Story 6.6: Document Custom Provider Implementation

As a developer,
I want documentation on implementing custom providers,
So that I can extend the boilerplate safely.

**Acceptance Criteria:**

**Given** ARCHITECTURE.md exists
**When** I read the extension section
**Then** each interface has implementation example
**And** registration pattern is documented

---

## Epic 7: Sample Module (Note)

Developer has complete reference implementation demonstrating all patterns to copy for new domains.

### Story 7.1: Create Note Domain Entity

As a developer,
I want an example entity with validation,
So that I can understand how to create domain entities.

**Acceptance Criteria:**

**Given** `internal/domain/note/entity.go` exists
**When** I review the code
**Then** Note entity has ID, Title, Content, CreatedAt, UpdatedAt
**And** validation method returns domain errors
**And** entity is documented with comments

### Story 7.2: Create Note SQL Migration

As a developer,
I want an example migration for Note table,
So that I can understand migration patterns.

**Acceptance Criteria:**

**Given** `db/migrations/YYYYMMDD_create_notes.up.sql` exists
**When** I run `make migrate-up`
**Then** notes table is created with proper columns
**And** down migration drops the table

### Story 7.3: Create Note Repository with sqlc

As a developer,
I want an example sqlc repository,
So that I can understand type-safe query patterns.

**Acceptance Criteria:**

**Given** `db/queries/note.sql` exists
**When** I run `make gen`
**Then** `internal/infra/postgres/note/queries.go` is generated
**And** queries include: CreateNote, GetNote, ListNotes, UpdateNote, DeleteNote

### Story 7.4: Create Note Usecase

As a developer,
I want an example usecase implementation,
So that I can understand business logic patterns.

**Acceptance Criteria:**

**Given** `internal/usecase/note/usecase.go` exists
**When** I review the code
**Then** NoteUsecase depends on NoteRepository interface
**And** methods include: Create, Get, List, Update, Delete
**And** business logic validates before persistence

### Story 7.5: Create Note HTTP Handler

As a developer,
I want an example HTTP handler,
So that I can understand handler patterns.

**Acceptance Criteria:**

**Given** `internal/interface/http/note/handler.go` exists
**When** I review the code
**Then** NoteHandler depends on NoteUsecase interface
**And** routes: POST, GET, GET/:id, PUT/:id, DELETE/:id
**And** uses response envelope pattern

### Story 7.6: Create Note Unit Tests

As a developer,
I want example unit tests,
So that I can understand testing patterns.

**Acceptance Criteria:**

**Given** `*_test.go` files exist in note packages
**When** I run `make test`
**Then** tests pass with ≥70% coverage
**And** table-driven tests with AAA pattern are used
**And** mocks are used for dependencies

### Story 7.7: Create Note Integration Test

As a developer,
I want an example integration test,
So that I can understand E2E testing patterns.

**Acceptance Criteria:**

**Given** `internal/interface/http/note/handler_integration_test.go` exists
**When** I run `make test`
**Then** integration test hits real HTTP endpoints
**And** test uses test database
**And** cleanup happens after test

### Story 7.8: Create ARCHITECTURE.md

As a tech lead,
I want ARCHITECTURE.md documenting the project,
So that developers understand the design decisions.

**Acceptance Criteria:**

**Given** `ARCHITECTURE.md` exists in project root
**When** I read the document
**Then** Three Pillars philosophy is explained
**And** layer structure is documented
**And** patterns and conventions are listed

### Story 7.9: Create AGENTS.md

As a AI assistant user,
I want AGENTS.md as AI contract,
So that AI assistants follow consistent patterns.

**Acceptance Criteria:**

**Given** `AGENTS.md` exists in project root
**When** AI reads the document
**Then** DO/DON'T patterns are clear
**And** file structure conventions are listed
**And** testing requirements are documented

### Story 7.10: Document Copy Pattern for New Modules

As a developer,
I want documentation on copying example module,
So that I can create new domains quickly.

**Acceptance Criteria:**

**Given** README.md and AGENTS.md exist
**When** I follow the "Adding New Module" section
**Then** step-by-step guide is available
**And** checklist ensures all layers are created
**And** example `cp -r` commands are provided

---

## Summary

| Epic | Title | Stories | FRs |
|------|-------|---------|-----|
| 1 | Project Foundation & DX Setup | 5 | 9 |
| 2 | Configuration & Environment | 5 | 5 |
| 3 | HTTP API Core | 8 | 8 |
| 4 | Database & Persistence | 7 | 7 |
| 5 | Observability Suite | 9 | 9 |
| 6 | Extension Interfaces | 6 | 6 |
| 7 | Sample Module (Note) | 10 | 12 |
| **Total** | | **50** | **56** |

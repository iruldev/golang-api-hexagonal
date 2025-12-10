---
stepsCompleted: [1, 2, 3]
inputDocuments: ['docs/analysis/brainstorming-session-2025-12-10.md']
workflowType: 'research'
lastStep: 3
research_type: 'technical'
research_topic: 'golang-enterprise-boilerplate'
research_goals: 'Validate tech stack, discover gaps, collect references'
user_name: 'Gan'
date: '2025-12-10'
web_research_enabled: true
source_verification: true
---

# Technical Research: Enterprise Go Backend Boilerplate

**Researcher:** Gan  
**Date:** 2025-12-10  
**Research Type:** Technical Validation & Best Practices

## Executive Summary

Penelitian ini memvalidasi keputusan tech stack dan arsitektur dari brainstorming session untuk **Enterprise Golang Backend Boilerplate**. Hasil menunjukkan bahwa semua pilihan teknologi yang diambil sesuai dengan **state-of-the-art practices 2024-2025**. Beberapa gap teridentifikasi untuk ekspansi v1.1+.

### Key Findings
✅ **All tech stack choices validated** - chi, sqlc+pgx, zap, koanf, asynq, OpenTelemetry  
✅ **Clean/Hexagonal architecture** - Well-documented patterns in Go ecosystem  
✅ **Testing strategy** - testify + testcontainers adalah industry standard  
⚠️ **Gaps identified** - Secret management integration, advanced rate limiting  

---

## Table of Contents

1. [Architecture & Boilerplate Patterns](#1-architecture--boilerplate-patterns)
2. [Tech Stack Validation](#2-tech-stack-validation)
3. [Observability Best Practices](#3-observability-best-practices)
4. [Testing Strategies](#4-testing-strategies)
5. [Time & Timezone Safety](#5-time--timezone-safety)
6. [Gaps & Recommendations](#6-gaps--recommendations)
7. [References](#7-references)

---

## 1. Architecture & Boilerplate Patterns

### Industry Validation

Clean Architecture dan Hexagonal Architecture (Ports & Adapters) adalah pattern yang well-established untuk Go microservices:

| Pattern | Key Principle | Go Adoption |
|---------|--------------|-------------|
| **Clean Architecture** | Dependency direction inward | High - multiple production boilerplates |
| **Hexagonal Architecture** | Domain core pure, infra at edges | High - well-documented |
| **DDD-light** | Domain focus without full DDD complexity | Medium - practical compromise |

### Folder Structure Consensus

Standard enterprise Go structure yang align dengan pilihan kita:

```
internal/
├── domain/          # Pure business logic, entities
├── usecase/         # Application services, orchestration
├── infra/           # Database, external services
└── interface/       # HTTP handlers, gRPC, CLI
```

> **Source Validation:** Multiple 2024-2025 boilerplates follow this exact pattern [[stackademic.com](https://stackademic.com), [medium.com](https://medium.com)]

### Critical Insight
> "Hexagonal Architecture keeps the domain core pure, pushing frameworks and databases to the edges" - This aligns exactly with our interface → usecase → domain → infra design.

---

## 2. Tech Stack Validation

### HTTP Framework: chi ✅ VALIDATED

| Aspect | Finding |
|--------|---------|
| **Go Idiomatic** | chi is built on `net/http`, most idiomatic choice |
| **Performance** | Patricia Radix trie router, low memory allocation |
| **vs Gin/Echo** | Less magic, more composable, better for clean architecture |
| **Production Use** | Used by major companies, active maintenance |

> **Verdict:** chi is the optimal choice for enterprise boilerplate requiring minimal magic and standard library compatibility.

### Database: sqlc + pgx ✅ VALIDATED

| Aspect | Finding |
|--------|---------|
| **Type Safety** | sqlc generates compile-time validated Go from SQL |
| **Performance** | pgx is fastest PostgreSQL driver for Go |
| **Best Practices** | Parameterized queries, explicit SQL, reviewable |
| **vs GORM** | More control, explicit queries, better for regulated environments |

**Key Best Practices Discovered:**
- Use `pgxpool.Pool` for connection pooling (NOT `pgx.Conn`)
- Configure `MaxConns`, `MinConns`, `MaxConnLifetime`
- Separate concerns with repository layer
- Use `sql_package: pgx/v5` in sqlc config

### Logging: zap ✅ VALIDATED

| Aspect | Finding |
|--------|---------|
| **Performance** | Zero-allocation, high-throughput |
| **Structure** | JSON output, machine-parseable |
| **Production** | `zap.NewProduction()` with sampling |
| **Industry Standard** | De-facto enterprise Go logger |

**Production Checklist:**
- Always `defer logger.Sync()`
- Set log levels (Info+ for production)
- Use fields for structured context
- Never log sensitive data (PII, secrets)

### Config: koanf ✅ VALIDATED

| Aspect | Finding |
|--------|---------|
| **Flexibility** | Multiple providers (ENV, file, flags, Vault) |
| **vs Viper** | Lighter, fewer dependencies, cleaner API |
| **Best Practice** | Layered config: defaults → file → ENV → flags |

**Key Pattern:** Load → Bind to struct → Validate → Fail fast

### Jobs: asynq ✅ VALIDATED

| Aspect | Finding |
|--------|---------|
| **Features** | Retry, DLQ, scheduling, priority queues |
| **Observability** | Web UI (AsynqMon), Prometheus integration |
| **Production** | At-least-once delivery, crash recovery |
| **Comparison** | Go equivalent of Sidekiq/Celery |

**Production Best Practices:**
- Design idempotent tasks
- Implement visibility timeout
- Use priority queues strategically
- Monitor with Prometheus metrics

### Observability: OpenTelemetry + Prometheus + Grafana ✅ VALIDATED

| Component | Role |
|-----------|------|
| **OpenTelemetry** | Unified traces, metrics, logs collection |
| **Prometheus** | Metrics storage and alerting |
| **Grafana** | Visualization and dashboards |

**Key Practices:**
- Early OTEL initialization in application lifecycle
- Follow semantic conventions for attributes
- ~1-5% latency overhead (acceptable)
- Use Prometheus v3 for optimal OTEL integration

---

## 3. Observability Best Practices

### Golden Signals Implementation

| Signal | Metric Example | Implementation |
|--------|---------------|----------------|
| **Latency** | `http_request_duration_seconds` | Histogram with route labels |
| **Traffic** | `http_requests_total` | Counter per endpoint |
| **Errors** | `http_errors_total` | Counter by status code |
| **Saturation** | `goroutines_count`, `db_pool_usage` | Gauge metrics |

### Health Endpoints

| Endpoint | Purpose | Checks |
|----------|---------|--------|
| `/healthz` | Liveness | Process alive |
| `/readyz` | Readiness | DB, Redis, dependencies |

> **Critical:** These MUST be different - `/healthz` should NOT check dependencies

### Tracing Integration

```
HTTP Request → otelhttp middleware → span created
↓
DB Call → pgx tracing → child span
↓
External API → http client wrapper → child span
```

---

## 4. Testing Strategies

### Layered Testing Approach

| Layer | Tool | Focus |
|-------|------|-------|
| **Unit** | testify/assert, mock | Domain logic, usecase |
| **Integration** | testcontainers | DB, cache, real dependencies |
| **E2E** | httptest + testcontainers | Full request flow |

### testify Best Practices

- Use `assert` for non-fatal, `require` for fatal assertions
- Table-driven tests for multiple scenarios
- `testify/mock` for dependency isolation
- Don't over-mock - mock only expensive dependencies

### testcontainers Best Practices

- Use specific modules (PostgreSQL, Redis)
- Share containers within test suites for performance
- Use `TestMain` for package-level setup/teardown
- Seed data before tests for predictability

---

## 5. Time & Timezone Safety

### Core Principle

> **ALWAYS store and process time in UTC. Convert to local only at display boundary.**

### Implementation Checklist

| Area | Recommendation |
|------|---------------|
| **Database** | `TIMESTAMP WITH TIME ZONE`, store UTC |
| **Go Code** | `time.Time` always in UTC |
| **Package** | Create `timeutil` with `NowUTC()`, `Parse()`, `Format()` |
| **Embedded TZDB** | Use `time/tzdata` package for consistent deployment |
| **Parsing** | Use `time.RFC3339` for external input |

### Clock Abstraction Pattern

```go
type Clock interface {
    Now() time.Time
}

type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now().UTC() }

type FakeClock struct{ fixed time.Time }
func (f FakeClock) Now() time.Time { return f.fixed }
```

---

## 6. Gaps & Recommendations

### Identified Gaps

| Gap | Priority | Recommendation |
|-----|----------|----------------|
| **Secret Management** | High | Add hook for Vault/AWS Secrets Manager |
| **Rate Limiting** | Medium | Add rate limiter abstraction (in-memory/Redis) |
| **Circuit Breaker** | Medium | Document pattern for external calls |
| **gRPC Support** | Low | Document as v1.1+ expansion |
| **GraphQL Support** | Low | Document as v1.1+ expansion |

### Additional Recommendations from Research

1. **Context Propagation** - Ensure trace context flows through all layers
2. **Connection Pool Monitoring** - Add metrics for DB pool health
3. **Log Rotation** - Consider lumberjack for file-based logs
4. **Migrations** - Use golang-migrate for schema evolution

---

## 7. References

### Architecture & Patterns
- Clean Architecture in Go [2024 Updated] - pkritiotis.io
- Hexagonal Architecture in Go - medium.com (multiple articles)
- Go Microservice Starter Kit (Fiber + Ent) - gumroad.com

### Tech Stack
- chi router - github.com/go-chi/chi
- sqlc Best Practices - haykot.dev, medium.com
- pgxpool Configuration - hexacluster.ai, betterstack.com
- zap Logging - signoz.io, dash0.com
- koanf Documentation - github.com/knadh/koanf

### Observability
- OpenTelemetry Go - opentelemetry.io
- Prometheus Best Practices - grafana.com
- OpenTelemetry + Prometheus + Grafana stack - medium.com

### Testing
- testcontainers-go - testcontainers.org
- testify Best Practices - betterstack.com
- Integration Testing Patterns - stackademic.com

### Time & Timezone
- Golang Timezone Best Practices - labex.io, dev.to
- time/tzdata Package - reintech.io

---

## Conclusion

Penelitian ini mengkonfirmasi bahwa **semua keputusan tech stack dari brainstorming session sudah align dengan best practices 2024-2025**:

| Decision | Status | Notes |
|----------|--------|-------|
| chi router | ✅ Validated | Most idiomatic choice |
| sqlc + pgx | ✅ Validated | Enterprise standard |
| zap logging | ✅ Validated | High-performance structured logging |
| koanf config | ✅ Validated | Clean layered configuration |
| asynq jobs | ✅ Validated | Feature-rich, production-ready |
| OpenTelemetry | ✅ Validated | Industry standard observability |
| testify + testcontainers | ✅ Validated | Best testing stack for Go |
| UTC-first time | ✅ Validated | Universal best practice |

**Next Step:** Transform brainstorming + research output into PRD document.

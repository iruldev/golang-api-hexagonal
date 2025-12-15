---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments: []
workflowType: 'research'
lastStep: 5
research_type: 'technical'
research_topic: 'Go Golden Template Best Practices'
research_goals: 'CI/CD quality gates, best practices, project structure, observability, security'
user_name: 'Gan'
date: '2025-12-15'
web_research_enabled: true
source_verification: true
---

# Technical Research: Go Golden Template Best Practices

**Date:** 2025-12-15
**Research Type:** Technical
**Researcher:** AI Research Partner
**Participant:** Gan

---

## Executive Summary

Penelitian komprehensif ini mencakup 5 area kritis untuk membangun "golden template" production-grade Go backend:

1. **CI/CD Quality Gates** - golangci-lint v2, depguard, 80% coverage
2. **Go Best Practices** - Uber style guide, production patterns
3. **Project Structure** - Hexagonal architecture real-world patterns
4. **Observability** - OpenTelemetry dengan otelhttp/otelchi
5. **Security** - OWASP guidelines, secret management, govulncheck

**Key Findings:**
- golangci-lint v2 dengan strict mode sebagai standar industri
- 80% coverage target yang realistis dan meaningful
- Hexagonal architecture dengan import boundaries via depguard
- OTel dengan context propagation dan W3C Trace Context
- Security scanning dengan govulncheck + gosec dalam CI

---

## Table of Contents

1. [CI/CD Quality Gates](#1-cicd-quality-gates)
2. [Go Backend Best Practices](#2-go-backend-best-practices)
3. [Enterprise Project Structure](#3-enterprise-project-structure)
4. [Observability Standards](#4-observability-standards)
5. [Security Best Practices](#5-security-best-practices)
6. [Recommendations Summary](#6-recommendations-summary)
7. [Sources](#sources)

---

## 1. CI/CD Quality Gates

### 1.1 golangci-lint v2 Configuration

**Key Findings [High Confidence]**

golangci-lint v2 introduces improved configuration schema dengan sections yang lebih jelas:

| Section | Purpose |
|---------|---------|
| `run` | Concurrency, timeout, directories to skip |
| `output` | Output format (JSON, text) |
| `linters` | Enable/disable specific linters |
| `settings` | Tune individual linter behavior |
| `exclusions` | Ignore specific warnings |
| `formatters` | Code formatting configuration |

**Recommended Linters for Production:**

| Linter | Purpose |
|--------|---------|
| `exhaustive` | Ensure switch covers all enum values |
| `forbidigo` | Enforce consistent package usage |
| `wrapcheck` | Verify errors from 3rd party are wrapped |
| `gosec` | Security vulnerability detection |
| `errcheck` | Catch unchecked errors |
| `staticcheck` | Comprehensive static analysis |
| `gci` | Import statement organization |
| `unused` | Detect unused code |
| `bodyclose` | HTTP response body closure |
| `contextcheck` | Proper context usage |

**Best Practices:**
- Start incrementally: enable core linters, fix issues, add more
- Use `linters.default: standard` atau `all` dengan exclusions
- Set appropriate `timeout` untuk CI (3-5 menit typical)
- Use `--new-from-merge-base=main` untuk focus pada new code saja

**Source:** [golangci-lint.run](https://golangci-lint.run), [reliasoftware.com](https://reliasoftware.com)

---

### 1.2 depguard Import Boundary Enforcement

**Key Findings [High Confidence]**

depguard adalah linter untuk enforce import boundaries dalam architectural layers:

**Modes:**
- **Strict Mode**: Everything denied unless explicitly allowed
- **Lax Mode**: Everything allowed unless explicitly denied

**Configuration untuk Hexagonal Architecture:**

```yaml
# .golangci.yml - depguard settings
linters-settings:
  depguard:
    rules:
      domain:
        files:
          - "**/internal/domain/**"
        deny:
          - pkg: "github.com/*/internal/usecase"
            desc: "domain cannot import usecase"
          - pkg: "github.com/*/internal/interface"
            desc: "domain cannot import interface"
          - pkg: "github.com/*/internal/infra"
            desc: "domain cannot import infra"
            
      usecase:
        files:
          - "**/internal/usecase/**"
        deny:
          - pkg: "github.com/*/internal/interface"
            desc: "usecase cannot import interface"
          - pkg: "github.com/*/internal/infra"
            desc: "usecase cannot import infra"
```

**Source:** [go.dev/pkg/depguard](https://pkg.go.dev/github.com/OpenPeeDeeP/depguard), [github.com/OpenPeeDeeP/depguard](https://github.com/OpenPeeDeeP/depguard)

---

### 1.3 Test Coverage Policy

**Key Findings [High Confidence]**

| Coverage Level | Assessment |
|----------------|------------|
| **< 50%** | Low code quality, insufficient testing |
| **70-80%** | Reasonable, achievable target |
| **> 80%** | High-quality codebase indicator |
| **90%+** | Reserved for critical components |

**Best Practices:**
- **Target 80%** sebagai balance antara quality dan practicality
- Focus on **meaningful tests** - critical paths, edge cases, error handling
- Use `go test -coverprofile=coverage.out` untuk generate report
- Enforce via CI dengan tools seperti `go-test-coverage`
- Exclude generated code dan trivial getters/setters

**CI Integration:**
```yaml
- name: Test with coverage
  run: go test -race -coverprofile=coverage.out ./...
  
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage ${COVERAGE}% is below 80% threshold"
      exit 1
    fi
```

**Source:** [dev.to](https://dev.to), [medium.com](https://medium.com)

---

### 1.4 Pre-commit vs CI

**Recommendation [Medium Confidence]**

| Approach | Pros | Cons |
|----------|------|------|
| **Pre-commit** | Catch early, fast feedback | Can be bypassed, slower commits |
| **CI only** | Single source of truth, mandatory | Late feedback |
| **Both** | Defense in depth | More complexity |

**Best Practice:** Use **both** dengan pre-commit as "recommended but byppassable":
- Pre-commit: Quick lint check (`golangci-lint run --fast`)
- CI: Full lint, test, coverage, security scans (mandatory, blocking)

---

## 2. Go Backend Best Practices

### 2.1 Uber Go Style Guide

**Key Findings [High Confidence]**

Uber's style guide adalah de facto standard untuk production Go code:

**Guidelines:**
- Verify interface compliance dengan `var _ Interface = (*Struct)(nil)`
- Use appropriate receiver types (pointer vs value)
- Copy slices/maps at boundaries to prevent mutation
- Use `defer` for cleanup (unlocks, file closes)
- Size channels appropriately (1 for signals, buffered for batches)

**Error Handling:**
- Use specific error types, wrap with context
- Handle errors only once
- Never use `panic` for recoverable errors
- Name errors with `Err` prefix: `ErrNotFound`

**Performance:**
- Prefer `strconv` over `fmt` for conversions
- Avoid repeated string-to-byte conversions
- Specify container capacity when known

**Testing:**
- Use table-driven tests
- Use functional options for configurable components

**Source:** [github.com/uber-go/guide](https://github.com/uber-go/guide)

---

### 2.2 Production Readiness Checklist

**Key Areas:**
- **Stateless Services** - Facilitate horizontal scaling
- **Message Queues** - Decouple services (Kafka, RabbitMQ)
- **Observability** - Prometheus metrics, Grafana, distributed tracing
- **Connection Pooling** - Efficient database connections
- **Circuit Breakers** - Prevent cascading failures
- **Graceful Degradation** - Design for failure

**Source:** [bytebytego.com](https://bytebytego.com)

---

## 3. Enterprise Project Structure

### 3.1 Hexagonal Architecture Patterns

**Key Findings [High Confidence]**

**Core Layers:**

```
internal/
├── domain/           # Entities, repository interfaces (ports)
│   └── {module}/
│       ├── entity.go
│       ├── errors.go
│       └── repository.go
│
├── usecase/          # Application business logic
│   └── {module}/
│       └── usecase.go
│
├── adapters/         # Interface implementations
│   ├── http/         # Driving adapter (primary)
│   ├── grpc/
│   └── repository/   # Driven adapter (secondary)
│       └── postgres/
│
└── infrastructure/   # External services
    ├── database/
    └── cache/
```

**Ports & Adapters:**
- **Driving Ports (Primary)**: Exposed by core, used by external (HTTP handlers)
- **Driven Ports (Secondary)**: Services core needs from outside (repositories)
- **Driving Adapters**: Implement driving ports (API controllers)
- **Driven Adapters**: Implement driven ports (database implementations)

**Dependency Rule:** Dependencies point inward. Outer layers implement interfaces defined by inner layers.

**Source:** [dev.to](https://dev.to), [medium.com](https://medium.com)

---

### 3.2 Multi-Entrypoint Monolith

**Pattern untuk server/worker/scheduler/CLI:**

```
cmd/
├── server/main.go      # HTTP API entry point
├── worker/main.go      # Background job processor
├── scheduler/main.go   # Cron job scheduler
└── cli/main.go         # CLI tool

internal/               # Shared business logic
├── app/                # Application wiring
├── domain/
├── usecase/
└── infra/
```

**Key Insight:** Worker, scheduler, dan CLI adalah **entry points**, bukan layers. Mereka menggunakan usecases yang sama dengan HTTP server.

**Source:** [github.com/golang-standards/project-layout](https://github.com/golang-standards/project-layout)

---

## 4. Observability Standards

### 4.1 OpenTelemetry Best Practices

**Key Findings [High Confidence]**

**chi Router Integration:**

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// Wrap router with otelhttp
handler := otelhttp.NewHandler(router, "my-service")

// Or use as middleware
router.Use(func(next http.Handler) http.Handler {
    return otelhttp.NewHandler(next, "my-service")
})
```

**Best Practices:**
1. **Use otelhttp** - Official package untuk HTTP instrumentation
2. **Context propagation** - W3C Trace Context standard
3. **Custom spans** - Add untuk business logic operations
4. **Meaningful names** - Descriptive service names
5. **Filtering** - Exclude health checks dari tracing
6. **Sampling** - Implement untuk high-throughput apps

**Performance:** Overhead typically < 1ms per request

**Source:** [opentelemetry.io](https://opentelemetry.io), [uptrace.dev](https://uptrace.dev)

---

### 4.2 Structured Logging with Zap

**Key Findings [High Confidence]**

**Best Practices:**
- Use `zap.NewProduction()` untuk production
- Prefer `zap.Logger` over `SugaredLogger` for performance
- Add contextual fields (trace_id, request_id)
- Log at appropriate levels

**trace_id/request_id Correlation:**

```go
// Middleware to propagate request_id
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Logger with request context
func LoggerFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
    if requestID, ok := ctx.Value("request_id").(string); ok {
        return logger.With(zap.String("request_id", requestID))
    }
    return logger
}
```

**Source:** [betterstack.com](https://betterstack.com), [medium.com](https://medium.com)

---

## 5. Security Best Practices

### 5.1 OWASP Guidelines for Go

**Key Findings [High Confidence]**

| OWASP Risk | Go Mitigation |
|------------|---------------|
| Broken Access Control | Implement RBAC, server-side validation |
| Cryptographic Failures | Use strong algorithms, proper key management |
| Injection | Validate input, use `html/template` |
| Security Misconfiguration | Avoid defaults, production-specific configs |
| Vulnerable Components | Update dependencies regularly |
| Authentication Failures | OAuth2/JWT, bcrypt for passwords |

**Source:** [medium.com](https://medium.com), [securityium.com](https://securityium.com)

---

### 5.2 Secret Management

**Best Practices:**
- **Never hardcode** secrets in source code
- **Environment variables** for simple apps (with caution)
- **Secret managers** for production (HashiCorp Vault, AWS Secrets Manager)
- **Runtime integration** - Sync secrets to env vars at runtime
- **Memory security** - Clear sensitive data after use

**Source:** [gitguardian.com](https://gitguardian.com)

---

### 5.3 Dependency Scanning

**Tools:**

| Tool | Purpose |
|------|---------|
| `govulncheck` | Official Go vulnerability scanner |
| `gosec` | Static security analysis |
| `gitleaks` | Secret detection |
| `trivy` | Container vulnerability scanning |

**CI Integration:**
```yaml
- name: Security scan
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
    
- name: Static security analysis
  run: |
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec ./...
```

**Source:** [go.dev/security](https://go.dev/security), [hackernoon.com](https://hackernoon.com)

---

## 6. Recommendations Summary

### Immediate Actions

| Area | Recommendation | Priority |
|------|----------------|----------|
| **Linting** | golangci-lint v2 dengan strict mode | 🔴 High |
| **Boundaries** | depguard untuk import enforcement | 🔴 High |
| **Coverage** | 80% target dengan CI enforcement | 🔴 High |
| **Style** | Adopt Uber Go style guide | 🟡 Medium |
| **Tracing** | otelhttp + context propagation | 🟡 Medium |
| **Logging** | Zap + trace_id correlation | 🟡 Medium |
| **Security** | govulncheck + gosec in CI | 🔴 High |

### golangci.yml Gold Standard Config

```yaml
version: "2"

run:
  timeout: 5m
  concurrency: 4

linters:
  default: standard
  enable:
    - exhaustive
    - wrapcheck
    - gosec
    - bodyclose
    - contextcheck
    - depguard
    - errcheck
    - staticcheck
    - unused

linters-settings:
  depguard:
    rules:
      domain-layer:
        files:
          - "**/internal/domain/**"
        deny:
          - pkg: "github.com/$PROJECT/internal/usecase"
          - pkg: "github.com/$PROJECT/internal/interface"
          - pkg: "github.com/$PROJECT/internal/infra"
      usecase-layer:
        files:
          - "**/internal/usecase/**"
        deny:
          - pkg: "github.com/$PROJECT/internal/interface"
          - pkg: "github.com/$PROJECT/internal/infra"
```

---

## Sources

**CI/CD & Quality:**
- golangci-lint.run - Official documentation
- github.com/OpenPeeDeeP/depguard - Import boundary enforcement
- dev.to - Coverage best practices

**Best Practices:**
- github.com/uber-go/guide - Uber Go Style Guide
- bytebytego.com - Production readiness patterns

**Architecture:**
- golang-standards/project-layout - Standard Go project layout
- dev.to, medium.com - Hexagonal architecture examples

**Observability:**
- opentelemetry.io - Official OTel documentation
- uptrace.dev - Go OTel best practices
- betterstack.com - Zap logging guide

**Security:**
- go.dev/security - Go security resources
- gitguardian.com - Secret management
- hackernoon.com - govulncheck guide

---

*Research completed on 2025-12-15 | Technical Research Workflow*

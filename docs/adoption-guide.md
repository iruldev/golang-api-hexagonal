# Adoption Guide for New Developers

**Last Updated:** 2026-01-04  
**Estimated Total Read Time:** ~80 minutes

---

Welcome to the golang-api-hexagonal project! This comprehensive guide will help you become productive quickly by understanding the architecture, implementing your first feature, and following best practices.

## Table of Contents

1. [Project Overview](#project-overview) (~5 min)
2. [Architecture Deep Dive](#architecture-deep-dive) (~15 min)
3. [Local Development Setup](#local-development-setup) (~10 min hands-on)
4. [Your First Feature Tutorial](#your-first-feature-tutorial) (~30 min hands-on)
5. [Testing Guide](#testing-guide) (~10 min)
6. [Code Review Checklist](#code-review-checklist) (~5 min)
7. [FAQ and Troubleshooting](#faq-and-troubleshooting) (~5 min)

---

## Project Overview

**Estimated reading time: ~5 minutes**

### What is This Project?

The **golang-api-hexagonal** project is a production-ready Go API that demonstrates **Hexagonal Architecture** (also known as Ports and Adapters). It's designed as a battle-tested boilerplate for building scalable, maintainable, and testable Go services.

### Key Features

| Feature | Description |
|---------|-------------|
| **Hexagonal Architecture** | Clear separation between business logic and infrastructure |
| **Layer Enforcement** | CI-enforced boundaries via golangci-lint depguard |
| **Dependency Injection** | Uber Fx for clean, testable wiring |
| **Comprehensive Testing** | Unit, integration, and contract tests with 80% coverage |
| **Observability** | OpenTelemetry tracing, Prometheus metrics, structured logging |
| **Security** | JWT authentication, rate limiting, security headers |
| **Cloud-Native** | Health probes, graceful shutdown, Kubernetes-ready |

### Technology Stack

| Category | Technology | Version |
|----------|------------|---------|
| Language | Go | 1.25.5 |
| Router | Chi | v5.2.3 |
| Database | PostgreSQL + pgx | v5.7.6 |
| DI | Uber Fx | v1.24.0 |
| Tracing | OpenTelemetry | v1.39.0 |
| Metrics | Prometheus | v1.23.2 |
| Linting | golangci-lint | v1.64.8 |

### Quick Reference

```bash
make setup       # First-time setup
make run         # Start server
make test        # Run unit tests
make lint        # Run linter
make ci          # Full CI pipeline
```

> 📚 **Learn More:** See [README.md](../README.md) for detailed setup instructions.

---

## Architecture Deep Dive

**Estimated reading time: ~15 minutes**

This section explains the Hexagonal Architecture pattern used in this project and how to work within its constraints.

### The Hexagonal Model

```
┌─────────────────────────────────────────────────────────────────┐
│                        transport/http                           │
│                      (Inbound Adapters)                         │
│            handler/ │ middleware/ │ contract/                   │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                             app/                                │
│                      (Application Layer)                        │
│                user/ │ audit/ │ auth.go                         │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                           domain/                               │
│                       (Business Core)                           │
│       User │ Audit │ ID │ Pagination │ Querier │ TxManager      │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                            infra/                               │
│                      (Outbound Adapters)                        │
│          postgres/ │ config/ │ observability/ │ fx/             │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

#### 🔴 Domain Layer (`internal/domain/`)

The **core business logic** - pure Go with zero external dependencies.

**Contains:**
- Entities (User, Audit, etc.)
- Value Objects (ID, Pagination)
- Repository Interfaces (Ports)
- Domain Errors

**Example - Domain Entity:**
```go
// internal/domain/user.go
package domain

import "time"

// User represents the core user entity.
// Note: NO JSON tags - transport layer handles serialization.
type User struct {
    ID        string
    Email     string
    FirstName string
    LastName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Example - Repository Port:**
```go
// internal/domain/user.go
package domain

import "context"

// UserRepository defines the contract for user persistence.
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    List(ctx context.Context, params ListParams) ([]*User, error)
}
```

#### 🟡 Application Layer (`internal/app/`)

**Use cases and orchestration** - implements business logic without knowing about HTTP or databases.

**Contains:**
- Use Cases (CreateUser, GetUser, ListUsers)
- Application Services
- Input/Output DTOs for use cases

**Example - Use Case:**
```go
// internal/app/user/create_user.go
package user

import (
    "context"
    "fmt"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type CreateUserUseCase struct {
    repo  domain.UserRepository
    audit AuditService
}

func NewCreateUserUseCase(repo domain.UserRepository, audit AuditService) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo, audit: audit}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*domain.User, error) {
    user := &domain.User{
        Email:     input.Email,
        FirstName: input.FirstName,
        LastName:  input.LastName,
    }
    
    if err := uc.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user in repository: %w", err)
    }
    
    return user, nil
}
```

#### 🔵 Transport Layer (`internal/transport/http/`)

**Inbound adapters** - handles HTTP requests and translates to/from domain objects.

**Contains:**
- HTTP Handlers
- Middleware (auth, logging, metrics)
- Request/Response DTOs with JSON tags
- Contract definitions

**Example - Handler:**
```go
// internal/transport/http/handler/user_handler.go
package handler

import (
    "encoding/json"
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/app/user"
)

type UserHandler struct {
    createUseCase *user.CreateUserUseCase
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Handle error
        return
    }
    
    result, err := h.createUseCase.Execute(r.Context(), user.CreateUserInput{
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    })
    if err != nil {
        // Handle error
        return
    }
    
    json.NewEncoder(w).Encode(toUserResponse(result))
}
```

#### 🟢 Infrastructure Layer (`internal/infra/`)

**Outbound adapters** - implements repository interfaces and external integrations.

**Contains:**
- Database repositories (PostgreSQL)
- Configuration loading
- Observability setup (logging, metrics, tracing)
- Fx modules for dependency injection

**Example - Repository Implementation:**
```go
// internal/infra/postgres/user_repo.go
package postgres

import (
    "context"
    "fmt"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
)

type userRepository struct {
    queries *sqlcgen.Queries
}

func NewUserRepository(queries *sqlcgen.Queries) domain.UserRepository {
    return &userRepository{queries: queries}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
    _, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
    })
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }
    return nil
}
```

### Layer Boundary Rules

> ⚠️ **CRITICAL:** These rules are enforced by depguard in CI. Violations will fail the build.

| Layer | CAN Import | CANNOT Import |
|-------|------------|---------------|
| `domain/` | stdlib ONLY | slog, otel, uuid, http, pgx, app, transport, infra |
| `app/` | domain | slog, otel, uuid, http, pgx, transport, infra |
| `transport/` | domain, app, chi, jwt, validator | pgx, infra |
| `infra/` | domain, shared, pgx, otel | transport |
| `infra/fx/` | ALL internal packages | (wiring layer exception) |

### Key Architecture Rules

1. **Domain entities have NO JSON tags** - Transport layer adds serialization
2. **App services do NOT log** - Return errors, let callers handle logging
3. **Transport handlers do NOT call database** - Use app layer use cases
4. **Infra repositories do NOT return HTTP errors** - Return domain errors

> 📚 **Learn More:** See [ADR-001: Hexagonal Architecture](adr/ADR-001-hexagonal-architecture.md) and [ADR-002: Layer Boundary Enforcement](adr/ADR-002-layer-boundary-enforcement.md)

---

## Local Development Setup

**Estimated time: ~10 minutes (hands-on)**

### Prerequisites

- [ ] Go 1.25.5 or later
- [ ] Docker and Docker Compose
- [ ] PostgreSQL client (optional, for debugging)

### Quick Start

```bash
# 1. Clone and enter the project
git clone <repository-url>
cd golang-api-hexagonal

# 2. Run the quick-start setup
make quick-start

# 3. Verify the API is running
curl http://localhost:8080/healthz
```

### Step-by-Step Setup

If you prefer manual setup:

```bash
# 1. Copy environment file
cp .env.example .env.local

# 2. Start PostgreSQL
make infra-up

# 3. Run database migrations
make migrate-up

# 4. Start the server
make run
```

### Verify Setup

```bash
# Health check
curl http://localhost:8080/healthz

# Create a test user (requires JWT - see README)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt>" \
  -d '{"email": "test@example.com", "first_name": "Test", "last_name": "User"}'
```

> 📚 **Detailed Setup:** See the [README.md Quick Start section](../README.md#quick-start)

---

## Your First Feature Tutorial

**Estimated time: ~30 minutes (hands-on)**

This tutorial walks you through adding a new feature following the Hexagonal Architecture pattern. We'll add a simple endpoint to get system information.

### What We'll Build

A new endpoint: `GET /api/v1/system/info` that returns system information.

### Step 1: Define the Domain (~5 min)

First, define the domain entity and repository interface.

**Create `internal/domain/system.go`:**

```go
package domain

import "context"

// SystemInfo represents system information.
type SystemInfo struct {
    Version   string
    GoVersion string
    BuildTime string
}

// SystemInfoProvider defines the contract for system info retrieval.
type SystemInfoProvider interface {
    GetInfo(ctx context.Context) (*SystemInfo, error)
}
```

> 💡 **Note:** Domain layer uses stdlib only - no external imports!

### Step 2: Create the Application Use Case (~5 min)

Create the use case that orchestrates the business logic.

**Create `internal/app/system/get_info.go`:**

```go
package system

import (
    "context"
    "fmt"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// GetInfoUseCase retrieves system information.
type GetInfoUseCase struct {
    provider domain.SystemInfoProvider
}

// NewGetInfoUseCase creates a new GetInfoUseCase.
func NewGetInfoUseCase(provider domain.SystemInfoProvider) *GetInfoUseCase {
    return &GetInfoUseCase{provider: provider}
}

// Execute retrieves system information.
func (uc *GetInfoUseCase) Execute(ctx context.Context) (*domain.SystemInfo, error) {
    info, err := uc.provider.GetInfo(ctx)
    if err != nil {
        return nil, fmt.Errorf("get system info: %w", err)
    }
    return info, nil
}
```

### Step 3: Implement the Infrastructure Adapter (~5 min)

Implement the `SystemInfoProvider` interface in the infrastructure layer.

**Create `internal/infra/system/provider.go`:**

```go
package system

import (
    "context"
    "runtime"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// Version information (set at build time)
var (
    Version   = "dev"
    BuildTime = "unknown"
)

type infoProvider struct{}

// NewInfoProvider creates a new SystemInfoProvider.
func NewInfoProvider() domain.SystemInfoProvider {
    return &infoProvider{}
}

func (p *infoProvider) GetInfo(ctx context.Context) (*domain.SystemInfo, error) {
    return &domain.SystemInfo{
        Version:   Version,
        GoVersion: runtime.Version(),
        BuildTime: BuildTime,
    }, nil
}
```

### Step 4: Create the HTTP Handler (~5 min)

Create the handler with request/response DTOs.

**Create `internal/transport/http/handler/system_handler.go`:**

```go
package handler

import (
    "encoding/json"
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/app/system"
)

// SystemInfoResponse is the API response for system info.
type SystemInfoResponse struct {
    Version   string `json:"version"`
    GoVersion string `json:"go_version"`
    BuildTime string `json:"build_time"`
}

// SystemHandler handles system-related HTTP endpoints.
type SystemHandler struct {
    getInfoUseCase *system.GetInfoUseCase
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(getInfoUseCase *system.GetInfoUseCase) *SystemHandler {
    return &SystemHandler{getInfoUseCase: getInfoUseCase}
}

// GetInfo handles GET /api/v1/system/info.
func (h *SystemHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
    info, err := h.getInfoUseCase.Execute(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(SystemInfoResponse{
        Version:   info.Version,
        GoVersion: info.GoVersion,
        BuildTime: info.BuildTime,
    })
}
```

### Step 5: Wire with Uber Fx (~5 min)

Register the new components with the dependency injection container.

**Update `internal/infra/fx/module.go`:**

```go
// Add these imports
import (
    systemapp "github.com/iruldev/golang-api-hexagonal/internal/app/system"
    systeminfra "github.com/iruldev/golang-api-hexagonal/internal/infra/system"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/handler"
)

// Add to the Module
var Module = fx.Options(
    // ... existing providers ...
    
    // System feature
    fx.Provide(systeminfra.NewInfoProvider),
    fx.Provide(systemapp.NewGetInfoUseCase),
    fx.Provide(handler.NewSystemHandler),
)
```

**Update `internal/transport/http/router.go`:**

```go
// Add route registration
r.Route("/api/v1/system", func(r chi.Router) {
    r.Get("/info", systemHandler.GetInfo)
})
```

### Step 6: Write Tests (~5 min)

Create unit tests using table-driven patterns.

**Create `internal/app/system/get_info_test.go`:**

```go
package system_test

import (
    "context"
    "errors"
    "testing"

    "github.com/iruldev/golang-api-hexagonal/internal/app/system"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Note: In a real workflow, generate mocks with `make mocks` instead of writing manual structs.
// See docs/testing-guide.md for details.
type mockProvider struct {
    info *domain.SystemInfo
    err  error
}

func (m *mockProvider) GetInfo(ctx context.Context) (*domain.SystemInfo, error) {
    return m.info, m.err
}

func TestGetInfoUseCase_Execute(t *testing.T) {
    tests := []struct {
        name      string
        provider  *mockProvider
        wantInfo  *domain.SystemInfo
        wantErr   bool
    }{
        {
            name: "success",
            provider: &mockProvider{
                info: &domain.SystemInfo{
                    Version:   "1.0.0",
                    GoVersion: "go1.25.5",
                    BuildTime: "2026-01-04",
                },
            },
            wantInfo: &domain.SystemInfo{
                Version:   "1.0.0",
                GoVersion: "go1.25.5",
                BuildTime: "2026-01-04",
            },
        },
        {
            name: "error",
            provider: &mockProvider{
                err: errors.New("provider error"),
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc := system.NewGetInfoUseCase(tt.provider)
            got, err := uc.Execute(context.Background())

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.wantInfo, got)
        })
    }
}
```

### Step 7: Verify Your Feature

```bash
# Run linter (checks layer boundaries)
make lint

# Run tests
make test

# Start server and test endpoint
make run

# In another terminal
curl http://localhost:8080/api/v1/system/info
```

> 📚 **More Patterns:** See [docs/patterns.md](patterns.md) and [docs/copy-paste-kit/](copy-paste-kit/) for additional templates.

---

## Testing Guide

**Estimated reading time: ~10 minutes**

### Test Types

| Type | File Pattern | Command | Purpose |
|------|--------------|---------|---------|
| Unit | `*_test.go` | `make test` | Fast, isolated tests |
| Integration | `*_integration_test.go` | `make test-integration` | Tests with real dependencies |
| All | - | `make test-all` | Complete test suite |

### Required Patterns

**Table-Driven Tests (Required):**

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   SomeInput
        want    SomeOutput
        wantErr bool
    }{
        {name: "success", input: validInput, want: expectedOutput},
        {name: "error case", input: invalidInput, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**Coverage Requirements:**

- Domain + App layers: **80% minimum** (enforced by CI)
- New features: Must include unit tests

### Running Tests

```bash
# Run all unit tests
make test

# Run with race detection
make test-shuffle

# Run integration tests (requires Docker)
make test-integration

# Check coverage
make test-coverage
```

> 📚 **Detailed Guide:** See [docs/testing-guide.md](testing-guide.md) for comprehensive testing patterns.

---

## Code Review Checklist

**Estimated time: ~5 minutes per review**

Use this checklist when submitting or reviewing PRs.

### ✅ Layer Boundaries

- [ ] Domain layer imports **stdlib only** (no external packages)
- [ ] App layer does NOT import transport or infra
- [ ] Transport layer does NOT import infra (except fx)
- [ ] `make lint` passes with **0 errors**

### ✅ Error Handling

- [ ] Errors wrapped with context: `fmt.Errorf("context: %w", err)`
- [ ] Domain errors use error codes (e.g., `USR-001`)
- [ ] No bare `return err` without context
- [ ] Panics are NOT used for error handling

### ✅ Context Propagation

- [ ] `context.Context` is **first parameter** in all functions
- [ ] Context is NOT stored in structs
- [ ] External calls use `context.WithTimeout`

### ✅ Testing

- [ ] **Table-driven tests** used for all test cases
- [ ] Coverage ≥80% for domain and app layers
- [ ] New features include unit tests
- [ ] Edge cases and error paths tested

### ✅ Code Quality

- [ ] Functions are ≤100 lines where possible
- [ ] Cyclomatic complexity ≤15 per function
- [ ] No commented-out code
- [ ] Magic numbers are named constants

### ✅ Documentation

- [ ] Exported functions have **godoc comments**
- [ ] Complex logic has inline comments
- [ ] README updated if adding new features

### Quick Reference Table

| Check | Command | Expected |
|-------|---------|----------|
| Linting | `make lint` | 0 errors |
| Tests | `make test` | All pass |
| Coverage | `make test-coverage` | ≥80% domain+app |
| Full CI | `make ci` | All pass |

---

## FAQ and Troubleshooting

**Estimated reading time: ~5 minutes**

### Common Questions

#### Q: Where do I add a new entity?

**A:** Create it in `internal/domain/`. Remember: no JSON tags, stdlib only.

#### Q: Where do I add a new API endpoint?

**A:** Follow the [Your First Feature Tutorial](#your-first-feature-tutorial):
1. Domain entity in `internal/domain/`
2. Use case in `internal/app/`
3. Handler in `internal/transport/http/handler/`
4. Wire in `internal/infra/fx/`

#### Q: Why is my import failing linting?

**A:** You're probably violating layer boundaries. Check the [Layer Boundary Rules](#layer-boundary-rules) section.

#### Q: How do I add a new external dependency?

**A:** Add it to `go.mod` and update `.golangci.yml` if it needs to be allowed in specific layers.

### Common Issues

#### "depguard: import not allowed" Error

```
internal/domain/user.go:5:2: import 'github.com/google/uuid' is not allowed
```

**Solution:** Domain layer cannot import external packages. Move the UUID logic to the infrastructure layer.

#### "goleak: leaked goroutine" Error

```
goleak: Leaked goroutine: goroutine 42 [running]:
```

**Solution:** Check for unclosed resources:
- Database connections not closed
- HTTP clients without timeout
- Channels not drained

#### Tests Pass Locally, Fail in CI

**Common causes:**
1. **Race conditions:** Run `make test-shuffle` locally
2. **Time-dependent tests:** Use proper synchronization
3. **Container startup:** Check Docker is running

#### Container Failed to Start

```
failed to start postgres: Cannot connect to Docker daemon
```

**Solution:**
1. Ensure Docker is running: `docker ps`
2. Check Docker permissions
3. For macOS: Restart Docker Desktop

---

## Next Steps

After completing this guide:

1. ✅ Run the full CI pipeline: `make ci`
2. ✅ Review the [Architecture Documentation](architecture.md)
3. ✅ Explore [Copy-Paste Patterns](patterns.md) for common use cases
4. ✅ Read the [ADRs](adr/index.md) for architectural decisions
5. ✅ Check the [Runbooks](runbooks/index.md) for operational scenarios

### Additional Resources

| Resource | Description |
|----------|-------------|
| [README.md](../README.md) | Quick start and project overview |
| [patterns.md](patterns.md) | Copy-paste code patterns |
| [testing-guide.md](testing-guide.md) | Comprehensive testing guide |
| [adr/index.md](adr/index.md) | Architecture Decision Records |
| [runbooks/index.md](runbooks/index.md) | Operational runbooks |

---

*Welcome to the team! If you have questions not covered here, check the FAQ or ask in your team's communication channel.*

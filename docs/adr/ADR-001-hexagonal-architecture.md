# ADR-001: Hexagonal Architecture

**Status:** Accepted
**Date:** 2026-01-04

## Context

The golang-api-hexagonal project needed an architectural foundation that would:

- Enable high testability without requiring external dependencies (databases, APIs)
- Maintain clear separation between business logic and technical concerns
- Allow swapping implementations (e.g., PostgreSQL → MySQL) without changing business logic
- Support a growing codebase with multiple developers
- Enforce boundaries that prevent accidental coupling between layers

Traditional layered architectures often suffer from:
- Domain logic leaking into infrastructure code
- Difficulty testing business logic in isolation
- Tight coupling between layers making refactoring risky

## Decision

We adopt **Hexagonal Architecture** (also known as Ports and Adapters), organizing the codebase into concentric layers:

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

**Layer Responsibilities:**

| Layer | Purpose | Allowed Imports |
|-------|---------|-----------------|
| `domain/` | Core business entities, interfaces, validation | stdlib only |
| `app/` | Use cases, orchestration | domain, shared |
| `transport/` | HTTP handlers, middleware, DTOs | domain, app, shared |
| `infra/` | Database, config, observability | domain, shared |
| `infra/fx/` | DI wiring (special exception) | ALL layers |

**Key Principles:**
1. Domain layer has zero external dependencies
2. Dependencies point inward (outer layers depend on inner layers)
3. Interfaces are defined in inner layers, implemented in outer layers
4. Uber Fx handles dependency injection at the wiring layer

## Consequences

### Positive

- **Testability**: Domain logic can be unit tested without mocks for databases
- **Flexibility**: Implementations can be swapped via interface abstraction
- **Clarity**: Clear ownership of code based on location
- **Maintainability**: Changes to one layer don't cascade to others
- **Onboarding**: New developers understand where code belongs

### Negative

- **Boilerplate**: Interface definitions add some overhead
- **Learning curve**: Developers must understand layer rules
- **Indirection**: Additional abstraction layers to navigate

### Neutral

- Requires discipline to maintain boundaries
- Enforcement mechanism needed (see [ADR-002](./ADR-002-layer-boundary-enforcement.md))

## Related ADRs

- [ADR-002: Layer Boundary Enforcement](./ADR-002-layer-boundary-enforcement.md) - Enforces these architecture rules via golangci-lint
- [ADR-003: Resilience Patterns](./ADR-003-resilience-patterns.md) - Resilience patterns implemented in infra layer

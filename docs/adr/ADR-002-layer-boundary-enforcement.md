# ADR-002: Layer Boundary Enforcement with depguard

**Status:** Accepted
**Date:** 2026-01-04

## Context

The Hexagonal Architecture defined in [ADR-001](./ADR-001-hexagonal-architecture.md) establishes clear layer boundaries:

- Domain layer can only use stdlib
- App layer cannot import transport or infra
- Transport layer cannot import infra directly

Without enforcement, these rules will inevitably be violated:

- Developers may not be aware of boundaries
- Time pressure leads to shortcuts
- Layer violations are subtle and often missed in code review
- Once a violation exists, more tend to follow

Manual code review is insufficient for consistent enforcement.

## Decision

We use **golangci-lint's depguard** rules in `.golangci.yml` to enforce layer boundaries at build/CI time.

**Configuration Structure:**

```yaml
linters:
  settings:
    depguard:
      rules:
        domain-layer:
          list-mode: strict
          files:
            - "**/internal/domain/**/*.go"
          allow:
            - $gostd
            - github.com/iruldev/golang-api-hexagonal/internal/domain
          deny:
            - pkg: github.com/iruldev/golang-api-hexagonal/internal/app
              desc: "Domain layer cannot import app layer"
            - pkg: github.com/iruldev/golang-api-hexagonal/internal/transport
              desc: "Domain layer cannot import transport layer"
            # ... additional deny rules
```

**Enforcement Layers:**

| Rule Name | Target Files | Allowed Imports |
|-----------|--------------|-----------------|
| `domain-layer` | `internal/domain/**/*.go` | stdlib only |
| `app-layer` | `internal/app/**/*.go` | domain, app, shared |
| `shared-layer` | `internal/shared/**/*.go` | domain, shared |
| `transport-layer` | `internal/transport/**/*.go` | domain, app, shared, transport, external HTTP packages |
| `infra-layer` | `internal/infra/{config,postgres,resilience,observability}/**/*.go` | domain, shared, infra, external DB packages |
| `wiring-layer` | `internal/infra/fx/**/*.go` | ALL (special exception) |

**Test Files:**
- Separate rules with `list-mode: lax` for test files
- Allow test utilities (testify, testcontainers) in tests

## Consequences

### Positive

- **Immediate feedback**: Violations caught at lint time, not production
- **CI enforcement**: PRs with violations automatically fail
- **Documentation**: Error messages explain why import is forbidden
- **Consistency**: Rules apply equally to all developers
- **Discoverability**: New developers learn boundaries from lint errors

### Negative

- **Initial friction**: Developers may need to refactor code to comply
- **Configuration complexity**: `.golangci.yml` has many rules to maintain
- **False positives**: May block legitimate use cases requiring exceptions

### Neutral

- Requires updating configuration when adding new packages
- Rules must be documented alongside architecture

## Related ADRs

- [ADR-001: Hexagonal Architecture](./ADR-001-hexagonal-architecture.md) - Defines the layers being enforced

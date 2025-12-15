# Policy Pack

This directory is the **single source of truth** for all project-wide configurations.

## Files

| File | Purpose |
|------|---------|
| `golangci.yml` | golangci-lint v2 configuration (includes depguard) |
| `depguard.yml` | Layer boundary rules placeholder (rules moved inline) |
| `error-codes.yml` | Public error code registry |
| `log-fields.yml` | Approved log field names for structured logging |

## Layer Boundary Enforcement (depguard)

The depguard linter enforces hexagonal architecture layer boundaries:

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| `domain` | (nothing) | usecase, interface, infra |
| `usecase` | domain | interface, infra |
| `interface` | domain, usecase | infra |
| `infra` | domain | - |

**Error messages include helpful descriptions explaining the violation.**

### Example Violation

```
import 'internal/infra/postgres' is not allowed from list 'usecase-layer': 
usecase layer cannot import infra - use dependency injection instead
```

### Cross-Cutting Packages

These packages are allowed from all layers:
- `internal/ctxutil` - Context utilities (claims, request ID)
- `internal/observability` - Logging, metrics, tracing
- `internal/runtimeutil` - Runtime utilities
- `internal/testing` - Test utilities
- `internal/config` - Configuration loading

## Usage

All CI and local tooling reads from this directory:

```bash
# Linting with policy pack
make lint  # Uses policy/golangci.yml

# Layer boundary validation happens automatically via depguard
```

## Adding New Policies

When adding new policy files:

1. Create the file in this directory
2. Update the table above
3. Update `Makefile` if a new make target is needed
4. Document usage in `AGENTS.md` for AI agent context

## Related Documentation

- [Architecture Decisions](../docs/architecture-decisions.md) - Policy Pack architectural decision
- [Project Context](../project_context.md) - Critical patterns and anti-patterns

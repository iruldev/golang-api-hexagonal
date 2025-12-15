# Policy Pack

This directory is the **single source of truth** for all project-wide configurations.

## Files

| File | Purpose |
|------|---------|
| `golangci.yml` | golangci-lint v2 configuration |
| `depguard.yml` | Layer boundary rules (hexagonal architecture) |
| `error-codes.yml` | Public error code registry |
| `log-fields.yml` | Approved log field names for structured logging |

## Usage

All CI and local tooling reads from this directory:

```bash
# Linting with policy pack
make lint  # Uses policy/golangci.yml

# Layer boundary validation (Story 1.2)
depguard → reads policy/depguard.yml
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

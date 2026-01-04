# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) documenting key architectural decisions for the golang-api-hexagonal project.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](./ADR-001-hexagonal-architecture.md) | Hexagonal Architecture | Accepted | 2026-01-04 |
| [ADR-002](./ADR-002-layer-boundary-enforcement.md) | Layer Boundary Enforcement with depguard | Accepted | 2026-01-04 |
| [ADR-003](./ADR-003-resilience-patterns.md) | Resilience Patterns (Circuit Breaker, Retry, Timeout) | Accepted | 2026-01-04 |
| [ADR-004](./ADR-004-rfc7807-error-handling.md) | RFC 7807 Error Handling | Accepted | 2026-01-04 |
| [ADR-005](./ADR-005-idempotency-key-implementation.md) | Idempotency Key Implementation | Accepted | 2026-01-04 |

## Creating New ADRs

1. Copy the [template](./template.md)
2. Name the file `ADR-XXX-short-title.md` (use next available number)
3. Fill in all sections: Context, Decision, Consequences
4. Set status to "Proposed" for discussion or "Accepted" for finalized decisions
5. Update this index with the new ADR

## Status Definitions

- **Proposed** - Under discussion, not yet finalized
- **Accepted** - Decision has been made and is in effect
- **Superseded** - Replaced by a newer decision (link to replacement)
- **Deprecated** - No longer recommended but not replaced

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Michael Nygard's Original Blog Post](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)

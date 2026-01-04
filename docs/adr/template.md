# ADR Template

Use this template when documenting architectural decisions.

---

```markdown
# ADR-XXX: [Decision Title]

**Status:** Accepted | Superseded | Deprecated
**Date:** YYYY-MM-DD
**Supersedes:** [ADR-XXX](./ADR-XXX-title.md) (if applicable)
**Superseded by:** [ADR-XXX](./ADR-XXX-title.md) (if applicable)

## Context

Describe the issue that motivates this decision. Include:
- What problem are we solving?
- What constraints exist?
- What forces are at play?

## Decision

Describe the decision that was made. Use active voice.
- What approach are we taking?
- What alternatives were considered?

## Consequences

What becomes easier or more difficult because of this decision?

### Positive
- List benefits

### Negative
- List tradeoffs

### Neutral
- Other impacts

## Related ADRs

- [ADR-XXX](./ADR-XXX-title.md) - Description of relationship
```

---

## Guidelines

1. **One decision per ADR** - Keep ADRs focused on a single decision
2. **Immutable once accepted** - Don't modify accepted ADRs; create new ones to supersede
3. **Use clear language** - Avoid vague terms like "might", "possibly", "should consider"
4. **Date format** - Always use YYYY-MM-DD format
5. **Status values** - Use only: Accepted, Superseded, Deprecated, Proposed

## Numbering

- ADRs are numbered sequentially: ADR-001, ADR-002, etc.
- Leading zeros maintain sort order in file listings
- Never reuse ADR numbers, even for superseded decisions

# ADR-003: Resilience Patterns (Circuit Breaker, Retry, Timeout)

**Status:** Accepted
**Date:** 2026-01-04

## Context

Production services face various failure scenarios:

- **Downstream service failures**: APIs, databases becoming unavailable
- **Transient errors**: Network blips, temporary overload
- **Cascading failures**: One failing service overwhelming others
- **Resource exhaustion**: Connection pools depleted while waiting

Without resilience patterns:

- A slow database blocks all requests
- Failed external API calls are not retried
- Downstream failures cascade to the entire system
- No automatic recovery when services restore

The service needed automatic failure handling without manual intervention.

## Decision

We implement a **resilience package** (`internal/infra/resilience/`) with three core patterns:

### Circuit Breaker (using sony/gobreaker)

```go
type CircuitBreaker interface {
    Execute(ctx context.Context, fn func() (any, error)) (any, error)
    State() State  // closed, open, half-open
    Name() string
}
```

**Behavior:**
- **Closed**: Normal operation, requests pass through
- **Open**: After consecutive failures exceed threshold, reject requests immediately (return `ErrCircuitOpen`)
- **Half-open**: After timeout, allow limited requests to test recovery

**Configuration:**
```go
CircuitBreakerConfig{
    MaxRequests:      3,          // Requests in half-open state
    Interval:         60s,        // Counting interval
    Timeout:          30s,        // Time before half-open attempt
    FailureThreshold: 5,          // Failures before opening
}
```

### Retry with Exponential Backoff

**Behavior:**
- Retry failed operations up to max attempts
- Delay increases exponentially: `initial_delay * multiplier^attempt`
- Jitter added to prevent thundering herd

### Timeout Wrapper

**Behavior:**
- `context.WithTimeout` applied to operations
- Per-operation type configuration (database: 5s, external: 10s)
- Context propagation through all layers

### Integration Pattern

Resilience patterns compose as middleware:
```
Timeout → Retry → CircuitBreaker → Actual Operation
```

## Consequences

### Positive

- **Auto-recovery**: System recovers when dependencies restore
- **Fail-fast**: Broken circuits prevent wasted resources
- **Predictable latency**: Timeouts bound response times
- **Observability**: Metrics for circuit state, retry counts

### Negative

- **Complexity**: Additional abstraction layer
- **Tuning required**: Thresholds need monitoring and adjustment
- **Learning curve**: Developers must understand patterns

### Neutral

- Patterns are opt-in per operation
- Configuration via environment/config file

## Related ADRs

- [ADR-001: Hexagonal Architecture](./ADR-001-hexagonal-architecture.md) - Resilience lives in infra layer per hexagonal rules

# Runbook: Circuit Breaker Open

**Severity:** 🟠 High
**Last Updated:** 2026-01-04
**Owner:** DevOps Team
**Related ADRs:** [ADR-003: Resilience Patterns](../adr/ADR-003-resilience-patterns.md)

---

## Overview

This runbook addresses scenarios where the circuit breaker has opened due to downstream service failures. When open, the circuit breaker rejects requests immediately to protect the system from cascading failures.

**Impact:** Requests to affected downstream services fail fast with `RES-001` error code. Service operates in degraded mode.

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] Prometheus/Grafana dashboard access
  - [ ] Downstream service access (for diagnosis)

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] `curl` for API testing
  - [ ] Access to monitoring dashboards

- **Knowledge Required:**
  - Circuit breaker pattern (open/half-open/closed states)
  - sony/gobreaker library behavior
  - Resilience package configuration

---

## Symptoms

### 🔔 Alerts

| Alert Name | Condition | Dashboard Link |
|------------|-----------|----------------|
| `CircuitBreakerOpen` | `circuit_breaker_state{state="open"} == 1` | Resilience Dashboard |
| `HighRES001Errors` | `error_code_total{code="RES-001"}` count > threshold | Error Dashboard |
| `DownstreamUnhealthy` | Downstream service health failing | Dependencies Dashboard |

### 📊 Metrics

- `circuit_breaker_state{name="xxx", state="open"}` = 1
- `circuit_breaker_operation_duration_seconds_count{result="rejected"}` increasing
- `circuit_breaker_transitions_total` increasing
- Error responses with code `RES-001` (Circuit Open)

### 🐛 User Reports

- "Service temporarily unavailable" errors
- Specific features not working
- Error message mentions "circuit breaker" or "RES-001"

---

## Investigation Steps

### Step 1: Identify Which Circuit Breaker is Open

```bash
# Check circuit breaker metrics
curl -s http://${API_HOST}:8081/metrics | grep circuit_breaker_state

# Example output:
# circuit_breaker_state{name="database", state="closed"} 1
# circuit_breaker_state{name="external_api", state="open"} 1
```

**What to look for:** Identify which circuit breaker(s) show `state="open"`

### Step 2: Check Circuit Breaker History

```bash
# View recent state transitions in logs
kubectl logs -l app=api-server --tail=200 | grep -i "circuit"

# Look for patterns like:
# Circuit breaker 'database' state: closed -> open
# Circuit breaker 'external_api' opened after 5 failures
```

### Step 3: Investigate Downstream Service

```bash
# If database circuit breaker:
curl -s http://${API_HOST}:8080/readyz | jq .

# If external API circuit breaker:
# Direct test to downstream service
curl -v http://${DOWNSTREAM_HOST}/health

# Check downstream service pods
kubectl get pods -l app=${DOWNSTREAM_APP} -o wide
```

### Step 4: Review Error Patterns

```bash
# Check error rates before circuit opened
curl -s http://${API_HOST}:8081/metrics | grep -E "(error|failure)"

# Review application logs for root cause
kubectl logs -l app=api-server --tail=500 | grep -E "(error|failed|timeout)"
```

### Step 5: Check Circuit Breaker Configuration

```bash
# View current configuration
kubectl exec -it $(kubectl get pod -l app=api-server -o jsonpath='{.items[0].metadata.name}') -- \
  env | grep -i CIRCUIT

# Expected settings:
# CB_FAILURE_THRESHOLD=5
# CB_TIMEOUT=30s
# CB_MAX_REQUESTS=3
```

---

## Resolution Steps

### Immediate Mitigation

> ⚡ **Time-Critical:** Complete within 10 minutes

1. **Wait for half-open transition** (automatic after timeout):
   ```bash
   # Circuit breaker timeout is typically 30s
   # Monitor state transition:
   watch -n 5 'curl -s http://${API_HOST}:8081/metrics | grep circuit_breaker_state'
   ```

2. **If downstream is healthy, manually trigger recovery:**
   ```bash
   # Restart pods to reset circuit breaker state
   kubectl rollout restart deployment/api-server
   ```

### Root Cause Fix

> 🔧 **Standard Resolution:** Fix the underlying downstream issue

1. **Database circuit breaker open:**
   - Follow [Database Connection Failure](./database-connection-failure.md) runbook
   - Check database health and connectivity
   - Review connection pool settings

2. **External API circuit breaker open:**
   - Contact external service team
   - Check external service status page
   - Review network connectivity to external service
   - Check API rate limits

3. **Internal service circuit breaker open:**
   - Check internal service health
   - Review service logs for errors
   - Scale up internal service if overloaded

### Configuration Tuning (if needed)

```yaml
# Adjust circuit breaker settings in config
resilience:
  circuit_breaker:
    failure_threshold: 5      # Failures before opening
    timeout: 30s              # Time in open state before half-open
    max_requests: 3           # Requests allowed in half-open state
    interval: 10s             # Time window for counting failures
```

---

## Rollback Procedure

### When to Rollback

- Recent deployment changed circuit breaker configuration
- Circuit breaker is overly sensitive (opens too easily)
- Configuration change caused false positives

### Rollback Steps

1. **Rollback deployment:**
   ```bash
   kubectl rollout undo deployment/api-server
   ```

2. **Verify circuit breaker state:**
   ```bash
   curl -s http://${API_HOST}:8081/metrics | grep circuit_breaker_state
   ```

3. **Temporarily disable circuit breaker (emergency only):**
   ```bash
   # Set very high threshold to effectively disable
   kubectl set env deployment/api-server CB_FAILURE_THRESHOLD=1000
   ```
   
   ⚠️ **Warning:** Only use in emergency. Re-enable circuit breaker once root cause is fixed.

### Verify Rollback Success

- [ ] Circuit breaker in closed state
- [ ] No `RES-001` errors in logs
- [ ] Downstream service healthy
- [ ] API requests succeeding

---

## Verification

### Confirm Resolution

- [ ] `circuit_breaker_state{state="closed"}` for all circuit breakers
- [ ] No new `RES-001` errors for 5 minutes
- [ ] Downstream service responding successfully
- [ ] API latency returned to baseline

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| `circuit_breaker_state{state="open"}` | 0 | Resilience Dashboard |
| `circuit_breaker_operation_duration_seconds_count{result="success"}` | Increasing | Resilience Dashboard |
| `error_code_total{code="RES-001"}` | Stable (not increasing) | Error Dashboard |
| Downstream service latency | < threshold | Dependencies Dashboard |

### Observation Period

⏱️ **Recommended:** Monitor for 15 minutes to ensure circuit stays closed

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call Engineer | First response | Immediate |
| L2 | Senior DevOps | Circuit reopens after fix | 15 min |
| L3 | Platform Team | Systemic downstream failure | 10 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: Circuit keeps reopening, downstream recovery unclear
- 🚨 Escalate to L3 if: Multiple circuit breakers open, cascading failures

---

## Related Runbooks

- [Database Connection Failure](./database-connection-failure.md) - If database circuit breaker opens
- [Deployment Rollback](./deployment-rollback.md) - If deployment caused circuit breaker issues
- [Memory/CPU Spike](./memory-cpu-spike.md) - Resource exhaustion can cause downstream failures

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | DevOps Team | Initial creation |

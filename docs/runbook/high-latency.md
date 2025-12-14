# Runbook: High Latency

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `HighLatency`, `HighLatencyCritical` |
| **Severity** | warning / critical |
| **Component** | http |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for elevated HTTP response latency.

**What triggers this alert:**
- **HighLatency (Warning):** p95 latency > 500ms for 5 minutes
- **HighLatencyCritical (Critical):** p95 latency > 1 second for 2 minutes

**Business impact:**
- Poor user experience with slow page loads
- API timeouts causing client-side errors
- Potential cascade failures in dependent services

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the app runs via `make dev` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal, use `ps aux | grep server`).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert in PagerDuty/Slack
- [ ] **Assess** - Check which endpoints are slow
- [ ] **Investigate** - Follow diagnosis steps below
- [ ] **Mitigate** - Scale resources or optimize queries
- [ ] **Communicate** - Update #incidents if user-facing impact
- [ ] **Document** - Record findings and actions
- [ ] **Resolve** - Confirm latency returns to normal

---

## Symptoms

### Observable Indicators
- Slow API responses reported by users
- Client-side timeouts
- Increased queue depth in dependent services

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `http_request_duration_seconds` p95 | < 100ms | > 500ms (warning), > 1s (critical) |
| `db_query_duration_seconds` p95 | < 50ms | > 100ms |

### Related Alerts
- `DBSlowQueries` - Often fires together
- `HighErrorRate` - Timeouts may cause 5xx errors
- `JobQueueBacklog` - If jobs are timing out

---

## Diagnosis

### Step 1: Identify Slow Endpoints

**Goal:** Find which endpoints are causing the latency

```bash
# Check Prometheus for slow endpoints
curl -s 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,path))' | jq

# Alternative: Check raw metrics
curl -s http://localhost:8080/metrics | grep http_request_duration
```

**Expected outcome:** List of endpoints sorted by p95 latency

### Step 2: Check Database Performance

**Goal:** Identify slow database queries

```bash
# Check active queries in PostgreSQL
docker exec -it postgres psql -U postgres -c "
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC
LIMIT 10;"
```

### Step 3: Check Application Resources

**Goal:** Identify resource constraints

```bash
# Check CPU and memory
docker stats golang-api --no-stream

# Check Go runtime metrics
curl -s http://localhost:8080/metrics | grep go_
```

### Step 4: Check External Dependencies

**Goal:** Verify external service latency

```bash
# Check external API response times in logs
docker logs golang-api 2>&1 | grep -i "external\|timeout" | tail -20
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Slow DB queries | High query duration | EXPLAIN ANALYZE | Add indexes, optimize queries |
| N+1 queries | Many small queries | Log query count | Use joins or batch queries |
| Resource contention | High CPU/memory | `docker stats` | Scale horizontally or vertically |
| External API slow | Timeout in logs | Check external status | Add caching, increase timeout |
| GC pressure | GC pause spikes | Go metrics | Reduce allocations, tune GC |

---

## Remediation

### Immediate Actions

1. **If database issue:** Check slow queries
   ```sql
   -- Find slow queries
   SELECT pid, now() - query_start AS duration, query
   FROM pg_stat_activity
   WHERE state = 'active' AND now() - query_start > interval '1 second';
   
   -- Kill long-running query if necessary
   SELECT pg_cancel_backend(pid);
   ```

2. **If resource contention:** Scale resources
   ```bash
   # Scale horizontally
   kubectl scale deployment golang-api --replicas=3
   
   # Or increase resource limits
   kubectl patch deployment golang-api -p '{"spec":{"template":{"spec":{"containers":[{"name":"golang-api","resources":{"limits":{"cpu":"2","memory":"4Gi"}}}]}}}}'
   ```

3. **If external API slow:** Enable caching
   - Check if caching layer is available
   - Consider circuit breaker pattern

### Post-Incident Actions

1. **Root Cause Analysis:** Identify which queries or operations are slow
2. **Prevention:** Add query indexes, implement caching
3. **Monitoring:** Set up query duration alerts

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Warning** | Investigate within 30 minutes | Escalate after 1 hour |
| **Critical** | Immediate investigation | Escalate after 15 minutes |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** Team lead / Senior engineer
3. **Level 2 Escalation:** DBA (if database-related)

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | First response |
| Team Lead | @team-lead | After 15 min (critical) |
| DBA | @dba-team | If database-related |

---

## References

- [Grafana HTTP Dashboard](http://localhost:3000/d/http)
- [DB Slow Queries Runbook](db-slow-queries.md)
- [PostgreSQL EXPLAIN Documentation](https://www.postgresql.org/docs/current/sql-explain.html)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |

# Runbook: Database Connection Exhausted

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `DBConnectionExhausted` |
| **Severity** | warning |
| **Component** | database |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for database connectivity issues and connection pool exhaustion.

**What triggers this alert:**
- **DBConnectionExhausted (Warning):** /readyz endpoint failures > 20% for 5 minutes

**Business impact:**
- API requests failing due to database unavailability
- Data read/write operations affected
- Potential user-facing errors

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the app runs via `make dev` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal, use `ps aux | grep server`).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert
- [ ] **Assess** - Check readiness endpoint and connection pool
- [ ] **Investigate** - Follow diagnosis steps below
- [ ] **Mitigate** - Restart service or increase connections
- [ ] **Communicate** - Update team if widespread impact
- [ ] **Document** - Record findings and root cause
- [ ] **Resolve** - Confirm DB connectivity is restored

---

## Symptoms

### Observable Indicators
- API requests timing out or returning 5xx errors
- `/readyz` endpoint returning 503
- "connection pool exhausted" errors in logs

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `/readyz` failure rate | 0% | > 20% |
| `db_pool_connections_in_use` *(future)* | < 80% of max | > 90% |
| PostgreSQL `pg_stat_activity` | < max_connections | approaching limit |

> **Note:** The `db_pool_connections_in_use` metric is planned but not yet implemented. Currently, the alert uses `/readyz` endpoint failure rate as a proxy for database connectivity issues.

### Related Alerts
- `HighErrorRate` - Likely fires together
- `HighLatency` - DB issues cause latency
- `DBSlowQueries` - May be root cause of connection exhaustion

---

## Diagnosis

### Step 1: Check Database Status

**Goal:** Verify PostgreSQL is running and accessible

```bash
# Check PostgreSQL status
docker exec -it postgres pg_isready

# For external database
pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER
```

**Expected outcome:** "accepting connections"

### Step 2: Check Connection Count

**Goal:** Verify current connection usage

```bash
# Check active connections in PostgreSQL
docker exec -it postgres psql -U postgres -c "
SELECT count(*) as total_connections,
       count(*) FILTER (WHERE state = 'active') as active,
       count(*) FILTER (WHERE state = 'idle') as idle
FROM pg_stat_activity;"

# Check max connections setting
docker exec -it postgres psql -U postgres -c "SHOW max_connections;"
```

### Step 3: Check Application Connection Pool

**Goal:** Identify if application is leaking connections

```bash
# Check application metrics
curl -s http://localhost:8080/metrics | grep db_pool

# Check for connection-related errors in logs
docker logs golang-api 2>&1 | grep -i "connection\|pool\|pgx" | tail -20
```

### Step 4: Identify Long-Running Queries

**Goal:** Find queries holding connections

```sql
-- Run on PostgreSQL
SELECT pid, now() - pg_stat_activity.query_start AS duration,
       query, state, wait_event_type
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC
LIMIT 10;
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Connection leak | Growing connections | Pool never decreases | Find/fix leak, restart |
| Too many clients | Max connections hit | `pg_stat_activity` count | Increase limit or scale |
| Long-running query | Connections held | Check query duration | Kill query, optimize |
| Network issue | Connection timeouts | Network diagnostics | Check DNS/firewall |
| Slow queries | Pool backup | High query duration | Optimize queries |

---

## Remediation

### Immediate Actions

1. **Restart application to release connections:**
   ```bash
   docker restart golang-api
   
   # For Kubernetes
   kubectl rollout restart deployment/golang-api
   ```

2. **Kill idle connections in PostgreSQL:**
   ```sql
   -- Terminate idle connections older than 10 minutes
   SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE state = 'idle'
     AND query_start < now() - interval '10 minutes'
     AND pid != pg_backend_pid();
   ```

3. **Increase max connections (temporary):**
   ```sql
   -- Increase PostgreSQL max_connections (requires restart)
   ALTER SYSTEM SET max_connections = 200;
   SELECT pg_reload_conf();
   ```

4. **Adjust application pool settings:**
   ```bash
   # Update environment variables
   export DB_MAX_OPEN_CONNS=30
   export DB_MAX_IDLE_CONNS=10
   ```

### Post-Incident Actions

1. **Root Cause Analysis:** Identify why connections were exhausted
2. **Prevention:** Implement connection leak detection
3. **Monitoring:** Add connection pool metrics dashboard

---

## Configuration Reference

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `DB_MAX_OPEN_CONNS` | 20 | Maximum open connections to database |
| `DB_MAX_IDLE_CONNS` | 5 | Maximum idle connections in pool |
| `DB_CONN_MAX_LIFETIME` | 30m | Maximum lifetime of a connection |
| `DB_CONN_MAX_IDLE_TIME` | 5m | Maximum idle time before closing |

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Warning** | Investigate within 30 minutes | Escalate after 1 hour |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** DBA / Senior engineer
3. **Level 2 Escalation:** Engineering manager

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | First response |
| DBA | @dba-team | If connection limit adjustments needed |
| Team Lead | @team-lead | If requires code changes |

---

## References

- [PostgreSQL Connection Pooling](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [pgx Connection Pool Documentation](https://github.com/jackc/pgx)
- [High Latency Runbook](high-latency.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |

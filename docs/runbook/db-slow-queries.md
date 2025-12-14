# Runbook: Database Slow Queries

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `DBSlowQueries` |
| **Severity** | warning |
| **Component** | database |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for slow database query performance.

**What triggers this alert:**
- **DBSlowQueries (Warning):** API p95 latency > 500ms for 5 minutes (proxy for DB latency)

**Business impact:**
- Slow API responses affecting user experience
- Potential cascade effects on dependent services
- Database resource contention

> **Deployment Context:** Database commands (PostgreSQL) apply to all environments. Docker commands for Docker deployments; for local development (`make dev`), check application logs in the terminal.

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert
- [ ] **Assess** - Check which queries/endpoints are slow
- [ ] **Investigate** - Find slow queries using diagnosis steps
- [ ] **Mitigate** - Kill long-running queries if needed
- [ ] **Communicate** - Update team if user-facing
- [ ] **Document** - Record slow queries found
- [ ] **Resolve** - Confirm latency returns to normal

---

## Symptoms

### Observable Indicators
- API endpoints responding slowly
- Database CPU usage elevated
- Connection pool filling up due to slow queries

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `http_request_duration_seconds` p95 | < 100ms | > 500ms |
| PostgreSQL query duration | < 50ms | > 100ms |
| Table sequential scans | Low | High (missing index) |

### Related Alerts
- `HighLatency` - Often fires together
- `DBConnectionExhausted` - Slow queries hold connections
- `HighErrorRate` - Timeouts may cause errors

---

## Diagnosis

### Step 1: Enable and Check Slow Query Logging

**Goal:** Identify which queries are slow

```sql
-- Enable slow query logging (logs queries > 500ms)
ALTER SYSTEM SET log_min_duration_statement = 500;
SELECT pg_reload_conf();

-- Check PostgreSQL logs
-- Docker:
docker logs postgres 2>&1 | grep "duration:" | tail -20
```

### Step 2: Identify Currently Running Long Queries

**Goal:** Find queries currently executing for too long

```sql
-- Find long-running queries
SELECT pid,
       now() - pg_stat_activity.query_start AS duration,
       usename,
       query,
       state
FROM pg_stat_activity
WHERE state != 'idle'
  AND now() - pg_stat_activity.query_start > interval '1 second'
ORDER BY duration DESC
LIMIT 10;
```

### Step 3: Analyze Query Execution Plan

**Goal:** Understand why a query is slow

```sql
-- Get execution plan for a slow query
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
<your_slow_query_here>;
```

**Look for:**
- Sequential scans on large tables (missing index)
- High actual vs estimated rows (stale statistics)
- Nested loops with high row counts

### Step 4: Check Table Statistics

**Goal:** Identify tables with missing indexes or stale stats

```sql
-- Check for sequential scans vs index scans
SELECT schemaname, relname, 
       seq_scan, seq_tup_read,
       idx_scan, idx_tup_fetch,
       n_live_tup
FROM pg_stat_user_tables
WHERE seq_scan > 1000
ORDER BY seq_tup_read DESC
LIMIT 10;

-- Check if statistics are stale
SELECT schemaname, relname, last_analyze, last_autoanalyze
FROM pg_stat_user_tables
ORDER BY last_analyze ASC NULLS FIRST;
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Missing index | Sequential scan | EXPLAIN shows Seq Scan | Add appropriate index |
| N+1 queries | Many small queries | Check query count | Use joins or batch |
| Lock contention | Waiting on locks | Check wait_event_type | Optimize transactions |
| Large result sets | High memory usage | Check rows returned | Add pagination |
| Stale statistics | Bad query plans | Check last_analyze | Run ANALYZE |

---

## Remediation

### Immediate Actions

1. **Kill long-running queries (if blocking):**
   ```sql
   -- Cancel query gracefully
   SELECT pg_cancel_backend(pid);
   
   -- Force terminate if needed
   SELECT pg_terminate_backend(pid);
   ```

2. **Update table statistics:**
   ```sql
   -- Analyze specific table
   ANALYZE table_name;
   
   -- Analyze entire database
   ANALYZE;
   ```

3. **Add missing index (if identified):**
   ```sql
   -- Create index (may take time on large tables)
   CREATE INDEX CONCURRENTLY idx_table_column 
   ON table_name(column_name);
   ```

4. **Increase work_mem temporarily:**
   ```sql
   -- For current session only
   SET work_mem = '256MB';
   ```

### Post-Incident Actions

1. **Root Cause Analysis:** Document which queries were slow and why
2. **Prevention:** Add indexes, optimize queries in application
3. **Monitoring:** Set up query performance monitoring

---

## Useful PostgreSQL Commands

```sql
-- Check index usage
SELECT indexrelname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Check table sizes
SELECT relname, pg_size_pretty(pg_total_relation_size(relid))
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 10;

-- Check for unused indexes
SELECT indexrelname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;

-- Check for bloated tables
SELECT schemaname, relname, n_dead_tup, n_live_tup,
       round(n_dead_tup::numeric / (n_live_tup + 1) * 100, 2) as dead_pct
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY dead_pct DESC;
```

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Warning** | Investigate within 30 minutes | Escalate after 1 hour |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** DBA
3. **Level 2 Escalation:** Database architect

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | First response |
| DBA | @dba-team | For index creation or major changes |

---

## References

- [PostgreSQL EXPLAIN Documentation](https://www.postgresql.org/docs/current/sql-explain.html)
- [Use The Index, Luke](https://use-the-index-luke.com/)
- [High Latency Runbook](high-latency.md)
- [DB Connection Exhausted Runbook](db-connection-exhausted.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |

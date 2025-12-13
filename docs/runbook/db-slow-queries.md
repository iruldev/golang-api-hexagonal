# Runbook: Database Slow Queries Alert

## Overview
This runbook addresses alerts for slow database query performance.

## Alert Covered
- `DBSlowQueries` - Warning: API p95 latency > 500ms for 5 minutes

## Investigation Steps

### 1. Enable Query Logging
```sql
-- In PostgreSQL, enable slow query logging
ALTER SYSTEM SET log_min_duration_statement = 500;
SELECT pg_reload_conf();
```

### 2. Identify Slow Queries
```sql
-- Find long-running queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY duration DESC
LIMIT 10;
```

### 3. Check Query Plan
```sql
EXPLAIN ANALYZE <slow_query>;
```

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Missing index | Sequential scan | Add appropriate index |
| N+1 queries | Many small queries | Use joins or batch |
| Lock contention | Waiting on locks | Optimize transactions |
| Large result sets | High memory | Add pagination |

## Resolution Actions

1. **If missing index**: Add index based on EXPLAIN output
2. **If N+1**: Optimize application code
3. **If lock contention**: Review transaction scope

## Useful PostgreSQL Commands
```sql
-- Check table statistics
SELECT schemaname, relname, seq_scan, idx_scan, n_live_tup
FROM pg_stat_user_tables
ORDER BY seq_scan DESC;

-- Check index usage
SELECT indexrelname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

## Escalation
If unresolved after 15 minutes, escalate to DBA.

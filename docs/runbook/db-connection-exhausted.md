# Runbook: Database Connection Exhausted Alert

## Overview
This runbook addresses alerts for database connectivity issues.

## Alert Covered
- `DBConnectionExhausted` - Warning: /readyz failures > 20% for 5 minutes

## Investigation Steps

### 1. Check Database Status
```bash
# Check PostgreSQL status
docker exec -it postgres pg_isready

# Check active connections
docker exec -it postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"
```

### 2. Check Connection Pool
```bash
# Check application connection pool metrics
curl -s http://localhost:8080/metrics | grep db_pool
```

### 3. Check Database Resources
- CPU utilization
- Memory usage
- Disk I/O
- Max connections setting

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Connection leak | Growing connections | Find and fix leak |
| Too many clients | Max connections | Increase limit or scale |
| Database overload | Slow queries | Optimize queries |
| Network issue | Timeout | Check network |

## Resolution Actions

1. **If connection leak**: Restart service, fix code
2. **If max connections**: Increase `max_connections` in PostgreSQL
3. **If overload**: Scale database or optimize queries

## Configuration
Check these environment variables:
- `DB_MAX_OPEN_CONNS` - Default: 20
- `DB_MAX_IDLE_CONNS` - Default: 5
- `DB_CONN_MAX_LIFETIME` - Default: 30m

## Escalation
If unresolved after 15 minutes, escalate to DBA or on-call engineer.

# Runbook: High Latency Alert

## Overview
This runbook addresses alerts for elevated HTTP response latency.

## Alerts Covered
- `HighLatency` - Warning: p95 latency > 500ms for 5 minutes
- `HighLatencyCritical` - Critical: p95 latency > 1 second for 2 minutes

## Investigation Steps

### 1. Identify Slow Endpoints
```bash
# Check Prometheus for slow endpoints
curl -s 'http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,path))'
```

### 2. Check Database Performance
- Review slow query logs
- Check connection pool usage
- Verify database CPU/memory

### 3. Check Application Resources
- CPU utilization
- Memory pressure
- Garbage collection pauses

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Slow DB queries | High query time | Add indexes, optimize queries |
| N+1 queries | Many small queries | Batch or join queries |
| Resource contention | High CPU/memory | Scale resources |
| External API slow | Timeout errors | Add caching, increase timeout |

## Resolution Actions

1. **If database issue**: Check `DBSlowQueries` runbook
2. **If resource issue**: Scale horizontally or vertically
3. **If code issue**: Profile and optimize

## Escalation
If unresolved after 15 minutes, escalate to on-call engineer.

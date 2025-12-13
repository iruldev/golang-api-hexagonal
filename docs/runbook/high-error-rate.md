# Runbook: High Error Rate Alert

## Overview
This runbook addresses alerts for elevated HTTP error rates (5xx responses).

## Alerts Covered
- `HighErrorRate` - Warning: 5xx errors > 5% for 5 minutes
- `HighErrorRateCritical` - Critical: 5xx errors > 10% for 2 minutes

## Investigation Steps

### 1. Check Application Logs
```bash
# View recent error logs
docker logs golang-api 2>&1 | grep -i error | tail -50
```

### 2. Check Metrics Dashboard
- Open Grafana dashboard
- Review HTTP error rate panel
- Identify affected endpoints

### 3. Check Dependencies
- Database connectivity
- External service availability
- Resource exhaustion (CPU, memory)

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Database down | Connection timeouts | Check DB health |
| Out of memory | OOM kills in logs | Increase resources |
| Panic in handler | Stack traces | Fix bug, deploy |
| External API failure | Timeout errors | Check external service |

## Resolution Actions

1. **If database issue**: Check `DBConnectionExhausted` runbook
2. **If memory issue**: Scale up or fix memory leak
3. **If code bug**: Rollback or hotfix

## Escalation
If unresolved after 15 minutes, escalate to on-call engineer.

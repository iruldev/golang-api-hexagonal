# Runbook: Service Down Alert

## Overview
This runbook addresses the critical alert when the service is unreachable.

## Alert Covered
- `ServiceDown` - Critical: `up == 0` for 1 minute

## Investigation Steps

### 1. Check Service Status
```bash
# Check if container is running
docker ps | grep golang-api

# Check container logs
docker logs golang-api --tail 100
```

### 2. Check Kubernetes (if applicable)
```bash
kubectl get pods -l app=golang-api
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### 3. Check System Resources
```bash
# Check disk space
df -h

# Check memory
free -m

# Check processes
ps aux | grep golang-api
```

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Container crashed | Restart loop | Check logs for panic |
| OOM killed | Memory exhaustion | Increase memory limit |
| Disk full | Write errors | Clear disk space |
| Network issue | Connection refused | Check network/firewall |
| Port conflict | Bind error | Check port availability |

## Resolution Actions

1. **If crashed**: Check logs, fix bug, restart
2. **If OOM**: Increase memory limits
3. **If network**: Verify network configuration

## Escalation
**IMMEDIATE ESCALATION** - This is a critical alert. Page on-call engineer immediately.

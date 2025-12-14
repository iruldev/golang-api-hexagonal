# Runbook: Service Down

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `ServiceDown` |
| **Severity** | critical |
| **Component** | http |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses the critical alert when the service is completely unreachable.

**What triggers this alert:**
- **ServiceDown (Critical):** `up == 0` for 1 minute (service not responding to Prometheus scrapes)

**Business impact:**
- Complete service outage
- All API requests failing
- Immediate user impact

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the app runs via `make dev` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal, use `ps aux | grep server`).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Immediately acknowledge in PagerDuty
- [ ] **Assess** - Verify service is actually down, not just metrics issue
- [ ] **Investigate** - Check container status and logs
- [ ] **Mitigate** - Restart service or rollback
- [ ] **Communicate** - Notify #incidents channel immediately
- [ ] **Document** - Record start time and actions taken
- [ ] **Resolve** - Confirm service is healthy and accepting traffic

---

## Symptoms

### Observable Indicators
- API endpoints returning connection refused or timeout
- Health check endpoints not responding
- Load balancer marking all instances unhealthy

### Metrics to Check
| Metric | Normal | Alert State |
|--------|--------|-------------|
| `up{job="golang-api"}` | 1 | 0 |
| `/healthz` response | 200 | No response |

### Related Alerts
- `HighErrorRate` - May fire before complete outage
- `DBConnectionExhausted` - If DB caused the crash

---

## Diagnosis

### Step 1: Check Service Status

**Goal:** Verify the service container state

```bash
# Check if container is running
docker ps -a | grep golang-api

# Check container logs
docker logs golang-api --tail 100

# For Kubernetes
kubectl get pods -l app=golang-api
kubectl describe pod <pod-name>
```

**Expected outcome:** Container status and recent logs

### Step 2: Check for Crash Reason

**Goal:** Find why the service crashed

```bash
# Check for OOM kill
docker inspect golang-api | grep -i oom
dmesg | grep -i "out of memory"

# Check exit code
docker inspect golang-api --format='{{.State.ExitCode}}'

# For Kubernetes
kubectl logs <pod-name> --previous
```

### Step 3: Check System Resources

**Goal:** Identify resource exhaustion

```bash
# Check disk space
df -h

# Check memory
free -m

# Check CPU
top -bn1 | head -5
```

### Step 4: Check Network

**Goal:** Verify network connectivity

```bash
# Check if port is in use
netstat -tlnp | grep 8080

# Check firewall
iptables -L -n | grep 8080

# Verify DNS resolution
nslookup postgres
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Container crashed | Exit code != 0 | Check logs for panic | Fix bug, restart |
| OOM killed | Exit code 137 | OOM in dmesg | Increase memory limit |
| Disk full | Write errors | `df -h` shows 100% | Clear disk space |
| Network issue | Connection refused | Check firewall/DNS | Fix network config |
| Port conflict | Bind error in logs | Check port usage | Kill conflicting process |
| Config error | Startup failure | Check config logs | Fix configuration |

---

## Remediation

### Immediate Actions

1. **Restart the service:**
   ```bash
   # Docker
   docker restart golang-api
   
   # Docker Compose
   docker-compose restart golang-api
   
   # Kubernetes
   kubectl rollout restart deployment/golang-api
   ```

2. **If OOM killed:** Increase memory limits
   ```bash
   # Docker Compose - edit docker-compose.yaml
   # Then restart
   docker-compose up -d golang-api
   
   # Kubernetes
   kubectl patch deployment golang-api -p '{"spec":{"template":{"spec":{"containers":[{"name":"golang-api","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
   ```

3. **If disk full:** Clear disk space
   ```bash
   # Remove old containers and images
   docker system prune -a
   
   # Clear old logs
   truncate -s 0 /var/log/*.log
   ```

4. **If recent deployment caused crash:** Rollback
   ```bash
   # Kubernetes
   kubectl rollout undo deployment/golang-api
   
   # Docker - use previous image tag
   docker-compose down
   docker-compose up -d
   ```

### Post-Incident Actions

1. **Root Cause Analysis:** Document why the crash occurred
2. **Prevention:** Add resource limits, improve health checks
3. **Monitoring:** Ensure adequate pre-crash alerting

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Critical** | Immediate investigation | Escalate after 5 minutes |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** Team lead / Senior engineer (5 min)
3. **Level 2 Escalation:** Engineering manager (15 min)
4. **Executive Escalation:** CTO (if major outage > 30 min)

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | Immediate |
| Team Lead | @team-lead | After 5 minutes |
| Engineering Manager | @eng-manager | After 15 minutes |

---

## References

- [Docker Troubleshooting](https://docs.docker.com/config/daemon/)
- [Kubernetes Pod Debugging](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-pod-replication-controller/)
- [High Error Rate Runbook](high-error-rate.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |

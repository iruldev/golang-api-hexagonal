# Runbook: Memory/CPU Spike

**Severity:** 🟠 High
**Last Updated:** 2026-01-04
**Owner:** DevOps Team
**Related ADRs:** [ADR-003: Resilience Patterns](../adr/ADR-003-resilience-patterns.md), [ADR-001: Hexagonal Architecture](../adr/ADR-001-hexagonal-architecture.md)

---

## Overview

This runbook addresses scenarios where application pods experience excessive memory or CPU usage, potentially leading to OOMKilled events, CPU throttling, or degraded performance.

**Impact:** Slow response times, request timeouts, pod restarts, potential service unavailability.

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] Prometheus/Grafana dashboard access
  - [ ] Go pprof access (if enabled)

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] `go tool pprof` (for profiling)
  - [ ] Grafana dashboards

- **Knowledge Required:**
  - Kubernetes resource limits and requests
  - Go memory management and garbage collection
  - Application performance profiling

---

## Symptoms

### 🔔 Alerts

| Alert Name | Condition | Dashboard Link |
|------------|-----------|----------------|
| `PodMemoryHigh` | `container_memory_usage_bytes` > 80% limit | K8s Resources Dashboard |
| `PodCPUThrottled` | `container_cpu_cfs_throttled_seconds` > threshold | K8s Resources Dashboard |
| `PodOOMKilled` | Pod restart reason = OOMKilled | K8s Events Dashboard |
| `HighLatency` | `http_request_duration_seconds_p99` > SLA | API Dashboard |

### 📊 Metrics

- `container_memory_usage_bytes` approaching or at limit
- `container_cpu_usage_seconds_total` at or near limit
- `container_cpu_cfs_throttled_seconds_total` increasing
- `go_memstats_alloc_bytes` growing without bounds
- Pod restart count increasing

### 🐛 User Reports

- API requests timing out
- Slow response times
- Intermittent 503 errors
- "Connection reset" errors

---

## Investigation Steps

### Step 1: Identify Affected Pods

```bash
# Check pod resource usage
kubectl top pods -l app=api-server --sort-by=memory

# Check for OOMKilled events
kubectl get pods -l app=api-server -o wide
kubectl describe pod <pod-name> | grep -A10 "Last State"

# Check recent pod events
kubectl get events --field-selector involvedObject.kind=Pod --sort-by='.lastTimestamp' | tail -20
```

**What to look for:**
- Pods using > 80% of memory limit
- Pods with OOMKilled restart reason
- High restart counts

### Step 2: Check Resource Limits

```bash
# View pod resource configuration
kubectl get deployment api-server -o yaml | grep -A20 resources

# Expected output shows limits and requests:
# resources:
#   limits:
#     memory: "512Mi"
#     cpu: "500m"
#   requests:
#     memory: "256Mi"
#     cpu: "100m"
```

### Step 3: Analyze Memory Usage (if pprof enabled)

```bash
# Get heap profile
curl -o heap.prof http://${POD_IP}:6060/debug/pprof/heap

# Analyze heap profile
go tool pprof heap.prof
# In pprof: top 10, list <function>

# Get goroutine profile (for goroutine leaks)
curl -o goroutine.prof http://${POD_IP}:6060/debug/pprof/goroutine
go tool pprof goroutine.prof
```

**What to look for:**
- Functions allocating large amounts of memory
- Goroutine count growing unbounded
- Memory not being freed (potential leak)

### Step 4: Analyze CPU Usage (if pprof enabled)

```bash
# 30-second CPU profile
curl -o cpu.prof "http://${POD_IP}:6060/debug/pprof/profile?seconds=30"

# Analyze CPU profile
go tool pprof -http=:8080 cpu.prof
# View flame graph in browser
```

**What to look for:**
- Hot functions consuming excessive CPU
- Unexpected functions in the profile
- Inefficient algorithms or loops

### Step 5: Check Application Logs

```bash
# Look for error patterns
kubectl logs -l app=api-server --tail=500 | grep -iE "(memory|oom|panic|goroutine)"

# Check for slow queries or operations
kubectl logs -l app=api-server --tail=500 | grep -iE "(slow|timeout|exceeded)"
```

---

## Resolution Steps

### Immediate Mitigation

> ⚡ **Time-Critical:** Complete within 10 minutes

1. **Scale horizontally to distribute load:**
   ```bash
   kubectl scale deployment api-server --replicas=5
   ```

2. **Restart unhealthy pods:**
   ```bash
   # Delete specific OOMKilled pod (K8s will recreate)
   kubectl delete pod <oom-killed-pod-name>

   # Or rolling restart all
   kubectl rollout restart deployment/api-server
   ```

3. **Temporarily increase resource limits (if cluster has capacity):**
   ```bash
   kubectl set resources deployment api-server --limits=memory=1Gi,cpu=1
   ```

### Root Cause Fix

> 🔧 **Standard Resolution:** Complete after stabilization

1. **Memory leak:**
   - Identify leaking code path from pprof analysis
   - Fix code to properly release resources
   - Ensure contexts are cancelled and connections closed

2. **Inefficient memory usage:**
   - Optimize data structures
   - Use object pools for frequently allocated objects
   - Reduce JSON/data copy operations

3. **CPU-intensive operations:**
   - Optimize hot code paths
   - Add caching for repeated computations
   - Consider async processing for heavy operations

4. **Goroutine leak:**
   - Ensure all goroutines have exit conditions
   - Use context cancellation properly
   - Check for blocking channel operations

5. **Insufficient resources:**
   - Increase resource limits in deployment
   - Request more cluster resources
   - Implement horizontal pod autoscaling

### Configuration Changes

```yaml
# Update deployment resources
spec:
  containers:
  - name: api-server
    resources:
      limits:
        memory: "1Gi"
        cpu: "1000m"
      requests:
        memory: "512Mi"
        cpu: "200m"

# Add Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## Rollback Procedure

### When to Rollback

- Recent deployment introduced memory/CPU regression
- New feature causing resource exhaustion
- Library upgrade causing performance issues

### Rollback Steps

1. **Rollback deployment:**
   ```bash
   kubectl rollout undo deployment/api-server
   ```

2. **Monitor resource usage:**
   ```bash
   watch -n 5 'kubectl top pods -l app=api-server'
   ```

3. **Verify rollback:**
   ```bash
   kubectl rollout status deployment/api-server
   ```

### Verify Rollback Success

- [ ] Memory usage returned to normal
- [ ] No OOMKilled events
- [ ] CPU usage under limits
- [ ] Response latency normalized

---

## Verification

### Confirm Resolution

- [ ] `container_memory_usage_bytes` < 70% of limits
- [ ] `container_cpu_usage_seconds` < 80% of limits
- [ ] No new OOMKilled events for 15 minutes
- [ ] API response latency within SLA
- [ ] All pods running (no restarts)

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| `container_memory_usage_bytes` | < 70% of limit | K8s Resources |
| `container_cpu_cfs_throttled_seconds` | Stable / 0 | K8s Resources |
| `go_memstats_heap_alloc_bytes` | Stable | Go Runtime |
| `go_goroutines` | Stable (no unbounded growth) | Go Runtime |
| `http_request_duration_seconds_p99` | < SLA | API Dashboard |

### Observation Period

⏱️ **Recommended:** Monitor for 30 minutes to ensure stability

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call Engineer | First response | Immediate |
| L2 | Senior DevOps | Pods keep getting OOMKilled | 15 min |
| L3 | Platform/SRE Team | Cluster resource exhaustion | 10 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: Issue persists after scaling, memory leak suspected
- 🚨 Escalate to L3 if: Cluster nodes affected, multiple services impacted

---

## Related Runbooks

- [Database Connection Failure](./database-connection-failure.md) - Resource issues can affect DB connections
- [Circuit Breaker Open](./circuit-breaker-open.md) - Slow responses can trigger circuit breakers
- [Deployment Rollback](./deployment-rollback.md) - If recent deployment caused resource issues

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | DevOps Team | Initial creation |

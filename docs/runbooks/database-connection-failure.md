# Runbook: Database Connection Failure

**Severity:** 🔴 Critical
**Last Updated:** 2026-01-04
**Owner:** DevOps Team
**Related ADRs:** [ADR-003: Resilience Patterns](../adr/ADR-003-resilience-patterns.md), [ADR-001: Hexagonal Architecture](../adr/ADR-001-hexagonal-architecture.md)

---

## Overview

This runbook addresses complete or partial loss of database connectivity. This is a **critical** incident as it affects all API operations requiring persistent storage.

**Impact:** All CRUD operations fail, API returns 503 errors, health checks fail.

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] PostgreSQL admin access
  - [ ] Prometheus/Grafana dashboard access
  - [ ] Cloud provider console (for managed DB)

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] `psql` PostgreSQL client
  - [ ] `make` (for project commands)

- **Knowledge Required:**
  - PostgreSQL connection pooling (pgx)
  - Kubernetes pod networking
  - Application health check endpoints

---

## Symptoms

### 🔔 Alerts

| Alert Name | Condition | Dashboard Link |
|------------|-----------|----------------|
| `DatabaseConnectionFailed` | `db_connection_errors > 0` for 1 min | Grafana DB Dashboard |
| `HealthCheckFailing` | `/readyz` returns non-200 | K8s Health Dashboard |
| `HighErrorRate` | `http_requests_total{status="503"}` spike | API Dashboard |

### 📊 Metrics

- `db_query_duration_seconds` timeout or NaN
- `api_request_duration_seconds{status="503"}` increasing

### 🐛 User Reports

- "Service Unavailable" (503) errors
- "Database connection failed" in error responses
- API requests timing out

---

## Investigation Steps

### Step 1: Verify Service Health

```bash
# Check API health endpoints
curl -s http://${API_HOST}:8080/healthz
curl -s http://${API_HOST}:8080/readyz | jq .

# Check pod status
kubectl get pods -l app=api-server -o wide
```

**Expected (Healthy):** `/healthz` returns 200, `/readyz` shows database: healthy
**Abnormal:** `/readyz` returns 503, database shows unhealthy

### Step 2: Check Database Connectivity

```bash
# Test direct database connection
PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;"

# Check from within a pod
kubectl exec -it $(kubectl get pod -l app=api-server -o jsonpath='{.items[0].metadata.name}') -- \
  sh -c 'nc -zv ${DB_HOST} 5432'
```

**Expected:** Connection successful, returns "1"
**Abnormal:** Connection refused, timeout, authentication failed

### Step 3: Check Connection Pool Metrics

```bash
# Get pool metrics from Prometheus
curl -s http://${API_HOST}:8081/metrics | grep db_pool

# Expected metrics:
# db_pool_connections_total (if implemented)
# db_pool_connections_in_use (if implemented)
```

**What to look for:**
- `connections_in_use` at `max_connections` indicates pool exhaustion
- `connections_idle` = 0 with high usage indicates connection leak

### Step 4: Check Database Server

```bash
# Check database server status (if self-hosted)
kubectl get pods -l app=postgres -o wide

# Check database logs
kubectl logs -l app=postgres --tail=100

# For managed database, check cloud console for:
# - Connection limits
# - CPU/Memory usage
# - Storage capacity
```

### Step 5: Network Diagnostics

```bash
# DNS resolution
kubectl exec -it $(kubectl get pod -l app=api-server -o jsonpath='{.items[0].metadata.name}') -- \
  nslookup ${DB_HOST}

# Network policy check
kubectl get networkpolicies -o wide
```

---

## Resolution Steps

### Immediate Mitigation

> ⚡ **Time-Critical:** Complete within 5 minutes

1. **Scale down to reduce connection pressure:**
   ```bash
   kubectl scale deployment api-server --replicas=1
   ```

2. **Restart pods to refresh connection pool:**
   ```bash
   kubectl rollout restart deployment/api-server
   ```

3. **If database is unresponsive, check/restart database:**
   ```bash
   kubectl rollout restart statefulset/postgres
   # OR for managed DB, use cloud console to restart
   ```

### Root Cause Fix

> 🔧 **Standard Resolution:** Complete after stabilization

1. **Connection pool exhaustion:**
   - Increase `max_connections` in config
   - Check for connection leaks in code
   - Review query performance (slow queries hold connections)

2. **Database overload:**
   - Scale database vertically (more CPU/RAM)
   - Add read replicas for read-heavy loads
   - Optimize slow queries

3. **Network issues:**
   - Check security groups/firewall rules
   - Verify network policies allow database access
   - Check DNS resolution

4. **Authentication issues:**
   - Rotate database credentials
   - Check secret mounting in pods
   - Verify DATABASE_URL environment variable

### Configuration Changes

```yaml
# Update connection pool settings in config
database:
  max_connections: 50  # Increase from 25
  max_idle_connections: 10
  connection_timeout: 10s
  max_connection_lifetime: 5m
```

---

## Rollback Procedure

### When to Rollback

- Recent deployment changed database configuration
- New code has connection leak
- Migration caused database issues

### Rollback Steps

1. **Rollback deployment:**
   ```bash
   kubectl rollout undo deployment/api-server
   ```

2. **Verify rollback:**
   ```bash
   kubectl rollout status deployment/api-server
   ```

3. **If database migration caused issue:**
   ```bash
   # Check migration status
   make migrate-status

   # Rollback last migration
   make migrate-down
   ```

### Verify Rollback Success

- [ ] `/readyz` returns 200
- [ ] Database connection metrics normalized
- [ ] No 503 errors in logs
- [ ] API requests succeeding

---

## Verification

### Confirm Resolution

- [ ] Database connection check returns success
- [ ] `/readyz` shows database: healthy
- [ ] No `DB-xxx` error codes in logs for 5 minutes
- [ ] API latency returned to baseline
- [ ] All replicas running and healthy

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| `db_health_status` | 1 | DB Dashboard |
| `db_query_duration_seconds_p99` | < 100ms | DB Dashboard |
| `http_requests_total{status="503"}` | 0 | API Dashboard |
| `health_check_status{probe="readiness"}` | 1 | K8s Dashboard |

### Observation Period

⏱️ **Recommended:** Monitor for 30 minutes before closing incident

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call Engineer | First response | Immediate |
| L2 | Senior DevOps | No resolution in 15 min | 10 min |
| L3 | Database Admin / SRE Lead | Data integrity concern | 5 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: Database pod won't start, connection pool not recovering
- 🚨 Escalate to L3 if: Data corruption suspected, cannot connect to database at all

---

## Related Runbooks

- [Circuit Breaker Open](./circuit-breaker-open.md) - May trigger if DB issues cause cascading failures
- [Deployment Rollback](./deployment-rollback.md) - If recent deployment caused issue
- [Memory/CPU Spike](./memory-cpu-spike.md) - Database issues can cause resource exhaustion

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | DevOps Team | Initial creation |

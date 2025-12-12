# Grafana Dashboard Configuration

This directory contains the Grafana configuration for the Golang API service observability.

## Directory Structure

```
deploy/grafana/
├── dashboards/
│   └── service.json       # Main service dashboard
├── provisioning/
│   ├── dashboards/
│   │   └── dashboards.yml # Dashboard auto-provisioning
│   └── datasources/
│       └── prometheus.yml # Prometheus datasource config
└── README.md              # This file
```

## Quick Start

```bash
# Start all services including Grafana
docker compose up -d

# Access Grafana
open http://localhost:3000
# Default credentials: admin/admin
```

## Dashboard Panels

The **Golang API Service** dashboard includes:

| Panel | Metric | Description |
|-------|--------|-------------|
| HTTP Request Rate | `http_requests_total` | Requests/sec by status code |
| HTTP Latency | `http_request_duration_seconds` | p50/p95/p99 percentiles |
| HTTP Error Rate | `http_requests_total{status=~"5.."}` | 5xx error percentage |
| Job Processing Rate | `job_processed_total` | Jobs/sec by status |
| Job Duration | `job_duration_seconds` | p50/p95 percentiles |

## Configuration

### Dashboard UID

The dashboard uses UID `golang-api-service`. If deploying multiple services to the same Grafana instance, modify the UID in `dashboards/service.json`:

```json
{
  "uid": "your-unique-service-name",
  "title": "Your Service Name"
}
```

### Datasource UID

The dashboard references datasource UID `prometheus`. This is configured in `provisioning/datasources/prometheus.yml`.

### Production Credentials

> ⚠️ **Warning**: Default credentials (admin/admin) are for development only.

For production, set credentials via environment variables:

```yaml
# docker-compose.override.yaml
services:
  grafana:
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
```

Or use Grafana's provisioning for admin user:
```yaml
# provisioning/access-control/admin.yaml
apiVersion: 1
users:
  - name: admin
    password: ${GRAFANA_ADMIN_PASSWORD}
```

## Customization

### Adding New Panels

1. Edit the dashboard in Grafana UI
2. Export as JSON via Share → Export → Save to file
3. Replace `dashboards/service.json`

Note: Provisioned dashboards are read-only. To make editable:
```json
{
  "editable": true
}
```

### Future Metrics (Not Yet Implemented)

The following metrics are planned for future stories:
- Database connection pool metrics
- Database query latency
- Asynq queue depth and pending jobs

## Troubleshooting

### Dashboard shows "No Data"

1. Ensure the application is running and generating metrics
2. Check Prometheus is scraping: http://localhost:9090/targets
3. Verify datasource connection in Grafana

### Datasource Connection Failed

Ensure the datasource UID matches between:
- `provisioning/datasources/prometheus.yml` → `uid: prometheus`
- `dashboards/service.json` → `"uid": "prometheus"`

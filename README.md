# Backend Service Golang Boilerplate

Enterprise-grade "golden template" untuk membangun backend services di Go dengan **observability-first architecture**.

## Quick Start

```bash
# Clone repository
git clone https://github.com/iruldev/golang-api-hexagonal.git
cd golang-api-hexagonal

# Download dependencies
go mod download

# Verify compilation
go build ./...
```

## Documentation

- [Architecture](docs/architecture.md) - Design decisions and patterns
- [PRD](docs/prd.md) - Product requirements

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24.x |
| Router | chi v5 |
| Database | PostgreSQL + pgx v5 |
| Query | sqlc |
| Logger | zap |
| Config | koanf v2 |
| Tracing | OpenTelemetry |

## Project Structure

```
├── cmd/server/         # Application entry point
├── internal/
│   ├── app/            # Application wiring
│   ├── config/         # Configuration
│   ├── domain/         # Business entities
│   ├── usecase/        # Business logic
│   ├── infra/          # Infrastructure adapters
│   ├── interface/      # HTTP handlers
│   ├── observability/  # Logging/tracing
│   └── runtimeutil/    # Utilities
├── db/
│   ├── migrations/     # SQL migrations
│   └── queries/        # sqlc queries
└── docs/               # Documentation
```

## License

MIT

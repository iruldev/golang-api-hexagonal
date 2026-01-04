# Copy-Paste Kit

📋 Ready-to-use code templates for implementing new features in the golang-api-hexagonal project.

## Quick Start

1. **Choose a template** from the list below
2. **Copy the file** to the appropriate location
3. **Find and replace** `{Entity}` and `{entity}` with your entity name
4. **Follow the TODO comments** in each file

## Template Files

| Template | Location | Description |
|----------|----------|-------------|
| [new-entity.go.example](./new-entity.go.example) | `internal/domain/` | Domain entity + repository interface |
| [new-usecase.go.example](./new-usecase.go.example) | `internal/app/{entity}/` | Use cases (Create, Get, List) |
| [new-handler.go.example](./new-handler.go.example) | `internal/transport/http/handler/` | HTTP handlers with validation |
| [new-repository.go.example](./new-repository.go.example) | `internal/infra/postgres/` | PostgreSQL repository + SQL queries |

## Workflow untuk Entity Baru

### Step 1: Domain Layer
```bash
# Copy entity template (rename from .example to .go)
cp docs/copy-paste-kit/new-entity.go.example internal/domain/product.go

# Edit: Replace {Entity} → Product, {entity} → product
# Add entity fields, validation, domain errors
```

### Step 2: Infrastructure Layer
```bash
# Copy repository template
cp docs/copy-paste-kit/new-repository.go internal/infra/postgres/product_repo.go

# Create SQL migration
touch internal/infra/postgres/migrations/000006_add_products.up.sql
touch internal/infra/postgres/migrations/000006_add_products.down.sql

# Add sqlc queries
touch internal/infra/postgres/queries/product.sql

# Generate sqlc code
make sqlc
```

### Step 3: Application Layer
```bash
# Create directory and copy template
mkdir -p internal/app/product
cp docs/copy-paste-kit/new-usecase.go internal/app/product/create_product.go

# Split into separate files if needed
# - create_product.go
# - get_product.go
# - list_products.go
```

### Step 4: Transport Layer
```bash
# Copy handler template
cp docs/copy-paste-kit/new-handler.go internal/transport/http/handler/product.go

# Add contract DTOs to internal/transport/http/contract/product.go

# Register routes in internal/transport/http/router/router.go
```

### Step 5: Generate & Verify
```bash
# Generate mocks
make mocks

# Run linter
make lint

# Run tests
make test
```

## Existing Kit Contents

| File/Folder | Description |
|-------------|-------------|
| `ci-workflow.yml` | GitHub Actions CI workflow snippet |
| `domain-errors/` | Domain error patterns |
| `makefile-snippets.md` | Makefile target patterns |
| `testutil/` | Test utility patterns (containers, fixtures) |

## Related Documentation

- [Patterns](../patterns.md) - Complete code pattern documentation
- [Testing Patterns](../testing-patterns.md) - Test writing conventions
- [Architecture](../../_bmad-output/planning-artifacts/architecture.md) - Layer boundaries

## Tips

### Find & Replace Checklist

When using templates, replace these placeholders:

- `{Entity}` → PascalCase entity name (e.g., `Product`)
- `{entity}` → lowercase entity name (e.g., `product`)
- `{Field}` → Field name in errors (e.g., `Price`)
- `{ENT}` → Error code prefix (e.g., `PROD`)

### Naming Convention Quick Reference

| Element | Pattern | Example |
|---------|---------|---------|
| Entity struct | PascalCase, singular | `Product` |
| Repository interface | `{Entity}Repository` | `ProductRepository` |
| Use case struct | `{Action}{Entity}UseCase` | `CreateProductUseCase` |
| Handler struct | `{Entity}Handler` | `ProductHandler` |
| DB table | snake_case, plural | `products` |
| Error codes | `{PREFIX}-{NUM}` | `PROD-001` |

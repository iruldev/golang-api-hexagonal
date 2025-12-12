# Architecture

This document describes the architectural decisions and patterns used in this Go API boilerplate project.

## Three Pillars

This project is built on three core principles that guide all design decisions:

### 1. Simplicity

> "Simplicity is prerequisite for reliability" — Edsger W. Dijkstra

- **Choose boring technology**: Standard library preferred over frameworks
- **Minimal dependencies**: Only add dependencies when they provide significant value
- **Clear code over clever code**: Readability trumps conciseness
- **Explicit over implicit**: Make behavior obvious and predictable

### 2. Consistency

> "A foolish consistency is the hobgoblin of little minds" — but a *wise* consistency enables scalability.

- **Predictable patterns**: Same problems solved the same way across the codebase
- **Naming conventions**: Follow Go idioms and project standards
- **File structure**: Consistent organization across all modules
- **Error handling**: Uniform error patterns throughout

### 3. Pragmatism

> "Perfect is the enemy of good" — but good enough is the friend of shipped.

- **80/20 rule**: Focus on the 20% that delivers 80% of value
- **YAGNI**: Don't build for hypothetical future requirements
- **Iterative improvement**: Start simple, refactor when needed
- **Real-world focus**: Solve actual problems, not theoretical ones

---

## Layer Structure

This project follows a hexagonal (ports and adapters) architecture:

```
internal/
├── domain/           # Business entities and rules
├── usecase/          # Application business logic
├── interface/        # External adapters (HTTP, CLI, etc)
└── infra/            # Infrastructure implementations
```

### Domain Layer (`internal/domain/`)

The core business logic that has no external dependencies.

```
domain/
└── note/
    ├── entity.go       # Note struct and validation
    ├── errors.go       # Domain-specific errors
    └── repository.go   # Repository interface (port)
```

**Responsibilities:**
- Define business entities with validation rules
- Define domain-specific errors
- Define repository interfaces (ports)

**Dependencies:** None (except standard library and UUID)

### Use Case Layer (`internal/usecase/`)

Application-specific business rules that orchestrate domain logic.

```
usecase/
└── note/
    └── usecase.go      # NoteUsecase with CRUD operations
```

**Responsibilities:**
- Implement business workflows
- Coordinate between domain entities
- Apply business rules before persistence

**Dependencies:** Domain layer only

### Interface Layer (`internal/interface/`)

Adapters for external communication (HTTP, gRPC, CLI, etc).

```
interface/
└── http/
    ├── router.go           # Chi router setup
    ├── middleware/         # HTTP middleware
    ├── handlers/           # Health, example handlers
    ├── response/           # Response envelope pattern
    └── note/
        ├── handler.go      # Note HTTP handlers
        └── dto.go          # Request/Response DTOs
```

**Responsibilities:**
- Handle HTTP requests/responses
- Transform DTOs to/from domain objects
- Apply middleware (auth, logging, etc)

**Dependencies:** Use case layer

### Infrastructure Layer (`internal/infra/`)

Concrete implementations of external services.

```
infra/
└── postgres/
    ├── postgres.go     # Database connection
    └── note/           # SQLC-generated repository
```

**Responsibilities:**
- Database connections and queries
- External service clients
- File system access

**Dependencies:** Domain interfaces (implements ports)

---

## Patterns

### Repository Pattern

Repositories abstract data persistence, allowing the domain to remain pure.

```go
// Domain defines the interface (port)
type Repository interface {
    Create(ctx context.Context, note *Note) error
    Get(ctx context.Context, id uuid.UUID) (*Note, error)
    List(ctx context.Context, limit, offset int) ([]*Note, int64, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Infrastructure implements the interface (adapter)
type PostgresNoteRepository struct {
    queries *sqlc.Queries
}
```

### Response Envelope Pattern

All API responses use a consistent envelope for predictable parsing.

```json
// Success response
{
  "success": true,
  "data": { ... }
}

// Error response
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Note not found"
  }
}
```

### Error Handling

Errors flow from domain → usecase → handler with proper HTTP mapping.

```go
// Domain errors
var ErrNoteNotFound = errors.New("note: not found")
var ErrEmptyTitle = errors.New("note: title cannot be empty")

// Handler maps domain errors to HTTP
func handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, note.ErrNoteNotFound):
        response.NotFound(w, err.Error())
    case errors.Is(err, note.ErrEmptyTitle):
        response.ValidationError(w, err.Error())
    default:
        response.InternalServerError(w, "An unexpected error occurred")
    }
}
```

### Testing Patterns

- **Table-driven tests**: Use test tables for comprehensive coverage
- **AAA pattern**: Arrange, Act, Assert structure
- **Mocks**: Interface-based mocking for unit tests
- **Integration tests**: Build-tagged tests for E2E scenarios

```go
func TestUsecase_Create(t *testing.T) {
    tests := []struct {
        name    string
        title   string
        wantErr error
    }{
        {"valid", "Title", nil},
        {"empty title", "", note.ErrEmptyTitle},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            repo := &MockRepository{}
            uc := NewUsecase(repo)
            
            // Act
            _, err := uc.Create(ctx, tt.title, "content")
            
            // Assert
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Conventions

### Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case | `note_handler.go` |
| Packages | lowercase | `note`, `postgres` |
| Types | PascalCase | `NoteHandler`, `CreateNoteRequest` |
| Functions | PascalCase (exported) | `NewHandler()` |
| Functions | camelCase (private) | `handleError()` |
| Variables | camelCase | `noteID`, `pageSize` |
| Constants | PascalCase or SCREAMING_CASE | `MaxTitleLength` |

### Package Organization

- One domain per package under `domain/`
- One usecase per domain under `usecase/`
- HTTP handlers grouped by domain under `interface/http/`
- Infrastructure implementations under `infra/`

### File Structure per Domain

```
domain/example/
├── entity.go         # Main entity struct
├── entity_test.go    # Entity unit tests
├── errors.go         # Domain errors
└── repository.go     # Repository interface

usecase/example/
├── usecase.go        # Business logic
└── usecase_test.go   # Unit tests with mocks

interface/http/example/
├── handler.go                   # HTTP handlers
├── handler_test.go              # Handler unit tests
├── handler_integration_test.go  # Integration tests
└── dto.go                       # Request/Response DTOs
```

---

## Quick Reference

### Adding a New Domain

1. Create domain entity: `internal/domain/[name]/entity.go`
2. Create domain errors: `internal/domain/[name]/errors.go`
3. Create repository interface: `internal/domain/[name]/repository.go`
4. Create SQL migration: `db/migrations/`
5. Create SQLC queries: `db/queries/[name].sql`
6. Generate SQLC: `make sqlc`
7. Create usecase: `internal/usecase/[name]/usecase.go`
8. Create HTTP handler: `internal/interface/http/[name]/handler.go`
9. Wire up in router

### Key Commands

```bash
make build      # Build the application
make test       # Run all tests
make lint       # Run linter
make sqlc       # Generate SQLC code
make migrate    # Run database migrations
```

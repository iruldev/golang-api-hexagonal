# Architecture Documentation

Dokumentasi arsitektur untuk **Backend Service Golang Boilerplate**.

---

## Three Pillars

Proyek ini dibangun di atas tiga prinsip inti:

### 1. Simplicity
- **KISS**: Solusi paling sederhana yang bisa bekerja
- **YAGNI**: Tidak membangun untuk kebutuhan hipotetis di masa depan
- **Iterative improvement**: Mulai sederhana, refactor saat dibutuhkan

### 2. Observability
- Tracing, metrics, dan logging sebagai warga kelas satu
- Setiap operasi penting dapat dilacak dan diukur
- Alerting dan runbook untuk incident response

### 3. Testability
- Dependency injection untuk mockability
- Unit tests untuk domain dan usecase
- Integration tests dengan testcontainers

---

## Layer Structure (Hexagonal Architecture)

```
┌─────────────────────────────────────────────────────────────────────┐
│                               cmd/                                   │
│            (Entry Points: server, worker, scheduler, bplat)          │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Interface Layer                               │
│                  internal/interface/ (Adapters)                      │
│                                                                      │
│   ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐       │
│   │      HTTP       │ │      gRPC       │ │    GraphQL      │       │
│   │    (chi v5)     │ │                 │ │    (gqlgen)     │       │
│   └────────┬────────┘ └────────┬────────┘ └────────┬────────┘       │
└────────────┼───────────────────┼───────────────────┼────────────────┘
             │                   │                   │
             └───────────────────┼───────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Use Case Layer                               │
│                      internal/usecase/                               │
│                                                                      │
│   - Orchestrates domain entities                                    │
│   - Applies business rules                                          │
│   - Depends only on domain layer                                    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Domain Layer                                 │
│                       internal/domain/                               │
│                                                                      │
│   - Entities with validation                                        │
│   - Repository interfaces (ports)                                   │
│   - Domain-specific errors                                          │
│   - NO external dependencies                                        │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
                                 ▲
                                 │
┌─────────────────────────────────────────────────────────────────────┐
│                      Infrastructure Layer                            │
│                        internal/infra/                               │
│                                                                      │
│   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐              │
│   │PostgreSQL│ │  Redis   │ │  Kafka   │ │ RabbitMQ │              │
│   │(pgx,sqlc)│ │(go-redis)│ │ (sarama) │ │(amqp091) │              │
│   └──────────┘ └──────────┘ └──────────┘ └──────────┘              │
│                                                                      │
│   - Implements domain interfaces                                    │
│   - External service integrations                                   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Layer Dependencies

| Layer | Can Depend On | Cannot Depend On |
|-------|--------------|------------------|
| **Domain** | Nothing | All other layers |
| **Use Case** | Domain | Interface, Infrastructure |
| **Interface** | Domain, Use Case | Infrastructure |
| **Infrastructure** | Domain | Use Case, Interface |

**Dependency Rule:** Dependencies point inward. Inner layers define interfaces (ports), outer layers implement them (adapters).

---

## Domain Layer (`internal/domain/`)

**Responsibility:** Business entities, rules, and repository interfaces.

### Entity Pattern

```go
// internal/domain/note/entity.go
type Note struct {
    ID        string
    Title     string
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (n *Note) Validate() error {
    if n.Title == "" {
        return ErrEmptyTitle
    }
    return nil
}
```

### Repository Interface (Port)

```go
// internal/domain/note/repository.go
type Repository interface {
    Create(ctx context.Context, note *Note) error
    GetByID(ctx context.Context, id string) (*Note, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*Note, error)
}
```

### Domain Errors

```go
// internal/domain/note/errors.go
var (
    ErrNoteNotFound = errors.New("note: not found")
    ErrEmptyTitle   = errors.New("note: title cannot be empty")
)
```

---

## Use Case Layer (`internal/usecase/`)

**Responsibility:** Application business logic orchestrating domain entities.

```go
// internal/usecase/note/usecase.go
type UseCase struct {
    repo   note.Repository
    logger observability.Logger
}

func New(repo note.Repository, logger observability.Logger) *UseCase {
    return &UseCase{repo: repo, logger: logger}
}

func (uc *UseCase) Create(ctx context.Context, title, content string) (*note.Note, error) {
    n := &note.Note{
        ID:        uuid.New().String(),
        Title:     title,
        Content:   content,
        CreatedAt: time.Now(),
    }
    
    if err := n.Validate(); err != nil {
        return nil, err
    }
    
    if err := uc.repo.Create(ctx, n); err != nil {
        return nil, err
    }
    
    return n, nil
}
```

---

## Interface Layer (`internal/interface/`)

**Responsibility:** Adapters for external communication.

### HTTP Handler Pattern

```go
// internal/interface/http/note/handler.go
type Handler struct {
    usecase *note.UseCase
    logger  observability.Logger
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateNoteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }
    
    n, err := h.usecase.Create(r.Context(), req.Title, req.Content)
    if err != nil {
        handleError(w, err)
        return
    }
    
    response.Created(w, toNoteResponse(n))
}
```

### Error Mapping

```go
func handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, note.ErrNoteNotFound):
        response.NotFound(w, err.Error())
    case errors.Is(err, note.ErrEmptyTitle):
        response.ValidationError(w, err.Error())
    default:
        response.InternalError(w)
    }
}
```

---

## Infrastructure Layer (`internal/infra/`)

**Responsibility:** Implements domain interfaces with external services.

### Repository Implementation

```go
// internal/infra/postgres/note_repository.go
type NoteRepository struct {
    pool *pgxpool.Pool
}

func (r *NoteRepository) Create(ctx context.Context, n *note.Note) error {
    _, err := r.pool.Exec(ctx, 
        "INSERT INTO notes (id, title, content, created_at) VALUES ($1, $2, $3, $4)",
        n.ID, n.Title, n.Content, n.CreatedAt,
    )
    return err
}

func (r *NoteRepository) GetByID(ctx context.Context, id string) (*note.Note, error) {
    row := r.pool.QueryRow(ctx, "SELECT id, title, content, created_at FROM notes WHERE id = $1", id)
    
    var n note.Note
    if err := row.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, note.ErrNoteNotFound
        }
        return nil, err
    }
    
    return &n, nil
}
```

---

## Background Worker Architecture

Worker adalah **entry point sekunder**, bukan layer arsitektur kelima.

```
cmd/worker/main.go
       │
       ▼
internal/worker/worker.go (Asynq server setup)
       │
       ▼
internal/worker/registry.go (Task handler registration)
       │
       ▼
internal/worker/tasks/*.go (Task definitions)
       │
       ▼
internal/usecase/*  (Business logic - same as HTTP)
       │
       ▼
internal/domain/*   (Entities)
```

### Async Job Patterns

| Pattern | Use Case | Location |
|---------|----------|----------|
| **Fire-and-Forget** | Analytics, audit logs | `patterns/fireandforget.go` |
| **Scheduled** | Periodic cleanup, reports | `patterns/scheduled.go` |
| **Fanout** | One event → multiple handlers | `patterns/fanout.go` |
| **Idempotency** | Prevent duplicate processing | `idempotency/handler.go` |

---

## Security Architecture

### Authentication Flow

```
Request → APIKey/JWT Middleware → Auth Context → Handler → UseCase
```

| Component | Location | Purpose |
|-----------|----------|---------|
| JWT Auth | `middleware/jwt.go` | Token validation |
| API Key Auth | `middleware/apikey.go` | Service-to-service |
| RBAC | `middleware/rbac.go` | Role/permission check |
| Auth Context | `middleware/auth.go` | Unified auth interface |

### Middleware Order

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.RequestID)
    r.Use(middleware.SecurityHeaders)
    r.Use(middleware.AuthMiddleware(jwtAuth))  // Authentication first
    r.Use(middleware.RequireRole("admin"))     // Authorization second
    // ... handlers
})
```

---

## Observability Architecture

### Components

| Component | Technology | Location |
|-----------|------------|----------|
| **Logging** | Zap | `internal/observability/logger.go` |
| **Tracing** | OpenTelemetry | `internal/observability/tracer.go` |
| **Metrics** | Prometheus | `internal/observability/metrics.go` |
| **Audit** | Structured logs | `internal/observability/audit.go` |

### Instrumentation Points

- HTTP handlers (via middleware)
- gRPC interceptors
- Database queries
- Background jobs
- External service calls

---

## Conventions

### Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case | `note_handler.go` |
| Packages | lowercase | `note`, `postgres` |
| Types | PascalCase | `NoteHandler`, `CreateNoteRequest` |
| Functions | PascalCase (exported) | `NewHandler()` |

### File Organization

```
internal/interface/http/{domain}/
├── handler.go            # HTTP handlers
├── handler_test.go       # Handler tests
├── dto.go                # Request/Response DTOs
└── routes.go             # Optional: route registration
```

---

## Quick Reference: Adding New Domain

1. Create entity: `internal/domain/{name}/entity.go`
2. Create errors: `internal/domain/{name}/errors.go`
3. Create repository interface: `internal/domain/{name}/repository.go`
4. Create migration: `db/migrations/`
5. Create queries: `db/queries/{name}.sql`
6. Run: `make sqlc`
7. Create usecase: `internal/usecase/{name}/usecase.go`
8. Create handler: `internal/interface/http/{name}/handler.go`
9. Register routes in `router.go`

---

*Generated by BMad Method document-project workflow on 2025-12-15*

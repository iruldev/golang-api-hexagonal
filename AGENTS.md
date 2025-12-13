# AI Assistant Guide (AGENTS.md)

This document serves as a contract between AI assistants (e.g., GitHub Copilot, Claude, ChatGPT) and the codebase. Following these guidelines ensures consistent, high-quality contributions.

---

## ✅ DO

### Architecture

- **DO** follow hexagonal (ports and adapters) architecture
- **DO** respect layer boundaries: domain → usecase → interface → infra
- **DO** define interfaces in the domain layer (ports)
- **DO** implement interfaces in the infra layer (adapters)

### Code Style

- **DO** use standard Go idioms and conventions
- **DO** prefer the standard library over third-party packages
- **DO** write clear, self-documenting code
- **DO** add comments for non-obvious logic
- **DO** use meaningful variable and function names

### Patterns

- **DO** use the repository pattern for data access
- **DO** use the response envelope pattern for HTTP responses
- **DO** use table-driven tests with AAA pattern
- **DO** use dependency injection via constructors
- **DO** use sentinel errors for domain errors

### Testing

- **DO** write unit tests for all new code
- **DO** use mocks for external dependencies
- **DO** maintain ≥70% test coverage
- **DO** write integration tests for HTTP handlers

---

## ❌ DON'T

### Architecture

- **DON'T** import from `interface/` or `infra/` in the domain layer
- **DON'T** bypass layers (e.g., handler calling repo directly)
- **DON'T** put business logic in handlers
- **DON'T** put HTTP concerns in use cases

### Code Style

- **DON'T** use `panic` for error handling (except truly unrecoverable)
- **DON'T** ignore error returns
- **DON'T** use global state
- **DON'T** write clever code over clear code

### Patterns

- **DON'T** create new patterns without referencing existing ones
- **DON'T** skip validation in domain entities
- **DON'T** return raw database errors to HTTP layer
- **DON'T** use magic numbers/strings

### Testing

- **DON'T** skip tests for "simple" code
- **DON'T** write tests that depend on external services
- **DON'T** use `time.Sleep` in tests (use mocks/channels)
- **DON'T** leave commented-out test code

---

## 📁 File Structure Conventions

### Per Domain Structure

```
internal/
├── domain/{name}/
│   ├── entity.go           # Main entity struct with Validate()
│   ├── entity_test.go      # Entity unit tests
│   ├── errors.go           # Domain-specific sentinel errors
│   └── repository.go       # Repository interface (port)
│
├── usecase/{name}/
│   ├── usecase.go          # Business logic with repo dependency
│   └── usecase_test.go     # Unit tests with mock repository
│
├── interface/http/{name}/
│   ├── handler.go                    # HTTP handlers
│   ├── handler_test.go               # Handler unit tests
│   ├── handler_integration_test.go   # Integration tests (build-tagged)
│   └── dto.go                        # Request/Response DTOs
│
└── infra/postgres/{name}/
    └── (sqlc-generated files)
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case.go | `note_handler.go` |
| Packages | lowercase | `note`, `postgres` |
| Types | PascalCase | `NoteHandler` |
| Exported funcs | PascalCase | `NewHandler()` |
| Private funcs | camelCase | `handleError()` |
| Variables | camelCase | `noteID` |
| Constants | PascalCase | `MaxTitleLength` |
| Errors | Err prefix | `ErrNoteNotFound` |

### Database Conventions

| Element | Location |
|---------|----------|
| Migrations | `db/migrations/YYYYMMDDHHMMSS_description.{up,down}.sql` |
| SQLC queries | `db/queries/{name}.sql` |
| Generated code | `internal/infra/postgres/{name}/` |

---

## 🧪 Testing Requirements

### Unit Tests

```go
// Required: Table-driven test with AAA pattern
func TestUsecase_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"valid", "Title", nil},
        {"empty", "", ErrEmptyTitle},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            repo := &MockRepository{}
            uc := NewUsecase(repo)
            
            // Act
            _, err := uc.Create(ctx, tt.input, "content")
            
            // Assert
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Coverage Requirements

| Layer | Minimum Coverage |
|-------|------------------|
| Domain | 90% |
| Use Case | 80% |
| Handler | 70% |
| Overall | 70% |

### Mock Pattern

```go
// Mock repository for testing
type MockRepository struct {
    CreateFunc func(ctx context.Context, n *Note) error
    GetFunc    func(ctx context.Context, id uuid.UUID) (*Note, error)
    // ... other methods
}

func (m *MockRepository) Create(ctx context.Context, n *Note) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, n)
    }
    return nil
}
```

### Integration Tests

```go
//go:build integration
// +build integration

func TestHandler_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup httptest.NewServer with real router
    srv := httptest.NewServer(router)
    defer srv.Close()
    
    // Make real HTTP requests
    resp, _ := http.Get(srv.URL + "/api/v1/notes")
    // Assert response
}
```

---

## 🔧 Common Tasks

### Adding a New Domain

1. Create entity: `internal/domain/{name}/entity.go`
2. Create errors: `internal/domain/{name}/errors.go`
3. Create repository interface: `internal/domain/{name}/repository.go`
4. Create migration: `db/migrations/{timestamp}_{description}.up.sql`
5. Create SQLC queries: `db/queries/{name}.sql`
6. Run `make sqlc`
7. Create usecase: `internal/usecase/{name}/usecase.go`
8. Create HTTP handler: `internal/interface/http/{name}/handler.go`
9. Create DTOs: `internal/interface/http/{name}/dto.go`
10. Register routes in router
11. Write tests for all layers

### Error Handling Flow

```
Domain Error (ErrNoteNotFound)
    ↓
Usecase (returns domain error)
    ↓
Handler (maps to HTTP status)
    ↓
Response (JSON envelope with error code)
```

### Adding a New Async Job

> **For comprehensive async job documentation, see [`docs/async-jobs.md`](docs/async-jobs.md)**

#### Step 1: Choose Your Pattern

| Scenario | Pattern | Package |
|----------|---------|---------|
| Non-critical background (analytics, audit) | Fire-and-Forget | `internal/worker/patterns/fireandforget.go` |
| Periodic tasks (cleanup, reports) | Scheduled | `internal/worker/patterns/scheduled.go` |
| Event → multiple handlers | Fanout | `internal/worker/patterns/fanout.go` |
| Critical operations (payments, orders) | Standard + Idempotency | `internal/worker/idempotency/` |

#### Step 2: Create Task Type

Add to `internal/worker/tasks/types.go`:

```go
const (
    Type{Name} = "{domain}:{action}"  // e.g., TypeEmailSend = "email:send"
)
```

#### Step 3: Create Task Handler

Create `internal/worker/tasks/{name}.go`:

```go
// 1. Payload struct
type {Name}Payload struct {
    ID uuid.UUID `json:"id"`
}

// 2. Task constructor
func New{Name}Task(id uuid.UUID) (*asynq.Task, error) {
    payload, err := json.Marshal({Name}Payload{ID: id})
    if err != nil {
        return nil, fmt.Errorf("marshal payload: %w", err)
    }
    return asynq.NewTask(Type{Name}, payload, asynq.MaxRetry(3)), nil
}

// 3. Handler struct with dependencies
type {Name}Handler struct {
    logger *zap.Logger
    // Add: repo, usecase, etc.
}

func New{Name}Handler(logger *zap.Logger) *{Name}Handler {
    return &{Name}Handler{logger: logger}
}

// 4. Handle method with validation
func (h *{Name}Handler) Handle(ctx context.Context, t *asynq.Task) error {
    var p {Name}Payload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal: %v: %w", err, asynq.SkipRetry)
    }
    if p.ID == uuid.Nil {
        return fmt.Errorf("id required: %w", asynq.SkipRetry)
    }
    // Process task
    return nil
}
```

#### Step 4: Register Handler

Add to `cmd/worker/main.go`:

```go
handler := tasks.New{Name}Handler(logger)
srv.HandleFunc(tasks.Type{Name}, handler.Handle)
```

#### Step 5: Write Tests

Create `internal/worker/tasks/{name}_test.go` with tests for:
- Valid payload handling
- Invalid/empty payload (SkipRetry)
- Missing required fields (SkipRetry)
- Happy path processing

#### Async Job Creation Checklist

- [ ] Task type constant in `internal/worker/tasks/types.go`
- [ ] Payload struct with JSON tags
- [ ] Task constructor (`New{Name}Task`) with default options
- [ ] Handler struct with dependencies
- [ ] `Handle` method with validation
- [ ] `SkipRetry` for validation errors
- [ ] Handler registered in `cmd/worker/main.go`
- [ ] Unit tests in `internal/worker/tasks/{name}_test.go`

#### Copy Commands

```bash
# Copy reference task file
cp internal/worker/tasks/note_archive.go internal/worker/tasks/{name}.go
cp internal/worker/tasks/note_archive_test.go internal/worker/tasks/{name}_test.go

# Replace placeholders (macOS)
sed -i '' 's/NoteArchive/{Name}/g' internal/worker/tasks/{name}.go
sed -i '' 's/note:archive/{domain}:{action}/g' internal/worker/tasks/{name}.go
sed -i '' 's/NoteID/YourFieldID/g' internal/worker/tasks/{name}.go

# Linux: use sed -i without quotes: sed -i 's/NoteArchive/{Name}/g' ...

# Add type constant to types.go
# Manually add: Type{Name} = "{domain}:{action}"

# Register handler in cmd/worker/main.go
# Manually add: srv.HandleFunc(tasks.Type{Name}, handler.Handle)
```

#### Queue Selection Guide

| Priority | Queue | Weight | When to Use |
|----------|-------|--------|-------------|
| High | `critical` | 6 | User-facing (email, notifications) |
| Normal | `default` | 3 | Business logic (archival, sync) |
| Low | `low` | 1 | Analytics, cleanup, batch jobs |

#### Pattern Selection Decision Table

| Your Scenario | Use This Pattern | Queue |
|---------------|------------------|-------|
| Non-critical, best-effort | `patterns.FireAndForget()` | `low` |
| Scheduled cleanup/reports | `patterns.RegisterScheduledJobs()` | `default` |
| Single event → multiple actions | `patterns.Fanout()` | per-handler |
| Prevent duplicate processing | `idempotency.IdempotentHandler()` | any |
| Critical with confirmation | Standard enqueue | `critical` |

---

## 📋 Checklist for Code Review

Before submitting code, verify:

- [ ] Follows hexagonal architecture
- [ ] No layer violations
- [ ] Uses existing patterns
- [ ] Has unit tests
- [ ] Uses table-driven tests
- [ ] Follows AAA pattern
- [ ] Has meaningful test names
- [ ] Domain has validation
- [ ] Errors are sentinel errors
- [ ] HTTP uses response envelope
- [ ] No global state
- [ ] No `panic` for error handling
- [ ] Comments for non-obvious logic
- [ ] Follows naming conventions

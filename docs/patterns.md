# Copy-Paste Code Patterns

Panduan lengkap untuk menambahkan fitur baru ke dalam project golang-api-hexagonal menggunakan pola-pola yang sudah established. Semua contoh kode dapat langsung di-copy-paste dengan modifikasi minimal.

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Domain Entity Pattern](#domain-entity-pattern)
3. [Repository Interface Pattern](#repository-interface-pattern)
4. [Use Case Pattern](#use-case-pattern)
5. [HTTP Handler Pattern](#http-handler-pattern)
6. [Error Handling Pattern](#error-handling-pattern)
7. [Request/Response DTO Pattern](#requestresponse-dto-pattern)
8. [Middleware Pattern](#middleware-pattern)
9. [RFC 7807 Error Response Pattern](#rfc-7807-error-response-pattern)
10. [Transaction Management Pattern](#transaction-management-pattern)
11. [Audit Event Recording Pattern](#audit-event-recording-pattern)
12. [Test Patterns](#test-patterns)
13. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Quick Reference

### File Structure untuk Entity Baru

```
internal/
├── domain/
│   ├── {entity}.go           # Entity + Repository interface
│   └── {entity}_test.go      # Entity validation tests
├── app/{entity}/
│   ├── create_{entity}.go    # Create use case
│   ├── create_{entity}_test.go
│   ├── get_{entity}.go       # Get use case  
│   ├── get_{entity}_test.go
│   ├── list_{entity}.go      # List use case (optional)
│   └── list_{entity}_test.go
├── transport/http/
│   ├── handler/{entity}.go   # HTTP handlers
│   ├── handler/{entity}_*_test.go
│   └── contract/             # Request/Response DTOs
├── infra/postgres/
│   ├── {entity}_repo.go      # Repository implementation
│   └── {entity}_repo_test.go
└── infra/postgres/migrations/
    └── {version}_{entity}.up.sql
```

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Entity struct | PascalCase, singular | `User`, `Order`, `Product` |
| Repository interface | `{Entity}Repository` | `UserRepository`, `OrderRepository` |
| Use case struct | `{Action}{Entity}UseCase` | `CreateUserUseCase`, `GetOrderUseCase` |
| Handler struct | `{Entity}Handler` | `UserHandler`, `OrderHandler` |
| Error codes | `{CATEGORY}-{NUMBER}` | `USR-001`, `ORD-001`, `PROD-001` |
| Table names | snake_case, plural | `users`, `orders`, `products` |
| Migration files | `{version}_{action}_{entity}.sql` | `000006_add_orders.up.sql` |

### Layer Boundaries (enforced by depguard)

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| `domain/` | stdlib only | app, transport, infra, external |
| `app/` | domain | transport, infra |
| `transport/http/` | domain, app | infra (except through DI) |
| `infra/` | domain, app | transport |

---

## Domain Entity Pattern

**Location:** `internal/domain/{entity}.go`

### Template

```go
package domain

import (
	"context"
	"strings"
	"time"
)

// == ENTITY DEFINITION ==

// {Entity} represents a domain entity for {entity} data.
// This entity follows hexagonal architecture principles - no external dependencies.
type {Entity} struct {
	ID        ID
	// TODO: Add your entity fields here
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the {Entity} entity has valid required fields.
// Returns a domain error if validation fails.
func (e {Entity}) Validate() error {
	if strings.TrimSpace(e.Name) == "" {
		return ErrInvalid{Field}
	}
	// TODO: Add more validation rules
	return nil
}

// == REPOSITORY INTERFACE ==

// {Entity}Repository defines the interface for {entity} persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
//
//go:generate mockgen -destination=../testutil/mocks/{entity}_repository_mock.go -package=mocks github.com/iruldev/golang-api-hexagonal/internal/domain {Entity}Repository
type {Entity}Repository interface {
	// Create stores a new {entity}.
	Create(ctx context.Context, q Querier, entity *{Entity}) error

	// GetByID retrieves a {entity} by its ID.
	GetByID(ctx context.Context, q Querier, id ID) (*{Entity}, error)

	// List retrieves {entity}s with pagination.
	// Returns the slice of {entity}s, total count of matching items, and any error.
	List(ctx context.Context, q Querier, params ListParams) ([]{Entity}, int, error)
}
```

### Contoh Penggunaan (User Entity)

```go
// Source: internal/domain/user.go

package domain

import (
	"context"
	"strings"
	"time"
)

// User represents a domain entity for user data.
type User struct {
	ID        ID
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the User entity has valid required fields.
func (u User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return ErrInvalidEmail
	}
	if strings.TrimSpace(u.FirstName) == "" {
		return ErrInvalidFirstName
	}
	if strings.TrimSpace(u.LastName) == "" {
		return ErrInvalidLastName
	}
	return nil
}

//go:generate mockgen -destination=../testutil/mocks/user_repository_mock.go -package=mocks github.com/iruldev/golang-api-hexagonal/internal/domain UserRepository
type UserRepository interface {
	Create(ctx context.Context, q Querier, user *User) error
	GetByID(ctx context.Context, q Querier, id ID) (*User, error)
	List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
}
```

---

## Repository Interface Pattern

**Location:** `internal/domain/{entity}.go` (interface) dan `internal/infra/postgres/{entity}_repo.go` (implementation)

### Template Interface

```go
// {Entity}Repository defines the interface for {entity} persistence operations.
//
//go:generate mockgen -destination=../testutil/mocks/{entity}_repository_mock.go -package=mocks github.com/iruldev/golang-api-hexagonal/internal/domain {Entity}Repository
type {Entity}Repository interface {
	// Create stores a new {entity}.
	Create(ctx context.Context, q Querier, entity *{Entity}) error

	// GetByID retrieves a {entity} by its ID.
	GetByID(ctx context.Context, q Querier, id ID) (*{Entity}, error)

	// List retrieves {entity}s with pagination.
	List(ctx context.Context, q Querier, params ListParams) ([]{Entity}, int, error)

	// Update modifies an existing {entity}.
	Update(ctx context.Context, q Querier, entity *{Entity}) error

	// Delete removes a {entity} by ID.
	Delete(ctx context.Context, q Querier, id ID) error
}
```

### Template Implementation

**Location:** `internal/infra/postgres/{entity}_repo.go`

```go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
)

const (
	pgUniqueViolation = "23505"
)

// {Entity}Repo implements domain.{Entity}Repository for PostgreSQL.
type {Entity}Repo struct{}

// New{Entity}Repo creates a new {Entity}Repo instance.
func New{Entity}Repo() *{Entity}Repo {
	return &{Entity}Repo{}
}

// Create stores a new {entity} in the database.
func (r *{Entity}Repo) Create(ctx context.Context, q domain.Querier, entity *domain.{Entity}) error {
	const op = "{entity}Repo.Create"

	dbtx, err := getDBTX(q)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	// Parse domain.ID to uuid
	uid, err := uuid.Parse(string(entity.ID))
	if err != nil {
		return fmt.Errorf("%s: parse ID: %w", op, err)
	}

	// TODO: Prepare params for sqlc generated query
	params := sqlcgen.Create{Entity}Params{
		ID:        pgtype.UUID{Bytes: uid, Valid: true},
		Name:      entity.Name,
		CreatedAt: pgtype.Timestamptz{Time: entity.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: entity.UpdatedAt, Valid: true},
	}

	if err := queries.Create{Entity}(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			// TODO: Handle specific constraint violations
			return fmt.Errorf("%s: %w", op, domain.Err{Entity}AlreadyExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetByID retrieves a {entity} by its ID.
func (r *{Entity}Repo) GetByID(ctx context.Context, q domain.Querier, id domain.ID) (*domain.{Entity}, error) {
	const op = "{entity}Repo.GetByID"

	dbtx, err := getDBTX(q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("%s: parse ID: %w", op, err)
	}

	dbEntity, err := queries.Get{Entity}ByID(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, domain.Err{Entity}NotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Map DB entity to domain entity
	uuidVal, err := uuid.FromBytes(dbEntity.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid from db: %w", op, err)
	}

	return &domain.{Entity}{
		ID:        domain.ID(uuidVal.String()),
		Name:      dbEntity.Name,
		CreatedAt: dbEntity.CreatedAt.Time,
		UpdatedAt: dbEntity.UpdatedAt.Time,
	}, nil
}

// Ensure {Entity}Repo implements domain.{Entity}Repository at compile time.
var _ domain.{Entity}Repository = (*{Entity}Repo)(nil)
```

---

## Use Case Pattern

**Location:** `internal/app/{entity}/create_{entity}.go`

### Template

```go
package {entity}

import (
	"context"
	"errors"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/audit"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// Create{Entity}Request represents the input data for creating a new {entity}.
type Create{Entity}Request struct {
	ID   domain.ID
	Name string
	// TODO: Add other request fields
	
	// RequestID correlates this operation with the HTTP request.
	RequestID string
	// ActorID identifies who is performing this action.
	ActorID domain.ID
}

// Create{Entity}Response represents the result of creating a new {entity}.
type Create{Entity}Response struct {
	{Entity} domain.{Entity}
}

// Create{Entity}UseCase handles the business logic for creating a new {entity}.
type Create{Entity}UseCase struct {
	{entity}Repo domain.{Entity}Repository
	auditService *audit.AuditService
	idGen        domain.IDGenerator
	txManager    domain.TxManager
	db           domain.Querier
}

// NewCreate{Entity}UseCase creates a new instance of Create{Entity}UseCase.
func NewCreate{Entity}UseCase(
	{entity}Repo domain.{Entity}Repository,
	auditService *audit.AuditService,
	idGen domain.IDGenerator,
	txManager domain.TxManager,
	db domain.Querier,
) *Create{Entity}UseCase {
	return &Create{Entity}UseCase{
		{entity}Repo: {entity}Repo,
		auditService: auditService,
		idGen:        idGen,
		txManager:    txManager,
		db:           db,
	}
}

// Execute processes the create {entity} request.
func (uc *Create{Entity}UseCase) Execute(ctx context.Context, req Create{Entity}Request) (Create{Entity}Response, error) {
	// 1. Create domain entity with generated ID
	id := req.ID
	if id.IsEmpty() {
		id = uc.idGen.NewID()
	}
	now := time.Now().UTC()
	entity := &domain.{Entity}{
		ID:        id,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 2. Validate using domain rules
	if err := entity.Validate(); err != nil {
		return Create{Entity}Response{}, &app.AppError{
			Op:      "Create{Entity}",
			Code:    app.CodeValidationError,
			Message: "Validation failed",
			Err:     err,
		}
	}

	// 3. Execute in transaction
	if err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
		// Create the entity
		if err := uc.{entity}Repo.Create(ctx, tx, entity); err != nil {
			if errors.Is(err, domain.Err{Entity}AlreadyExists) {
				return &app.AppError{
					Op:      "Create{Entity}",
					Code:    app.CodeConflict,
					Message: "{Entity} already exists",
					Err:     err,
				}
			}
			return &app.AppError{
				Op:      "Create{Entity}",
				Code:    app.CodeInternalError,
				Message: "Failed to create {entity}",
				Err:     err,
			}
		}

		// Record audit event
		auditInput := audit.AuditEventInput{
			EventType:  domain.Event{Entity}Created,
			ActorID:    req.ActorID,
			EntityType: "{entity}",
			EntityID:   entity.ID,
			Payload:    entity,
			RequestID:  req.RequestID,
		}

		if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
			return &app.AppError{
				Op:      "Create{Entity}",
				Code:    app.CodeInternalError,
				Message: "Failed to record audit event",
				Err:     err,
			}
		}

		return nil
	}); err != nil {
		return Create{Entity}Response{}, err
	}

	return Create{Entity}Response{{Entity}: *entity}, nil
}
```

---

## HTTP Handler Pattern

**Location:** `internal/transport/http/handler/{entity}.go`

### Template

```go
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/{entity}"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// == INTERFACES (for dependency injection) ==

type create{Entity}Executor interface {
	Execute(ctx context.Context, req {entity}.Create{Entity}Request) ({entity}.Create{Entity}Response, error)
}

type get{Entity}Executor interface {
	Execute(ctx context.Context, req {entity}.Get{Entity}Request) ({entity}.Get{Entity}Response, error)
}

type list{Entity}sExecutor interface {
	Execute(ctx context.Context, req {entity}.List{Entity}sRequest) ({entity}.List{Entity}sResponse, error)
}

// == HANDLER ==

// {Entity}Handler handles {entity}-related HTTP requests.
type {Entity}Handler struct {
	createUC     create{Entity}Executor
	getUC        get{Entity}Executor
	listUC       list{Entity}sExecutor
	resourcePath string
}

// New{Entity}Handler creates a new {Entity}Handler.
func New{Entity}Handler(
	createUC create{Entity}Executor,
	getUC get{Entity}Executor,
	listUC list{Entity}sExecutor,
	resourcePath string,
) *{Entity}Handler {
	return &{Entity}Handler{
		createUC:     createUC,
		getUC:        getUC,
		listUC:       listUC,
		resourcePath: resourcePath,
	}
}

// Create{Entity} handles POST /api/v1/{entity}s.
func (h *{Entity}Handler) Create{Entity}(w http.ResponseWriter, r *http.Request) {
	// 1. Decode and validate request
	var req contract.Create{Entity}Request
	if errs := contract.ValidateRequestBody(r, &req); len(errs) > 0 {
		contract.WriteValidationError(w, r, errs)
		return
	}

	// 2. Generate UUID v7 at transport boundary
	id, err := uuid.NewV7()
	if err != nil {
		contract.WriteProblemJSON(w, r, &app.AppError{
			Op:      "Create{Entity}",
			Code:    app.CodeInternalError,
			Message: "Failed to generate ID",
			Err:     err,
		})
		return
	}

	// 3. Extract context values for audit trail
	reqID := ctxutil.GetRequestID(r.Context())
	var actorID domain.ID
	if authCtx := app.GetAuthContext(r.Context()); authCtx != nil {
		actorID = domain.ID(authCtx.SubjectID)
	}

	// 4. Map to app layer request
	appReq := {entity}.Create{Entity}Request{
		ID:        domain.ID(id.String()),
		Name:      req.Name,
		RequestID: reqID,
		ActorID:   actorID,
	}

	// 5. Execute use case
	resp, err := h.createUC.Execute(r.Context(), appReq)
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	// 6. Map to response and write
	{entity}Resp := contract.To{Entity}Response(resp.{Entity})
	
	// Set Location header for 201 Created
	location := fmt.Sprintf("%s/%s", h.resourcePath, resp.{Entity}.ID)
	w.Header().Set("Location", location)

	_ = contract.WriteJSON(w, http.StatusCreated, contract.DataResponse[contract.{Entity}Response]{Data: {entity}Resp})
}

// Get{Entity} handles GET /api/v1/{entity}s/{id}.
func (h *{Entity}Handler) Get{Entity}(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	// Validate UUID format and version
	parsedID, err := uuid.Parse(idParam)
	if err != nil {
		contract.WriteValidationError(w, r, []contract.ValidationError{
			{Field: "id", Message: "must be a valid UUID"},
		})
		return
	}
	if parsedID.Version() != 7 {
		contract.WriteValidationError(w, r, []contract.ValidationError{
			{Field: "id", Message: "must be UUID v7 (time-ordered)"},
		})
		return
	}

	resp, err := h.getUC.Execute(r.Context(), {entity}.Get{Entity}Request{ID: domain.ID(parsedID.String())})
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	{entity}Resp := contract.To{Entity}Response(resp.{Entity})
	_ = contract.WriteJSON(w, http.StatusOK, contract.DataResponse[contract.{Entity}Response]{Data: {entity}Resp})
}

// List{Entity}s handles GET /api/v1/{entity}s.
func (h *{Entity}Handler) List{Entity}s(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 20

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			contract.WriteValidationError(w, r, []contract.ValidationError{
				{Field: "page", Message: "must be a positive integer", Code: contract.CodeValOutOfRange},
			})
			return
		}
		page = p
	}

	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 {
			contract.WriteValidationError(w, r, []contract.ValidationError{
				{Field: "pageSize", Message: "must be a positive integer", Code: contract.CodeValOutOfRange},
			})
			return
		}
		if ps > 100 {
			ps = 100 // Cap at 100
		}
		pageSize = ps
	}

	req := {entity}.List{Entity}sRequest{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.listUC.Execute(r.Context(), req)
	if err != nil {
		contract.WriteProblemJSON(w, r, err)
		return
	}

	listResp := contract.NewList{Entity}sResponse(resp.{Entity}s, page, pageSize, resp.TotalCount)
	_ = contract.WriteJSON(w, http.StatusOK, listResp)
}
```

---

## Error Handling Pattern

### Error Flow

```
Domain Error → App Layer (wrap with AppError) → Transport Layer → RFC 7807 Response
```

### Domain Errors

**Location:** `internal/domain/errors.go`

```go
package domain

import "errors"

// Domain-specific errors for {entity}.
// Prefix with Err{Entity} for namespacing.
var (
	Err{Entity}NotFound      = errors.New("{entity} not found")
	Err{Entity}AlreadyExists = errors.New("{entity} already exists")
	ErrInvalid{Field}        = errors.New("invalid {field}")
)
```

### App Layer Error Wrapping

**Location:** `internal/app/{entity}/*.go`

```go
// Always wrap domain errors with AppError for proper HTTP mapping
if errors.Is(err, domain.Err{Entity}NotFound) {
	return Response{}, &app.AppError{
		Op:      "Get{Entity}",
		Code:    app.CodeNotFound,      // Maps to 404
		Message: "{Entity} not found",
		Err:     err,
	}
}

// For internal errors
return Response{}, &app.AppError{
	Op:      "Get{Entity}",
	Code:    app.CodeInternalError,     // Maps to 500
	Message: "Failed to retrieve {entity}",
	Err:     err,
}
```

### Error Code Registry

**Location:** `internal/transport/http/contract/errors.go`

```go
// Error codes follow {CATEGORY}-{NUMBER} format
const (
	// {ENTITY} Domain Errors ({ENT}-xxx)
	Code{Ent}{Entity}NotFound      = "{ENT}-001"
	Code{Ent}{Entity}AlreadyExists = "{ENT}-002"
	Code{Ent}Invalid{Field}        = "{ENT}-003"
)
```

---

## Request/Response DTO Pattern

**Location:** `internal/transport/http/contract/{entity}.go`

### Template

```go
package contract

import (
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// == REQUEST DTOs ==

// Create{Entity}Request represents the HTTP request for creating a {entity}.
type Create{Entity}Request struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	// TODO: Add other fields with validation tags
}

// Update{Entity}Request represents the HTTP request for updating a {entity}.
type Update{Entity}Request struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

// == RESPONSE DTOs ==

// {Entity}Response represents the HTTP response for a {entity}.
type {Entity}Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// To{Entity}Response converts a domain {entity} to the HTTP response format.
func To{Entity}Response(e domain.{Entity}) {Entity}Response {
	return {Entity}Response{
		ID:        string(e.ID),
		Name:      e.Name,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// == LIST RESPONSE ==

// List{Entity}sResponse represents the HTTP response for listing {entity}s.
type List{Entity}sResponse struct {
	Data       []{Entity}Response `json:"data"`
	Pagination PaginationMeta     `json:"pagination"`
}

// NewList{Entity}sResponse creates a paginated list response.
func NewList{Entity}sResponse(entities []domain.{Entity}, page, pageSize, totalCount int) List{Entity}sResponse {
	data := make([]{Entity}Response, len(entities))
	for i, e := range entities {
		data[i] = To{Entity}Response(e)
	}

	return List{Entity}sResponse{
		Data: data,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: (totalCount + pageSize - 1) / pageSize,
		},
	}
}
```

---

## Middleware Pattern

**Location:** `internal/transport/http/middleware/{middleware}.go`

### Template

```go
package middleware

import (
	"net/http"
)

// {Name}Config holds configuration for {name} middleware.
type {Name}Config struct {
	Enabled bool
	// TODO: Add config fields
}

// New{Name}Middleware creates a new {name} middleware.
func New{Name}Middleware(config {Name}Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Pre-processing (before handler)
			// TODO: Add middleware logic

			// Call the next handler
			next.ServeHTTP(w, r)

			// Post-processing (after handler)
			// Note: Cannot modify response body after next.ServeHTTP
		})
	}
}
```

### Contoh: Context Value Middleware

```go
// RequestIDMiddleware adds a request ID to the context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get existing or generate new request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), ctxkey.RequestID, reqID)
		
		// Set response header
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

---

## RFC 7807 Error Response Pattern

**Location:** `internal/transport/http/contract/problem.go`

### Menulis Error Response

```go
// Di handler, gunakan WriteProblemJSON untuk semua error response
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	result, err := h.useCase.Execute(r.Context(), req)
	if err != nil {
		// AppError akan otomatis dikonversi ke RFC 7807
		contract.WriteProblemJSON(w, r, err)
		return
	}
	// ... success response
}
```

### Response Format

```json
{
	"type": "https://api.example.com/problems/user-not-found",
	"title": "User Not Found",
	"status": 404,
	"detail": "User with ID xyz was not found",
	"code": "USR-001",
	"request_id": "req_abc123",
	"trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

### Validation Error Response

```json
{
	"type": "https://api.example.com/problems/validation-error",
	"title": "Validation Error",
	"status": 400,
	"detail": "One or more fields failed validation",
	"code": "VAL-001",
	"request_id": "req_abc123",
	"trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
	"validation_errors": [
		{
			"field": "email",
			"message": "must be a valid email address",
			"code": "VAL-002"
		}
	]
}
```

---

## Transaction Management Pattern

### Template

```go
// Semua operasi yang perlu atomic harus dibungkus dalam WithTx
if err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
	// Operation 1 - menggunakan tx sebagai Querier
	if err := uc.entityRepo.Create(ctx, tx, entity); err != nil {
		return &app.AppError{
			Op:      "CreateEntity",
			Code:    app.CodeInternalError,
			Message: "Failed to create entity",
			Err:     err,
		}
	}

	// Operation 2 - masih dalam transaction yang sama
	if err := uc.relatedRepo.Create(ctx, tx, relatedEntity); err != nil {
		return &app.AppError{
			Op:      "CreateEntity",
			Code:    app.CodeInternalError,
			Message: "Failed to create related entity",
			Err:     err,
		}
	}

	// Operation 3 - Audit event juga dalam transaction
	if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
		return &app.AppError{
			Op:      "CreateEntity",
			Code:    app.CodeInternalError,
			Message: "Failed to record audit event",
			Err:     err,
		}
	}

	// Return nil = commit transaction
	return nil
}); err != nil {
	// Error returned = rollback transaction
	return Response{}, err
}
```

---

## Audit Event Recording Pattern

### Template

```go
// Di dalam use case, setelah operasi utama
auditInput := audit.AuditEventInput{
	EventType:  domain.Event{Entity}Created, // atau Updated, Deleted
	ActorID:    req.ActorID,                  // Dari JWT claims
	EntityType: "{entity}",                   // String literal
	EntityID:   entity.ID,                    // ID dari entity
	Payload:    entity,                       // Entity atau diff data
	RequestID:  req.RequestID,                // Untuk correlation
}

if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
	return &app.AppError{
		Op:      "Create{Entity}",
		Code:    app.CodeInternalError,
		Message: "Failed to record audit event",
		Err:     err,
	}
}
```

### Event Types

**Location:** `internal/domain/audit.go`

```go
const (
	Event{Entity}Created EventType = "{entity}.created"
	Event{Entity}Updated EventType = "{entity}.updated"
	Event{Entity}Deleted EventType = "{entity}.deleted"
)
```

---

## Test Patterns

### Unit Test Template

```go
func Test{Entity}_{Method}_{Scenario}(t *testing.T) {
	tests := []struct {
		name    string
		input   InputType
		want    OutputType
		wantErr error
	}{
		{
			name:  "valid input succeeds",
			input: validInput,
			want:  expectedOutput,
		},
		{
			name:    "invalid input returns error",
			input:   invalidInput,
			wantErr: ErrExpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Function(tt.input)
			
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

### Handler Test Template

```go
func TestHandler_Create{Entity}_Success(t *testing.T) {
	// Setup mocks
	mockUC := new(MockCreate{Entity}UseCase)
	handler := NewHandler(mockUC)

	// Setup expectation
	mockUC.On("Execute", mock.Anything, mock.AnythingOfType("{entity}.Create{Entity}Request")).
		Return({entity}.Create{Entity}Response{{Entity}: expectedEntity}, nil)

	// Create request
	body := `{"name": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/{entity}s", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	rec := httptest.NewRecorder()

	// Execute
	handler.Create{Entity}(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUC.AssertExpectations(t)
}
```

### Lihat juga: [Testing Patterns](./testing-patterns.md)

---

## Anti-Patterns to Avoid

### ❌ DO NOT: Mix Layers

```go
// BAD: Handler importing infra package directly
import "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"

// GOOD: Handler uses interfaces defined in domain
import "github.com/iruldev/golang-api-hexagonal/internal/domain"
```

### ❌ DO NOT: Skip Error Wrapping

```go
// BAD: Return raw error without context
return CreateResponse{}, err

// GOOD: Wrap with AppError for proper HTTP mapping
return CreateResponse{}, &app.AppError{
	Op:      "CreateEntity",
	Code:    app.CodeInternalError,
	Message: "Failed to create entity",
	Err:     err,
}
```

### ❌ DO NOT: Add External Dependencies to Domain Layer

```go
// BAD: Domain importing external package
package domain
import "github.com/jackc/pgx/v5" // ❌

// GOOD: Domain uses only stdlib
package domain
import "context" // ✅
```

### ❌ DO NOT: Forget mockgen Directive

```go
// BAD: Repository interface without mockgen
type EntityRepository interface { ... }

// GOOD: Repository with mockgen directive
//go:generate mockgen -destination=../testutil/mocks/entity_repository_mock.go -package=mocks ...
type EntityRepository interface { ... }
```

### ❌ DO NOT: Use Non-Transaction Operations for Multi-Step Writes

```go
// BAD: Multiple writes without transaction
uc.repo.Create(ctx, uc.db, entity)     // Step 1
uc.auditService.Record(ctx, uc.db, ...) // Step 2 - may fail, leaving inconsistent state

// GOOD: Wrap in transaction
uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
	if err := uc.repo.Create(ctx, tx, entity); err != nil { return err }
	if err := uc.auditService.Record(ctx, tx, ...); err != nil { return err }
	return nil
})
```

---

## Related Documentation

- [Architecture](../_bmad-output/planning-artifacts/architecture.md) - Architectural decisions
- [Testing Patterns](./testing-patterns.md) - Testing conventions
- [Error Codes](./error-codes.md) - Complete error code registry
- [Copy-Paste Kit](./copy-paste-kit/README.md) - Ready-to-use template files

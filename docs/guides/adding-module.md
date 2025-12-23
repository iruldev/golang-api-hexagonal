# Guide: Adding a New Module

This guide provides step-by-step instructions for adding a new business module to the golang-api-hexagonal project. It follows hexagonal architecture principles and uses the Users module as a reference implementation.

## Table of Contents

- [Overview](#overview)
- [Step 1: Create Domain Entity & Repository Interface](#step-1-create-domain-entity--repository-interface)
- [Step 2: Create Database Migration](#step-2-create-database-migration)
- [Step 3: Implement Repository](#step-3-implement-repository)
- [Step 4: Create Use Cases](#step-4-create-use-cases)
- [Step 5: Create DTOs/Contracts](#step-5-create-dtoscontracts)
- [Step 6: Create HTTP Handlers](#step-6-create-http-handlers)
- [Step 7: Wire Routes in Router](#step-7-wire-routes-in-router)
- [Step 8: Add Audit Events](#step-8-add-audit-events)
- [Step 9: Write Tests](#step-9-write-tests)
- [Quick Reference Checklist](#quick-reference-checklist)

---

## Overview

Adding a new module (e.g., "orders") involves creating components across all architectural layers while following strict import rules. The Users module serves as the canonical reference implementation.

> [!IMPORTANT]
> **Layer boundaries are enforced by CI via golangci-lint depguard rules.** Violations will fail the build. Read [docs/architecture.md](../architecture.md) for details.

**Estimated Time:** 2-4 hours for a basic CRUD module

**Prerequisites:**
- Familiarity with Go and hexagonal architecture concepts
- Project running locally (`make run`)
- Database accessible (`make migrate`)

---

## Step 1: Create Domain Entity & Repository Interface

The domain layer defines business entities and contracts. It must have **zero external dependencies** - only Go stdlib.

### 1.1 Define the Entity

Create `internal/domain/order.go`:

```go
package domain

import (
    "context"
    "strings"
    "time"
)

// Order represents a domain entity for order data.
// This entity follows hexagonal architecture principles - no external dependencies.
type Order struct {
    ID          ID
    CustomerID  ID
    Status      string
    TotalAmount int64 // cents
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

> [!CAUTION]
> **Use `type ID string` for identifiers — NEVER `uuid.UUID`!** UUID parsing/generation happens at transport and infra boundaries, not in domain.

### 1.2 Add Validation

Add a `Validate()` method to enforce domain rules:

```go
// Validate checks if the Order entity has valid required fields.
// Returns a domain error if validation fails.
func (o Order) Validate() error {
    if o.ID.IsEmpty() {
        return ErrInvalidID
    }

    if o.CustomerID.IsEmpty() {
        return ErrInvalidCustomerID
    }

    if strings.TrimSpace(o.Status) == "" {
        return ErrInvalidOrderStatus
    }

    if o.TotalAmount < 0 {
        return ErrInvalidOrderAmount
    }

    return nil
}
```

### 1.3 Define Repository Interface

Define the repository contract in the same file. **All methods must accept a `Querier` parameter** to support both connection pool and transaction usage:

```go
// OrderRepository defines the interface for order persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
type OrderRepository interface {
    // Create stores a new order.
    Create(ctx context.Context, q Querier, order *Order) error

    // GetByID retrieves an order by its ID.
    GetByID(ctx context.Context, q Querier, id ID) (*Order, error)

    // List retrieves orders with pagination.
    // Returns the slice of orders, total count of matching orders, and any error.
    List(ctx context.Context, q Querier, params ListParams) ([]Order, int, error)
}
```

### 1.4 Add Domain Errors

Add sentinel errors in `internal/domain/errors.go`:

```go
var (
    // ErrOrderNotFound is returned when an order cannot be found.
    ErrOrderNotFound = errors.New("order not found")

    // ErrInvalidOrderStatus is returned when the order status is empty or invalid.
    ErrInvalidOrderStatus = errors.New("invalid order status")

    // ErrInvalidOrderAmount is returned when the order amount is negative.
    ErrInvalidOrderAmount = errors.New("invalid order amount")

    // ErrInvalidCustomerID is returned when the customer ID is empty or invalid.
    ErrInvalidCustomerID = errors.New("invalid customer ID")
)
```

> [!TIP]
> Use the `Err` prefix for all domain errors. This is the Go convention for sentinel errors.

**Reference:** [internal/domain/user.go](../../internal/domain/user.go)

---

## Step 2: Create Database Migration

### 2.1 Generate Migration File

Use the Makefile command to create a new migration:

```bash
make migrate-create name=create_orders
```

This generates a timestamped file like `migrations/20251223100000_create_orders.sql`.

### 2.2 Write Schema

Edit the generated migration file:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
```

> [!WARNING]
> **Primary Key:** Use `id UUID PRIMARY KEY` — **NO DEFAULT**. The application provides UUID v7, not the database.

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Tables | snake_case, plural | `orders`, `audit_events` |
| Columns | snake_case | `created_at`, `customer_id` |
| Foreign Keys | `{table_singular}_id` | `user_id`, `order_id` |
| Regular Index | `idx_{table}_{column}` | `idx_orders_status` |
| Unique Index | `uniq_{table}_{column}` | `uniq_users_email` |

**Reference:** [migrations/20251217000000_create_users.sql](../../migrations/20251217000000_create_users.sql)

---

## Step 3: Implement Repository

The infrastructure layer implements domain interfaces and handles database operations.

### 3.1 Create Repository Struct

Create `internal/infra/postgres/order_repo.go`:

```go
package postgres

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// OrderRepo implements domain.OrderRepository for PostgreSQL.
type OrderRepo struct{}

// NewOrderRepo creates a new OrderRepo instance.
func NewOrderRepo() *OrderRepo {
    return &OrderRepo{}
}
```

### 3.2 Implement Interface Methods

Implement each repository method with proper error wrapping:

```go
// Create stores a new order in the database.
func (r *OrderRepo) Create(ctx context.Context, q domain.Querier, order *domain.Order) error {
    const op = "orderRepo.Create"

    // Parse domain.ID to uuid.UUID at repository boundary
    id, err := uuid.Parse(string(order.ID))
    if err != nil {
        return fmt.Errorf("%s: parse ID: %w", op, err)
    }

    customerID, err := uuid.Parse(string(order.CustomerID))
    if err != nil {
        return fmt.Errorf("%s: parse CustomerID: %w", op, err)
    }

    _, err = q.Exec(ctx, `
        INSERT INTO orders (id, customer_id, status, total_amount, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, id, customerID, order.Status, order.TotalAmount, order.CreatedAt, order.UpdatedAt)

    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}

// GetByID retrieves an order by its ID.
// It returns domain.ErrOrderNotFound if no order exists with the given ID.
func (r *OrderRepo) GetByID(ctx context.Context, q domain.Querier, id domain.ID) (*domain.Order, error) {
    const op = "orderRepo.GetByID"

    // Parse domain.ID to uuid.UUID
    uid, err := uuid.Parse(string(id))
    if err != nil {
        return nil, fmt.Errorf("%s: parse ID: %w", op, err)
    }

    row := q.QueryRow(ctx, `
        SELECT id, customer_id, status, total_amount, created_at, updated_at
        FROM orders WHERE id = $1
    `, uid)

    // Type assert to rowScanner interface
    scanner, ok := row.(rowScanner)
    if !ok {
        return nil, fmt.Errorf("%s: invalid querier type", op)
    }

    var order domain.Order
    var dbID, dbCustomerID uuid.UUID
    err = scanner.Scan(&dbID, &dbCustomerID, &order.Status, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    order.ID = domain.ID(dbID.String())
    order.CustomerID = domain.ID(dbCustomerID.String())
    return &order, nil
}
```

> [!IMPORTANT]
> **UUID Conversion:** Convert `domain.ID` ↔ `uuid.UUID` at the repository boundary. Domain layer never sees `uuid.UUID`.

### 3.3 Add Compile-Time Check

Ensure the repository implements the interface at compile time:

```go
// Ensure OrderRepo implements domain.OrderRepository at compile time.
var _ domain.OrderRepository = (*OrderRepo)(nil)
```

### Error Wrapping Pattern

Always wrap errors with an operation string (`op`) for debugging:

```go
const op = "orderRepo.Create"
// ...
return fmt.Errorf("%s: %w", op, err)
```

**Reference:** [internal/infra/postgres/user_repo.go](../../internal/infra/postgres/user_repo.go)

---

## Step 4: Create Use Cases

The application layer contains business logic. It **cannot** import transport, infra, logging, or external packages.

### 4.1 Create Use Case Package

Create the directory structure:

```
internal/app/order/
├── create_order.go
├── get_order.go
└── list_orders.go
```

### 4.2 Implement Create Use Case

Create `internal/app/order/create_order.go`:

```go
// Package order provides use cases for order-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package order

import (
    "context"
    "errors"
    "time"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/app/audit"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateOrderRequest represents the input data for creating a new order.
type CreateOrderRequest struct {
    ID          domain.ID
    CustomerID  domain.ID
    Status      string
    TotalAmount int64
    // RequestID correlates this operation with the HTTP request.
    // Transport layer extracts from context and passes here.
    RequestID string
    // ActorID identifies who is performing this action.
    // Transport layer extracts from JWT claims and passes here.
    ActorID domain.ID
}

// CreateOrderResponse represents the result of creating a new order.
type CreateOrderResponse struct {
    Order domain.Order
}

// CreateOrderUseCase handles the business logic for creating a new order.
type CreateOrderUseCase struct {
    orderRepo    domain.OrderRepository
    auditService *audit.AuditService
    idGen        domain.IDGenerator
    txManager    domain.TxManager
    db           domain.Querier
}

// NewCreateOrderUseCase creates a new instance of CreateOrderUseCase.
func NewCreateOrderUseCase(
    orderRepo domain.OrderRepository,
    auditService *audit.AuditService,
    idGen domain.IDGenerator,
    txManager domain.TxManager,
    db domain.Querier,
) *CreateOrderUseCase {
    return &CreateOrderUseCase{
        orderRepo:    orderRepo,
        auditService: auditService,
        idGen:        idGen,
        txManager:    txManager,
        db:           db,
    }
}
```

### 4.3 Implement Execute Method

```go
// Execute processes the create order request.
// It validates the input and creates the order.
// Returns AppError with appropriate Code for domain errors.
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (CreateOrderResponse, error) {
    // Create a new order entity with generated ID
    id := req.ID
    if id.IsEmpty() {
        id = uc.idGen.NewID()
    }
    now := time.Now().UTC()
    order := &domain.Order{
        ID:          id,
        CustomerID:  req.CustomerID,
        Status:      req.Status,
        TotalAmount: req.TotalAmount,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // Validate the order entity using domain rules
    if err := order.Validate(); err != nil {
        return CreateOrderResponse{}, &app.AppError{
            Op:      "CreateOrder",
            Code:    app.CodeValidationError,
            Message: "Validation failed",
            Err:     err,
        }
    }

    // Execute logic within a transaction
    if err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
        // Create the order in the repository
        if err := uc.orderRepo.Create(ctx, tx, order); err != nil {
            return &app.AppError{
                Op:      "CreateOrder",
                Code:    app.CodeInternalError,
                Message: "Failed to create order",
                Err:     err,
            }
        }

        // Record audit event (same transaction context)
        auditInput := audit.AuditEventInput{
            EventType:  domain.EventOrderCreated,
            ActorID:    req.ActorID,
            EntityType: "order",
            EntityID:   order.ID,
            Payload:    order,
            RequestID:  req.RequestID,
        }

        if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
            return &app.AppError{
                Op:      "CreateOrder",
                Code:    app.CodeInternalError,
                Message: "Failed to record audit event",
                Err:     err,
            }
        }

        return nil
    }); err != nil {
        return CreateOrderResponse{}, err
    }

    return CreateOrderResponse{Order: *order}, nil
}
```

> [!NOTE]
> **No logging in app layer!** Use tracing context for observability. Logging is handled at transport and infra layers only.

### Transaction Pattern

Always use `TxManager.WithTx()` for operations that need atomicity:

```go
err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
    // All repository calls use `tx` (transaction)
    if err := uc.repo.Create(ctx, tx, entity); err != nil {
        return err
    }
    return uc.auditService.Record(ctx, tx, auditInput)
})
```

**Reference:** [internal/app/user/create_user.go](../../internal/app/user/create_user.go)

---

## Step 5: Create DTOs/Contracts

The transport layer defines HTTP request/response structures.

### 5.1 Request Structs

Create `internal/transport/http/contract/order.go`:

```go
package contract

import (
    "time"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateOrderRequest represents the HTTP body for creating an order.
type CreateOrderRequest struct {
    CustomerID  string `json:"customerId" validate:"required,uuid"`
    Status      string `json:"status" validate:"required,oneof=pending confirmed shipped delivered"`
    TotalAmount int64  `json:"totalAmount" validate:"required,min=0"`
}
```

> [!TIP]
> Use `validate` struct tags for request validation. Common validators: `required`, `email`, `min`, `max`, `oneof`, `uuid`.

### 5.2 Response Structs

```go
// OrderResponse represents an order in HTTP responses.
type OrderResponse struct {
    ID          string    `json:"id"`
    CustomerID  string    `json:"customerId"`
    Status      string    `json:"status"`
    TotalAmount int64     `json:"totalAmount"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

### 5.3 Mapper Functions

```go
// ToOrderResponse converts a domain.Order into a response DTO.
func ToOrderResponse(o domain.Order) OrderResponse {
    return OrderResponse{
        ID:          string(o.ID),
        CustomerID:  string(o.CustomerID),
        Status:      o.Status,
        TotalAmount: o.TotalAmount,
        CreatedAt:   o.CreatedAt,
        UpdatedAt:   o.UpdatedAt,
    }
}

// ToOrderResponses converts a slice of domain.Order into []OrderResponse.
func ToOrderResponses(orders []domain.Order) []OrderResponse {
    responses := make([]OrderResponse, len(orders))
    for i, o := range orders {
        responses[i] = ToOrderResponse(o)
    }
    return responses
}
```

### 5.4 Pagination Response (if applicable)

```go
// ListOrdersResponse represents the list orders response body.
type ListOrdersResponse struct {
    Data       []OrderResponse    `json:"data"`
    Pagination PaginationResponse `json:"pagination"`
}

// NewListOrdersResponse creates a list response from domain data.
func NewListOrdersResponse(orders []domain.Order, page, pageSize, totalItems int) ListOrdersResponse {
    return ListOrdersResponse{
        Data:       ToOrderResponses(orders),
        Pagination: NewPaginationResponse(page, pageSize, totalItems),
    }
}
```

**Reference:** [internal/transport/http/contract/user.go](../../internal/transport/http/contract/user.go)

---

## Step 6: Create HTTP Handlers

Handlers bridge HTTP requests to use case execution.

### 6.1 Handler Struct

Create `internal/transport/http/handler/order.go`:

```go
// Package handler provides HTTP handlers for the API.
package handler

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/app/order"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

type createOrderExecutor interface {
    Execute(ctx context.Context, req order.CreateOrderRequest) (order.CreateOrderResponse, error)
}

type getOrderExecutor interface {
    Execute(ctx context.Context, req order.GetOrderRequest) (order.GetOrderResponse, error)
}

type listOrdersExecutor interface {
    Execute(ctx context.Context, req order.ListOrdersRequest) (order.ListOrdersResponse, error)
}

// OrderHandler handles order-related HTTP requests.
type OrderHandler struct {
    createUC createOrderExecutor
    getUC    getOrderExecutor
    listUC   listOrdersExecutor
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(
    createUC createOrderExecutor,
    getUC getOrderExecutor,
    listUC listOrdersExecutor,
) *OrderHandler {
    return &OrderHandler{
        createUC: createUC,
        getUC:    getUC,
        listUC:   listUC,
    }
}
```

### 6.2 Implement Handler Methods

```go
// CreateOrder handles POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req contract.CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        contract.WriteProblemJSON(w, r, &app.AppError{
            Op:      "CreateOrder",
            Code:    app.CodeValidationError,
            Message: "Invalid request body",
            Err:     err,
        })
        return
    }

    // Validate request
    if errs := contract.Validate(req); len(errs) > 0 {
        contract.WriteValidationError(w, r, errs)
        return
    }

    // Generate UUID v7 at transport boundary
    id, err := uuid.NewV7()
    if err != nil {
        contract.WriteProblemJSON(w, r, &app.AppError{
            Op:      "CreateOrder",
            Code:    app.CodeInternalError,
            Message: "Failed to generate order ID",
            Err:     err,
        })
        return
    }

    // Map to app layer request
    appReq := order.CreateOrderRequest{
        ID:          domain.ID(id.String()),
        CustomerID:  domain.ID(req.CustomerID),
        Status:      req.Status,
        TotalAmount: req.TotalAmount,
    }

    // Execute use case
    resp, err := h.createUC.Execute(r.Context(), appReq)
    if err != nil {
        contract.WriteProblemJSON(w, r, err)
        return
    }

    // Map to response
    orderResp := contract.ToOrderResponse(resp.Order)
    _ = contract.WriteJSON(w, http.StatusCreated, contract.DataResponse[contract.OrderResponse]{Data: orderResp})
}
```

### 6.3 GetOrder Handler

```go
// GetOrder handles GET /api/v1/orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    idParam := chi.URLParam(r, "id")

    // Validate UUID format
    parsedID, err := uuid.Parse(idParam)
    if err != nil || parsedID.Version() != 7 {
        contract.WriteProblemJSON(w, r, &app.AppError{
            Op:      "GetOrder",
            Code:    app.CodeValidationError,
            Message: "Invalid order ID format",
            Err:     err,
        })
        return
    }

    // Execute use case
    resp, err := h.getUC.Execute(r.Context(), order.GetOrderRequest{ID: domain.ID(idParam)})
    if err != nil {
        contract.WriteProblemJSON(w, r, err)
        return
    }

    // Map to response
    orderResp := contract.ToOrderResponse(resp.Order)
    _ = contract.WriteJSON(w, http.StatusOK, contract.DataResponse[contract.OrderResponse]{Data: orderResp})
}
```

> [!IMPORTANT]
> **Error Mapping:** Use `contract.WriteProblemJSON()` to convert `AppError` to RFC 7807 Problem Details JSON.

**Reference:** [internal/transport/http/handler/user.go](../../internal/transport/http/handler/user.go)

---

## Step 7: Wire Routes in Router

### 7.1 Define Routes Interface

Add interface in `internal/transport/http/router.go`:

```go
// OrderRoutes defines the interface for order-related HTTP handlers.
// This interface breaks the import cycle between http and handler packages.
type OrderRoutes interface {
    CreateOrder(w stdhttp.ResponseWriter, r *stdhttp.Request)
    GetOrder(w stdhttp.ResponseWriter, r *stdhttp.Request)
    ListOrders(w stdhttp.ResponseWriter, r *stdhttp.Request)
}
```

### 7.2 Update Router Function Signature

Add the handler parameter to `NewRouter()`:

```go
func NewRouter(
    logger *slog.Logger,
    tracingEnabled bool,
    metricsReg *prometheus.Registry,
    httpMetrics metrics.HTTPMetrics,
    healthHandler, readyHandler stdhttp.Handler,
    userHandler UserRoutes,
    orderHandler OrderRoutes,  // Add this
    maxRequestSize int64,
    jwtConfig JWTConfig,
    rateLimitConfig RateLimitConfig,
) chi.Router {
```

### 7.3 Register Routes

Add route registrations inside the `/api/v1` group:

```go
r.Route("/api/v1", func(r chi.Router) {
    // ... JWT and rate limiting middleware ...

    // User routes
    if userHandler != nil {
        r.Post("/users", userHandler.CreateUser)
        r.Get("/users/{id}", userHandler.GetUser)
        r.Get("/users", userHandler.ListUsers)
    }

    // Order routes
    if orderHandler != nil {
        r.Post("/orders", orderHandler.CreateOrder)
        r.Get("/orders/{id}", orderHandler.GetOrder)
        r.Get("/orders", orderHandler.ListOrders)
    }
})
```

### 7.4 Update main.go

Wire the new handler in `cmd/api/main.go`:

```go
// Create repositories
orderRepo := postgres.NewOrderRepo()

// Create use cases
createOrderUC := order.NewCreateOrderUseCase(orderRepo, auditService, idGen, txManager, pool)
getOrderUC := order.NewGetOrderUseCase(orderRepo, pool)
listOrdersUC := order.NewListOrdersUseCase(orderRepo, pool)

// Create handlers
orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, listOrdersUC)

// Create router with all handlers
router := http.NewRouter(
    logger,
    tracingEnabled,
    metricsReg,
    httpMetrics,
    healthHandler,
    readyHandler,
    userHandler,
    orderHandler,  // Add here
    cfg.MaxRequestSize,
    jwtConfig,
    rateLimitConfig,
)
```

**Reference:** [internal/transport/http/router.go](../../internal/transport/http/router.go)

---

## Step 8: Add Audit Events

See [Adding Audit Events Guide](./adding-audit-events.md) for detailed instructions.

### 8.1 Define Event Type

Add to `internal/domain/audit.go`:

```go
const EventOrderCreated = "order.created"
const EventOrderUpdated = "order.updated"
const EventOrderDeleted = "order.deleted"
```

### 8.2 Record Events

Audit events are recorded within transactions in use case methods (shown in Step 4).

```go
auditInput := audit.AuditEventInput{
    EventType:  domain.EventOrderCreated,
    ActorID:    req.ActorID,
    EntityType: "order",
    EntityID:   order.ID,
    Payload:    order,  // Automatically PII-redacted
    RequestID:  req.RequestID,
}

if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
    return &app.AppError{...}
}
```

**Reference:** [internal/app/user/create_user.go](../../internal/app/user/create_user.go) (lines 110-128)

---

## Step 9: Write Tests

### 9.1 Domain Tests

Test entity validation in `internal/domain/order_test.go`:

```go
package domain

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestOrder_Validate(t *testing.T) {
    tests := []struct {
        name    string
        order   Order
        wantErr error
    }{
        {
            name:    "valid order",
            order:   Order{ID: "123", CustomerID: "456", Status: "pending", TotalAmount: 1000},
            wantErr: nil,
        },
        {
            name:    "empty ID",
            order:   Order{ID: "", CustomerID: "456", Status: "pending", TotalAmount: 1000},
            wantErr: ErrInvalidID,
        },
        {
            name:    "negative amount",
            order:   Order{ID: "123", CustomerID: "456", Status: "pending", TotalAmount: -100},
            wantErr: ErrInvalidOrderAmount,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.order.Validate()
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Coverage Target:** 100% for domain entities

### 9.2 Use Case Tests

Test business logic in `internal/app/order/create_order_test.go`:

```go
func TestCreateOrderUseCase_Execute(t *testing.T) {
    tests := []struct {
        name      string
        req       CreateOrderRequest
        setupMock func(*mockOrderRepo, *mockTxManager)
        wantErr   bool
        wantCode  string
    }{
        {
            name: "success",
            req: CreateOrderRequest{
                ID:          "valid-id",
                CustomerID:  "customer-id",
                Status:      "pending",
                TotalAmount: 1000,
            },
            setupMock: func(repo *mockOrderRepo, tm *mockTxManager) {
                tm.execFunc = func(fn func(domain.Querier) error) error {
                    return fn(nil)
                }
                repo.createErr = nil
            },
            wantErr: false,
        },
        {
            name: "validation error - negative amount",
            req: CreateOrderRequest{
                ID:          "valid-id",
                CustomerID:  "customer-id",
                Status:      "pending",
                TotalAmount: -100,
            },
            wantErr:  true,
            wantCode: app.CodeValidationError,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... setup and execute ...
        })
    }
}
```

**Coverage Target:** 90% for use cases

### 9.3 Handler Tests

Test HTTP handlers with testify + httptest:

```go
func TestOrderHandler_CreateOrder(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        setupMock  func(*mockCreateOrderExecutor)
        wantStatus int
    }{
        {
            name: "success",
            body: `{"customerId":"uuid-here","status":"pending","totalAmount":1000}`,
            setupMock: func(m *mockCreateOrderExecutor) {
                m.result = order.CreateOrderResponse{Order: domain.Order{ID: "new-id"}}
            },
            wantStatus: http.StatusCreated,
        },
        {
            name:       "invalid JSON",
            body:       `{invalid`,
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(tt.body))
            rec := httptest.NewRecorder()

            // ... setup handler and execute ...

            assert.Equal(t, tt.wantStatus, rec.Code)
        })
    }
}
```

### 9.4 Repository Integration Tests

Use testcontainers for database integration tests:

```go
//go:build integration

package postgres

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    // ... testcontainers imports ...
)

func TestOrderRepo_Create_Integration(t *testing.T) {
    ctx := context.Background()
    pool := setupTestDB(t) // Uses testcontainers
    defer pool.Close()

    repo := NewOrderRepo()
    order := &domain.Order{
        ID:          "test-id",
        CustomerID:  "customer-id",
        Status:      "pending",
        TotalAmount: 1000,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    err := repo.Create(ctx, pool, order)
    require.NoError(t, err)

    // Verify retrieval
    found, err := repo.GetByID(ctx, pool, order.ID)
    require.NoError(t, err)
    assert.Equal(t, order.Status, found.Status)
}
```

### Run Tests

```bash
# Unit tests
make test

# Integration tests (requires Docker)
go test -tags=integration ./internal/infra/postgres/...

# Coverage
make coverage
```

---

## Quick Reference Checklist

### Files to Create/Modify

| Layer | File | Action |
|-------|------|--------|
| Domain | `internal/domain/order.go` | **NEW** - Entity + repository interface |
| Domain | `internal/domain/errors.go` | **MODIFY** - Add domain errors |
| Domain | `internal/domain/audit.go` | **MODIFY** - Add event types |
| Migration | `migrations/*_create_orders.sql` | **NEW** - Database schema |
| Infra | `internal/infra/postgres/order_repo.go` | **NEW** - Repository implementation |
| App | `internal/app/order/create_order.go` | **NEW** - Create use case |
| App | `internal/app/order/get_order.go` | **NEW** - Get use case |
| App | `internal/app/order/list_orders.go` | **NEW** - List use case |
| Transport | `internal/transport/http/contract/order.go` | **NEW** - DTOs |
| Transport | `internal/transport/http/handler/order.go` | **NEW** - HTTP handlers |
| Transport | `internal/transport/http/router.go` | **MODIFY** - Add routes |
| Cmd | `cmd/api/main.go` | **MODIFY** - Wire dependencies |
| Tests | `internal/domain/order_test.go` | **NEW** - Domain tests |
| Tests | `internal/app/order/*_test.go` | **NEW** - Use case tests |

### Import Rules Quick Reference

| Layer | ✅ Can Import | ❌ CANNOT Import |
|-------|--------------|------------------|
| **Domain** | `$gostd` only | `slog`, `uuid`, `pgx`, `otel`, ANY external |
| **App** | `$gostd`, `internal/domain` | `slog`, `otel`, `uuid`, `net/http`, `pgx`, `transport`, `infra` |
| **Transport** | `domain`, `app`, `chi`, `uuid`, `stdlib`, `otel` | `pgx`, `internal/infra` |
| **Infra** | `domain`, `pgx`, `slog`, `otel`, `uuid`, everything | `app`, `transport` |

### Common Commands

```bash
# Create migration
make migrate-create name=create_orders

# Run migrations
make migrate

# Run tests
make test

# Run linter
make lint

# Local CI (full check)
make ci

# Verify file paths
ls internal/domain/order.go
ls internal/infra/postgres/order_repo.go
ls internal/app/order/
ls internal/transport/http/contract/order.go
ls internal/transport/http/handler/order.go
```

### Key Patterns Summary

1. **UUID Handling:** Generate UUID v7 at transport boundary, parse at infra boundary
2. **Error Flow:** Domain sentinel → Infra wrap with op → App `AppError` → Transport HTTP status
3. **Transactions:** Always use `TxManager.WithTx()` for multi-step operations
4. **Audit:** Record within transaction using `auditService.Record()`
5. **Validation:** Domain `Validate()` for business rules, struct tags for HTTP request validation

---

**Last Updated:** 2025-12-23

**Reference Implementation:** Users module in `internal/domain/user.go`, `internal/app/user/`, `internal/transport/http/handler/user.go`

// Package user provides use cases for user-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package user

import (
	"context"
	"errors"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/audit"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateUserRequest represents the input data for creating a new user.
type CreateUserRequest struct {
	ID        domain.ID
	FirstName string
	LastName  string
	Email     string
	// RequestID correlates this operation with the HTTP request.
	// Transport layer extracts from context and passes here.
	RequestID string
	// ActorID identifies who is performing this action.
	// Transport layer extracts from JWT claims and passes here.
	ActorID domain.ID
}

// CreateUserResponse represents the result of creating a new user.
type CreateUserResponse struct {
	User domain.User
}

// CreateUserUseCase handles the business logic for creating a new user.
type CreateUserUseCase struct {
	userRepo     domain.UserRepository
	auditService *audit.AuditService
	idGen        domain.IDGenerator
	txManager    domain.TxManager
	db           domain.Querier
}

// NewCreateUserUseCase creates a new instance of CreateUserUseCase.
func NewCreateUserUseCase(
	userRepo domain.UserRepository,
	auditService *audit.AuditService,
	idGen domain.IDGenerator,
	txManager domain.TxManager,
	db domain.Querier,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:     userRepo,
		auditService: auditService,
		idGen:        idGen,
		txManager:    txManager,
		db:           db,
	}
}

// Execute processes the create user request.
// It validates the input and creates the user.
// Returns AppError with appropriate Code for domain errors.
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
	// Create a new user entity with generated ID
	id := req.ID
	if id.IsEmpty() {
		id = uc.idGen.NewID()
	}
	now := time.Now().UTC()
	user := &domain.User{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate the user entity using domain rules
	if err := user.Validate(); err != nil {
		return CreateUserResponse{}, &app.AppError{
			Op:      "CreateUser",
			Code:    app.CodeValidationError,
			Message: "Validation failed",
			Err:     err,
		}
	}

	// Execute logic within a transaction
	if err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
		// Create the user in the repository
		if err := uc.userRepo.Create(ctx, tx, user); err != nil {
			if errors.Is(err, domain.ErrEmailAlreadyExists) {
				return &app.AppError{
					Op:      "CreateUser",
					Code:    app.CodeEmailExists,
					Message: "Email already exists",
					Err:     err,
				}
			}
			return &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeInternalError,
				Message: "Failed to create user",
				Err:     err,
			}
		}

		// Record audit event (same transaction context)
		// RequestID and ActorID come from request struct (passed by transport layer)
		auditInput := audit.AuditEventInput{
			EventType:  domain.EventUserCreated,
			ActorID:    req.ActorID,
			EntityType: "user",
			EntityID:   user.ID,
			Payload:    user,
			RequestID:  req.RequestID,
		}

		if err := uc.auditService.Record(ctx, tx, auditInput); err != nil {
			return &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeInternalError,
				Message: "Failed to record audit event",
				Err:     err,
			}
		}

		return nil
	}); err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{User: *user}, nil
}

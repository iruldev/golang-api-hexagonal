// Package user provides use cases for user-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package user

import (
	"context"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateUserRequest represents the input data for creating a new user.
type CreateUserRequest struct {
	FirstName string
	LastName  string
	Email     string
}

// CreateUserResponse represents the result of creating a new user.
type CreateUserResponse struct {
	User domain.User
}

// CreateUserUseCase handles the business logic for creating a new user.
type CreateUserUseCase struct {
	userRepo domain.UserRepository
	idGen    domain.IDGenerator
	db       domain.Querier
}

// NewCreateUserUseCase creates a new instance of CreateUserUseCase.
func NewCreateUserUseCase(userRepo domain.UserRepository, idGen domain.IDGenerator, db domain.Querier) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
		idGen:    idGen,
		db:       db,
	}
}

// Execute processes the create user request.
// It validates the input and creates the user.
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
	// Create a new user entity with generated ID
	user := &domain.User{
		ID:        uc.idGen.NewID(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	// Validate the user entity using domain rules
	if err := user.Validate(); err != nil {
		return CreateUserResponse{}, err
	}

	// Create the user in the repository
	if err := uc.userRepo.Create(ctx, uc.db, user); err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{User: *user}, nil
}

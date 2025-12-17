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
	Name  string
	Email string
}

// CreateUserResponse represents the result of creating a new user.
type CreateUserResponse struct {
	User domain.User
}

// CreateUserUseCase handles the business logic for creating a new user.
type CreateUserUseCase struct {
	userRepo domain.UserRepository
}

// NewCreateUserUseCase creates a new instance of CreateUserUseCase.
func NewCreateUserUseCase(userRepo domain.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
	}
}

// Execute processes the create user request.
// It validates the input, checks for existing email, and creates the user.
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
	// Create a new user entity
	user := domain.User{
		Name:  req.Name,
		Email: req.Email,
	}

	// Validate the user entity using domain rules
	if err := user.Validate(); err != nil {
		return CreateUserResponse{}, err
	}

	// Check if email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return CreateUserResponse{}, err
	}
	if !existingUser.ID.IsEmpty() {
		return CreateUserResponse{}, domain.ErrEmailExists
	}

	// Create the user in the repository
	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{User: createdUser}, nil
}

// Package user provides use cases for user-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package user

import (
	"context"
	"errors"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// GetUserRequest represents the input data for getting a user by ID.
type GetUserRequest struct {
	ID domain.ID
}

// GetUserResponse represents the result of getting a user.
type GetUserResponse struct {
	User domain.User
}

// GetUserUseCase handles the business logic for retrieving a user by ID.
type GetUserUseCase struct {
	userRepo domain.UserRepository
	db       domain.Querier
}

// NewGetUserUseCase creates a new instance of GetUserUseCase.
func NewGetUserUseCase(userRepo domain.UserRepository, db domain.Querier) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		db:       db,
	}
}

// Execute retrieves a user by ID.
// Returns AppError with Code=USER_NOT_FOUND if user doesn't exist.
func (uc *GetUserUseCase) Execute(ctx context.Context, req GetUserRequest) (GetUserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, uc.db, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return GetUserResponse{}, &app.AppError{
				Op:      "GetUser",
				Code:    app.CodeUserNotFound,
				Message: "User not found",
				Err:     err,
			}
		}
		return GetUserResponse{}, &app.AppError{
			Op:      "GetUser",
			Code:    app.CodeInternalError,
			Message: "Failed to get user",
			Err:     err,
		}
	}
	if user == nil {
		return GetUserResponse{}, &app.AppError{
			Op:      "GetUser",
			Code:    app.CodeInternalError,
			Message: "Failed to get user",
			Err:     errors.New("user repository returned nil user without error"),
		}
	}
	return GetUserResponse{User: *user}, nil
}

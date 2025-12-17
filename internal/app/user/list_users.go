// Package user provides use cases for user-related operations.
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
package user

import (
	"context"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// ListUsersRequest represents the input data for listing users with pagination.
type ListUsersRequest struct {
	Page     int
	PageSize int
}

// ListUsersResponse represents the result of listing users.
type ListUsersResponse struct {
	Users      []domain.User
	TotalCount int
	Page       int
	PageSize   int
}

// ListUsersUseCase handles the business logic for listing users with pagination.
type ListUsersUseCase struct {
	userRepo domain.UserRepository
	db       domain.Querier
}

// NewListUsersUseCase creates a new instance of ListUsersUseCase.
func NewListUsersUseCase(userRepo domain.UserRepository, db domain.Querier) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
		db:       db,
	}
}

// Execute lists users with pagination support.
// Returns paginated user list with total count for UI pagination.
func (uc *ListUsersUseCase) Execute(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error) {
	params := domain.ListParams{Page: req.Page, PageSize: req.PageSize}
	if params.Page <= 0 {
		params.Page = 1
	}

	users, totalCount, err := uc.userRepo.List(ctx, uc.db, params)
	if err != nil {
		return ListUsersResponse{}, &app.AppError{
			Op:      "ListUsers",
			Code:    app.CodeInternalError,
			Message: "Failed to list users",
			Err:     err,
		}
	}

	return ListUsersResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.Limit(),
	}, nil
}

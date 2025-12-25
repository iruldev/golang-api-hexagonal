package contract

import (
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// CreateUserRequest represents the HTTP body for creating a user.
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	FirstName string `json:"firstName" validate:"required,min=1,max=100"`
	LastName  string `json:"lastName" validate:"required,min=1,max=100"`
}

// UserResponse represents a user in HTTP responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToUserResponse converts a domain.User into a response DTO.
func ToUserResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:        string(u.ID),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToUserResponses converts a slice of domain.User into []UserResponse.
func ToUserResponses(users []domain.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = ToUserResponse(u)
	}
	return responses
}

// PaginationResponse represents pagination metadata in list responses.
type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// NewPaginationResponse builds pagination metadata with calculated total pages.
func NewPaginationResponse(page, pageSize, totalItems int) PaginationResponse {
	totalPages := 0
	if pageSize > 0 && totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}

	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// ListUsersResponse represents the list users response body.
type ListUsersResponse struct {
	Data       []UserResponse     `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// NewListUsersResponse creates a list response from domain data.
func NewListUsersResponse(users []domain.User, page, pageSize, totalItems int) ListUsersResponse {
	return ListUsersResponse{
		Data:       ToUserResponses(users),
		Pagination: NewPaginationResponse(page, pageSize, totalItems),
	}
}

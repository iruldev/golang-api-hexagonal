package handler

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
)

// MockCreateUserUseCase mocks the CreateUserUseCase.
type MockCreateUserUseCase struct {
	mock.Mock
}

func (m *MockCreateUserUseCase) Execute(ctx context.Context, req user.CreateUserRequest) (user.CreateUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(user.CreateUserResponse), args.Error(1)
}

// MockGetUserUseCase mocks the GetUserUseCase.
type MockGetUserUseCase struct {
	mock.Mock
}

func (m *MockGetUserUseCase) Execute(ctx context.Context, req user.GetUserRequest) (user.GetUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(user.GetUserResponse), args.Error(1)
}

// MockListUsersUseCase mocks the ListUsersUseCase.
type MockListUsersUseCase struct {
	mock.Mock
}

func (m *MockListUsersUseCase) Execute(ctx context.Context, req user.ListUsersRequest) (user.ListUsersResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(user.ListUsersResponse), args.Error(1)
}

// Helpers for creating test users.
var testUserResourcePath = httpTransport.BasePath + "/users"

func createTestUser() domain.User {
	return domain.User{
		ID:        domain.ID("019400a0-1234-7abc-8def-1234567890ab"),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

//go:build !integration

package user

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// mockIDGenerator is a test double for domain.IDGenerator.
type mockIDGenerator struct {
	nextID int
}

func newMockIDGenerator() *mockIDGenerator {
	return &mockIDGenerator{nextID: 1}
}

func (m *mockIDGenerator) NewID() domain.ID {
	id := domain.ID("user-" + strconv.Itoa(m.nextID))
	m.nextID++
	return id
}

// mockQuerier is a test double for domain.Querier.
type mockQuerier struct{}

func (m *mockQuerier) Exec(_ context.Context, _ string, _ ...any) (any, error)  { return nil, nil }
func (m *mockQuerier) Query(_ context.Context, _ string, _ ...any) (any, error) { return nil, nil }
func (m *mockQuerier) QueryRow(_ context.Context, _ string, _ ...any) any       { return nil }

// mockUserRepository is a test double for domain.UserRepository.
type mockUserRepository struct {
	users       map[domain.ID]domain.User
	createError error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[domain.ID]domain.User),
	}
}

func (m *mockUserRepository) Create(_ context.Context, _ domain.Querier, user *domain.User) error {
	if m.createError != nil {
		return m.createError
	}

	now := time.Unix(0, 0).UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users[user.ID] = *user
	return nil
}

func (m *mockUserRepository) GetByID(_ context.Context, _ domain.Querier, id domain.ID) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

func (m *mockUserRepository) List(_ context.Context, _ domain.Querier, _ domain.ListParams) ([]domain.User, int, error) {
	users := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, len(users), nil
}

func TestCreateUserUseCase_Execute(t *testing.T) {
	repoErr := errors.New("database error")

	tests := []struct {
		name       string
		req        CreateUserRequest
		setupMock  func(*mockUserRepository)
		wantCode   string
		wantErr    error
		wantUserID bool
	}{
		{
			name: "successfully creates a new user",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantErr:    nil,
			wantUserID: true,
		},
		{
			name: "fails with invalid email - returns AppError with VALIDATION_ERROR code",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantCode:   app.CodeValidationError,
			wantErr:    domain.ErrInvalidEmail,
			wantUserID: false,
		},
		{
			name: "fails with whitespace-only email - returns AppError with VALIDATION_ERROR code",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "   ",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantCode:   app.CodeValidationError,
			wantErr:    domain.ErrInvalidEmail,
			wantUserID: false,
		},
		{
			name: "fails with invalid first name - returns AppError with VALIDATION_ERROR code",
			req: CreateUserRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantCode:   app.CodeValidationError,
			wantErr:    domain.ErrInvalidFirstName,
			wantUserID: false,
		},
		{
			name: "fails with invalid last name - returns AppError with VALIDATION_ERROR code",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "",
				Email:     "john@example.com",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantCode:   app.CodeValidationError,
			wantErr:    domain.ErrInvalidLastName,
			wantUserID: false,
		},
		{
			name: "fails with email already exists - returns AppError with EMAIL_EXISTS code",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.createError = domain.ErrEmailAlreadyExists
			},
			wantCode:   app.CodeEmailExists,
			wantErr:    domain.ErrEmailAlreadyExists,
			wantUserID: false,
		},
		{
			name: "propagates repository create error - returns AppError with INTERNAL_ERROR code",
			req: CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.createError = repoErr
			},
			wantCode:   app.CodeInternalError,
			wantErr:    repoErr,
			wantUserID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			mockIDGen := newMockIDGenerator()
			mockDB := &mockQuerier{}
			tt.setupMock(mockRepo)

			useCase := NewCreateUserUseCase(mockRepo, mockIDGen, mockDB)
			resp, err := useCase.Execute(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				// Verify error is wrapped in AppError
				var appErr *app.AppError
				require.True(t, errors.As(err, &appErr), "expected AppError, got %T", err)
				assert.Equal(t, tt.wantCode, appErr.Code)
				// Verify underlying error is preserved
				assert.True(t, errors.Is(err, tt.wantErr), "expected wrapped error %v, got %v", tt.wantErr, appErr.Err)
				assert.True(t, resp.User.ID.IsEmpty())
			} else {
				require.NoError(t, err)
				if tt.wantUserID {
					assert.False(t, resp.User.ID.IsEmpty())
					assert.Equal(t, tt.req.FirstName, resp.User.FirstName)
					assert.Equal(t, tt.req.LastName, resp.User.LastName)
					assert.Equal(t, tt.req.Email, resp.User.Email)
				}
			}
		})
	}
}

func TestNewCreateUserUseCase(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockIDGen := newMockIDGenerator()
	mockDB := &mockQuerier{}
	useCase := NewCreateUserUseCase(mockRepo, mockIDGen, mockDB)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.userRepo)
	assert.Equal(t, mockIDGen, useCase.idGen)
	assert.Equal(t, mockDB, useCase.db)
}

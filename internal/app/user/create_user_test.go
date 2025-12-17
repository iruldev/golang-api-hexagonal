package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// mockUserRepository is a test double for domain.UserRepository.
// It allows us to control repository behavior in tests without database access.
type mockUserRepository struct {
	users       map[domain.ID]domain.User
	emailIndex  map[string]domain.ID
	nextID      int
	createError error
	getError    error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:      make(map[domain.ID]domain.User),
		emailIndex: make(map[string]domain.ID),
		nextID:     1,
	}
}

func (m *mockUserRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	if m.createError != nil {
		return domain.User{}, m.createError
	}

	user.ID = domain.ID(fmt.Sprintf("user-%d", m.nextID))
	m.nextID++

	now := time.Unix(0, 0).UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users[user.ID] = user
	m.emailIndex[user.Email] = user.ID

	return user, nil
}

func (m *mockUserRepository) GetByID(_ context.Context, id domain.ID) (domain.User, error) {
	if m.getError != nil {
		return domain.User{}, m.getError
	}

	user, exists := m.users[id]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(_ context.Context, email string) (domain.User, error) {
	if m.getError != nil {
		return domain.User{}, m.getError
	}

	id, exists := m.emailIndex[email]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}
	return m.users[id], nil
}

func (m *mockUserRepository) Update(_ context.Context, _ domain.User) error {
	return nil
}

func (m *mockUserRepository) Delete(_ context.Context, _ domain.ID) error {
	return nil
}

// addExistingUser adds a user directly to the mock repository for testing.
func (m *mockUserRepository) addExistingUser(user domain.User) {
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user.ID
}

func TestCreateUserUseCase_Execute(t *testing.T) {
	repoErr := assert.AnError

	tests := []struct {
		name       string
		req        CreateUserRequest
		setupMock  func(*mockUserRepository)
		wantErr    error
		wantUserID bool
	}{
		{
			name: "successfully creates a new user",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantErr:    nil,
			wantUserID: true,
		},
		{
			name: "fails with invalid email",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantErr:    domain.ErrInvalidEmail,
			wantUserID: false,
		},
		{
			name: "fails with whitespace-only email",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "   ",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantErr:    domain.ErrInvalidEmail,
			wantUserID: false,
		},
		{
			name: "fails with invalid name",
			req: CreateUserRequest{
				Name:  "",
				Email: "john@example.com",
			},
			setupMock:  func(_ *mockUserRepository) {},
			wantErr:    domain.ErrInvalidUserName,
			wantUserID: false,
		},
		{
			name: "fails when email already exists",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "existing@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.addExistingUser(domain.User{
					ID:    domain.ID("existing-user"),
					Name:  "Existing User",
					Email: "existing@example.com",
				})
			},
			wantErr:    domain.ErrEmailExists,
			wantUserID: false,
		},
		{
			name: "propagates repository get error (non-not-found)",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.getError = repoErr
			},
			wantErr:    repoErr,
			wantUserID: false,
		},
		{
			name: "propagates repository create error",
			req: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.createError = repoErr
			},
			wantErr:    repoErr,
			wantUserID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			tt.setupMock(mockRepo)

			useCase := NewCreateUserUseCase(mockRepo)
			resp, err := useCase.Execute(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.True(t, resp.User.ID.IsEmpty())
			} else {
				assert.NoError(t, err)
				if tt.wantUserID {
					assert.False(t, resp.User.ID.IsEmpty())
					assert.Equal(t, tt.req.Name, resp.User.Name)
					assert.Equal(t, tt.req.Email, resp.User.Email)
				}
			}
		})
	}
}

func TestNewCreateUserUseCase(t *testing.T) {
	mockRepo := newMockUserRepository()
	useCase := NewCreateUserUseCase(mockRepo)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.userRepo)
}

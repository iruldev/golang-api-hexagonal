//go:build !integration

package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// mockQuerierGetUser is a test double for domain.Querier.
type mockQuerierGetUser struct{}

func (m *mockQuerierGetUser) Exec(_ context.Context, _ string, _ ...any) (any, error) {
	return nil, nil
}
func (m *mockQuerierGetUser) Query(_ context.Context, _ string, _ ...any) (any, error) {
	return nil, nil
}
func (m *mockQuerierGetUser) QueryRow(_ context.Context, _ string, _ ...any) any { return nil }

// mockUserRepositoryGetUser is a test double for domain.UserRepository.
type mockUserRepositoryGetUser struct {
	users        map[domain.ID]domain.User
	getByIDError error
	returnNil    bool
}

func newMockUserRepositoryGetUser() *mockUserRepositoryGetUser {
	return &mockUserRepositoryGetUser{
		users: make(map[domain.ID]domain.User),
	}
}

func (m *mockUserRepositoryGetUser) Create(_ context.Context, _ domain.Querier, user *domain.User) error {
	now := time.Unix(0, 0).UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	m.users[user.ID] = *user
	return nil
}

func (m *mockUserRepositoryGetUser) GetByID(_ context.Context, _ domain.Querier, id domain.ID) (*domain.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	if m.returnNil {
		return nil, nil
	}
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

func (m *mockUserRepositoryGetUser) List(_ context.Context, _ domain.Querier, _ domain.ListParams) ([]domain.User, int, error) {
	users := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, len(users), nil
}

func TestGetUserUseCase_Execute(t *testing.T) {
	repoErr := errors.New("database connection failed")

	tests := []struct {
		name      string
		req       GetUserRequest
		setupMock func(*mockUserRepositoryGetUser)
		wantCode  string
		wantErr   bool
	}{
		{
			name: "successfully gets user by ID",
			req:  GetUserRequest{ID: "existing-id"},
			setupMock: func(m *mockUserRepositoryGetUser) {
				m.users["existing-id"] = domain.User{
					ID:        "existing-id",
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
				}
			},
			wantErr: false,
		},
		{
			name:      "returns USER_NOT_FOUND when user doesn't exist",
			req:       GetUserRequest{ID: "non-existent-id"},
			setupMock: func(_ *mockUserRepositoryGetUser) {},
			wantCode:  app.CodeUserNotFound,
			wantErr:   true,
		},
		{
			name: "propagates repository error as INTERNAL_ERROR",
			req:  GetUserRequest{ID: "some-id"},
			setupMock: func(m *mockUserRepositoryGetUser) {
				m.getByIDError = repoErr
			},
			wantCode: app.CodeInternalError,
			wantErr:  true,
		},
		{
			name: "handles nil user without error as INTERNAL_ERROR",
			req:  GetUserRequest{ID: "some-id"},
			setupMock: func(m *mockUserRepositoryGetUser) {
				m.returnNil = true
			},
			wantCode: app.CodeInternalError,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepositoryGetUser()
			mockDB := &mockQuerierGetUser{}
			tt.setupMock(mockRepo)

			useCase := NewGetUserUseCase(mockRepo, mockDB)
			resp, err := useCase.Execute(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				var appErr *app.AppError
				require.True(t, errors.As(err, &appErr), "expected AppError")
				assert.Equal(t, tt.wantCode, appErr.Code)
				assert.True(t, resp.User.ID.IsEmpty())
			} else {
				require.NoError(t, err)
				assert.False(t, resp.User.ID.IsEmpty())
				assert.Equal(t, "existing-id", string(resp.User.ID))
				assert.Equal(t, "John", resp.User.FirstName)
				assert.Equal(t, "Doe", resp.User.LastName)
				assert.Equal(t, "test@example.com", resp.User.Email)
			}
		})
	}
}

func TestNewGetUserUseCase(t *testing.T) {
	mockRepo := newMockUserRepositoryGetUser()
	mockDB := &mockQuerierGetUser{}
	useCase := NewGetUserUseCase(mockRepo, mockDB)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.userRepo)
	assert.Equal(t, mockDB, useCase.db)
}

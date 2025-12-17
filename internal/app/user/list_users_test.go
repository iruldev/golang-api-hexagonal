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

// mockQuerierListUsers is a test double for domain.Querier.
type mockQuerierListUsers struct{}

func (m *mockQuerierListUsers) Exec(_ context.Context, _ string, _ ...any) (any, error) {
	return nil, nil
}
func (m *mockQuerierListUsers) Query(_ context.Context, _ string, _ ...any) (any, error) {
	return nil, nil
}
func (m *mockQuerierListUsers) QueryRow(_ context.Context, _ string, _ ...any) any { return nil }

// mockUserRepositoryListUsers is a test double for domain.UserRepository.
type mockUserRepositoryListUsers struct {
	users     map[domain.ID]domain.User
	listError error
}

func newMockUserRepositoryListUsers() *mockUserRepositoryListUsers {
	return &mockUserRepositoryListUsers{
		users: make(map[domain.ID]domain.User),
	}
}

func (m *mockUserRepositoryListUsers) Create(_ context.Context, _ domain.Querier, user *domain.User) error {
	now := time.Unix(0, 0).UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	m.users[user.ID] = *user
	return nil
}

func (m *mockUserRepositoryListUsers) GetByID(_ context.Context, _ domain.Querier, id domain.ID) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

func (m *mockUserRepositoryListUsers) List(_ context.Context, _ domain.Querier, _ domain.ListParams) ([]domain.User, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	users := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, len(users), nil
}

func TestListUsersUseCase_Execute(t *testing.T) {
	repoErr := errors.New("database connection failed")

	tests := []struct {
		name           string
		req            ListUsersRequest
		setupMock      func(*mockUserRepositoryListUsers)
		wantUserCount  int
		wantTotalCount int
		wantPage       int
		wantCode       string
		wantErr        bool
	}{
		{
			name: "successfully lists users with pagination",
			req:  ListUsersRequest{Page: 1, PageSize: 10},
			setupMock: func(m *mockUserRepositoryListUsers) {
				m.users["user-1"] = domain.User{
					ID:        "user-1",
					Email:     "user1@example.com",
					FirstName: "John",
					LastName:  "Doe",
				}
				m.users["user-2"] = domain.User{
					ID:        "user-2",
					Email:     "user2@example.com",
					FirstName: "Jane",
					LastName:  "Doe",
				}
			},
			wantUserCount:  2,
			wantTotalCount: 2,
			wantPage:       1,
			wantErr:        false,
		},
		{
			name:           "returns empty list when no users exist",
			req:            ListUsersRequest{Page: 1, PageSize: 10},
			setupMock:      func(_ *mockUserRepositoryListUsers) {},
			wantUserCount:  0,
			wantTotalCount: 0,
			wantPage:       1,
			wantErr:        false,
		},
		{
			name: "propagates repository error as INTERNAL_ERROR",
			req:  ListUsersRequest{Page: 1, PageSize: 10},
			setupMock: func(m *mockUserRepositoryListUsers) {
				m.listError = repoErr
			},
			wantPage: 1,
			wantCode: app.CodeInternalError,
			wantErr:  true,
		},
		{
			name: "uses default page size when PageSize is 0",
			req:  ListUsersRequest{Page: 1, PageSize: 0},
			setupMock: func(m *mockUserRepositoryListUsers) {
				m.users["user-1"] = domain.User{
					ID:        "user-1",
					Email:     "user1@example.com",
					FirstName: "John",
					LastName:  "Doe",
				}
			},
			wantUserCount:  1,
			wantTotalCount: 1,
			wantPage:       1,
			wantErr:        false,
		},
		{
			name: "normalizes Page to 1 when Page is 0",
			req:  ListUsersRequest{Page: 0, PageSize: 10},
			setupMock: func(m *mockUserRepositoryListUsers) {
				m.users["user-1"] = domain.User{
					ID:        "user-1",
					Email:     "user1@example.com",
					FirstName: "John",
					LastName:  "Doe",
				}
			},
			wantUserCount:  1,
			wantTotalCount: 1,
			wantPage:       1,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepositoryListUsers()
			mockDB := &mockQuerierListUsers{}
			tt.setupMock(mockRepo)

			useCase := NewListUsersUseCase(mockRepo, mockDB)
			resp, err := useCase.Execute(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				var appErr *app.AppError
				require.True(t, errors.As(err, &appErr), "expected AppError")
				assert.Equal(t, tt.wantCode, appErr.Code)
			} else {
				require.NoError(t, err)
				assert.Len(t, resp.Users, tt.wantUserCount)
				assert.Equal(t, tt.wantTotalCount, resp.TotalCount)
				assert.Equal(t, tt.wantPage, resp.Page)
				// PageSize should be set via ListParams.Limit() (defaults to 20 if 0)
				if tt.req.PageSize == 0 {
					assert.Equal(t, 20, resp.PageSize) // default page size
				} else {
					assert.Equal(t, tt.req.PageSize, resp.PageSize)
				}
			}
		})
	}
}

func TestNewListUsersUseCase(t *testing.T) {
	mockRepo := newMockUserRepositoryListUsers()
	mockDB := &mockQuerierListUsers{}
	useCase := NewListUsersUseCase(mockRepo, mockDB)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.userRepo)
	assert.Equal(t, mockDB, useCase.db)
}

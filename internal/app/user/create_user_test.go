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
	"github.com/iruldev/golang-api-hexagonal/internal/app/audit"
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

// mockAuditService is a simplified test double for audit.AuditService.
// Since AuditService is a concrete type, we create a real one with mock dependencies.
type mockAuditDeps struct {
	repo     *mockAuditEventRepository
	redactor *mockRedactor
	idGen    *mockIDGenerator
}

func newMockAuditService() (*audit.AuditService, *mockAuditDeps) {
	repo := newMockAuditEventRepository()
	redactor := newMockRedactor()
	idGen := &mockIDGenerator{nextID: 100}
	deps := &mockAuditDeps{repo: repo, redactor: redactor, idGen: idGen}
	return audit.NewAuditService(repo, redactor, idGen), deps
}

// mockAuditEventRepository is a test double for domain.AuditEventRepository.
type mockAuditEventRepository struct {
	events      []*domain.AuditEvent
	createError error
}

func newMockAuditEventRepository() *mockAuditEventRepository {
	return &mockAuditEventRepository{
		events: make([]*domain.AuditEvent, 0),
	}
}

func (m *mockAuditEventRepository) Create(_ context.Context, _ domain.Querier, event *domain.AuditEvent) error {
	if m.createError != nil {
		return m.createError
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditEventRepository) ListByEntityID(_ context.Context, _ domain.Querier, _ string, _ domain.ID, _ domain.ListParams) ([]domain.AuditEvent, int, error) {
	return nil, 0, nil
}

// mockRedactor is a test double for domain.Redactor.
type mockRedactor struct{}

func newMockRedactor() *mockRedactor { return &mockRedactor{} }

func (m *mockRedactor) RedactMap(data map[string]any) map[string]any { return data }
func (m *mockRedactor) Redact(data any) any                          { return data }

// mockTxManager is a test double for domain.TxManager.
type mockTxManager struct{}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(tx domain.Querier) error) error {
	return fn(&mockQuerier{})
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
			mockAudit, _ := newMockAuditService()
			mockTx := &mockTxManager{}
			tt.setupMock(mockRepo)

			useCase := NewCreateUserUseCase(mockRepo, mockAudit, mockIDGen, mockTx, mockDB)
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

func TestCreateUserUseCase_Execute_AuditEventRecorded(t *testing.T) {
	t.Run("records audit event on successful user creation", func(t *testing.T) {
		mockRepo := newMockUserRepository()
		mockIDGen := newMockIDGenerator()
		mockDB := &mockQuerier{}
		mockAudit, deps := newMockAuditService()
		mockTx := &mockTxManager{}

		useCase := NewCreateUserUseCase(mockRepo, mockAudit, mockIDGen, mockTx, mockDB)
		req := CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			RequestID: "req-123",
			ActorID:   domain.ID("actor-456"),
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		assert.False(t, resp.User.ID.IsEmpty())

		// Verify audit event was recorded
		require.Len(t, deps.repo.events, 1)
		event := deps.repo.events[0]
		assert.Equal(t, domain.EventUserCreated, event.EventType)
		assert.Equal(t, "user", event.EntityType)
		assert.Equal(t, resp.User.ID, event.EntityID)
		assert.Equal(t, req.RequestID, event.RequestID)
		assert.Equal(t, req.ActorID, event.ActorID)
	})

	t.Run("returns error when audit recording fails", func(t *testing.T) {
		mockRepo := newMockUserRepository()
		mockIDGen := newMockIDGenerator()
		mockDB := &mockQuerier{}
		mockAudit, deps := newMockAuditService()
		// Configure audit repo to fail
		deps.repo.createError = errors.New("audit database error")
		mockTx := &mockTxManager{}

		useCase := NewCreateUserUseCase(mockRepo, mockAudit, mockIDGen, mockTx, mockDB)
		req := CreateUserRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			RequestID: "req-123",
			ActorID:   domain.ID("actor-456"),
		}

		_, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		var appErr *app.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, app.CodeInternalError, appErr.Code)
		assert.Contains(t, appErr.Message, "audit")
	})
}

func TestNewCreateUserUseCase(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockIDGen := newMockIDGenerator()
	mockDB := &mockQuerier{}
	mockAudit, _ := newMockAuditService()
	mockTx := &mockTxManager{}
	useCase := NewCreateUserUseCase(mockRepo, mockAudit, mockIDGen, mockTx, mockDB)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.userRepo)
	assert.Equal(t, mockAudit, useCase.auditService)
	assert.Equal(t, mockIDGen, useCase.idGen)
	assert.Equal(t, mockTx, useCase.txManager)
	assert.Equal(t, mockDB, useCase.db)
}

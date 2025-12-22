//go:build !integration

package audit

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// -- Mocks --

// mockIDGenerator is a test double for domain.IDGenerator.
type mockIDGenerator struct {
	nextID int
}

func newMockIDGenerator() *mockIDGenerator {
	return &mockIDGenerator{nextID: 1}
}

func (m *mockIDGenerator) NewID() domain.ID {
	id := domain.ID("audit-" + strconv.Itoa(m.nextID))
	m.nextID++
	return id
}

// mockQuerier is a test double for domain.Querier.
type mockQuerier struct{}

func (m *mockQuerier) Exec(_ context.Context, _ string, _ ...any) (any, error)  { return nil, nil }
func (m *mockQuerier) Query(_ context.Context, _ string, _ ...any) (any, error) { return nil, nil }
func (m *mockQuerier) QueryRow(_ context.Context, _ string, _ ...any) any       { return nil }

// mockRedactor is a test double for domain.Redactor.
type mockRedactor struct {
	redactCalled bool
	redactResult any
}

func newMockRedactor() *mockRedactor {
	return &mockRedactor{}
}

func (m *mockRedactor) RedactMap(data map[string]any) map[string]any {
	m.redactCalled = true
	if m.redactResult != nil {
		if result, ok := m.redactResult.(map[string]any); ok {
			return result
		}
	}
	// Return a copy of data with PII fields marked as redacted
	if data == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range data {
		result[k] = v
	}
	return result
}

func (m *mockRedactor) Redact(data any) any {
	m.redactCalled = true
	if m.redactResult != nil {
		return m.redactResult
	}
	// Return the data as-is for testing
	return data
}

// mockAuditEventRepository is a test double for domain.AuditEventRepository.
type mockAuditEventRepository struct {
	events      []*domain.AuditEvent
	createError error
	listResult  []domain.AuditEvent
	listCount   int
	listError   error
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

func (m *mockAuditEventRepository) ListByEntityID(_ context.Context, _ domain.Querier, entityType string, entityID domain.ID, _ domain.ListParams) ([]domain.AuditEvent, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	// Filter events if list result is not set
	if m.listResult != nil {
		return m.listResult, m.listCount, nil
	}
	var result []domain.AuditEvent
	for _, e := range m.events {
		if e.EntityType == entityType && e.EntityID == entityID {
			result = append(result, *e)
		}
	}
	return result, len(result), nil
}

// -- Tests --

func TestNewAuditService(t *testing.T) {
	mockRepo := newMockAuditEventRepository()
	mockRedactor := newMockRedactor()
	mockIDGen := newMockIDGenerator()

	svc := NewAuditService(mockRepo, mockRedactor, mockIDGen)

	assert.NotNil(t, svc)
	assert.Equal(t, mockRepo, svc.repo)
	assert.Equal(t, mockRedactor, svc.redactor)
	assert.Equal(t, mockIDGen, svc.idGen)
}

func TestAuditService_Record(t *testing.T) {
	repoErr := errors.New("database error")

	tests := []struct {
		name         string
		input        AuditEventInput
		setupMock    func(*mockAuditEventRepository, *mockRedactor)
		wantCode     string
		wantErr      bool
		checkCalled  bool
		verifyEvents bool
	}{
		{
			name: "successfully records audit event",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{"email": "[email protected]"},
				RequestID:  "req-789",
			},
			setupMock:    func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
		{
			name: "successfully records audit event with nil payload",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    nil,
				RequestID:  "req-789",
			},
			setupMock: func(_ *mockAuditEventRepository, r *mockRedactor) {
				r.redactResult = nil
			},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
		{
			name: "successfully records audit event with empty actorID (system action)",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID(""), // Empty for system actions
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{"action": "system"},
				RequestID:  "req-789",
			},
			setupMock:    func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
		{
			name: "uses requestID from input struct correctly",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{},
				RequestID:  "custom-request-id-123",
			},
			setupMock:    func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
		{
			name: "uses actorID from input struct correctly",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("custom-actor-id-456"),
				EntityType: "user",
				EntityID:   domain.ID("user-789"),
				Payload:    map[string]any{},
				RequestID:  "req-789",
			},
			setupMock:    func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
		{
			name: "fails with empty event type - returns VALIDATION_ERROR",
			input: AuditEventInput{
				EventType:  "",
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{},
				RequestID:  "req-789",
			},
			setupMock: func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantCode:  app.CodeValidationError,
			wantErr:   true,
		},
		{
			name: "fails with empty entity type - returns VALIDATION_ERROR",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{},
				RequestID:  "req-789",
			},
			setupMock: func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantCode:  app.CodeValidationError,
			wantErr:   true,
		},
		{
			name: "fails with empty entity ID - returns VALIDATION_ERROR",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID(""),
				Payload:    map[string]any{},
				RequestID:  "req-789",
			},
			setupMock: func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantCode:  app.CodeValidationError,
			wantErr:   true,
		},
		{
			name: "fails when repository returns error - returns INTERNAL_ERROR",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{"email": "[email protected]"},
				RequestID:  "req-789",
			},
			setupMock: func(repo *mockAuditEventRepository, _ *mockRedactor) {
				repo.createError = repoErr
			},
			wantCode:    app.CodeInternalError,
			wantErr:     true,
			checkCalled: true,
		},
		{
			name: "verifies redact is called on payload",
			input: AuditEventInput{
				EventType:  domain.EventUserCreated,
				ActorID:    domain.ID("actor-123"),
				EntityType: "user",
				EntityID:   domain.ID("user-456"),
				Payload:    map[string]any{"email": "[email protected]", "password": "secret"},
				RequestID:  "req-789",
			},
			setupMock:    func(_ *mockAuditEventRepository, _ *mockRedactor) {},
			wantErr:      false,
			checkCalled:  true,
			verifyEvents: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockAuditEventRepository()
			mockRedactor := newMockRedactor()
			mockIDGen := newMockIDGenerator()
			mockDB := &mockQuerier{}
			tt.setupMock(mockRepo, mockRedactor)

			svc := NewAuditService(mockRepo, mockRedactor, mockIDGen)
			err := svc.Record(context.Background(), mockDB, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				var appErr *app.AppError
				require.True(t, errors.As(err, &appErr), "expected AppError, got %T", err)
				assert.Equal(t, tt.wantCode, appErr.Code)
			} else {
				require.NoError(t, err)
				if tt.checkCalled {
					assert.True(t, mockRedactor.redactCalled, "expected redactor.Redact to be called")
				}
				if tt.verifyEvents {
					require.Len(t, mockRepo.events, 1)
					event := mockRepo.events[0]
					assert.Equal(t, tt.input.EventType, event.EventType)
					assert.Equal(t, tt.input.ActorID, event.ActorID)
					assert.Equal(t, tt.input.EntityType, event.EntityType)
					assert.Equal(t, tt.input.EntityID, event.EntityID)
					assert.Equal(t, tt.input.RequestID, event.RequestID)
					assert.NotEmpty(t, event.ID)
					assert.False(t, event.Timestamp.IsZero())
				}
			}
		})
	}
}

func TestAuditService_ListByEntity(t *testing.T) {
	listErr := errors.New("list error")

	tests := []struct {
		name       string
		entityType string
		entityID   domain.ID
		params     domain.ListParams
		setupMock  func(*mockAuditEventRepository)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "successfully lists audit events",
			entityType: "user",
			entityID:   domain.ID("user-123"),
			params:     domain.ListParams{Page: 1, PageSize: 20},
			setupMock: func(repo *mockAuditEventRepository) {
				repo.listResult = []domain.AuditEvent{
					{ID: domain.ID("audit-1"), EventType: domain.EventUserCreated, EntityType: "user", EntityID: domain.ID("user-123")},
					{ID: domain.ID("audit-2"), EventType: domain.EventUserUpdated, EntityType: "user", EntityID: domain.ID("user-123")},
				}
				repo.listCount = 2
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "returns empty list when no events found",
			entityType: "user",
			entityID:   domain.ID("user-999"),
			params:     domain.ListParams{Page: 1, PageSize: 20},
			setupMock: func(repo *mockAuditEventRepository) {
				repo.listResult = []domain.AuditEvent{}
				repo.listCount = 0
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:       "propagates repository error",
			entityType: "user",
			entityID:   domain.ID("user-123"),
			params:     domain.ListParams{Page: 1, PageSize: 20},
			setupMock: func(repo *mockAuditEventRepository) {
				repo.listError = listErr
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockAuditEventRepository()
			mockRedactor := newMockRedactor()
			mockIDGen := newMockIDGenerator()
			mockDB := &mockQuerier{}
			tt.setupMock(mockRepo)

			svc := NewAuditService(mockRepo, mockRedactor, mockIDGen)
			events, count, err := svc.ListByEntity(context.Background(), mockDB, tt.entityType, tt.entityID, tt.params)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
				assert.Len(t, events, tt.wantCount)
			}
		})
	}
}

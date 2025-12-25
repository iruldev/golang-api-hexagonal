//go:build !integration

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
)

// mockRepoForIDOR is a minimal mock repository for IDOR testing
type mockRepoForIDOR struct {
	mock.Mock
}

func (m *mockRepoForIDOR) Create(ctx context.Context, db domain.Querier, u *domain.User) error {
	return m.Called(ctx, db, u).Error(0)
}

func (m *mockRepoForIDOR) GetByID(ctx context.Context, db domain.Querier, id domain.ID) (*domain.User, error) {
	args := m.Called(ctx, db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	u := args.Get(0).(domain.User)
	return &u, args.Error(1)
}

func (m *mockRepoForIDOR) List(ctx context.Context, db domain.Querier, p domain.ListParams) ([]domain.User, int, error) {
	args := m.Called(ctx, db, p)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

type mockQuerier struct{}

func (m *mockQuerier) Exec(ctx context.Context, query string, args ...any) (any, error) {
	return nil, nil
}

func (m *mockQuerier) Query(ctx context.Context, query string, args ...any) (any, error) {
	return nil, nil
}

func (m *mockQuerier) QueryRow(ctx context.Context, query string, args ...any) any {
	return nil
}

func (m *mockQuerier) Ping(ctx context.Context) error {
	return nil
}

// TestIntegration_IDORPrevention verifies that the Router + Handler + UseCase chain
// correctly enforces authorization rules (blocking access to other users' data).
// This satisfies Story 2.7 AC #1: "integration tests verify IDOR prevention".
func TestIntegration_IDORPrevention(t *testing.T) {
	// 1. Setup Dependencies
	mockRepo := new(mockRepoForIDOR)
	mockDB := &mockQuerier{}
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()

	// Real UseCase (NOT mocked) to verify the logic inside UseCase is called
	realUseCase := user.NewGetUserUseCase(mockRepo, mockDB, logger)

	// Mock other use cases required by NewUserHandler
	mockCreateUC := new(MockCreateUserUseCase) // from user_test.go
	mockListUC := new(MockListUsersUseCase)    // from user_test.go

	// Real Handler
	userHandler := NewUserHandler(mockCreateUC, realUseCase, mockListUC, httpTransport.BasePath+"/users")
	healthHandler := NewHealthHandler()
	readyHandler := NewReadyHandler(mockDB, logger)

	// 2. Setup Router with JWT Enabled
	jwtSecret := []byte("test-secret-key-12345") // 32 bytes not required for HS256 but good practice
	jwtConfig := httpTransport.JWTConfig{
		Enabled:   true,
		Secret:    jwtSecret,
		ClockSkew: time.Minute,
	}

	r := httpTransport.NewRouter(
		logger,
		false,
		metricsReg,
		httpMetrics,
		healthHandler,
		readyHandler,
		userHandler,
		1024,
		jwtConfig,
		httpTransport.RateLimitConfig{RequestsPerSecond: 100},
	)

	// 3. Define Test Data
	userA_ID := uuid.Must(uuid.NewV7()).String()
	userB_ID := uuid.Must(uuid.NewV7()).String()

	// Setup Mock: If use case reaches Repo, it returns User B.
	// But we expect it to BLOCK before reaching Repo.
	// We'll mock it anyway just in case it fails open (which would be a bug).
	mockRepo.On("GetByID", mock.Anything, mock.Anything, domain.ID(userB_ID)).
		Return(domain.User{ID: domain.ID(userB_ID)}, nil).
		Maybe() // Should NOT be called if auth works

	// 4. Create and Sign JWT for User A
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userA_ID,
		"role": app.RoleUser,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	assert.NoError(t, err)

	// 5. Execute Request: User A requests User B
	req := httptest.NewRequest(http.MethodGet, httpTransport.BasePath+"/users/"+userB_ID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// 6. Assertions
	// Expect 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected 403 Forbidden when User A accesses User B")

	// Verify Check: Ensure repo was NOT called (proving access denied at UseCase layer)
	mockRepo.AssertNotCalled(t, "GetByID")
}

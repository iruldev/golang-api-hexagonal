//go:build !integration

package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// =============================================================================
// RequestID and ActorID Propagation Tests
// =============================================================================

func TestUserHandler_CreateUser_PropagatesRequestIDAndActorID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	expectedUser := createTestUser()

	// Capture the request to verify RequestID and ActorID
	var capturedReq user.CreateUserRequest
	mockCreateUC.On("Execute", mock.Anything, mock.MatchedBy(func(req user.CreateUserRequest) bool {
		capturedReq = req
		return true
	})).Return(user.CreateUserResponse{User: expectedUser}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context with RequestID (simulating middleware)
	ctx := req.Context()
	ctx = ctxutil.SetRequestID(ctx, "test-request-123")

	// Setup context with AuthContext (simulating auth middleware)
	authCtx := &app.AuthContext{SubjectID: "actor-user-456", Role: app.RoleUser}
	ctx = app.SetAuthContext(ctx, authCtx)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert response is successful
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Assert RequestID was propagated
	assert.Equal(t, "test-request-123", capturedReq.RequestID, "RequestID should be extracted from context")

	// Assert ActorID was propagated
	assert.Equal(t, domain.ID("actor-user-456"), capturedReq.ActorID, "ActorID should be extracted from AuthContext")

	mockCreateUC.AssertExpectations(t)
}

func TestUserHandler_CreateUser_EmptyActorIDWhenNoAuthContext(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	expectedUser := createTestUser()

	// Capture the request to verify ActorID is empty
	var capturedReq user.CreateUserRequest
	mockCreateUC.On("Execute", mock.Anything, mock.MatchedBy(func(req user.CreateUserRequest) bool {
		capturedReq = req
		return true
	})).Return(user.CreateUserResponse{User: expectedUser}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context with RequestID but NO AuthContext (unauthenticated request)
	ctx := req.Context()
	ctx = ctxutil.SetRequestID(ctx, "anon-request-789")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert success
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Assert RequestID was propagated
	assert.Equal(t, "anon-request-789", capturedReq.RequestID)

	// Assert ActorID is empty (system/anonymous)
	assert.True(t, capturedReq.ActorID.IsEmpty(), "ActorID should be empty for unauthenticated requests")

	mockCreateUC.AssertExpectations(t)
}

//go:build !integration

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
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

// Helpers for creating test users
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

// =============================================================================
// CreateUser Handler Tests
// =============================================================================

func TestUserHandler_CreateUser_Success(t *testing.T) {
	// Setup mocks
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	expectedUser := createTestUser()

	mockCreateUC.On("Execute", mock.Anything, mock.MatchedBy(func(req user.CreateUserRequest) bool {
		_, err := uuid.Parse(req.ID.String())
		return err == nil &&
			req.Email == "test@example.com" &&
			req.FirstName == "John" &&
			req.LastName == "Doe"
	})).Return(user.CreateUserResponse{User: expectedUser}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "John", data["firstName"])
	assert.Equal(t, "Doe", data["lastName"])
	idStr := data["id"].(string)
	parsedID, err := uuid.Parse(idStr)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsedID.Version())

	mockCreateUC.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidEmail(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with invalid email
	body := `{"email":"invalid-email","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
	assert.NotEmpty(t, problemResp.ValidationErrors)

	// Find email validation error
	foundEmailError := false
	for _, ve := range problemResp.ValidationErrors {
		if ve.Field == "email" {
			foundEmailError = true
			break
		}
	}
	assert.True(t, foundEmailError, "Expected validation error for email field")
}

func TestUserHandler_CreateUser_DuplicateEmail(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	mockCreateUC.On("Execute", mock.Anything, mock.Anything).
		Return(user.CreateUserResponse{}, &app.AppError{
			Op:      "CreateUser",
			Code:    app.CodeEmailExists,
			Message: "Email already exists",
			Err:     domain.ErrEmailAlreadyExists,
		})

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request
	body := `{"email":"existing@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusConflict, problemResp.Status)
	assert.Equal(t, app.CodeEmailExists, problemResp.Code)

	mockCreateUC.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidJSONBody(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with invalid JSON
	body := `{"email": invalid}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))
}

// =============================================================================
// GetUser Handler Tests
// =============================================================================

func TestUserHandler_GetUser_Success(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	expectedUser := createTestUser()
	userID := string(expectedUser.ID)

	mockGetUC.On("Execute", mock.Anything, user.GetUserRequest{ID: expectedUser.ID}).
		Return(user.GetUserResponse{User: expectedUser}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with chi router context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
	rr := httptest.NewRecorder()

	// Set up chi router context with URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, userID, data["id"])
	assert.Equal(t, "test@example.com", data["email"])

	mockGetUC.AssertExpectations(t)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	userID := "019400a0-1234-7abc-8def-1234567890ab"

	mockGetUC.On("Execute", mock.Anything, user.GetUserRequest{ID: domain.ID(userID)}).
		Return(user.GetUserResponse{}, &app.AppError{
			Op:      "GetUser",
			Code:    app.CodeUserNotFound,
			Message: "User not found",
			Err:     domain.ErrUserNotFound,
		})

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with chi router context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, problemResp.Status)
	assert.Equal(t, app.CodeUserNotFound, problemResp.Code)

	mockGetUC.AssertExpectations(t)
}

func TestUserHandler_GetUser_InvalidUUID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-uuid", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
}

func TestUserHandler_GetUser_InvalidUUIDVersion(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	nonV7 := uuid.New().String() // uuid v4 by default
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+nonV7, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nonV7)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))
}

func TestUserHandler_GetUser_EmptyID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with empty ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
	assert.NotEmpty(t, problemResp.ValidationErrors)
	assert.Equal(t, "id", problemResp.ValidationErrors[0].Field)
}

func TestUserHandler_GetUser_MixedCaseUUID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	// Valid v7 UUID
	uuidStr := "019400a0-1234-7abc-8def-1234567890ab"
	userID := domain.ID(uuidStr)
	expectedUser := createTestUser()
	expectedUser.ID = userID

	// Expect the LOWERCASE ID to be passed to the use case
	mockGetUC.On("Execute", mock.Anything, user.GetUserRequest{ID: userID}).
		Return(user.GetUserResponse{User: expectedUser}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with MIXED CASE UUID
	mixedCaseUUID := "019400A0-1234-7ABC-8DEF-1234567890AB"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+mixedCaseUUID, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", mixedCaseUUID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	mockGetUC.AssertExpectations(t)
}

func TestUserHandler_GetUser_NilUUID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Nil UUID (all zeros) - technically valid format but version 0, so should fail v7 check
	nilUUID := "00000000-0000-0000-0000-000000000000"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+nilUUID, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nilUUID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, "must be UUID v7 (time-ordered)", problemResp.ValidationErrors[0].Message)
}

// =============================================================================
// ListUsers Handler Tests
// =============================================================================

func TestUserHandler_ListUsers_Success(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	users := []domain.User{createTestUser()}
	totalCount := 1

	mockListUC.On("Execute", mock.Anything, user.ListUsersRequest{Page: 1, PageSize: 10}).
		Return(user.ListUsersResponse{
			Users:      users,
			TotalCount: totalCount,
			Page:       1,
			PageSize:   10,
		}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request with query params
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&pageSize=10", nil)
	rr := httptest.NewRecorder()

	// Execute
	h.ListUsers(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp contract.ListUsersResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Data, 1)
	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 10, resp.Pagination.PageSize)
	assert.Equal(t, 1, resp.Pagination.TotalItems)

	mockListUC.AssertExpectations(t)
}

func TestUserHandler_ListUsers_DefaultPagination(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	users := []domain.User{}
	totalCount := 0

	// Default page=1, pageSize=20
	mockListUC.On("Execute", mock.Anything, user.ListUsersRequest{Page: 1, PageSize: 20}).
		Return(user.ListUsersResponse{
			Users:      users,
			TotalCount: totalCount,
			Page:       1,
			PageSize:   20,
		}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request without query params (uses defaults)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()

	// Execute
	h.ListUsers(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp contract.ListUsersResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 20, resp.Pagination.PageSize)

	mockListUC.AssertExpectations(t)
}

func TestUserHandler_ListUsers_MaxPageSize(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	// Request pageSize=500, should be capped at 100
	mockListUC.On("Execute", mock.Anything, user.ListUsersRequest{Page: 1, PageSize: 100}).
		Return(user.ListUsersResponse{
			Users:      []domain.User{},
			TotalCount: 0,
			Page:       1,
			PageSize:   100,
		}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&pageSize=500", nil)
	rr := httptest.NewRecorder()

	// Execute
	h.ListUsers(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	mockListUC.AssertExpectations(t)
}

// =============================================================================
// Helper Functions
// =============================================================================

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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	// Create request
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context with RequestID (simulating middleware)
	ctx := req.Context()
	ctx = middleware.SetRequestID(ctx, "test-request-123")

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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC)

	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context with RequestID but NO AuthContext (unauthenticated request)
	ctx := req.Context()
	ctx = middleware.SetRequestID(ctx, "anon-request-789")
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

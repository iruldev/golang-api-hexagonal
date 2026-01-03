//go:build !integration

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with chi router context
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/"+userID, nil)
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with chi router context
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/"+userID, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, problemResp.Status)
	assert.Equal(t, contract.CodeUsrNotFound, problemResp.Code) // Story 2.3: New taxonomy

	mockGetUC.AssertExpectations(t)
}

func TestUserHandler_GetUser_InvalidUUID(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/invalid-uuid", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
}

func TestUserHandler_GetUser_InvalidUUIDVersion(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	nonV7 := "550e8400-e29b-41d4-a716-446655440000" // uuid v4 format
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/"+nonV7, nil)
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with empty ID
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/", nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with MIXED CASE UUID
	mixedCaseUUID := "019400A0-1234-7ABC-8DEF-1234567890AB"
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/"+mixedCaseUUID, nil)
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Nil UUID (all zeros) - technically valid format but version 0, so should fail v7 check
	nilUUID := "00000000-0000-0000-0000-000000000000"
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"/"+nilUUID, nil)
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", nilUUID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute
	h.GetUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, "must be UUID v7 (time-ordered)", problemResp.ValidationErrors[0].Message)
}

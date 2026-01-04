//go:build !integration

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with query params
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"?page=1&pageSize=10", nil)
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request without query params (uses defaults)
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath, nil)
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"?page=1&pageSize=500", nil)
	rr := httptest.NewRecorder()

	// Execute
	h.ListUsers(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	mockListUC.AssertExpectations(t)
}

func TestUserHandler_ListUsers_InvalidPage(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Invalid page (0)
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"?page=0", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)
	assert.Equal(t, contract.CodeValOutOfRange, problemResp.ValidationErrors[0].Code)
}

func TestUserHandler_ListUsers_InvalidPageSize(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Invalid pageSize (negative)
	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"?pageSize=-1", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)
	assert.Equal(t, contract.CodeValOutOfRange, problemResp.ValidationErrors[0].Code)
}

// Added for strict verification of defaults.
func TestUserHandler_ListUsers_ExplicitDefaultPagination(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	mockListUC.On("Execute", mock.Anything, user.ListUsersRequest{Page: 1, PageSize: 20}).
		Return(user.ListUsersResponse{
			Users:      []domain.User{},
			TotalCount: 0,
			Page:       1,
			PageSize:   20,
		}, nil)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	req := httptest.NewRequest(http.MethodGet, testUserResourcePath+"?page=1&pageSize=20", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockListUC.AssertExpectations(t)
}

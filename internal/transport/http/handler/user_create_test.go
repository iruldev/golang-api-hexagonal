//go:build !integration

package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
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

	// Verify Location header is set correctly (Story 4.6)
	location := rr.Header().Get("Location")
	assert.NotEmpty(t, location, "Location header should be set on 201 Created")
	assert.Equal(t, testUserResourcePath+"/"+idStr, location, "Location should point to created resource")

	mockCreateUC.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidEmail(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with invalid email
	body := `{"email":"invalid-email","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
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

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request
	body := `{"email":"existing@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusConflict, problemResp.Status)
	assert.Equal(t, contract.CodeUsrEmailExists, problemResp.Code) // Story 2.3: New taxonomy

	mockCreateUC.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidJSONBody(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with invalid JSON
	body := `{"email": invalid}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))
}

func TestUserHandler_CreateUser_UnknownField(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with unknown field
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe","unknownField":"val"}`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
	assert.NotEmpty(t, problemResp.ValidationErrors)

	// Check for unknown field error
	assert.Equal(t, "unknownField", problemResp.ValidationErrors[0].Field)
	assert.Contains(t, problemResp.ValidationErrors[0].Message, "unknown field")
}

func TestUserHandler_CreateUser_TrailingData(t *testing.T) {
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)

	h := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, testUserResourcePath)

	// Create request with trailing data
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}extra`
	req := httptest.NewRequest(http.MethodPost, testUserResourcePath, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute
	h.CreateUser(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/problem+json", rr.Header().Get("Content-Type"))

	var problemResp testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problemResp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, problemResp.Status)
	assert.Equal(t, app.CodeValidationError, problemResp.Code)
	assert.NotEmpty(t, problemResp.ValidationErrors)

	// Check for trailing data error
	assert.Equal(t, "body", problemResp.ValidationErrors[0].Field)
	assert.Contains(t, problemResp.ValidationErrors[0].Message, "request body contains trailing data")
}

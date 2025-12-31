package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

func TestNewProblem(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		title      string
		detail     string
		wantStatus int
		wantTitle  string
		wantDetail string
	}{
		{
			name:       "creates 400 bad request problem",
			status:     http.StatusBadRequest,
			title:      "Bad Request",
			detail:     "Invalid input provided",
			wantStatus: http.StatusBadRequest,
			wantTitle:  "Bad Request",
			wantDetail: "Invalid input provided",
		},
		{
			name:       "creates 404 not found problem",
			status:     http.StatusNotFound,
			title:      "Not Found",
			detail:     "Resource does not exist",
			wantStatus: http.StatusNotFound,
			wantTitle:  "Not Found",
			wantDetail: "Resource does not exist",
		},
		{
			name:       "creates 500 internal error problem",
			status:     http.StatusInternalServerError,
			title:      "Internal Server Error",
			detail:     "An unexpected error occurred",
			wantStatus: http.StatusInternalServerError,
			wantTitle:  "Internal Server Error",
			wantDetail: "An unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problem := NewProblem(tt.status, tt.title, tt.detail)

			assert.NotNil(t, problem)
			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantDetail, problem.Detail)
			assert.NotEmpty(t, problem.Type, "Type should be set")
		})
	}
}

func TestNewProblemWithType(t *testing.T) {
	typeURI := "https://example.com/problems/custom-error"
	problem := NewProblemWithType(typeURI, http.StatusBadRequest, "Custom Error", "A custom error occurred")

	assert.Equal(t, typeURI, problem.Type)
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.Equal(t, "Custom Error", problem.Title)
	assert.Equal(t, "A custom error occurred", problem.Detail)
}

func TestNewFieldValidationProblem(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)

	fieldErrors := []FieldError{
		{Field: "email", Message: "must be a valid email address", Code: "VAL-002"},
		{Field: "firstName", Message: "must not be empty", Code: "VAL-003"},
	}

	problem := NewFieldValidationProblem(req, fieldErrors)

	assert.NotNil(t, problem)
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.Equal(t, "Validation Error", problem.Title)
	assert.Equal(t, "One or more fields failed validation", problem.Detail)
	assert.Equal(t, app.CodeValidationError, problem.Code)
	assert.Equal(t, "/api/v1/users", problem.Instance)
	assert.Len(t, problem.Errors, 2)
	assert.Equal(t, "email", problem.Errors[0].Field)
	assert.Equal(t, "VAL-002", problem.Errors[0].Code)
}

func TestNewFieldValidationProblem_NilRequest(t *testing.T) {
	fieldErrors := []FieldError{
		{Field: "email", Message: "must be a valid email address"},
	}

	problem := NewFieldValidationProblem(nil, fieldErrors)

	assert.NotNil(t, problem)
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.Empty(t, problem.Instance)
	assert.Empty(t, problem.RequestID)
}

func TestFromAppError(t *testing.T) {
	tests := []struct {
		name       string
		appErr     *app.AppError
		wantStatus int
		wantTitle  string
		wantCode   string
	}{
		{
			name: "maps user not found error",
			appErr: &app.AppError{
				Op:      "GetUser",
				Code:    app.CodeUserNotFound,
				Message: "User with ID 123 not found",
			},
			wantStatus: http.StatusNotFound,
			wantTitle:  "User Not Found",
			wantCode:   CodeUsrNotFound,
		},
		{
			name: "maps email exists error",
			appErr: &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeEmailExists,
				Message: "Email already exists",
			},
			wantStatus: http.StatusConflict,
			wantTitle:  "Email Already Exists",
			wantCode:   CodeUsrEmailExists,
		},
		{
			name: "maps validation error",
			appErr: &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeValidationError,
				Message: "Validation failed",
			},
			wantStatus: http.StatusBadRequest,
			wantTitle:  "Required Field Missing",
			wantCode:   CodeValRequired, // Default mapping for generic validation error
		},
		{
			name: "maps internal error",
			appErr: &app.AppError{
				Op:      "CreateUser",
				Code:    app.CodeInternalError,
				Message: "Database connection failed",
			},
			wantStatus: http.StatusInternalServerError,
			wantTitle:  "Internal Server Error",
			wantCode:   CodeSysInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
			problem := FromAppError(req, tt.appErr)

			assert.NotNil(t, problem)
			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, tt.wantTitle, problem.Title)
			assert.Equal(t, tt.wantCode, problem.Code)
			assert.Equal(t, "/api/v1/users/123", problem.Instance)
		})
	}
}

func TestFromAppError_NilError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	problem := FromAppError(req, nil)

	assert.NotNil(t, problem)
	assert.Equal(t, http.StatusInternalServerError, problem.Status)
	assert.Equal(t, "Internal Server Error", problem.Title)
}

func TestFromAppError_NilRequest(t *testing.T) {
	appErr := &app.AppError{
		Op:      "GetUser",
		Code:    app.CodeUserNotFound,
		Message: "User not found",
	}

	problem := FromAppError(nil, appErr)

	assert.NotNil(t, problem)
	assert.Equal(t, http.StatusNotFound, problem.Status)
	assert.Empty(t, problem.Instance)
}

func TestFromAppError_HidesInternalDetails(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	appErr := &app.AppError{
		Op:      "CreateUser",
		Code:    app.CodeInternalError,
		Message: "connection refused to database at 192.168.1.100:5432",
	}

	problem := FromAppError(req, appErr)

	assert.NotNil(t, problem)
	assert.Equal(t, "An internal error occurred", problem.Detail)
	assert.NotContains(t, problem.Detail, "192.168.1.100")
}

func TestFromDomainError(t *testing.T) {
	tests := []struct {
		name       string
		domainErr  *domainerrors.DomainError
		wantStatus int
		wantTitle  string
		wantCode   string
	}{
		{
			name: "maps user not found error",
			domainErr: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeUserNotFound,
				Message: "User not found",
			},
			wantStatus: http.StatusNotFound,
			wantTitle:  "User Not Found",
			wantCode:   CodeUsrNotFound,
		},
		{
			name: "maps email exists error",
			domainErr: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeEmailExists,
				Message: "Email already in use",
			},
			wantStatus: http.StatusConflict,
			wantTitle:  "Email Already Exists",
			wantCode:   CodeUsrEmailExists,
		},
		{
			name: "maps invalid email error",
			domainErr: &domainerrors.DomainError{
				Code:    domainerrors.ErrCodeInvalidEmail,
				Message: "Invalid email format",
			},
			wantStatus: http.StatusBadRequest,
			wantTitle:  "Validation Error", // Note: Title comes from CodeValInvalidEmail
			wantCode:   CodeValInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
			problem := FromDomainError(req, tt.domainErr)

			assert.NotNil(t, problem)
			assert.Equal(t, tt.wantStatus, problem.Status)
			// Title check for invalid email needs adjustment to legacy behavior or new behavior
			// FromDomainError uses GetErrorCodeInfo(code).Title
			// CodeValInvalidEmail has title "Invalid Email"
			if tt.wantCode == CodeValInvalidEmail {
				assert.Equal(t, "Invalid Email", problem.Title)
			} else {
				assert.Equal(t, tt.wantTitle, problem.Title)
			}
			assert.Equal(t, tt.wantCode, problem.Code)
		})
	}
}

func TestFromDomainError_NilError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	problem := FromDomainError(req, nil)

	assert.NotNil(t, problem)
	assert.Equal(t, http.StatusInternalServerError, problem.Status)
}

func TestFromDomainError_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		code      domainerrors.ErrorCode
		wantField string
	}{
		{
			name:      "invalid email maps to email field",
			code:      domainerrors.ErrCodeInvalidEmail,
			wantField: "email",
		},
		{
			name:      "invalid first name maps to firstName field",
			code:      domainerrors.ErrCodeInvalidFirstName,
			wantField: "firstName",
		},
		{
			name:      "invalid last name maps to lastName field",
			code:      domainerrors.ErrCodeInvalidLastName,
			wantField: "lastName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", nil)
			domainErr := &domainerrors.DomainError{
				Code:    tt.code,
				Message: "Validation failed",
			}

			problem := FromDomainError(req, domainErr)

			require.Len(t, problem.Errors, 1)
			assert.Equal(t, tt.wantField, problem.Errors[0].Field)
		})
	}
}

func TestProblem_JSONMarshaling(t *testing.T) {
	problem := NewFieldValidationProblem(nil, []FieldError{
		{Field: "email", Message: "must be a valid email address", Code: "VAL-002"},
	})
	// Manually set other fields to match test expectation
	problem.Type = "https://api.example.com/problems/validation-error"
	problem.Instance = "/api/v1/users"
	problem.Detail = "Email format is invalid"
	problem.Code = "VAL-001"
	problem.RequestID = "req_abc123"
	problem.TraceID = "4bf92f3577b34da6a3ce929d0e0e4736"

	jsonBytes, err := json.Marshal(problem)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	// Verify RFC 7807 required fields
	assert.Equal(t, "https://api.example.com/problems/validation-error", unmarshaled["type"])
	assert.Equal(t, "Validation Error", unmarshaled["title"])
	assert.Equal(t, float64(400), unmarshaled["status"])
	assert.Equal(t, "Email format is invalid", unmarshaled["detail"])
	assert.Equal(t, "/api/v1/users", unmarshaled["instance"])

	// Verify extension fields
	assert.Equal(t, "VAL-001", unmarshaled["code"])
	assert.Equal(t, "req_abc123", unmarshaled["request_id"])
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", unmarshaled["trace_id"])

	// Verify errors array
	errors, ok := unmarshaled["errors"].([]interface{})
	require.True(t, ok)
	require.Len(t, errors, 1)

	errorObj, ok := errors[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "email", errorObj["field"])
	assert.Equal(t, "must be a valid email address", errorObj["message"])
	assert.Equal(t, "VAL-002", errorObj["code"])

	// Verify backward compatibility field (AC3)
	legacyErrors, ok := unmarshaled["validation_errors"].([]interface{})
	require.True(t, ok, "validation_errors field must be present for backward compatibility")
	require.Len(t, legacyErrors, 1)

	legacyObj, ok := legacyErrors[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "email", legacyObj["field"])
	assert.Equal(t, "must be a valid email address", legacyObj["message"])
}

func TestProblem_OmitsEmptyFields(t *testing.T) {
	problem := NewProblem(404, "Not Found", "")
	problem.Type = "https://api.example.com/problems/not-found"

	jsonBytes, err := json.Marshal(problem)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	// Required fields should be present
	assert.Contains(t, unmarshaled, "type")
	assert.Contains(t, unmarshaled, "title")
	assert.Contains(t, unmarshaled, "status")

	// Optional fields should be omitted when empty
	assert.NotContains(t, unmarshaled, "detail")
	assert.NotContains(t, unmarshaled, "instance")
	assert.NotContains(t, unmarshaled, "code")
	assert.NotContains(t, unmarshaled, "request_id")
	assert.NotContains(t, unmarshaled, "trace_id")
	assert.NotContains(t, unmarshaled, "errors")
}

func TestWriteProblem(t *testing.T) {
	problem := NewProblem(http.StatusNotFound, "Not Found", "User not found")
	problem.Type = problemTypeURL(ProblemTypeNotFoundSlug)
	problem.Code = "USR-001"

	w := httptest.NewRecorder()
	WriteProblem(w, problem)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, ContentTypeProblemJSON, w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Not Found", response["title"])
	assert.Equal(t, float64(404), response["status"])
	assert.Equal(t, "USR-001", response["code"])
}

func TestWriteProblem_NilProblem(t *testing.T) {
	w := httptest.NewRecorder()
	WriteProblem(w, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, ContentTypeProblemJSON, w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Internal Server Error", response["title"])
}

func TestFieldError_JSONMarshaling(t *testing.T) {
	fieldErr := FieldError{
		Field:   "email",
		Message: "must be a valid email address",
		Code:    "VAL-002",
	}

	jsonBytes, err := json.Marshal(fieldErr)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "email", unmarshaled["field"])
	assert.Equal(t, "must be a valid email address", unmarshaled["message"])
	assert.Equal(t, "VAL-002", unmarshaled["code"])
}

func TestFieldError_OmitsEmptyCode(t *testing.T) {
	fieldErr := FieldError{
		Field:   "email",
		Message: "must be a valid email address",
		// Code is intentionally empty
	}

	jsonBytes, err := json.Marshal(fieldErr)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Contains(t, unmarshaled, "field")
	assert.Contains(t, unmarshaled, "message")
	assert.NotContains(t, unmarshaled, "code")
}

//go:build !integration

package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

func TestCreateUserRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         CreateUserRequest
		expectValid bool
		errorField  string
	}{
		{
			name: "valid request",
			req: CreateUserRequest{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectValid: true,
		},
		{
			name: "invalid email",
			req: CreateUserRequest{
				Email:     "invalid-email",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectValid: false,
			errorField:  "email",
		},
		{
			name: "missing firstName",
			req: CreateUserRequest{
				Email:    "test@example.com",
				LastName: "Doe",
			},
			expectValid: false,
			errorField:  "firstName",
		},
		{
			name: "missing lastName",
			req: CreateUserRequest{
				Email:     "test@example.com",
				FirstName: "John",
			},
			expectValid: false,
			errorField:  "lastName",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs := Validate(tt.req)
			if tt.expectValid {
				assert.Empty(t, errs)
				return
			}

			require.NotEmpty(t, errs)
			assert.Equal(t, tt.errorField, errs[0].Field)
		})
	}
}

func TestUserResponse_JSONSerialization(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 12, 18, 10, 30, 0, 0, time.UTC)
	user := domain.User{
		ID:        domain.ID("019400a0-1234-7abc-8def-1234567890ab"),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := ToUserResponse(user)
	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"id":`)
	assert.Contains(t, jsonStr, `"email":`)
	assert.Contains(t, jsonStr, `"firstName":`)
	assert.Contains(t, jsonStr, `"lastName":`)
	assert.Contains(t, jsonStr, `"createdAt":`)
	assert.Contains(t, jsonStr, `"updatedAt":`)
	assert.Contains(t, jsonStr, "2025-12-18T10:30:00Z")
}

func TestPaginationResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		page       int
		pageSize   int
		totalItems int
		wantPages  int
	}{
		{"exact division", 1, 10, 100, 10},
		{"partial page", 1, 10, 95, 10},
		{"single page", 1, 20, 15, 1},
		{"empty results", 1, 10, 0, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPaginationResponse(tt.page, tt.pageSize, tt.totalItems)
			assert.Equal(t, tt.wantPages, p.TotalPages)
			assert.Equal(t, tt.page, p.Page)
			assert.Equal(t, tt.pageSize, p.PageSize)
			assert.Equal(t, tt.totalItems, p.TotalItems)
		})
	}
}

func TestNewListUsersResponse(t *testing.T) {
	t.Parallel()

	users := []domain.User{
		{
			ID:        domain.ID("1"),
			Email:     "a@example.com",
			FirstName: "Alice",
			LastName:  "Anderson",
		},
		{
			ID:        domain.ID("2"),
			Email:     "b@example.com",
			FirstName: "Bob",
			LastName:  "Brown",
		},
	}

	resp := NewListUsersResponse(users, 2, 10, 25)

	require.Len(t, resp.Data, 2)
	assert.Equal(t, "1", resp.Data[0].ID)
	assert.Equal(t, "Alice", resp.Data[0].FirstName)
	assert.Equal(t, "Brown", resp.Data[1].LastName)

	assert.Equal(t, 2, resp.Pagination.Page)
	assert.Equal(t, 10, resp.Pagination.PageSize)
	assert.Equal(t, 25, resp.Pagination.TotalItems)
	assert.Equal(t, 3, resp.Pagination.TotalPages)
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	err := WriteJSON(rec, http.StatusCreated, DataResponse[string]{Data: "ok"})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var payload map[string]any
	decodeErr := json.NewDecoder(rec.Body).Decode(&payload)
	require.NoError(t, decodeErr)
	assert.Equal(t, "ok", payload["data"])
}

func TestValidateRequestBody_HTTPFlow_InvalidEmail(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dto CreateUserRequest
		errs := ValidateRequestBody(r, &dto)
		if len(errs) > 0 {
			WriteValidationError(w, r, errs)
			return
		}
		_ = WriteJSON(w, http.StatusOK, DataResponse[string]{Data: "ok"})
	})

	body := strings.NewReader(`{"email":"bad","firstName":"John","lastName":"Doe"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var problem ProblemDetail
	err := json.NewDecoder(rec.Body).Decode(&problem)
	require.NoError(t, err)
	require.Len(t, problem.ValidationErrors, 1)
	assert.Equal(t, "email", problem.ValidationErrors[0].Field)
}

func TestValidateRequestBody_InvalidEmail(t *testing.T) {
	t.Parallel()

	body := strings.NewReader(`{"email":"invalid","firstName":"John","lastName":"Doe"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)

	var dto CreateUserRequest
	errs := ValidateRequestBody(req, &dto)

	require.NotEmpty(t, errs)
	assert.Equal(t, "email", errs[0].Field)
	assert.Equal(t, "must be a valid email address", errs[0].Message)
}

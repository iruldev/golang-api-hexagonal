package response_test

import (
	"net/http"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// TestMapError_Completeness ensures that ALL defined domain error codes
// are correctly mapped to specific HTTP status codes.
// This prevents the "Silent Mapper Failures" issue where a new code relies
// on the default 500 handler silently.
func TestMapError_Completeness(t *testing.T) {
	allCodes := domainerrors.GetAllCodes()
	if len(allCodes) == 0 {
		t.Fatal("GetAllCodes() returned empty list, expected defined codes")
	}

	for _, code := range allCodes {
		// Create a domain error with this code
		err := domainerrors.NewDomain(code, "test error")

		// Map it
		status, mappedCode := response.MapError(err)

		// Verify mapped code matches input code
		if mappedCode != code {
			t.Errorf("Code mismatch for %s: got %s, want %s", code, mappedCode, code)
		}

		// Verify status is NOT 500, UNLESS the code is explicitly CodeInternalError.
		// If MapError falls through to default, it returns 500.
		// We want to ensure specific mapping for everything else.
		if code != domainerrors.CodeInternalError && status == http.StatusInternalServerError {
			t.Errorf("CRITICAL: Code %s is not mapped to a specific HTTP status (got 500 Internal Server Error)", code)
		}
	}
}

// TestMapError_LegacySentinels ensures legacy sentinel errors still map correctly.
func TestMapError_LegacySentinels(t *testing.T) {
	tests := []struct {
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{domain.ErrNotFound, http.StatusNotFound, domainerrors.CodeNotFound},
		{domain.ErrValidation, http.StatusUnprocessableEntity, domainerrors.CodeValidationError},
		{domain.ErrUnauthorized, http.StatusUnauthorized, domainerrors.CodeUnauthorized},
		{domain.ErrForbidden, http.StatusForbidden, domainerrors.CodeForbidden},
		{domain.ErrConflict, http.StatusConflict, domainerrors.CodeConflict},
		{domain.ErrInternal, http.StatusInternalServerError, domainerrors.CodeInternalError},
		{domain.ErrTimeout, http.StatusGatewayTimeout, domainerrors.CodeTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			status, code := response.MapError(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("Status mismatch for %v: got %d, want %d", tt.err, status, tt.expectedStatus)
			}
			if code != tt.expectedCode {
				t.Errorf("Code mismatch for %v: got %s, want %s", tt.err, code, tt.expectedCode)
			}
		})
	}
}

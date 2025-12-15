package errors

import (
	"regexp"
	"testing"
)

// TestAllCodesAreUpperSnake verifies all code constants follow UPPER_SNAKE_CASE format.
func TestAllCodesAreUpperSnake(t *testing.T) {
	// UPPER_SNAKE_CASE pattern: uppercase letters and underscores only
	upperSnakePattern := regexp.MustCompile(`^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$`)

	for code := range allCodes {
		t.Run(code, func(t *testing.T) {
			if !upperSnakePattern.MatchString(code) {
				t.Errorf("Code %q does not match UPPER_SNAKE_CASE pattern", code)
			}
		})
	}
}

// TestCodesDoNotHaveErrPrefix verifies codes do not use ERR_ prefix.
func TestCodesDoNotHaveErrPrefix(t *testing.T) {
	for code := range allCodes {
		t.Run(code, func(t *testing.T) {
			if len(code) >= 4 && code[:4] == "ERR_" {
				t.Errorf("Code %q should not have ERR_ prefix", code)
			}
		})
	}
}

// TestCodeValues verifies that specific constants map to expected string values.
// This ensures no accidental changes to critical public constants.
func TestCodeValues(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"NotFound", CodeNotFound, "NOT_FOUND"},
		{"ValidationError", CodeValidationError, "VALIDATION_ERROR"},
		{"Unauthorized", CodeUnauthorized, "UNAUTHORIZED"},
		{"Forbidden", CodeForbidden, "FORBIDDEN"},
		{"Conflict", CodeConflict, "CONFLICT"},
		{"InternalError", CodeInternalError, "INTERNAL_ERROR"},
		{"Timeout", CodeTimeout, "TIMEOUT"},
		{"RateLimitExceeded", CodeRateLimitExceeded, "RATE_LIMIT_EXCEEDED"},
		{"BadRequest", CodeBadRequest, "BAD_REQUEST"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, tc.got)
			}
			// Verify it exists in the registry
			if !IsValidCode(tc.got) {
				t.Errorf("Code %q is not in the registry", tc.got)
			}
		})
	}
}

func TestNewDomain_InvalidCode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// Should panic
	NewDomain("INVALID_CODE_XYZ", "message")
}

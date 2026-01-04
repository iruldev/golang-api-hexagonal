package contract

import (
	"net/http"
	"regexp"
	"testing"
)

// codeFormatRegex validates the new taxonomy format: {CATEGORY}-{NNN}.
var codeFormatRegex = regexp.MustCompile(`^[A-Z]+-\d{3}$`)

// TestAllCodesHaveValidFormat verifies all defined codes follow {CATEGORY}-{NNN} format.
func TestAllCodesHaveValidFormat(t *testing.T) {
	codes := []struct {
		name string
		code string
	}{
		// AUTH codes
		{"CodeAuthExpiredToken", CodeAuthExpiredToken},
		{"CodeAuthInvalidToken", CodeAuthInvalidToken},
		{"CodeAuthMissingToken", CodeAuthMissingToken},
		{"CodeAuthInvalidCredentials", CodeAuthInvalidCredentials},

		// AUTHZ codes
		{"CodeAuthzForbidden", CodeAuthzForbidden},
		{"CodeAuthzInsufficientPermissions", CodeAuthzInsufficientPermissions},

		// VAL codes
		{"CodeValRequired", CodeValRequired},
		{"CodeValInvalidFormat", CodeValInvalidFormat},
		{"CodeValInvalidEmail", CodeValInvalidEmail},
		{"CodeValTooShort", CodeValTooShort},
		{"CodeValTooLong", CodeValTooLong},
		{"CodeValOutOfRange", CodeValOutOfRange},
		{"CodeValInvalidType", CodeValInvalidType},
		{"CodeValInvalidJSON", CodeValInvalidJSON},
		{"CodeValRequestTooLarge", CodeValRequestTooLarge},
		{"CodeValInvalidUUID", CodeValInvalidUUID},
		{"CodeValIdempotencyConflict", CodeValIdempotencyConflict},

		// USR codes
		{"CodeUsrNotFound", CodeUsrNotFound},
		{"CodeUsrEmailExists", CodeUsrEmailExists},
		{"CodeUsrInvalidField", CodeUsrInvalidField},

		// DB codes
		{"CodeDBConnection", CodeDBConnection},
		{"CodeDBQuery", CodeDBQuery},
		{"CodeDBTransaction", CodeDBTransaction},

		// SYS codes
		{"CodeSysInternal", CodeSysInternal},
		{"CodeSysUnavailable", CodeSysUnavailable},
		{"CodeSysConfig", CodeSysConfig},

		// RATE codes
		{"CodeRateLimitExceeded", CodeRateLimitExceeded},

		// RES codes
		{"CodeResCircuitOpen", CodeResCircuitOpen},
		{"CodeResBulkheadFull", CodeResBulkheadFull},
		{"CodeResTimeoutExceeded", CodeResTimeoutExceeded},
		{"CodeResMaxRetriesExceeded", CodeResMaxRetriesExceeded},
	}

	for _, tc := range codes {
		t.Run(tc.name, func(t *testing.T) {
			if !codeFormatRegex.MatchString(tc.code) {
				t.Errorf("code %s (%q) does not match format {CATEGORY}-{NNN}", tc.name, tc.code)
			}
		})
	}
}

// TestGetErrorCodeInfo verifies metadata is returned correctly for all codes.
func TestGetErrorCodeInfo(t *testing.T) {
	tests := []struct {
		code           string
		expectedStatus int
		expectedTitle  string
		expectedCat    string
	}{
		// AUTH
		{CodeAuthExpiredToken, http.StatusUnauthorized, "Token Expired", "AUTH"},
		{CodeAuthInvalidToken, http.StatusUnauthorized, "Invalid Token", "AUTH"},
		{CodeAuthMissingToken, http.StatusUnauthorized, "Missing Token", "AUTH"},
		{CodeAuthInvalidCredentials, http.StatusUnauthorized, "Invalid Credentials", "AUTH"},

		// AUTHZ
		{CodeAuthzForbidden, http.StatusForbidden, "Forbidden", "AUTHZ"},
		{CodeAuthzInsufficientPermissions, http.StatusForbidden, "Insufficient Permissions", "AUTHZ"},

		// VAL
		{CodeValRequired, http.StatusBadRequest, "Required Field Missing", "VAL"},
		{CodeValInvalidFormat, http.StatusBadRequest, "Invalid Format", "VAL"},
		{CodeValInvalidEmail, http.StatusBadRequest, "Invalid Email", "VAL"},
		{CodeValRequestTooLarge, http.StatusRequestEntityTooLarge, "Request Too Large", "VAL"},
		{CodeValIdempotencyConflict, http.StatusConflict, "Idempotency Conflict", "VAL"},

		// USR
		{CodeUsrNotFound, http.StatusNotFound, "User Not Found", "USR"},
		{CodeUsrEmailExists, http.StatusConflict, "Email Already Exists", "USR"},
		{CodeUsrInvalidField, http.StatusBadRequest, "Invalid User Field", "USR"},

		// DB
		{CodeDBConnection, http.StatusServiceUnavailable, "Database Connection Failed", "DB"},
		{CodeDBQuery, http.StatusInternalServerError, "Database Query Failed", "DB"},
		{CodeDBTransaction, http.StatusInternalServerError, "Transaction Failed", "DB"},

		// SYS
		{CodeSysInternal, http.StatusInternalServerError, "Internal Server Error", "SYS"},
		{CodeSysUnavailable, http.StatusServiceUnavailable, "Service Unavailable", "SYS"},
		{CodeSysConfig, http.StatusInternalServerError, "Configuration Error", "SYS"},

		// RATE
		{CodeRateLimitExceeded, http.StatusTooManyRequests, "Rate Limit Exceeded", "RATE"},

		// RES
		{CodeResCircuitOpen, http.StatusServiceUnavailable, "Circuit Breaker Open", "RES"},
		{CodeResBulkheadFull, http.StatusServiceUnavailable, "Service Overloaded", "RES"},
		{CodeResTimeoutExceeded, http.StatusServiceUnavailable, "Operation Timeout", "RES"},
		{CodeResMaxRetriesExceeded, http.StatusServiceUnavailable, "Retry Limit Exceeded", "RES"},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			info := GetErrorCodeInfo(tc.code)

			if info.HTTPStatus != tc.expectedStatus {
				t.Errorf("HTTPStatus: got %d, want %d", info.HTTPStatus, tc.expectedStatus)
			}
			if info.Title != tc.expectedTitle {
				t.Errorf("Title: got %q, want %q", info.Title, tc.expectedTitle)
			}
			if info.Category != tc.expectedCat {
				t.Errorf("Category: got %q, want %q", info.Category, tc.expectedCat)
			}
			if info.ProblemTypeSlug == "" {
				t.Error("ProblemTypeSlug should not be empty")
			}
			if info.DetailTemplate == "" {
				t.Error("DetailTemplate should not be empty")
			}
		})
	}
}

// TestHTTPStatusForCode verifies HTTP status helper returns correct values.
func TestHTTPStatusForCode(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{CodeAuthExpiredToken, http.StatusUnauthorized},
		{CodeAuthzForbidden, http.StatusForbidden},
		{CodeValRequired, http.StatusBadRequest},
		{CodeUsrNotFound, http.StatusNotFound},
		{CodeUsrEmailExists, http.StatusConflict},
		{CodeDBConnection, http.StatusServiceUnavailable},
		{CodeSysInternal, http.StatusInternalServerError},
		{CodeRateLimitExceeded, http.StatusTooManyRequests},
		{CodeResCircuitOpen, http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			status := HTTPStatusForCode(tc.code)
			if status != tc.expected {
				t.Errorf("HTTPStatusForCode(%q): got %d, want %d", tc.code, status, tc.expected)
			}
		})
	}
}

// TestTitleForCode verifies title helper returns correct values.
func TestTitleForCode(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{CodeAuthExpiredToken, "Token Expired"},
		{CodeValRequired, "Required Field Missing"},
		{CodeUsrNotFound, "User Not Found"},
		{CodeRateLimitExceeded, "Rate Limit Exceeded"},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			title := TitleForCode(tc.code)
			if title != tc.expected {
				t.Errorf("TitleForCode(%q): got %q, want %q", tc.code, title, tc.expected)
			}
		})
	}
}

// TestProblemTypeForCode verifies problem type slug helper returns correct values.
func TestProblemTypeForCode(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{CodeAuthExpiredToken, ProblemTypeUnauthorizedSlug},
		{CodeAuthzForbidden, ProblemTypeForbiddenSlug},
		{CodeValRequired, ProblemTypeValidationErrorSlug},
		{CodeUsrNotFound, ProblemTypeNotFoundSlug},
		{CodeUsrEmailExists, ProblemTypeConflictSlug},
		{CodeRateLimitExceeded, ProblemTypeRateLimitSlug},
		{CodeSysUnavailable, ProblemTypeServiceUnavailableSlug},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			slug := ProblemTypeForCode(tc.code)
			if slug != tc.expected {
				t.Errorf("ProblemTypeForCode(%q): got %q, want %q", tc.code, slug, tc.expected)
			}
		})
	}
}

// TestUnknownCodeReturnsDefault verifies unknown codes return default internal error.
func TestUnknownCodeReturnsDefault(t *testing.T) {
	unknownCodes := []string{
		"UNKNOWN-001",
		"XYZ-999",
		"",
		"invalid",
		"ERR_UNKNOWN_CODE",
	}

	for _, code := range unknownCodes {
		t.Run(code, func(t *testing.T) {
			info := GetErrorCodeInfo(code)

			if info.HTTPStatus != http.StatusInternalServerError {
				t.Errorf("expected 500 for unknown code %q, got %d", code, info.HTTPStatus)
			}
			if info.Title != "Internal Server Error" {
				t.Errorf("expected 'Internal Server Error' for unknown code %q, got %q", code, info.Title)
			}
			if info.ProblemTypeSlug != ProblemTypeInternalErrorSlug {
				t.Errorf("expected internal error slug for unknown code %q, got %q", code, info.ProblemTypeSlug)
			}
		})
	}
}

// TestLegacyCodeMapping verifies legacy ERR_* codes map correctly to new taxonomy.
func TestLegacyCodeMapping(t *testing.T) {
	tests := []struct {
		legacyCode     string
		expectedStatus int
		expectedTitle  string
	}{
		// User domain
		{"ERR_USER_NOT_FOUND", http.StatusNotFound, "User Not Found"},
		{"ERR_USER_EMAIL_EXISTS", http.StatusConflict, "Email Already Exists"},

		// Validation
		{"ERR_VALIDATION", http.StatusBadRequest, "Required Field Missing"},
		{"ERR_USER_INVALID_EMAIL", http.StatusBadRequest, "Invalid Email"},

		// Auth
		{"ERR_UNAUTHORIZED", http.StatusUnauthorized, "Token Expired"},
		{"ERR_FORBIDDEN", http.StatusForbidden, "Forbidden"},

		// System
		{"ERR_INTERNAL", http.StatusInternalServerError, "Internal Server Error"},
		{"ERR_REQUEST_TOO_LARGE", http.StatusRequestEntityTooLarge, "Request Too Large"},
		{"ERR_RATE_LIMIT_EXCEEDED", http.StatusTooManyRequests, "Rate Limit Exceeded"},
	}

	for _, tc := range tests {
		t.Run(tc.legacyCode, func(t *testing.T) {
			info := GetErrorCodeInfo(tc.legacyCode)

			if info.HTTPStatus != tc.expectedStatus {
				t.Errorf("HTTPStatus: got %d, want %d", info.HTTPStatus, tc.expectedStatus)
			}
			if info.Title != tc.expectedTitle {
				t.Errorf("Title: got %q, want %q", info.Title, tc.expectedTitle)
			}
			// Legacy codes should preserve original code
			if info.Code != tc.legacyCode {
				t.Errorf("Code should be preserved: got %q, want %q", info.Code, tc.legacyCode)
			}
		})
	}
}

// TestCategoryForCode verifies category extraction works correctly.
func TestCategoryForCode(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{CodeAuthExpiredToken, "AUTH"},
		{CodeAuthzForbidden, "AUTHZ"},
		{CodeValRequired, "VAL"},
		{CodeUsrNotFound, "USR"},
		{CodeDBConnection, "DB"},
		{CodeSysInternal, "SYS"},
		{CodeRateLimitExceeded, "RATE"},
		{CodeResCircuitOpen, "RES"},
		// Legacy/unknown formats
		{"ERR_UNKNOWN", "SYS"},
		{"invalid", "SYS"},
		{"", "SYS"},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			cat := CategoryForCode(tc.code)
			if cat != tc.expected {
				t.Errorf("CategoryForCode(%q): got %q, want %q", tc.code, cat, tc.expected)
			}
		})
	}
}

// TestIsNewTaxonomyCode verifies format detection works correctly.
func TestIsNewTaxonomyCode(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		// Valid new format
		{"AUTH-001", true},
		{"VAL-100", true},
		{"RES-004", true},
		{"A-001", true},

		// Invalid formats
		{"ERR_USER_NOT_FOUND", false},
		{"AUTH-01", false},
		{"AUTH-1", false},
		{"AUTH001", false},
		{"-001", false},
		{"AUTH-", false},
		{"", false},
		{"AUTH-0001", false},
		{"AUTH-abc", false},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			result := IsNewTaxonomyCode(tc.code)
			if result != tc.expected {
				t.Errorf("IsNewTaxonomyCode(%q): got %t, want %t", tc.code, result, tc.expected)
			}
		})
	}
}

// TestTranslateLegacyCode verifies legacy code translation works correctly.
func TestTranslateLegacyCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Legacy codes that should translate
		{"ERR_USER_NOT_FOUND", CodeUsrNotFound},
		{"ERR_UNAUTHORIZED", CodeAuthExpiredToken},
		{"ERR_INTERNAL", CodeSysInternal},

		// New format codes should pass through unchanged
		{CodeAuthExpiredToken, CodeAuthExpiredToken},
		{CodeValRequired, CodeValRequired},

		// Unknown codes should pass through unchanged
		{"UNKNOWN", "UNKNOWN"},
		{"ERR_UNKNOWN_CODE", "ERR_UNKNOWN_CODE"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := TranslateLegacyCode(tc.input)
			if result != tc.expected {
				t.Errorf("TranslateLegacyCode(%q): got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestAllRegisteredCodesHaveInfo verifies every code in registry has complete metadata.
func TestAllRegisteredCodesHaveInfo(t *testing.T) {
	for code, info := range errorCodeInfoRegistry {
		t.Run(code, func(t *testing.T) {
			if info.Code == "" {
				t.Error("Code should not be empty")
			}
			if info.Category == "" {
				t.Error("Category should not be empty")
			}
			if info.Title == "" {
				t.Error("Title should not be empty")
			}
			if info.DetailTemplate == "" {
				t.Error("DetailTemplate should not be empty")
			}
			if info.HTTPStatus < 100 || info.HTTPStatus >= 600 {
				t.Errorf("HTTPStatus %d is not a valid HTTP status code", info.HTTPStatus)
			}
			if info.ProblemTypeSlug == "" {
				t.Error("ProblemTypeSlug should not be empty")
			}
		})
	}
}

// TestRESCodesMatchResiliencePackage verifies RES codes match resilience package values.
func TestRESCodesMatchResiliencePackage(t *testing.T) {
	// These values must match internal/infra/resilience/errors.go
	expectedCodes := map[string]string{
		"CircuitOpen":        "RES-001",
		"BulkheadFull":       "RES-002",
		"TimeoutExceeded":    "RES-003",
		"MaxRetriesExceeded": "RES-004",
	}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"CircuitOpen", CodeResCircuitOpen, expectedCodes["CircuitOpen"]},
		{"BulkheadFull", CodeResBulkheadFull, expectedCodes["BulkheadFull"]},
		{"TimeoutExceeded", CodeResTimeoutExceeded, expectedCodes["TimeoutExceeded"]},
		{"MaxRetriesExceeded", CodeResMaxRetriesExceeded, expectedCodes["MaxRetriesExceeded"]},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.code != tc.expected {
				t.Errorf("RES code mismatch for %s: got %q, want %q", tc.name, tc.code, tc.expected)
			}
		})
	}
}

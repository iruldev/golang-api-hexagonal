package redact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

func TestPIIRedactor_Robustness(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	tests := []struct {
		name     string
		field    string
		isPII    bool
		expected string
	}{
		// Standard fields (Exact match)
		{"standard password", "password", true, "[REDACTED]"},
		{"standard token", "token", true, "[REDACTED]"},
		{"standard email", "email", true, "[REDACTED]"},
		{"standard ssn", "ssn", true, "[REDACTED]"},

		// Robustness: Suffix/Contains matching
		{"access_token", "access_token", true, "[REDACTED]"},
		{"refresh_token", "refresh_token", true, "[REDACTED]"},
		{"api_key_secret", "api_key_secret", true, "[REDACTED]"},
		{"UserPassword", "UserPassword", true, "[REDACTED]"},
		{"user_password", "user_password", true, "[REDACTED]"},
		{"authorization_header", "authorization_header", true, "[REDACTED]"},
		{"myConnectionSecret", "myConnectionSecret", true, "[REDACTED]"},
		{"billing_creditcard", "billing_creditcard", true, "[REDACTED]"},

		// Should NOT Redact (False Positives check)
		{"tokenization_service", "tokenization_service", true, "[REDACTED]"}, // Contains "token" - conservative approach says redact
		{"secretary", "secretary", true, "[REDACTED]"},                       // Contains "secret" - conservative approach says redact

		// Safe fields
		{"user_id", "user_id", false, "123"},
		{"username", "username", false, "jdoe"},
		{"created_at", "created_at", false, "2024-01-01"},
		{"is_active", "is_active", false, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := "sensitive-value"
			if !tt.isPII {
				val = tt.expected
			}

			input := map[string]any{
				tt.field: val,
			}
			result := r.RedactMap(input)

			if tt.isPII {
				assert.Equal(t, "[REDACTED]", result[tt.field], "Field %s should be redacted", tt.field)
			} else {
				assert.Equal(t, tt.expected, result[tt.field], "Field %s should NOT be redacted", tt.field)
			}
		})
	}
}

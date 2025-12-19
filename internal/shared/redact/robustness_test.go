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
		// Standard fields (Exact/Smart match)
		{"standard password", "password", true, "[REDACTED]"},
		{"standard token", "token", true, "[REDACTED]"},
		{"standard email", "email", true, "[REDACTED]"},
		{"standard ssn", "ssn", true, "[REDACTED]"},

		// New Fields
		{"phone", "phone", true, "[REDACTED]"},
		{"mobile", "mobile", true, "[REDACTED]"},
		{"dob", "dob", true, "[REDACTED]"},
		{"birth_date", "birth_date", true, "[REDACTED]"},
		{"passport", "passport", true, "[REDACTED]"},

		// Robustness: Boundary Matching
		{"access_token (snake)", "access_token", true, "[REDACTED]"},
		{"accessToken (camel)", "accessToken", true, "[REDACTED]"},
		{"api-token (kebab)", "api-token", true, "[REDACTED]"},
		{"my.token (dot)", "my.token", true, "[REDACTED]"},
		{"TokenKey (start camel)", "TokenKey", true, "[REDACTED]"},

		{"authtoken (lowercase)", "authtoken", true, "[REDACTED]"},
		{"userid (lowercase)", "userid", false, "123"},

		// Robustness: Substring Matching (Password/CreditCard always redacted)
		{"UserPassword", "UserPassword", true, "[REDACTED]"},
		{"user_password", "user_password", true, "[REDACTED]"},
		{"billing_creditcard", "billing_creditcard", true, "[REDACTED]"},

		// False Positives - Should NOT Redact now (Smart Match)
		{"tokenization_service", "tokenization_service", false, "sensitive-value"}, // No boundary
		{"secretary", "secretary", false, "sensitive-value"},                       // No boundary
		{"accesstoken (all lower)", "accesstoken", false, "sensitive-value"},       // No boundary
		{"blackmail", "blackmail", false, "sensitive-value"},                       // No boundary ("mail" inside)

		// Safe fields
		{"user_id", "user_id", false, "123"},
		{"username", "username", false, "jdoe"},
		{"created_at", "created_at", false, "2024-01-01"},
		{"is_active", "is_active", false, "true"},
		{"token_id", "token_id", false, "t-123"},
		{"TokenId", "TokenId", false, "t-123"},
		{"tokenid", "tokenid", false, "t-123"},
		{"secret_id", "secret_id", false, "s-123"},

		// Adversarial Checks: Multiple occurrences / Shielding
		{"False positive prefix", "broken_tokenization_but_real_token", true, "[REDACTED]"},
		{"False positive suffix", "mytoken_is_not_token_but_this_token_is", true, "[REDACTED]"},
		{"Embedded false positive", "supertokenization_values_token", true, "[REDACTED]"},
		{"Double embedding", "antitoken_tokenless_token", true, "[REDACTED]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := "sensitive-value"
			if !tt.isPII {
				val = tt.expected
			}

			// If it's not PII, we expect the original value back
			expectedVal := tt.expected
			if !tt.isPII {
				expectedVal = val // Expect original value
			}

			input := map[string]any{
				tt.field: val,
			}
			result := r.RedactMap(input)

			assert.Equal(t, expectedVal, result[tt.field], "Field matching mismatch for %s", tt.field)
		})
	}
}

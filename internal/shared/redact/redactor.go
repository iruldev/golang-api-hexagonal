// Package redact provides PII redaction utilities for audit event payloads.
package redact

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// RedactedValue is the placeholder for fully redacted PII fields.
const RedactedValue = "[REDACTED]"

// PII Field Patterns
const (
	// ExactMatchFields must match exactly (case-insensitive)
	fieldSSN   = "ssn"
	fieldEmail = "email"

	// ContainsMatchFields are redacted if they appear anywhere in the key (case-insensitive)
	fieldPassword      = "password"
	fieldToken         = "token"
	fieldSecret        = "secret"
	fieldAuthorization = "authorization"
	fieldCreditCard    = "creditcard"
	fieldCreditCardAlt = "credit_card"
)

// PIIRedactor implements domain.Redactor for PII redaction.
type PIIRedactor struct {
	emailMode string
}

// MaxRecursionDepth defines the maximum depth for recursive redaction to prevent stack overflow.
const MaxRecursionDepth = 100

// NewPIIRedactor creates a new PIIRedactor with the given configuration.
func NewPIIRedactor(cfg domain.RedactorConfig) *PIIRedactor {
	mode := strings.ToLower(strings.TrimSpace(cfg.EmailMode))
	if mode == "" {
		mode = domain.EmailModeFull
	}
	return &PIIRedactor{emailMode: mode}
}

// RedactMap processes a map and returns a new map with PII fields redacted.
// Original map is NOT modified.
// Returns nil if input is nil.
// RedactMap processes a map and returns a new map with PII fields redacted.
// Original map is NOT modified.
// Returns nil if input is nil.
func (r *PIIRedactor) RedactMap(data map[string]any) map[string]any {
	return r.redactMapInternal(data, 0)
}

// Redact processes any valid JSON type (map, slice, or primitive) and returns redacted copy.
// Original data is NOT modified.
func (r *PIIRedactor) Redact(data any) any {
	switch v := data.(type) {
	case map[string]any:
		return r.redactMapInternal(v, 0)
	case []any:
		return r.redactSlice(v, 0)
	default:
		return v
	}
}

func (r *PIIRedactor) redactMapInternal(data map[string]any, depth int) map[string]any {
	if data == nil {
		return nil
	}
	// Prevent stack overflow and PII leakage at extreme nesting depths.
	// Return empty map (fail-safe) rather than unredacted data.
	if depth > MaxRecursionDepth {
		return map[string]any{}
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = r.redactValue(k, v, depth)
	}
	return result
}

// redactValue processes a single value, redacting if it's a PII field or recursively processing nested structures.
func (r *PIIRedactor) redactValue(key string, value any, depth int) any {
	lowerKey := strings.ToLower(key)

	// Check if this key is a PII field
	if r.isPIIField(lowerKey) {
		return r.redactPIIValue(lowerKey, value)
	}

	// Recursively handle nested structures
	// At max depth, return empty structures to prevent PII leakage
	if depth >= MaxRecursionDepth {
		switch value.(type) {
		case map[string]any:
			return map[string]any{} // Fail-safe: empty map
		case []any:
			return []any{} // Fail-safe: empty slice
		default:
			return value // Primitive values are safe
		}
	}

	switch v := value.(type) {
	case map[string]any:
		return r.redactMapInternal(v, depth+1)
	case []any:
		return r.redactSlice(v, depth+1)
	default:
		return v
	}
}

// isPIIField checks if a field name matches known PII patterns.
// It uses a combination of exact matching and substring matching for robustness.
func (r *PIIRedactor) isPIIField(lowerKey string) bool {
	// 1. Exact matches for common short fields
	if lowerKey == fieldSSN || lowerKey == fieldEmail {
		return true
	}

	// 2. Substring matches for broader security (e.g., "user_password", "access_token", "MySecret")
	// Using generic terms that usually indicate sensitivity
	sensitiveTerms := []string{
		fieldPassword,
		fieldToken,
		fieldSecret,
		fieldAuthorization,
		fieldCreditCard,
		fieldCreditCardAlt,
	}

	for _, term := range sensitiveTerms {
		if strings.Contains(lowerKey, term) {
			return true
		}
	}

	return false
}

// redactPIIValue redacts a PII value based on the field type.
func (r *PIIRedactor) redactPIIValue(lowerKey string, value any) any {
	// Email has special handling for partial mode
	if lowerKey == "email" && r.emailMode == domain.EmailModePartial {
		// Only apply partial redaction if value is a string
		if strVal, ok := value.(string); ok {
			return r.partialRedactEmail(strVal)
		}
	}
	// All other PII fields get fully redacted
	return RedactedValue
}

// partialRedactEmail applies partial masking to an email address.
// Shows first 2 characters (or fewer if email local part is shorter) + domain.
// Example: "john.doe@example.com" -> "jo***@example.com"
func (r *PIIRedactor) partialRedactEmail(email string) string {
	// Use Index (first @) instead of LastIndex for safer parsing
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 {
		// No @ found or @ is first character, fully redact
		return RedactedValue
	}

	localPart := email[:atIndex]
	domainPart := email[atIndex:] // includes @

	// Show first 2 chars of local part (or all available if fewer)
	visibleChars := 2
	if len(localPart) < visibleChars {
		visibleChars = len(localPart)
	}

	return localPart[:visibleChars] + "***" + domainPart
}

// redactSlice processes a slice, recursively redacting any nested maps or slices.
func (r *PIIRedactor) redactSlice(slice []any, depth int) []any {
	if slice == nil {
		return nil
	}
	// Fail-safe: return empty slice at max depth to prevent PII leakage.
	if depth > MaxRecursionDepth {
		return []any{}
	}

	result := make([]any, len(slice))
	for i, v := range slice {
		switch item := v.(type) {
		case map[string]any:
			result[i] = r.redactMapInternal(item, depth+1)
		case []any:
			result[i] = r.redactSlice(item, depth+1)
		default:
			result[i] = item
		}
	}
	return result
}

// RedactAndMarshal converts input data to map/slice, applies redaction, and marshals to JSON bytes.
// Accepts map[string]any, []any, struct, or []byte (JSON).
// IMPORTANT: If input is a struct, fields MUST have `json` tags to be correctly handled and redacted.
// Returns the redacted data as JSON bytes suitable for AuditEvent.Payload.
func RedactAndMarshal(redactor domain.Redactor, data any) ([]byte, error) {
	// TODO: Performance optimization
	// This function performs a triple transformation (Struct -> JSON -> Map -> Redacted Map -> JSON).
	// For high-throughput audit logging, this might become a bottleneck.
	// Consider implementing a direct reflection-based mapper or code generation to skip the intermediate JSON/Map steps.

	if data == nil {
		return nil, nil
	}

	var container any

	switch v := data.(type) {
	case map[string]any:
		container = v
	case []any:
		container = v
	case []byte:
		// Parse JSON bytes into container (map or slice)
		if len(v) == 0 {
			return nil, nil
		}
		if err := json.Unmarshal(v, &container); err != nil {
			return nil, fmt.Errorf("redact: failed to unmarshal JSON bytes: %w", err)
		}
	default:
		// Assume it's a struct - marshal to JSON then unmarshal to generic container
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("redact: failed to marshal input to JSON: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &container); err != nil {
			return nil, fmt.Errorf("redact: failed to unmarshal JSON: %w", err)
		}
	}

	// Apply redaction
	redacted := redactor.Redact(container)

	// Marshal to JSON bytes
	result, err := json.Marshal(redacted)
	if err != nil {
		return nil, fmt.Errorf("redact: failed to marshal redacted data: %w", err)
	}

	return result, nil
}

// Compile-time interface check
var _ domain.Redactor = (*PIIRedactor)(nil)

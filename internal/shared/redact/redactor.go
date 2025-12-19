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
	// ContainsMatchFields are redacted if they appear anywhere in the key (case-insensitive)
	fieldPassword      = "password"
	fieldCreditCard    = "creditcard"
	fieldCreditCardAlt = "credit_card"

	// SmartMatchFields are redacted if they match as a whole word (snake_case, camelCase, or exact)
	fieldToken         = "token"
	fieldSecret        = "secret"
	fieldAuthorization = "authorization"
	fieldEmail         = "email"
	fieldSSN           = "ssn"
	fieldPhone         = "phone"
	fieldMobile        = "mobile"
	fieldDOB           = "dob"
	fieldBirthDate     = "birth_date"
	fieldPassport      = "passport"
	fieldAuthToken     = "authtoken" // Lowercase compound
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
	// Valid modes: full, partial
	if mode != domain.EmailModeFull && mode != domain.EmailModePartial {
		// Default to full redaction for safety if invalid/empty
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
		result[k] = r.redactValue(k, strings.ToLower(k), v, depth)
	}
	return result
}

// redactValue processes a single value, redacting if it's a PII field or recursively processing nested structures.
func (r *PIIRedactor) redactValue(key, lowerKey string, value any, depth int) any {
	// Check if this key is a PII field
	if r.isPIIField(key, lowerKey) {
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
// It uses a combination of substring matching and smart word boundary detection.
func (r *PIIRedactor) isPIIField(key, lowerKey string) bool {
	// 1. Unsafe Substrings: Always redact if these appear anywhere
	// "password" is almost never part of a non-sensitive field name
	unsafeTerms := []string{
		fieldPassword,
		fieldCreditCard,
		fieldCreditCardAlt,
	}

	for _, term := range unsafeTerms {
		if strings.Contains(lowerKey, term) {
			return true
		}
	}

	// 2. Safe Suffix Check:
	// If it ends with "id", it's likely an identifier, not the secret itself.
	// e.g. "token_id", "TokenID", "secretId"
	if strings.HasSuffix(lowerKey, "id") || strings.HasSuffix(lowerKey, "_id") {
		return false
	}

	// 3. Smart Matches: Redact only if it's a "whole word" match
	// Avoids false positives like "tokenization" or "secretary"
	smartTerms := []string{
		fieldToken,
		fieldSecret,
		fieldAuthorization,
		fieldEmail,
		fieldSSN,
		fieldPhone,
		fieldMobile,
		fieldDOB,
		fieldBirthDate,
		fieldPassport,
		fieldAuthToken,
	}

	for _, term := range smartTerms {
		if r.hasWord(key, lowerKey, term) {
			return true
		}
	}

	return false
}

// hasWord checks if 'term' exists in 'key' with proper boundaries (start, end, _, -, ., or CamelCase change).
// key: original mixed-case key
// lowerKey: lowercased key (optimization to avoid re-lowercasing)
// term: lowercased search term
func (r *PIIRedactor) hasWord(key, lowerKey, term string) bool {
	start := 0
	for {
		idx := strings.Index(lowerKey[start:], term)
		if idx == -1 {
			return false
		}

		// Adjust index relative to original string
		actualIdx := start + idx

		// Check boundaries for this match
		// Valid boundaries: Start of string, '_', '-', '.', or Digit
		// OR: CamelCase transition (prev char is lower, start of term is Upper in original key)
		isBoundaryBefore := true
		if actualIdx > 0 {
			prevChar := key[actualIdx-1]
			isBoundarySymbol := prevChar == '_' || prevChar == '-' || prevChar == '.' || (prevChar >= '0' && prevChar <= '9')

			// CamelCase check: "myToken" -> 'y' (lower) and 'T' (upper)
			// We need to check if key[actualIdx] is Upper
			isCamelBoundary := false
			if !isBoundarySymbol {
				// Check if current is upper (start of new word)
				if key[actualIdx] >= 'A' && key[actualIdx] <= 'Z' {
					isCamelBoundary = true
				}
			}

			if !isBoundarySymbol && !isCamelBoundary {
				isBoundaryBefore = false // Part of a previous word
			}
		}

		isBoundaryAfter := true
		endIdx := actualIdx + len(term)
		if endIdx < len(key) {
			nextChar := key[endIdx]
			isBoundarySymbol := nextChar == '_' || nextChar == '-' || nextChar == '.' || (nextChar >= '0' && nextChar <= '9')

			// CamelCase check: "TokenId" -> 'n' (lower) and 'I' (upper)
			isCamelBoundary := nextChar >= 'A' && nextChar <= 'Z'

			if !isBoundarySymbol && !isCamelBoundary {
				isBoundaryAfter = false // Part of suffix
			}
		}

		if isBoundaryBefore && isBoundaryAfter {
			return true
		}

		// If this wasn't a valid match, continue searching past this occurrence
		start = actualIdx + 1
	}
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

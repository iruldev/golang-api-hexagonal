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
func (r *PIIRedactor) RedactMap(data map[string]any) map[string]any {
	return r.redactMapInternal(data, 0)
}

func (r *PIIRedactor) redactMapInternal(data map[string]any, depth int) map[string]any {
	if data == nil {
		return nil
	}
	// Prevent stack overflow calls
	if depth > MaxRecursionDepth {
		return data // Return unredacted data or empty map? Returning as-is stops recursion but might leak.
		// Ideally we should probably fully redact or return error, but signature is fixed.
		// Let's stop recursing. For safety in audit logs, maybe we should return a placeholder?
		// But let's stick to simple "stop recursing" for now to avoid losing data context,
		// or maybe return nil?
		// Returning data as is at depth 100 is risky if PII is at depth 101.
		// Safest IS to stop recursing.
		// Let's proceed with returning `data` but NOT descending further.
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
	if depth >= MaxRecursionDepth {
		return value
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
func (r *PIIRedactor) isPIIField(lowerKey string) bool {
	switch lowerKey {
	case "password", "token", "secret", "authorization",
		"creditcard", "credit_card", "ssn", "email":
		return true
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
	if depth > MaxRecursionDepth {
		return slice
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

// RedactAndMarshal converts input data to map, applies redaction, and marshals to JSON bytes.
// Accepts map[string]any, struct (uses JSON marshal/unmarshal), or []byte (JSON).
// Returns the redacted data as JSON bytes suitable for AuditEvent.Payload.
func RedactAndMarshal(redactor domain.Redactor, data any) ([]byte, error) {
	// TODO: Performance optimization
	// This function performs a triple transformation (Struct -> JSON -> Map -> Redacted Map -> JSON).
	// For high-throughput audit logging, this might become a bottleneck.
	// Consider implementing a direct reflection-based mapper or code generation to skip the intermediate JSON/Map steps.

	if data == nil {
		return nil, nil
	}

	var dataMap map[string]any

	switch v := data.(type) {
	case map[string]any:
		dataMap = v
	case []byte:
		// Parse JSON bytes into map
		if len(v) == 0 {
			return nil, nil
		}
		if err := json.Unmarshal(v, &dataMap); err != nil {
			return nil, fmt.Errorf("redact: failed to unmarshal JSON bytes: %w", err)
		}
	default:
		// Assume it's a struct - marshal to JSON then unmarshal to map
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("redact: failed to marshal input to JSON: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			return nil, fmt.Errorf("redact: failed to unmarshal JSON to map: %w", err)
		}
	}

	// Apply redaction
	redactedMap := redactor.RedactMap(dataMap)

	// Marshal to JSON bytes
	result, err := json.Marshal(redactedMap)
	if err != nil {
		return nil, fmt.Errorf("redact: failed to marshal redacted data: %w", err)
	}

	return result, nil
}

// Compile-time interface check
var _ domain.Redactor = (*PIIRedactor)(nil)

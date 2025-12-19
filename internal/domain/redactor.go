// Package domain contains core business entities and interfaces.
// This file defines the Redactor interface for PII redaction.
package domain

// EmailMode constants define the supported pii redaction modes for email.
const (
	EmailModeFull    = "full"
	EmailModePartial = "partial"
)

// RedactorConfig holds configuration for PII redaction.
type RedactorConfig struct {
	// EmailMode specifies how email addresses should be redacted.
	// Options: EmailModeFull (replace with [REDACTED]) or EmailModePartial (show first 2 chars + domain).
	EmailMode string
}

// Redactor defines the interface for PII redaction.
// Implementations must ensure the original data is never modified.
type Redactor interface {
	// RedactMap processes a map and returns a new map with PII fields redacted.
	// Original map is NOT modified.
	// Returns nil if input is nil.
	RedactMap(data map[string]any) map[string]any

	// Redact processes any valid JSON type (map, slice, or primitive) and returns redacted copy.
	// Original data is NOT modified.
	Redact(data any) any
}

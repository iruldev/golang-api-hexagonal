package observability

import (
	"context"
	"strings"
	"time"
)

// AuditAction represents an audit action type.
type AuditAction string

const (
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
	ActionLogin  AuditAction = "login"
	ActionAccess AuditAction = "access"
)

// AuditEvent represents a security audit log entry.
type AuditEvent struct {
	Action    AuditAction    `json:"action"`
	Resource  string         `json:"resource"`
	ActorID   string         `json:"actor_id"`
	RequestID string         `json:"request_id,omitempty"`
	IPAddress string         `json:"ip_address,omitempty"`
	UserAgent string         `json:"user_agent,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Status    string         `json:"status"`          // "success" or "failure"
	Error     string         `json:"error,omitempty"` // Error message if any
	Timestamp time.Time      `json:"timestamp"`
}

// NewAuditEvent creates a new audit event.
// The actorID should be extracted from context by the caller.
// RequestID should be set manually by the caller if needed using middleware.
func NewAuditEvent(ctx context.Context, action AuditAction, resource, actorID string, metadata map[string]any) AuditEvent {
	return AuditEvent{
		Action:    action,
		Resource:  resource,
		ActorID:   actorID,
		Metadata:  MaskSensitive(metadata),
		Timestamp: time.Now(),
		Status:    "success", // Default to success
	}
}

// sensitiveKeys defines keys that should be masked.
var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"token":         {},
	"secret":        {},
	"authorization": {},
	"api_key":       {},
	"access_token":  {},
	"refresh_token": {},
	"cc_number":     {},
	"cvv":           {},
}

// MaskSensitive recursively masks sensitive keys in a map.
func MaskSensitive(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	output := make(map[string]any, len(input))
	for k, v := range input {
		// Check lowe-cased key against sensitive list
		if _, isSensitive := sensitiveKeys[strings.ToLower(k)]; isSensitive {
			output[k] = "***MASKED***"
			continue
		}

		// Recursively mask maps
		if nestedMap, ok := v.(map[string]any); ok {
			output[k] = MaskSensitive(nestedMap)
		} else {
			output[k] = v
		}
	}
	return output
}

package observability

import (
	"context"

	"go.uber.org/zap"
)

// LogAudit logs an audit event using the provided logger or the global logger.
// It ensures that the log entry is marked as an audit event for easy filtering.
func LogAudit(ctx context.Context, logger *zap.Logger, event AuditEvent) {
	if logger == nil {
		return // Or use a default/global logger if one was configured
	}

	// Enrich fields
	fields := []zap.Field{
		zap.String("event_type", "audit"),
		zap.String("audit_action", string(event.Action)),
		zap.String("audit_resource", event.Resource),
		zap.String("audit_actor_id", event.ActorID),
		zap.String("audit_status", event.Status),
		zap.Any("audit_metadata", event.Metadata),
	}

	if event.Error != "" {
		fields = append(fields, zap.String("audit_error", event.Error))
	}

	if event.RequestID != "" {
		fields = append(fields, zap.String("request_id", event.RequestID))
	}
	// IP/UserAgent should be added if present, typically from context or request

	logger.Info("Audit Event", fields...)
}

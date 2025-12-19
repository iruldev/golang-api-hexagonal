package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// AuditEventRepo implements domain.AuditEventRepository for PostgreSQL.
type AuditEventRepo struct{}

// NewAuditEventRepo creates a new AuditEventRepo instance.
func NewAuditEventRepo() *AuditEventRepo {
	return &AuditEventRepo{}
}

// Create stores a new audit event in the database.
func (r *AuditEventRepo) Create(ctx context.Context, q domain.Querier, event *domain.AuditEvent) error {
	const op = "auditEventRepo.Create"

	// Parse domain.ID to uuid.UUID at repository boundary
	id, err := uuid.Parse(string(event.ID))
	if err != nil {
		return fmt.Errorf("%s: parse ID: %w", op, err)
	}

	entityID, err := uuid.Parse(string(event.EntityID))
	if err != nil {
		return fmt.Errorf("%s: parse EntityID: %w", op, err)
	}

	// Handle nullable ActorID
	var actorID *uuid.UUID
	if !event.ActorID.IsEmpty() {
		parsed, err := uuid.Parse(string(event.ActorID))
		if err != nil {
			return fmt.Errorf("%s: parse ActorID: %w", op, err)
		}
		actorID = &parsed
	}

	_, err = q.Exec(ctx, `
		INSERT INTO audit_events (id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, event.EventType, actorID, event.EntityType, entityID, event.Payload, event.Timestamp, event.RequestID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ListByEntityID retrieves audit events for a specific entity.
// Results are ordered by timestamp DESC (newest first).
func (r *AuditEventRepo) ListByEntityID(ctx context.Context, q domain.Querier, entityType string, entityID domain.ID, params domain.ListParams) ([]domain.AuditEvent, int, error) {
	const op = "auditEventRepo.ListByEntityID"

	// Parse entityID to uuid.UUID
	eid, err := uuid.Parse(string(entityID))
	if err != nil {
		return nil, 0, fmt.Errorf("%s: parse entityID: %w", op, err)
	}

	// Get total count
	countRow := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_events 
		WHERE entity_type = $1 AND entity_id = $2
	`, entityType, eid)
	countScanner, ok := countRow.(rowScanner)
	if !ok {
		return nil, 0, fmt.Errorf("%s: invalid querier type for count", op)
	}

	var totalCount int
	if err := countScanner.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("%s: count: %w", op, err)
	}

	// If no results, return early
	if totalCount == 0 {
		return []domain.AuditEvent{}, 0, nil
	}

	// Get paginated results
	rows, err := q.Query(ctx, `
		SELECT id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id
		FROM audit_events
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY timestamp DESC, id DESC
		LIMIT $3 OFFSET $4
	`, entityType, eid, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("%s: query: %w", op, err)
	}

	scanner, ok := rows.(rowsScanner)
	if !ok {
		return nil, 0, fmt.Errorf("%s: invalid querier type for rows", op)
	}
	defer scanner.Close()

	var events []domain.AuditEvent
	for scanner.Next() {
		var event domain.AuditEvent
		var dbID, dbEntityID uuid.UUID
		var actorIDPtr *uuid.UUID

		if err := scanner.Scan(&dbID, &event.EventType, &actorIDPtr, &event.EntityType, &dbEntityID, &event.Payload, &event.Timestamp, &event.RequestID); err != nil {
			return nil, 0, fmt.Errorf("%s: scan: %w", op, err)
		}

		event.ID = domain.ID(dbID.String())
		event.EntityID = domain.ID(dbEntityID.String())
		if actorIDPtr != nil {
			event.ActorID = domain.ID(actorIDPtr.String())
		}
		// ActorID remains empty if NULL in DB (zero value)

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: rows: %w", op, err)
	}

	return events, totalCount, nil
}

// Ensure AuditEventRepo implements domain.AuditEventRepository at compile time.
var _ domain.AuditEventRepository = (*AuditEventRepo)(nil)

package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
)

// AuditEventRepo implements domain.AuditEventRepository for PostgreSQL.
type AuditEventRepo struct{}

// NewAuditEventRepo creates a new AuditEventRepo instance.
func NewAuditEventRepo() *AuditEventRepo {
	return &AuditEventRepo{}
}

// getDBTX extracts the underlying DBTX from domain.Querier.
func (r *AuditEventRepo) getDBTX(q domain.Querier) (sqlcgen.DBTX, error) {
	switch v := q.(type) {
	case *PoolQuerier:
		pool := v.pool.Pool()
		if pool == nil {
			return nil, fmt.Errorf("database not connected")
		}
		return pool, nil
	case *TxQuerier:
		return v.tx, nil
	default:
		return nil, fmt.Errorf("auditEventRepo: unsupported querier type: %T", q)
	}
}

// Create stores a new audit event in the database.
func (r *AuditEventRepo) Create(ctx context.Context, q domain.Querier, event *domain.AuditEvent) error {
	const op = "auditEventRepo.Create"

	dbtx, err := r.getDBTX(q)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

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
	var actorID pgtype.UUID
	if !event.ActorID.IsEmpty() {
		// Attempt to parse. If it's not a valid UUID, we treat it as NULL.
		if parsed, err := uuid.Parse(string(event.ActorID)); err == nil {
			actorID = pgtype.UUID{Bytes: parsed, Valid: true}
		} else {
			// Log the dropped ActorID for debugging integration issues
			observability.LoggerFromContext(ctx, slog.Default()).Warn("audit_event_repo: dropping invalid ActorID", "op", op, "actor_id", event.ActorID, "error", err, "request_id", event.RequestID)
			actorID = pgtype.UUID{Valid: false}
		}
	} else {
		actorID = pgtype.UUID{Valid: false}
	}

	params := sqlcgen.CreateAuditEventParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		EventType:  event.EventType,
		ActorID:    actorID,
		EntityType: event.EntityType,
		EntityID:   pgtype.UUID{Bytes: entityID, Valid: true},
		Payload:    event.Payload,
		Timestamp:  pgtype.Timestamptz{Time: event.Timestamp, Valid: true},
		RequestID:  pgtype.Text{String: event.RequestID, Valid: event.RequestID != ""},
	}

	if err := queries.CreateAuditEvent(ctx, params); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ListByEntityID retrieves audit events for a specific entity.
// Results are ordered by timestamp DESC (newest first).
func (r *AuditEventRepo) ListByEntityID(ctx context.Context, q domain.Querier, entityType string, entityID domain.ID, params domain.ListParams) ([]domain.AuditEvent, int, error) {
	const op = "auditEventRepo.ListByEntityID"

	dbtx, err := r.getDBTX(q)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	// Parse entityID to uuid.UUID
	eid, err := uuid.Parse(string(entityID))
	if err != nil {
		return nil, 0, fmt.Errorf("%s: parse entityID: %w", op, err)
	}

	// 1. Get total count
	count, err := queries.CountAuditEventsByEntity(ctx, sqlcgen.CountAuditEventsByEntityParams{
		EntityType: entityType,
		EntityID:   pgtype.UUID{Bytes: eid, Valid: true},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("%s: count: %w", op, err)
	}

	totalCount := int(count)

	// If no results, return early
	if totalCount == 0 {
		return []domain.AuditEvent{}, 0, nil
	}

	// 2. Get paginated results
	rows, err := queries.ListAuditEventsByEntity(ctx, sqlcgen.ListAuditEventsByEntityParams{
		EntityType: entityType,
		EntityID:   pgtype.UUID{Bytes: eid, Valid: true},
		Limit:      int32(params.Limit()),
		Offset:     int32(params.Offset()),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("%s: query: %w", op, err)
	}

	var events []domain.AuditEvent
	for _, row := range rows {
		// Convert generated struct back to domain struct
		evt := domain.AuditEvent{
			EventType:  row.EventType,
			EntityType: row.EntityType,
			Payload:    row.Payload,
			Timestamp:  row.Timestamp.Time,
			RequestID:  row.RequestID.String,
		}

		// UUID conversions
		var idUuid uuid.UUID
		copy(idUuid[:], row.ID.Bytes[:])
		evt.ID = domain.ID(idUuid.String())

		var eidUuid uuid.UUID
		copy(eidUuid[:], row.EntityID.Bytes[:])
		evt.EntityID = domain.ID(eidUuid.String())

		if row.ActorID.Valid {
			var aidUuid uuid.UUID
			copy(aidUuid[:], row.ActorID.Bytes[:])
			evt.ActorID = domain.ID(aidUuid.String())
		}

		events = append(events, evt)
	}

	return events, totalCount, nil
}

// Ensure AuditEventRepo implements domain.AuditEventRepository at compile time.
var _ domain.AuditEventRepository = (*AuditEventRepo)(nil)

-- Audit queries for sqlc
-- Story 5.4: Type-safe SQL queries for Audit module

-- name: CreateAuditEvent :exec
INSERT INTO audit_events (id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAuditEventsByEntity :many
SELECT id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id
FROM audit_events
WHERE entity_type = $1 AND entity_id = $2
ORDER BY timestamp DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: CountAuditEventsByEntity :one
SELECT COUNT(*) FROM audit_events
WHERE entity_type = $1 AND entity_id = $2;

-- name: GetAuditEventByID :one
SELECT id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id
FROM audit_events WHERE id = $1;

-- name: ListAuditEventsByRequestID :many
SELECT id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id
FROM audit_events
WHERE request_id = $1
ORDER BY timestamp DESC;

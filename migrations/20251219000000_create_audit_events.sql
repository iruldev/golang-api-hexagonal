-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_events (
    id uuid PRIMARY KEY,
    event_type varchar(100) NOT NULL,
    actor_id uuid,  -- NULL for system/unauthenticated events
    entity_type varchar(50) NOT NULL,
    entity_id uuid NOT NULL,
    payload jsonb NOT NULL,
    timestamp timestamptz NOT NULL,
    request_id varchar(64)
);

CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_entity_type ON audit_events(entity_type);
CREATE INDEX idx_audit_events_entity_time ON audit_events(entity_type, entity_id, timestamp DESC);
CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp DESC);
CREATE INDEX idx_audit_events_request_id ON audit_events(request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_events;
-- +goose StatementEnd

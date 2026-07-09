BEGIN;

CREATE TABLE IF NOT EXISTS audit_events
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),

    event_type     TEXT        NOT NULL,
    aggregate_type TEXT        NOT NULL,
    aggregate_id   TEXT        NOT NULL,

    payload        JSONB       NOT NULL DEFAULT '{}'::jsonb,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT audit_events_event_type_not_blank
        CHECK (char_length(btrim(event_type)) > 0),

    CONSTRAINT audit_events_aggregate_type_not_blank
        CHECK (char_length(btrim(aggregate_type)) > 0),

    CONSTRAINT audit_events_aggregate_id_not_blank
        CHECK (char_length(btrim(aggregate_id)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_audit_events_event_type
    ON audit_events (event_type);

CREATE INDEX IF NOT EXISTS idx_audit_events_aggregate
    ON audit_events (aggregate_type, aggregate_id);

CREATE INDEX IF NOT EXISTS idx_audit_events_created_at
    ON audit_events (created_at);

COMMIT;
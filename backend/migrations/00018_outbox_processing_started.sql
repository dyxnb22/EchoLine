-- +goose Up
ALTER TABLE outbox_events ADD COLUMN processing_started_at TIMESTAMPTZ;

CREATE INDEX outbox_events_processing_started_idx
    ON outbox_events (status, processing_started_at)
    WHERE status = 'processing';

-- +goose Down
DROP INDEX IF EXISTS outbox_events_processing_started_idx;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS processing_started_at;

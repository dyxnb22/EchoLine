-- +goose Up
ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS processing_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE outbox_events DROP COLUMN IF EXISTS processing_at;

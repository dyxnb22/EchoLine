-- +goose Up
ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS outbox_events_status_check;
ALTER TABLE outbox_events ADD CONSTRAINT outbox_events_status_check
    CHECK (status IN ('pending', 'processing', 'published', 'failed'));

-- +goose Down
ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS outbox_events_status_check;
ALTER TABLE outbox_events ADD CONSTRAINT outbox_events_status_check
    CHECK (status IN ('pending', 'published', 'failed'));

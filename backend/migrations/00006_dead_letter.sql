-- +goose Up
CREATE TABLE dead_letter_events (
    id UUID PRIMARY KEY,
    source_topic TEXT NOT NULL,
    payload JSONB NOT NULL,
    error_message TEXT NOT NULL DEFAULT '',
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX dead_letter_events_created_at_idx ON dead_letter_events (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS dead_letter_events;

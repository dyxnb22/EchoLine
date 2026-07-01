-- +goose Up
CREATE TABLE device_sync_cursors (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    last_seq BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, device_id, conversation_id)
);

CREATE INDEX device_sync_cursors_user_device_idx ON device_sync_cursors (user_id, device_id);

-- +goose Down
DROP TABLE IF EXISTS device_sync_cursors;

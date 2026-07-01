-- +goose Up
CREATE TABLE message_deliveries (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('sent', 'delivered', 'read')),
    acked_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (message_id, user_id, device_id)
);

CREATE INDEX message_deliveries_user_id_idx ON message_deliveries (user_id);

-- +goose Down
DROP TABLE IF EXISTS message_deliveries;

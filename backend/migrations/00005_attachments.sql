-- +goose Up
CREATE TABLE attachments (
    id UUID PRIMARY KEY,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    object_key TEXT NOT NULL,
    mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    checksum TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX attachments_message_id_idx ON attachments (message_id);
CREATE INDEX attachments_owner_id_idx ON attachments (owner_id);

-- +goose Down
DROP TABLE IF EXISTS attachments;

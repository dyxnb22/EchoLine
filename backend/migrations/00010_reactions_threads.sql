-- +goose Up
CREATE TABLE message_reactions (
    message_id UUID NOT NULL,
    user_id    UUID NOT NULL,
    emoji      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id, emoji)
);

ALTER TABLE messages ADD COLUMN IF NOT EXISTS parent_message_id UUID NULL REFERENCES messages(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE messages DROP COLUMN IF EXISTS parent_message_id;
DROP TABLE IF EXISTS message_reactions;

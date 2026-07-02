-- +goose Up
ALTER TABLE conversation_members ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE conversation_members DROP COLUMN IF EXISTS archived_at;

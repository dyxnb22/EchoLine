-- +goose Up
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS requires_entitlement BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE conversations DROP COLUMN IF EXISTS requires_entitlement;

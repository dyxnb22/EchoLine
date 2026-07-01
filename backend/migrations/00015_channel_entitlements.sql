-- +goose Up
CREATE TABLE IF NOT EXISTS channel_entitlements (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id      UUID        NOT NULL,
    channel_id   UUID        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'active',
    reference    TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,
    UNIQUE (user_id, channel_id)
);

CREATE INDEX IF NOT EXISTS channel_entitlements_user_idx ON channel_entitlements (user_id, status);

-- +goose Down
DROP TABLE IF EXISTS channel_entitlements;

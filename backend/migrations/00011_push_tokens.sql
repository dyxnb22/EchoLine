-- +goose Up
CREATE TABLE push_tokens (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    UUID        NOT NULL,
    device_id  TEXT        NOT NULL,
    token      TEXT        NOT NULL,
    platform   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);

-- +goose Down
DROP TABLE IF EXISTS push_tokens;

-- +goose Up
CREATE TABLE payment_ledger (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id      UUID        NOT NULL,
    amount_cents BIGINT      NOT NULL,
    currency     TEXT        NOT NULL DEFAULT 'USD',
    status       TEXT        NOT NULL DEFAULT 'pending',
    reference    TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ad_campaigns (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    channel_id UUID        NOT NULL,
    title      TEXT        NOT NULL DEFAULT '',
    status     TEXT        NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ad_impressions (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    campaign_id UUID        NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE encryption_key_bundles (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    UUID        NOT NULL,
    device_id  TEXT        NOT NULL,
    public_key TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);

-- +goose Down
DROP TABLE IF EXISTS encryption_key_bundles;
DROP TABLE IF EXISTS ad_impressions;
DROP TABLE IF EXISTS ad_campaigns;
DROP TABLE IF EXISTS payment_ledger;

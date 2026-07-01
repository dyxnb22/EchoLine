-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    event_type   TEXT        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}',
    status       TEXT        NOT NULL DEFAULT 'pending',
    attempts     INT         NOT NULL DEFAULT 0,
    last_error   TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS webhook_deliveries_status_idx ON webhook_deliveries (status, created_at);

ALTER TABLE ad_campaigns ADD COLUMN IF NOT EXISTS budget_cents BIGINT NOT NULL DEFAULT 0;
ALTER TABLE ad_campaigns ADD COLUMN IF NOT EXISTS frequency_cap INT NOT NULL DEFAULT 3;

CREATE UNIQUE INDEX IF NOT EXISTS ad_impressions_campaign_user_day_idx
    ON ad_impressions (campaign_id, user_id, (created_at::date));

-- +goose Down
DROP INDEX IF EXISTS ad_impressions_campaign_user_day_idx;
ALTER TABLE ad_campaigns DROP COLUMN IF EXISTS frequency_cap;
ALTER TABLE ad_campaigns DROP COLUMN IF EXISTS budget_cents;
DROP TABLE IF EXISTS webhook_deliveries;
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;

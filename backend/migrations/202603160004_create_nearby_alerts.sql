BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

ALTER TABLE notification_preferences
    ADD COLUMN IF NOT EXISTS notify_on_nearby_issue_created BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS nearby_alert_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL,
    label TEXT,
    latitude DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    radius_m INT NOT NULL CHECK (radius_m BETWEEN 100 AND 5000),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nearby_alert_subscriptions_follower_updated_at
    ON nearby_alert_subscriptions(follower_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_nearby_alert_subscriptions_geog
    ON nearby_alert_subscriptions
    USING GIST ((ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography));

DROP TRIGGER IF EXISTS trg_nearby_alert_subscriptions_updated_at ON nearby_alert_subscriptions;
CREATE TRIGGER trg_nearby_alert_subscriptions_updated_at
BEFORE UPDATE ON nearby_alert_subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS nearby_alert_deliveries (
    id BIGSERIAL PRIMARY KEY,
    subscription_id UUID NOT NULL REFERENCES nearby_alert_subscriptions(id) ON DELETE CASCADE,
    follower_id UUID NOT NULL,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_nearby_alert_deliveries_subscription_issue UNIQUE (subscription_id, issue_id)
);

CREATE INDEX IF NOT EXISTS idx_nearby_alert_deliveries_issue_id ON nearby_alert_deliveries(issue_id);
CREATE INDEX IF NOT EXISTS idx_nearby_alert_deliveries_follower_created_at ON nearby_alert_deliveries(follower_id, created_at DESC);

COMMIT;

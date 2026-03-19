BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;
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

CREATE TABLE IF NOT EXISTS regions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE,
    name VARCHAR(150) NOT NULL,
    level VARCHAR(20) NOT NULL CHECK (level IN ('province', 'city', 'district', 'village')),
    parent_id BIGINT REFERENCES regions(id) ON DELETE SET NULL,
    geom GEOMETRY(MULTIPOLYGON, 4326) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_regions_parent_id ON regions(parent_id);
CREATE INDEX IF NOT EXISTS idx_regions_level ON regions(level);
CREATE INDEX IF NOT EXISTS idx_regions_geom ON regions USING GIST(geom);

DROP TRIGGER IF EXISTS trg_regions_updated_at ON regions;
CREATE TRIGGER trg_regions_updated_at
BEFORE UPDATE ON regions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY,
    anon_token_hash VARCHAR(128) NOT NULL UNIQUE,
    fingerprint_hash VARCHAR(128),
    trust_score INT NOT NULL DEFAULT 0,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    ban_reason TEXT,
    last_ip INET,
    last_user_agent TEXT,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_fingerprint_hash ON devices(fingerprint_hash);
CREATE INDEX IF NOT EXISTS idx_devices_is_banned ON devices(is_banned);

DROP TRIGGER IF EXISTS trg_devices_updated_at ON devices;
CREATE TRIGGER trg_devices_updated_at
BEFORE UPDATE ON devices
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100),
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'moderator', 'admin')),
    xp_points INT NOT NULL DEFAULT 0,
    rank_title VARCHAR(50) NOT NULL DEFAULT 'Warga Biasa',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS oauth_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL CHECK (provider IN ('google')),
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(128) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE TABLE IF NOT EXISTS device_consents (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    terms_version VARCHAR(32) NOT NULL,
    privacy_version VARCHAR(32),
    ip_address INET,
    user_agent TEXT,
    consented_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_consents_device_id ON device_consents(device_id);
CREATE INDEX IF NOT EXISTS idx_device_consents_user_id ON device_consents(user_id);

CREATE TABLE IF NOT EXISTS issues (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'fixed', 'archived', 'rejected', 'merged')),
    verification_status VARCHAR(20) NOT NULL DEFAULT 'unverified' CHECK (verification_status IN ('unverified', 'community_verified', 'admin_verified')),
    severity_current SMALLINT NOT NULL DEFAULT 1 CHECK (severity_current BETWEEN 1 AND 5),
    severity_max SMALLINT NOT NULL DEFAULT 1 CHECK (severity_max BETWEEN 1 AND 5),
    public_location GEOGRAPHY(POINT, 4326) NOT NULL,
    region_id BIGINT REFERENCES regions(id) ON DELETE SET NULL,
    road_name VARCHAR(200),
    road_type VARCHAR(32) CHECK (road_type IN ('national', 'provincial', 'city', 'village', 'unknown')),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    reopened_count INT NOT NULL DEFAULT 0 CHECK (reopened_count >= 0),
    casualty_count INT NOT NULL DEFAULT 0 CHECK (casualty_count >= 0),
    submission_count INT NOT NULL DEFAULT 0 CHECK (submission_count >= 0),
    photo_count INT NOT NULL DEFAULT 0 CHECK (photo_count >= 0),
    reaction_count INT NOT NULL DEFAULT 0 CHECK (reaction_count >= 0),
    share_count INT NOT NULL DEFAULT 0 CHECK (share_count >= 0),
    view_count INT NOT NULL DEFAULT 0 CHECK (view_count >= 0),
    unique_view_count INT NOT NULL DEFAULT 0 CHECK (unique_view_count >= 0),
    flag_count INT NOT NULL DEFAULT 0 CHECK (flag_count >= 0),
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    hidden_reason TEXT,
    merged_into_issue_id UUID REFERENCES issues(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_verification_status ON issues(verification_status);
CREATE INDEX IF NOT EXISTS idx_issues_region_id ON issues(region_id);
CREATE INDEX IF NOT EXISTS idx_issues_is_hidden ON issues(is_hidden);
CREATE INDEX IF NOT EXISTS idx_issues_public_location ON issues USING GIST(public_location);

DROP TRIGGER IF EXISTS trg_issues_updated_at ON issues;
CREATE TRIGGER trg_issues_updated_at
BEFORE UPDATE ON issues
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS issue_submissions (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    client_request_id UUID NOT NULL UNIQUE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'spam')),
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    region_id BIGINT REFERENCES regions(id) ON DELETE SET NULL,
    gps_accuracy_m NUMERIC(6, 2) CHECK (gps_accuracy_m IS NULL OR gps_accuracy_m >= 0),
    captured_at TIMESTAMPTZ,
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    severity SMALLINT NOT NULL DEFAULT 1 CHECK (severity BETWEEN 1 AND 5),
    has_casualty BOOLEAN NOT NULL DEFAULT FALSE,
    casualty_count INT NOT NULL DEFAULT 0 CHECK (casualty_count >= 0),
    note TEXT,
    source VARCHAR(20) NOT NULL DEFAULT 'web' CHECK (source IN ('web', 'pwa', 'admin')),
    moderation_note TEXT,
    moderated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    moderated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_issue_submissions_issue_id ON issue_submissions(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_submissions_device_id ON issue_submissions(device_id);
CREATE INDEX IF NOT EXISTS idx_issue_submissions_user_id ON issue_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_issue_submissions_region_id ON issue_submissions(region_id);
CREATE INDEX IF NOT EXISTS idx_issue_submissions_status ON issue_submissions(status);
CREATE INDEX IF NOT EXISTS idx_issue_submissions_location ON issue_submissions USING GIST(location);

DROP TRIGGER IF EXISTS trg_issue_submissions_updated_at ON issue_submissions;
CREATE TRIGGER trg_issue_submissions_updated_at
BEFORE UPDATE ON issue_submissions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS submission_media (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES issue_submissions(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL UNIQUE,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes INT NOT NULL CHECK (size_bytes > 0),
    width INT CHECK (width IS NULL OR width > 0),
    height INT CHECK (height IS NULL OR height > 0),
    sha256 CHAR(64),
    blurhash VARCHAR(255),
    metadata JSONB,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submission_media_submission_id ON submission_media(submission_id);
CREATE INDEX IF NOT EXISTS idx_submission_media_sha256 ON submission_media(sha256);
CREATE UNIQUE INDEX IF NOT EXISTS uq_submission_media_primary_per_submission
    ON submission_media(submission_id)
    WHERE is_primary = TRUE;

CREATE TABLE IF NOT EXISTS issue_reactions (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    reaction_type VARCHAR(20) NOT NULL CHECK (reaction_type IN ('angry', 'danger', 'upvote')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issue_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_reactions_reaction_type ON issue_reactions(reaction_type);

CREATE TABLE IF NOT EXISTS issue_flags (
    id BIGSERIAL PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('spam', 'hoax', 'off_topic', 'duplicate', 'abuse', 'other')),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (issue_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_flags_issue_id ON issue_flags(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_flags_reason ON issue_flags(reason);

CREATE TABLE IF NOT EXISTS submission_flags (
    id BIGSERIAL PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES issue_submissions(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    reason VARCHAR(20) NOT NULL CHECK (reason IN ('spam', 'hoax', 'off_topic', 'duplicate', 'abuse', 'other')),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (submission_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_submission_flags_submission_id ON submission_flags(submission_id);
CREATE INDEX IF NOT EXISTS idx_submission_flags_reason ON submission_flags(reason);

CREATE TABLE IF NOT EXISTS issue_status_history (
    id BIGSERIAL PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    from_status VARCHAR(20) CHECK (from_status IS NULL OR from_status IN ('open', 'fixed', 'archived', 'rejected', 'merged')),
    to_status VARCHAR(20) NOT NULL CHECK (to_status IN ('open', 'fixed', 'archived', 'rejected', 'merged')),
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_by_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_issue_status_history_issue_id ON issue_status_history(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_status_history_created_at ON issue_status_history(created_at);

CREATE TABLE IF NOT EXISTS issue_events (
    id BIGSERIAL PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_issue_events_issue_id_created_at
    ON issue_events(issue_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS issue_followers (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    follower_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_issue_followers_issue_follower UNIQUE (issue_id, follower_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_followers_issue_id ON issue_followers(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_followers_follower_id ON issue_followers(follower_id);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    follower_id UUID NOT NULL,
    event_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ,
    CONSTRAINT uq_notifications_event_follower UNIQUE (event_id, follower_id)
);

CREATE INDEX IF NOT EXISTS idx_notifications_issue_id ON notifications(issue_id);
CREATE INDEX IF NOT EXISTS idx_notifications_follower_created_at ON notifications(follower_id, created_at DESC);

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disabled_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_follower_id ON push_subscriptions(follower_id);
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_active_follower_id
    ON push_subscriptions(follower_id)
    WHERE disabled_at IS NULL;

DROP TRIGGER IF EXISTS trg_push_subscriptions_updated_at ON push_subscriptions;
CREATE TRIGGER trg_push_subscriptions_updated_at
BEFORE UPDATE ON push_subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS follower_auth_bindings (
    follower_id UUID PRIMARY KEY,
    device_token_hash CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_follower_auth_bindings_updated_at ON follower_auth_bindings;
CREATE TRIGGER trg_follower_auth_bindings_updated_at
BEFORE UPDATE ON follower_auth_bindings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS notification_preferences (
    follower_id UUID PRIMARY KEY,
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_photo_added BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_status_updated BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_severity_changed BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_casualty_reported BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_nearby_issue_created BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_notification_preferences_updated_at ON notification_preferences;
CREATE TRIGGER trg_notification_preferences_updated_at
BEFORE UPDATE ON notification_preferences
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

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
    USING GIST (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography);

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

CREATE TABLE IF NOT EXISTS moderation_actions (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action_type VARCHAR(32) NOT NULL CHECK (
        action_type IN (
            'accept_submission',
            'reject_submission',
            'hide_issue',
            'unhide_issue',
            'merge_issue',
            'mark_fixed',
            'reopen_issue',
            'reject_issue',
            'ban_device',
            'unban_device',
            'auto_hide_issue'
        )
    ),
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('issue', 'submission', 'device')),
    target_id UUID NOT NULL,
    admin_username VARCHAR(100),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_moderation_actions_actor_user_id ON moderation_actions(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_moderation_actions_target ON moderation_actions(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_moderation_actions_action_type ON moderation_actions(action_type);

CREATE TABLE IF NOT EXISTS issue_daily_stats (
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    stat_date DATE NOT NULL,
    views INT NOT NULL DEFAULT 0 CHECK (views >= 0),
    unique_views INT NOT NULL DEFAULT 0 CHECK (unique_views >= 0),
    shares INT NOT NULL DEFAULT 0 CHECK (shares >= 0),
    reactions INT NOT NULL DEFAULT 0 CHECK (reactions >= 0),
    submissions INT NOT NULL DEFAULT 0 CHECK (submissions >= 0),
    flags INT NOT NULL DEFAULT 0 CHECK (flags >= 0),
    PRIMARY KEY (issue_id, stat_date)
);

CREATE OR REPLACE VIEW issue_public_view AS
SELECT
    i.id,
    i.status,
    i.verification_status,
    i.severity_current,
    i.severity_max,
    i.public_location,
    i.region_id,
    i.road_name,
    i.road_type,
    i.first_seen_at,
    i.last_seen_at,
    i.resolved_at,
    i.reopened_count,
    i.casualty_count,
    i.submission_count,
    i.photo_count,
    i.reaction_count,
    i.share_count,
    i.view_count,
    i.unique_view_count,
    GREATEST(
        0,
        FLOOR(EXTRACT(EPOCH FROM (COALESCE(i.resolved_at, NOW()) - i.first_seen_at)) / 86400)
    )::INT AS age_days,
    (
        (COALESCE(i.unique_view_count, 0)::NUMERIC * 500) +
        (
            GREATEST(
                0,
                FLOOR(EXTRACT(EPOCH FROM (COALESCE(i.resolved_at, NOW()) - i.first_seen_at)) / 86400)
            )::NUMERIC * 100000
        )
    )::NUMERIC(18, 2) AS estimated_loss,
    i.created_at,
    i.updated_at
FROM issues i
WHERE i.is_hidden = FALSE
  AND i.status NOT IN ('rejected', 'merged');

COMMIT;

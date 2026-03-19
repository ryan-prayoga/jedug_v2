BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

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

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_notifications_issue_id ON notifications(issue_id);
CREATE INDEX IF NOT EXISTS idx_notifications_follower_created_at ON notifications(follower_id, created_at DESC);

COMMIT;

BEGIN;

CREATE TABLE IF NOT EXISTS issue_followers (
    id UUID PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    follower_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_issue_followers_issue_follower UNIQUE (issue_id, follower_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_followers_issue_id ON issue_followers(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_followers_follower_id ON issue_followers(follower_id);

COMMIT;

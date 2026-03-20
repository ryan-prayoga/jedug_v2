CREATE TABLE IF NOT EXISTS push_delivery_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID NOT NULL,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_push_delivery_jobs_event_follower UNIQUE (event_id, follower_id),
    CONSTRAINT chk_push_delivery_jobs_attempt_count_non_negative CHECK (attempt_count >= 0)
);

CREATE INDEX IF NOT EXISTS idx_push_delivery_jobs_follower_id
    ON push_delivery_jobs(follower_id);

CREATE INDEX IF NOT EXISTS idx_push_delivery_jobs_ready
    ON push_delivery_jobs(next_attempt_at, created_at)
    WHERE delivered_at IS NULL AND failed_at IS NULL;

DROP TRIGGER IF EXISTS trg_push_delivery_jobs_updated_at ON push_delivery_jobs;
CREATE TRIGGER trg_push_delivery_jobs_updated_at
BEFORE UPDATE ON push_delivery_jobs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


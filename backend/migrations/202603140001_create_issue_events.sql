BEGIN;

CREATE TABLE IF NOT EXISTS issue_events (
    id BIGSERIAL PRIMARY KEY,
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_issue_events_issue_id_created_at
    ON issue_events(issue_id, created_at DESC, id DESC);

COMMIT;

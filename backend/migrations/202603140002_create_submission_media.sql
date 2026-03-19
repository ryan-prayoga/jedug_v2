BEGIN;

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

ALTER TABLE submission_media
    ADD COLUMN IF NOT EXISTS width INT,
    ADD COLUMN IF NOT EXISTS height INT,
    ADD COLUMN IF NOT EXISTS blurhash VARCHAR(255),
    ADD COLUMN IF NOT EXISTS metadata JSONB,
    ADD COLUMN IF NOT EXISTS sort_order SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_submission_media_submission_id ON submission_media(submission_id);
CREATE INDEX IF NOT EXISTS idx_submission_media_sha256 ON submission_media(sha256);
CREATE UNIQUE INDEX IF NOT EXISTS uq_submission_media_primary_per_submission
    ON submission_media(submission_id)
    WHERE is_primary = TRUE;

COMMIT;

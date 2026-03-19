ALTER TABLE issue_submissions
    ADD COLUMN IF NOT EXISTS request_fingerprint CHAR(64) NOT NULL DEFAULT '';

ALTER TABLE issue_submissions
    ADD COLUMN IF NOT EXISTS created_issue BOOLEAN NOT NULL DEFAULT FALSE;

WITH ranked_submissions AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY issue_id
            ORDER BY reported_at ASC, created_at ASC, id ASC
        ) AS submission_rank
    FROM issue_submissions
)
UPDATE issue_submissions s
SET created_issue = (ranked_submissions.submission_rank = 1)
FROM ranked_submissions
WHERE ranked_submissions.id = s.id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'issue_submissions'::regclass
          AND conname = 'issue_submissions_client_request_id_key'
    ) THEN
        ALTER TABLE issue_submissions
            DROP CONSTRAINT issue_submissions_client_request_id_key;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_issue_submissions_device_client_request
    ON issue_submissions(device_id, client_request_id);

BEGIN;

CREATE TABLE IF NOT EXISTS report_upload_tickets (
    object_key TEXT PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    content_type VARCHAR(100) NOT NULL,
    size_bytes INT NOT NULL CHECK (size_bytes > 0),
    upload_mode VARCHAR(20) NOT NULL CHECK (upload_mode IN ('local', 'r2')),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_report_upload_tickets_device_issued_at
    ON report_upload_tickets(device_id, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_report_upload_tickets_issued_at
    ON report_upload_tickets(issued_at);

COMMIT;

ALTER TABLE issue_submissions
    ADD COLUMN IF NOT EXISTS district_name VARCHAR(120),
    ADD COLUMN IF NOT EXISTS regency_name VARCHAR(120),
    ADD COLUMN IF NOT EXISTS province_name VARCHAR(120);

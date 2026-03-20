#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${BACKEND_DIR}/.env"

if [[ -z "${DATABASE_URL:-}" && -f "${ENV_FILE}" ]]; then
    set -a
    # shellcheck disable=SC1090
    . "${ENV_FILE}"
    set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL is required (export it or set it in backend/.env)" >&2
    exit 1
fi

run_query() {
    local sql="$1"
    psql "${DATABASE_URL}" -X -v ON_ERROR_STOP=1 -At -F $'\t' -c "${sql}"
}

check_empty() {
    local label="$1"
    local sql="$2"
    local output
    output="$(run_query "${sql}")"
    if [[ -n "${output}" ]]; then
        echo "[FAIL] ${label}"
        printf '%s\n' "${output}"
        return 1
    fi
    echo "[OK] ${label}"
}

fail=0

check_empty "required extensions" "
WITH required(name) AS (
    VALUES ('postgis'), ('pgcrypto')
)
SELECT name
FROM required
WHERE NOT EXISTS (
    SELECT 1
    FROM pg_extension e
    WHERE e.extname = required.name
);
" || fail=1

check_empty "required tables and view" "
WITH required(name) AS (
    VALUES
        ('regions'),
        ('devices'),
        ('users'),
        ('oauth_accounts'),
        ('user_sessions'),
        ('device_consents'),
        ('issues'),
        ('issue_submissions'),
        ('submission_media'),
        ('report_upload_tickets'),
        ('issue_reactions'),
        ('issue_flags'),
        ('submission_flags'),
        ('issue_status_history'),
        ('issue_events'),
        ('issue_followers'),
        ('notifications'),
        ('push_subscriptions'),
        ('push_delivery_jobs'),
        ('follower_auth_bindings'),
        ('notification_preferences'),
        ('nearby_alert_subscriptions'),
        ('nearby_alert_deliveries'),
        ('moderation_actions'),
        ('issue_daily_stats'),
        ('issue_public_view')
)
SELECT name
FROM required
WHERE to_regclass('public.' || name) IS NULL;
" || fail=1

check_empty "required columns and types" "
WITH required(table_name, column_name, expected_type) AS (
    VALUES
        ('submission_media', 'width', 'integer'),
        ('submission_media', 'height', 'integer'),
        ('issue_events', 'id', 'bigint'),
        ('issue_submissions', 'district_name', 'character varying'),
        ('issue_submissions', 'regency_name', 'character varying'),
        ('issue_submissions', 'province_name', 'character varying'),
        ('notifications', 'event_id', 'bigint'),
        ('notification_preferences', 'notify_on_nearby_issue_created', 'boolean'),
        ('push_delivery_jobs', 'event_id', 'bigint'),
        ('push_delivery_jobs', 'attempt_count', 'integer'),
        ('report_upload_tickets', 'expires_at', 'timestamp with time zone')
)
SELECT r.table_name || '.' || r.column_name || ' expected=' || r.expected_type || ' actual=' || COALESCE(c.data_type, '<missing>')
FROM required r
LEFT JOIN information_schema.columns c
    ON c.table_schema = 'public'
   AND c.table_name = r.table_name
   AND c.column_name = r.column_name
WHERE c.data_type IS DISTINCT FROM r.expected_type;
" || fail=1

check_empty "required helper function" "
SELECT 'set_updated_at()'
WHERE to_regprocedure('public.set_updated_at()') IS NULL;
" || fail=1

check_empty "required indexes and constraints" "
WITH missing_named_objects AS (
    SELECT name
    FROM (
        VALUES
            ('idx_regions_geom'),
            ('idx_issues_public_location'),
            ('idx_issue_events_issue_id_created_at'),
            ('idx_issue_followers_issue_id'),
            ('idx_issue_followers_follower_id'),
            ('idx_notifications_follower_created_at'),
            ('idx_notifications_created_at'),
            ('idx_report_upload_tickets_device_issued_at'),
            ('idx_report_upload_tickets_issued_at'),
            ('idx_push_subscriptions_active_follower_id'),
            ('idx_push_subscriptions_disabled_at'),
            ('idx_push_subscriptions_active_updated_at'),
            ('idx_push_delivery_jobs_ready'),
            ('idx_push_delivery_jobs_delivered_at'),
            ('idx_push_delivery_jobs_failed_at'),
            ('idx_nearby_alert_subscriptions_geog')
    ) AS required(name)
    WHERE to_regclass('public.' || name) IS NULL
),
missing_semantic_constraints AS (
    SELECT 'issue_followers unique(issue_id, follower_id)' AS name
    WHERE NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_attribute a1 ON a1.attrelid = t.oid AND a1.attname = 'issue_id'
        JOIN pg_attribute a2 ON a2.attrelid = t.oid AND a2.attname = 'follower_id'
        WHERE t.relname = 'issue_followers'
          AND c.contype = 'u'
          AND c.conkey = ARRAY[a1.attnum, a2.attnum]::smallint[]
    )
    UNION ALL
    SELECT 'notifications unique(event_id, follower_id)' AS name
    WHERE NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_attribute a1 ON a1.attrelid = t.oid AND a1.attname = 'event_id'
        JOIN pg_attribute a2 ON a2.attrelid = t.oid AND a2.attname = 'follower_id'
        WHERE t.relname = 'notifications'
          AND c.contype = 'u'
          AND c.conkey = ARRAY[a1.attnum, a2.attnum]::smallint[]
    )
    UNION ALL
    SELECT 'nearby_alert_deliveries unique(subscription_id, issue_id)' AS name
    WHERE NOT EXISTS (
        SELECT 1
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_attribute a1 ON a1.attrelid = t.oid AND a1.attname = 'subscription_id'
        JOIN pg_attribute a2 ON a2.attrelid = t.oid AND a2.attname = 'issue_id'
        WHERE t.relname = 'nearby_alert_deliveries'
          AND c.contype = 'u'
          AND c.conkey = ARRAY[a1.attnum, a2.attnum]::smallint[]
    )
)
SELECT name FROM missing_named_objects
UNION ALL
SELECT name FROM missing_semantic_constraints;
" || fail=1

if [[ "${fail}" -ne 0 ]]; then
    echo "schema governance verification failed" >&2
    exit 1
fi

echo "schema governance verification passed"

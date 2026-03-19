#!/usr/bin/env bash

set -euo pipefail

if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL is required" >&2
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
        ('issue_reactions'),
        ('issue_flags'),
        ('submission_flags'),
        ('issue_status_history'),
        ('issue_events'),
        ('issue_followers'),
        ('notifications'),
        ('push_subscriptions'),
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
        ('notifications', 'event_id', 'bigint'),
        ('notification_preferences', 'notify_on_nearby_issue_created', 'boolean')
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
WITH required(name) AS (
    VALUES
        ('idx_regions_geom'),
        ('idx_issues_public_location'),
        ('idx_issue_events_issue_id_created_at'),
        ('idx_issue_followers_issue_id'),
        ('idx_issue_followers_follower_id'),
        ('uq_issue_followers_issue_follower'),
        ('idx_notifications_follower_created_at'),
        ('uq_notifications_event_follower'),
        ('idx_push_subscriptions_active_follower_id'),
        ('idx_nearby_alert_subscriptions_geog'),
        ('uq_nearby_alert_deliveries_subscription_issue')
)
SELECT name
FROM required
WHERE to_regclass('public.' || name) IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM pg_constraint c
      WHERE c.conname = required.name
  );
" || fail=1

if [[ "${fail}" -ne 0 ]]; then
    echo "schema governance verification failed" >&2
    exit 1
fi

echo "schema governance verification passed"

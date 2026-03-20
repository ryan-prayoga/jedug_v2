CREATE INDEX IF NOT EXISTS idx_notifications_created_at
    ON notifications(created_at);

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_disabled_at
    ON push_subscriptions(disabled_at)
    WHERE disabled_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_push_subscriptions_active_updated_at
    ON push_subscriptions(updated_at)
    WHERE disabled_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_push_delivery_jobs_delivered_at
    ON push_delivery_jobs(delivered_at)
    WHERE delivered_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_push_delivery_jobs_failed_at
    ON push_delivery_jobs(failed_at)
    WHERE failed_at IS NOT NULL;

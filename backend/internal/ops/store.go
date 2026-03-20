package ops

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const retentionAdvisoryLockKey int64 = 2026032003

type RetentionPolicy struct {
	NotificationsRetention             time.Duration
	PushSubscriptionsStaleAfter        time.Duration
	PushSubscriptionsDisabledRetention time.Duration
	PushDeliveryDeliveredRetention     time.Duration
	PushDeliveryFailedRetention        time.Duration
	UploadOrphanRetention              time.Duration
}

type CleanupSummary struct {
	Skipped                        bool
	NotificationsDeleted           int64
	PushSubscriptionsDisabled      int64
	PushSubscriptionsDeleted       int64
	PushDeliveriesDeliveredDeleted int64
	PushDeliveriesFailedDeleted    int64
	UploadOrphansDeleted           int64
}

type HealthSnapshot struct {
	PushReadyCount              int64
	PushFailedLast24H           int64
	NotificationsOverRetention  int64
	StalePushSubscriptions      int64
	DisabledPushSubscriptions   int64
	UploadOrphansOverRetention  int64
	IssueEventsEstimate         int64
	NotificationsEstimate       int64
	PushSubscriptionsEstimate   int64
	PushDeliveryJobsEstimate    int64
	ReportUploadTicketsEstimate int64
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) RunCleanup(ctx context.Context, policy RetentionPolicy) (*CleanupSummary, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var locked bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, retentionAdvisoryLockKey).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked {
		return &CleanupSummary{Skipped: true}, nil
	}

	now := time.Now().UTC()
	summary := &CleanupSummary{}

	if summary.NotificationsDeleted, err = execRowsAffected(ctx, tx, `
		DELETE FROM notifications
		WHERE created_at < $1
	`, now.Add(-policy.NotificationsRetention)); err != nil {
		return nil, err
	}

	if summary.PushSubscriptionsDisabled, err = execRowsAffected(ctx, tx, `
		UPDATE push_subscriptions
		SET disabled_at = NOW(),
		    updated_at = NOW()
		WHERE disabled_at IS NULL
		  AND updated_at < $1
	`, now.Add(-policy.PushSubscriptionsStaleAfter)); err != nil {
		return nil, err
	}

	if summary.PushSubscriptionsDeleted, err = execRowsAffected(ctx, tx, `
		DELETE FROM push_subscriptions
		WHERE disabled_at IS NOT NULL
		  AND disabled_at < $1
	`, now.Add(-policy.PushSubscriptionsDisabledRetention)); err != nil {
		return nil, err
	}

	if summary.PushDeliveriesDeliveredDeleted, err = execRowsAffected(ctx, tx, `
		DELETE FROM push_delivery_jobs
		WHERE delivered_at IS NOT NULL
		  AND delivered_at < $1
	`, now.Add(-policy.PushDeliveryDeliveredRetention)); err != nil {
		return nil, err
	}

	if summary.PushDeliveriesFailedDeleted, err = execRowsAffected(ctx, tx, `
		DELETE FROM push_delivery_jobs
		WHERE failed_at IS NOT NULL
		  AND failed_at < $1
	`, now.Add(-policy.PushDeliveryFailedRetention)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return summary, nil
}

type OrphanedUpload struct {
	ObjectKey  string
	UploadMode string
	IssuedAt   time.Time
}

func (s *Store) ListUploadOrphansBefore(ctx context.Context, cutoff time.Time, limit int) ([]OrphanedUpload, error) {
	rows, err := s.db.Query(ctx, `
		SELECT object_key, upload_mode, issued_at
		FROM report_upload_tickets
		WHERE issued_at < $1
		ORDER BY issued_at ASC
		LIMIT $2
	`, cutoff.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrphanedUpload, 0, limit)
	for rows.Next() {
		var item OrphanedUpload
		if err := rows.Scan(&item.ObjectKey, &item.UploadMode, &item.IssuedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) DeleteUploadOrphans(ctx context.Context, objectKeys []string) (int64, error) {
	if len(objectKeys) == 0 {
		return 0, nil
	}
	return execPoolRowsAffected(ctx, s.db, `
		DELETE FROM report_upload_tickets
		WHERE object_key = ANY($1::text[])
	`, objectKeys)
}

func (s *Store) HealthSnapshot(ctx context.Context, policy RetentionPolicy) (*HealthSnapshot, error) {
	now := time.Now().UTC()
	snapshot := &HealthSnapshot{}

	if err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM push_delivery_jobs WHERE delivered_at IS NULL AND failed_at IS NULL AND next_attempt_at <= NOW()),
			(SELECT COUNT(*) FROM push_delivery_jobs WHERE failed_at >= NOW() - INTERVAL '24 hours'),
			(SELECT COUNT(*) FROM notifications WHERE created_at < $1),
			(SELECT COUNT(*) FROM push_subscriptions WHERE disabled_at IS NULL AND updated_at < $2),
			(SELECT COUNT(*) FROM push_subscriptions WHERE disabled_at IS NOT NULL),
			(SELECT COUNT(*) FROM report_upload_tickets WHERE issued_at < $3)
	`, now.Add(-policy.NotificationsRetention), now.Add(-policy.PushSubscriptionsStaleAfter), now.Add(-policy.UploadOrphanRetention)).Scan(
		&snapshot.PushReadyCount,
		&snapshot.PushFailedLast24H,
		&snapshot.NotificationsOverRetention,
		&snapshot.StalePushSubscriptions,
		&snapshot.DisabledPushSubscriptions,
		&snapshot.UploadOrphansOverRetention,
	); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT relname, COALESCE(n_live_tup::bigint, 0)
		FROM pg_stat_user_tables
		WHERE schemaname = 'public'
		  AND relname = ANY($1::text[])
	`, []string{"issue_events", "notifications", "push_subscriptions", "push_delivery_jobs", "report_upload_tickets"})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name  string
			count int64
		)
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}

		switch name {
		case "issue_events":
			snapshot.IssueEventsEstimate = count
		case "notifications":
			snapshot.NotificationsEstimate = count
		case "push_subscriptions":
			snapshot.PushSubscriptionsEstimate = count
		case "push_delivery_jobs":
			snapshot.PushDeliveryJobsEstimate = count
		case "report_upload_tickets":
			snapshot.ReportUploadTicketsEstimate = count
		}
	}

	return snapshot, rows.Err()
}

func execRowsAffected(ctx context.Context, tx pgx.Tx, query string, args ...any) (int64, error) {
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func execPoolRowsAffected(ctx context.Context, db *pgxpool.Pool, query string, args ...any) (int64, error) {
	result, err := db.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

// notificationMessages maps event_type → [title, message] shown to followers.
var notificationMessages = map[string][2]string{
	"issue_created":     {"Laporan Baru Dibuat", "Laporan jalan rusak baru telah dibuat di area ini."},
	"photo_added":       {"Foto Baru Ditambahkan", "Foto baru ditambahkan ke laporan yang kamu ikuti."},
	"severity_changed":  {"Tingkat Keparahan Berubah", "Tingkat keparahan laporan yang kamu ikuti telah berubah."},
	"casualty_reported": {"Ada Korban Dilaporkan", "Ada korban yang dilaporkan pada laporan yang kamu ikuti."},
	"status_updated":    {"Status Laporan Diperbarui", "Status laporan yang kamu ikuti telah diperbarui."},
}

func notifTitleMessage(eventType string) (string, string) {
	if msg, ok := notificationMessages[eventType]; ok {
		return msg[0], msg[1]
	}
	return "Ada Pembaruan Laporan", "Ada pembaruan pada laporan yang kamu ikuti."
}

// DispatchNotificationsForEvent inserts a notification row for every follower of
// issueID, linked to the given eventID. It is idempotent via the unique constraint
// (event_id, follower_id). Callers must treat a non-nil error as non-fatal and log it.
// eventID is int64 because issue_events.id is BIGSERIAL in the production schema.
func DispatchNotificationsForEvent(ctx context.Context, db *pgxpool.Pool, issueID uuid.UUID, eventID int64, eventType string) error {
	title, message := notifTitleMessage(eventType)
	_, err := db.Exec(ctx, `
		INSERT INTO notifications (id, issue_id, follower_id, event_id, type, title, message)
		SELECT gen_random_uuid(), $1, follower_id, $2, $3, $4, $5
		FROM issue_followers
		WHERE issue_id = $1
		ON CONFLICT (event_id, follower_id) DO NOTHING
	`, issueID, eventID, eventType, title, message)
	return err
}

// NotificationRepository reads notification data for a given follower.
type NotificationRepository interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) error
}

type notificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, issue_id, event_id, type, title, message, created_at, read_at
		FROM notifications
		WHERE follower_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, followerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*domain.Notification, 0)
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.IssueID, &n.EventID, &n.Type, &n.Title, &n.Message, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		result = append(result, &n)
	}
	return result, rows.Err()
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE id = $1
		  AND follower_id = $2
	`, notificationID, followerID)
	return err
}

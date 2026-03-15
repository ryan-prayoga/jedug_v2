package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/sse"
)

func compactIssueLocationLabel(roadName, regionName *string, issueID uuid.UUID) string {
	if roadName != nil && strings.TrimSpace(*roadName) != "" {
		return strings.TrimSpace(*roadName)
	}
	if regionName != nil && strings.TrimSpace(*regionName) != "" {
		return strings.TrimSpace(*regionName)
	}
	shortID := issueID.String()
	if len(shortID) >= 8 {
		shortID = shortID[:8]
	}
	return "Issue #" + shortID
}

func notifTitleMessage(eventType, locationLabel string) (string, string) {
	switch eventType {
	case "issue_created":
		return "Laporan Baru Dibuat", "Laporan baru terdeteksi di " + locationLabel + "."
	case "photo_added":
		return "Foto Baru Ditambahkan", "Foto baru ditambahkan pada laporan di " + locationLabel + "."
	case "severity_changed":
		return "Tingkat Keparahan Berubah", "Keparahan laporan di " + locationLabel + " berubah."
	case "casualty_reported":
		return "Ada Korban Dilaporkan", "Ada laporan korban pada issue di " + locationLabel + "."
	case "status_updated":
		return "Status Laporan Diperbarui", "Status laporan di " + locationLabel + " telah diperbarui."
	default:
		return "Ada Pembaruan Laporan", "Ada pembaruan pada laporan di " + locationLabel + "."
	}
}

// ssePayload is the JSON shape sent in SSE notification events.
type ssePayload struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	EventID   int64     `json:"event_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// DispatchNotificationsForEvent inserts a notification row for every follower of
// issueID, linked to the given eventID. It is idempotent via the unique constraint
// (event_id, follower_id). After each successful insert it pushes a realtime SSE
// event to the follower via sse.Default. Callers must treat a non-nil error as
// non-fatal and log it.
// eventID is int64 because issue_events.id is BIGSERIAL in the production schema.
func DispatchNotificationsForEvent(ctx context.Context, db *pgxpool.Pool, issueID uuid.UUID, eventID int64, eventType string, excludeFollowerID *uuid.UUID) error {
	var (
		roadName   *string
		regionName *string
	)
	if queryErr := db.QueryRow(ctx, `
		SELECT i.road_name,
		       COALESCE(NULLIF(TRIM(r.name), ''), NULLIF(TRIM(sr.name), '')) AS region_name
		FROM issues i
		LEFT JOIN regions r ON r.id = i.region_id
		LEFT JOIN LATERAL (
			SELECT s.region_id
			FROM issue_submissions s
			WHERE s.issue_id = i.id
			  AND s.region_id IS NOT NULL
			ORDER BY s.reported_at DESC
			LIMIT 1
		) latest_sub ON TRUE
		LEFT JOIN regions sr ON sr.id = latest_sub.region_id
		WHERE i.id = $1
	`, issueID).Scan(&roadName, &regionName); queryErr != nil {
		roadName = nil
		regionName = nil
	}

	locationLabel := compactIssueLocationLabel(roadName, regionName, issueID)
	title, message := notifTitleMessage(eventType, locationLabel)
	rows, err := db.Query(ctx, `
		INSERT INTO notifications (id, issue_id, follower_id, event_id, type, title, message)
		SELECT gen_random_uuid(), $1, follower_id, $2, $3, $4, $5
		FROM issue_followers
		WHERE issue_id = $1
		  AND ($6::uuid IS NULL OR follower_id <> $6::uuid)
		ON CONFLICT (event_id, follower_id) DO NOTHING
		RETURNING id, issue_id, follower_id, type, title, message, created_at
	`, issueID, eventID, eventType, title, message, excludeFollowerID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			notifID    uuid.UUID
			issID      uuid.UUID
			followerID uuid.UUID
			notifType  string
			notifTitle string
			notifMsg   string
			createdAt  time.Time
		)
		if scanErr := rows.Scan(&notifID, &issID, &followerID, &notifType, &notifTitle, &notifMsg, &createdAt); scanErr != nil {
			continue
		}
		payload, jsonErr := json.Marshal(ssePayload{
			ID:        notifID.String(),
			IssueID:   issID.String(),
			EventID:   eventID,
			Type:      notifType,
			Title:     notifTitle,
			Message:   notifMsg,
			CreatedAt: createdAt,
		})
		if jsonErr != nil {
			continue
		}
		sse.Default.Push(followerID.String(), "event: notification\ndata: "+string(payload)+"\n\n")
	}
	return rows.Err()
}

// NotificationRepository reads notification data for a given follower.
type NotificationRepository interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) (*time.Time, bool, error)
	Delete(ctx context.Context, notificationID, followerID uuid.UUID) (bool, error)
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

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) (*time.Time, bool, error) {
	var readAt time.Time
	err := r.db.QueryRow(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE id = $1
		  AND follower_id = $2
		RETURNING read_at
	`, notificationID, followerID).Scan(&readAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &readAt, true, nil
}

func (r *notificationRepository) Delete(ctx context.Context, notificationID, followerID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM notifications
		WHERE id = $1
		  AND follower_id = $2
	`, notificationID, followerID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/sse"
)

type PushDelivery struct {
	FollowerID uuid.UUID
	IssueID    uuid.UUID
	EventID    int64
	Type       string
	Title      string
	Message    string
}

type NotificationPushNotifier interface {
	DeliverBatch(ctx context.Context, deliveries []PushDelivery) error
}

type dispatchTarget struct {
	FollowerID uuid.UUID
	SendInApp  bool
	SendPush   bool
}

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
	case "nearby_issue_created":
		return "Laporan Baru di Area Pantauan", "Ada laporan baru di sekitar area pantauanmu: " + locationLabel + "."
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
func DispatchNotificationsForEvent(
	ctx context.Context,
	db *pgxpool.Pool,
	pushNotifier NotificationPushNotifier,
	issueID uuid.UUID,
	eventID int64,
	eventType string,
	excludeFollowerID *uuid.UUID,
) error {
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
	targets, err := listDispatchTargets(ctx, db, issueID, eventType, excludeFollowerID)
	if err != nil {
		return err
	}

	inAppFollowerIDs := make([]uuid.UUID, 0, len(targets))
	pushDeliveries := make([]PushDelivery, 0, len(targets))
	for _, target := range targets {
		if target.SendInApp {
			inAppFollowerIDs = append(inAppFollowerIDs, target.FollowerID)
		}
		if target.SendPush {
			pushDeliveries = append(pushDeliveries, PushDelivery{
				FollowerID: target.FollowerID,
				IssueID:    issueID,
				EventID:    eventID,
				Type:       eventType,
				Title:      title,
				Message:    message,
			})
		}
	}

	if len(inAppFollowerIDs) > 0 {
		rows, err := db.Query(ctx, `
			INSERT INTO notifications (id, issue_id, follower_id, event_id, type, title, message)
			SELECT gen_random_uuid(), $1, follower_id, $2, $3, $4, $5
			FROM unnest($6::uuid[]) AS targets(follower_id)
			ON CONFLICT (event_id, follower_id) DO NOTHING
			RETURNING id, issue_id, follower_id, type, title, message, created_at
		`, issueID, eventID, eventType, title, message, inAppFollowerIDs)
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
			sse.Default.Push(
				followerID.String(),
				sse.FormatEvent("notification", payload, strconv.FormatInt(eventID, 10)),
			)
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	if pushNotifier != nil && len(pushDeliveries) > 0 {
		if err := pushNotifier.DeliverBatch(ctx, pushDeliveries); err != nil {
			log.Printf("[NOTIFICATION] push_delivery_error issue=%s event=%d error=%v", issueID, eventID, err)
		}
	}

	return nil
}

func listDispatchTargets(
	ctx context.Context,
	db *pgxpool.Pool,
	issueID uuid.UUID,
	eventType string,
	excludeFollowerID *uuid.UUID,
) ([]dispatchTarget, error) {
	eventPreferenceExpr := notificationEventPreferenceExpr(eventType)
	query := fmt.Sprintf(`
		SELECT f.follower_id,
			(COALESCE(p.notifications_enabled, TRUE)
				AND %s
				AND COALESCE(p.in_app_enabled, TRUE)) AS send_in_app,
			(COALESCE(p.notifications_enabled, TRUE)
				AND %s
				AND COALESCE(p.push_enabled, TRUE)) AS send_push
		FROM issue_followers f
		LEFT JOIN notification_preferences p ON p.follower_id = f.follower_id
		WHERE f.issue_id = $1
		  AND ($2::uuid IS NULL OR f.follower_id <> $2::uuid)
	`, eventPreferenceExpr, eventPreferenceExpr)

	rows, err := db.Query(ctx, query, issueID, excludeFollowerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]dispatchTarget, 0)
	for rows.Next() {
		var target dispatchTarget
		if err := rows.Scan(&target.FollowerID, &target.SendInApp, &target.SendPush); err != nil {
			return nil, err
		}
		if !target.SendInApp && !target.SendPush {
			continue
		}
		result = append(result, target)
	}

	return result, rows.Err()
}

func notificationEventPreferenceExpr(eventType string) string {
	switch eventType {
	case "photo_added":
		return "COALESCE(p.notify_on_photo_added, TRUE)"
	case "status_updated":
		return "COALESCE(p.notify_on_status_updated, TRUE)"
	case "severity_changed":
		return "COALESCE(p.notify_on_severity_changed, TRUE)"
	case "casualty_reported":
		return "COALESCE(p.notify_on_casualty_reported, TRUE)"
	case "nearby_issue_created":
		return "COALESCE(p.notify_on_nearby_issue_created, TRUE)"
	default:
		return "TRUE"
	}
}

// NotificationRepository reads notification data for a given follower.
type NotificationRepository interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error)
	GetByFollowerIDSinceEventID(ctx context.Context, followerID uuid.UUID, afterEventID int64, limit int) ([]*domain.Notification, error)
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

	return scanNotifications(rows)
}

func (r *notificationRepository) GetByFollowerIDSinceEventID(ctx context.Context, followerID uuid.UUID, afterEventID int64, limit int) ([]*domain.Notification, error) {
	if afterEventID <= 0 {
		return []*domain.Notification{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, issue_id, event_id, type, title, message, created_at, read_at
		FROM notifications
		WHERE follower_id = $1
		  AND event_id > $2
		ORDER BY event_id ASC, created_at ASC
		LIMIT $3
	`, followerID, afterEventID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func scanNotifications(rows pgx.Rows) ([]*domain.Notification, error) {
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

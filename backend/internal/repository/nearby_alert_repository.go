package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/sse"
)

type NearbyAlertCreateInput struct {
	FollowerID uuid.UUID
	Label      *string
	Latitude   float64
	Longitude  float64
	RadiusM    int
	Enabled    bool
}

type NearbyAlertRepository interface {
	ListByFollowerID(ctx context.Context, followerID uuid.UUID) ([]*domain.NearbyAlertSubscription, error)
	CountByFollowerID(ctx context.Context, followerID uuid.UUID) (int, error)
	Create(ctx context.Context, input NearbyAlertCreateInput) (*domain.NearbyAlertSubscription, error)
	Update(ctx context.Context, followerID, subscriptionID uuid.UUID, patch domain.NearbyAlertSubscriptionPatch) (*domain.NearbyAlertSubscription, error)
	Delete(ctx context.Context, followerID, subscriptionID uuid.UUID) (bool, error)
}

type nearbyAlertRepository struct {
	db *pgxpool.Pool
}

func NewNearbyAlertRepository(db *pgxpool.Pool) NearbyAlertRepository {
	return &nearbyAlertRepository{db: db}
}

func (r *nearbyAlertRepository) ListByFollowerID(ctx context.Context, followerID uuid.UUID) ([]*domain.NearbyAlertSubscription, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, follower_id, label, latitude, longitude, radius_m, enabled, created_at, updated_at
		FROM nearby_alert_subscriptions
		WHERE follower_id = $1
		ORDER BY updated_at DESC, created_at DESC
	`, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*domain.NearbyAlertSubscription, 0)
	for rows.Next() {
		item, scanErr := scanNearbyAlertSubscription(
			rows.Scan,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *nearbyAlertRepository) CountByFollowerID(ctx context.Context, followerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM nearby_alert_subscriptions
		WHERE follower_id = $1
	`, followerID).Scan(&count)
	return count, err
}

func (r *nearbyAlertRepository) Create(ctx context.Context, input NearbyAlertCreateInput) (*domain.NearbyAlertSubscription, error) {
	return scanNearbyAlertSubscription(r.db.QueryRow(ctx, `
		INSERT INTO nearby_alert_subscriptions (
			follower_id,
			label,
			latitude,
			longitude,
			radius_m,
			enabled
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, follower_id, label, latitude, longitude, radius_m, enabled, created_at, updated_at
	`, input.FollowerID, nullableTrimmedLabel(input.Label), input.Latitude, input.Longitude, input.RadiusM, input.Enabled).Scan)
}

func (r *nearbyAlertRepository) Update(ctx context.Context, followerID, subscriptionID uuid.UUID, patch domain.NearbyAlertSubscriptionPatch) (*domain.NearbyAlertSubscription, error) {
	labelProvided := patch.Label != nil
	labelValue := ""
	if patch.Label != nil {
		labelValue = strings.TrimSpace(*patch.Label)
	}

	row := r.db.QueryRow(ctx, `
		UPDATE nearby_alert_subscriptions
		SET label = CASE
				WHEN $3::boolean THEN NULLIF(BTRIM($4::text), '')
				ELSE label
			END,
			latitude = COALESCE($5, latitude),
			longitude = COALESCE($6, longitude),
			radius_m = COALESCE($7, radius_m),
			enabled = COALESCE($8, enabled),
			updated_at = NOW()
		WHERE follower_id = $1
		  AND id = $2
		RETURNING id, follower_id, label, latitude, longitude, radius_m, enabled, created_at, updated_at
	`,
		followerID,
		subscriptionID,
		labelProvided,
		labelValue,
		nullableFloat64Value(patch.Latitude),
		nullableFloat64Value(patch.Longitude),
		nullableIntValue(patch.RadiusM),
		nullableBoolValue(patch.Enabled),
	)

	item, err := scanNearbyAlertSubscription(row.Scan)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *nearbyAlertRepository) Delete(ctx context.Context, followerID, subscriptionID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM nearby_alert_subscriptions
		WHERE follower_id = $1
		  AND id = $2
	`, followerID, subscriptionID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

type nearbyAlertDispatchTarget struct {
	FollowerID  uuid.UUID
	Label       *string
	DistanceM   float64
	SendInApp   bool
	SendPush    bool
	MatchCount  int
	IssueID     uuid.UUID
	Title       string
	Message     string
}

func DispatchNearbyAlertsForIssueCreated(
	ctx context.Context,
	db *pgxpool.Pool,
	pushNotifier NotificationPushNotifier,
	issueID uuid.UUID,
	eventID int64,
	excludeFollowerID *uuid.UUID,
) error {
	issueLocation := loadNotificationIssueLocationLabel(ctx, db, issueID)
	targets, err := listNearbyAlertDispatchTargets(ctx, db, issueID, excludeFollowerID)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	inAppFollowerIDs := make([]uuid.UUID, 0, len(targets))
	inAppTitles := make([]string, 0, len(targets))
	inAppMessages := make([]string, 0, len(targets))
	pushDeliveries := make([]PushDelivery, 0, len(targets))

	for _, target := range targets {
		title, message := nearbyAlertTitleMessage(issueLocation, target.Label, target.MatchCount)
		if target.SendInApp {
			inAppFollowerIDs = append(inAppFollowerIDs, target.FollowerID)
			inAppTitles = append(inAppTitles, title)
			inAppMessages = append(inAppMessages, message)
		}
		if target.SendPush {
			pushDeliveries = append(pushDeliveries, PushDelivery{
				FollowerID: target.FollowerID,
				IssueID:    issueID,
				Type:       "nearby_issue_created",
				Title:      title,
				Message:    message,
			})
		}
	}

	if len(inAppFollowerIDs) > 0 {
		rows, err := db.Query(ctx, `
			INSERT INTO notifications (id, issue_id, follower_id, event_id, type, title, message)
			SELECT gen_random_uuid(), $1, targets.follower_id, $2, 'nearby_issue_created', targets.title, targets.message
			FROM unnest($3::uuid[], $4::text[], $5::text[]) AS targets(follower_id, title, message)
			ON CONFLICT (event_id, follower_id) DO NOTHING
			RETURNING id, issue_id, follower_id, type, title, message, created_at
		`, issueID, eventID, inAppFollowerIDs, inAppTitles, inAppMessages)
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
		if err := rows.Err(); err != nil {
			return err
		}
	}

	if pushNotifier != nil && len(pushDeliveries) > 0 {
		if err := pushNotifier.DeliverBatch(ctx, pushDeliveries); err != nil {
			return fmt.Errorf("deliver nearby alert push: %w", err)
		}
	}

	return nil
}

func listNearbyAlertDispatchTargets(
	ctx context.Context,
	db *pgxpool.Pool,
	issueID uuid.UUID,
	excludeFollowerID *uuid.UUID,
) ([]nearbyAlertDispatchTarget, error) {
	rows, err := db.Query(ctx, `
		WITH issue_geo AS (
			SELECT public_location
			FROM issues
			WHERE id = $1
		),
		matched AS (
			SELECT s.id AS subscription_id,
				s.follower_id,
				NULLIF(BTRIM(s.label), '') AS label,
				ST_Distance(
					ST_SetSRID(ST_MakePoint(s.longitude, s.latitude), 4326)::geography,
					issue_geo.public_location
				) AS distance_m,
				(COALESCE(p.notifications_enabled, TRUE)
					AND COALESCE(p.notify_on_nearby_issue_created, TRUE)
					AND COALESCE(p.in_app_enabled, TRUE)) AS send_in_app,
				(COALESCE(p.notifications_enabled, TRUE)
					AND COALESCE(p.notify_on_nearby_issue_created, TRUE)
					AND COALESCE(p.push_enabled, TRUE)) AS send_push
			FROM nearby_alert_subscriptions s
			CROSS JOIN issue_geo
			LEFT JOIN notification_preferences p ON p.follower_id = s.follower_id
			WHERE s.enabled = TRUE
			  AND ($2::uuid IS NULL OR s.follower_id <> $2::uuid)
			  AND ST_DWithin(
				ST_SetSRID(ST_MakePoint(s.longitude, s.latitude), 4326)::geography,
				issue_geo.public_location,
				s.radius_m
			  )
		),
		inserted AS (
			INSERT INTO nearby_alert_deliveries (subscription_id, follower_id, issue_id)
			SELECT subscription_id, follower_id, $1
			FROM matched
			ON CONFLICT (subscription_id, issue_id) DO NOTHING
			RETURNING subscription_id, follower_id
		)
		SELECT inserted.subscription_id,
			inserted.follower_id,
			matched.label,
			matched.distance_m,
			matched.send_in_app,
			matched.send_push
		FROM inserted
		JOIN matched USING (subscription_id, follower_id)
		ORDER BY inserted.follower_id, matched.distance_m ASC, inserted.subscription_id
	`, issueID, excludeFollowerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[uuid.UUID]*nearbyAlertDispatchTarget)
	order := make([]uuid.UUID, 0)
	for rows.Next() {
		var (
			subscriptionID uuid.UUID
			followerID    uuid.UUID
			label         *string
			distanceM     float64
			sendInApp     bool
			sendPush      bool
		)
		if err := rows.Scan(&subscriptionID, &followerID, &label, &distanceM, &sendInApp, &sendPush); err != nil {
			return nil, err
		}
		current, ok := grouped[followerID]
		if !ok {
			grouped[followerID] = &nearbyAlertDispatchTarget{
				FollowerID: followerID,
				Label:      label,
				DistanceM:  distanceM,
				SendInApp:  sendInApp,
				SendPush:   sendPush,
				MatchCount: 1,
			}
			order = append(order, followerID)
			continue
		}
		current.MatchCount++
		current.SendInApp = current.SendInApp || sendInApp
		current.SendPush = current.SendPush || sendPush
		if distanceM < current.DistanceM {
			current.DistanceM = distanceM
			current.Label = label
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]nearbyAlertDispatchTarget, 0, len(order))
	for _, followerID := range order {
		result = append(result, *grouped[followerID])
	}
	return result, nil
}

func nearbyAlertTitleMessage(issueLocation string, watchedLabel *string, matchCount int) (string, string) {
	label := ""
	if watchedLabel != nil {
		label = strings.TrimSpace(*watchedLabel)
	}
	title := "Laporan Baru di Area Pantauan"
	message := "Ada laporan baru di sekitar lokasi yang kamu pantau"
	if label != "" {
		title = "Laporan Baru di Sekitar " + label
		message = "Ada laporan baru di sekitar " + label
	}
	if issueLocation != "" {
		message += " — " + issueLocation
	}
	if matchCount > 1 {
		message += fmt.Sprintf(". Juga cocok dengan %d area pantauan lain.", matchCount-1)
		return title, message
	}
	return title, message + "."
}

func loadNotificationIssueLocationLabel(ctx context.Context, db *pgxpool.Pool, issueID uuid.UUID) string {
	var (
		roadName   *string
		regionName *string
	)
	_ = db.QueryRow(ctx, `
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
	`, issueID).Scan(&roadName, &regionName)
	return compactIssueLocationLabel(roadName, regionName, issueID)
}

func scanNearbyAlertSubscription(scan func(dest ...any) error) (*domain.NearbyAlertSubscription, error) {
	var item domain.NearbyAlertSubscription
	if err := scan(
		&item.ID,
		&item.FollowerID,
		&item.Label,
		&item.Latitude,
		&item.Longitude,
		&item.RadiusM,
		&item.Enabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func nullableTrimmedLabel(label *string) any {
	if label == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*label)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableFloat64Value(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableIntValue(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
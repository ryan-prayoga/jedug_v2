package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type NotificationPreferencesRepository interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.NotificationPreferences, error)
	Update(ctx context.Context, followerID uuid.UUID, patch domain.NotificationPreferencesPatch) (*domain.NotificationPreferences, error)
}

type notificationPreferencesRepository struct {
	db *pgxpool.Pool
}

func NewNotificationPreferencesRepository(db *pgxpool.Pool) NotificationPreferencesRepository {
	return &notificationPreferencesRepository{db: db}
}

func (r *notificationPreferencesRepository) GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.NotificationPreferences, error) {
	prefs, err := r.findByFollowerID(ctx, followerID)
	if err != nil {
		return nil, err
	}
	if prefs != nil {
		return prefs, nil
	}

	pushEnabledDefault, err := r.hasActivePushSubscription(ctx, followerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &domain.NotificationPreferences{
		FollowerID:               followerID,
		NotificationsEnabled:     true,
		InAppEnabled:             true,
		PushEnabled:              pushEnabledDefault,
		NotifyOnPhotoAdded:       true,
		NotifyOnStatusUpdated:    true,
		NotifyOnSeverityChanged:  true,
		NotifyOnCasualtyReported: true,
		CreatedAt:                now,
		UpdatedAt:                now,
	}, nil
}

func (r *notificationPreferencesRepository) Update(ctx context.Context, followerID uuid.UUID, patch domain.NotificationPreferencesPatch) (*domain.NotificationPreferences, error) {
	if err := r.ensureDefaults(ctx, followerID); err != nil {
		return nil, err
	}

	var prefs domain.NotificationPreferences
	err := r.db.QueryRow(ctx, `
		UPDATE notification_preferences
		SET notifications_enabled = COALESCE($2, notifications_enabled),
			in_app_enabled = COALESCE($3, in_app_enabled),
			push_enabled = COALESCE($4, push_enabled),
			notify_on_photo_added = COALESCE($5, notify_on_photo_added),
			notify_on_status_updated = COALESCE($6, notify_on_status_updated),
			notify_on_severity_changed = COALESCE($7, notify_on_severity_changed),
			notify_on_casualty_reported = COALESCE($8, notify_on_casualty_reported),
			updated_at = NOW()
		WHERE follower_id = $1
		RETURNING follower_id,
			notifications_enabled,
			in_app_enabled,
			push_enabled,
			notify_on_photo_added,
			notify_on_status_updated,
			notify_on_severity_changed,
			notify_on_casualty_reported,
			created_at,
			updated_at
	`,
		followerID,
		nullableBoolValue(patch.NotificationsEnabled),
		nullableBoolValue(patch.InAppEnabled),
		nullableBoolValue(patch.PushEnabled),
		nullableBoolValue(patch.NotifyOnPhotoAdded),
		nullableBoolValue(patch.NotifyOnStatusUpdated),
		nullableBoolValue(patch.NotifyOnSeverityChanged),
		nullableBoolValue(patch.NotifyOnCasualtyReported),
	).Scan(
		&prefs.FollowerID,
		&prefs.NotificationsEnabled,
		&prefs.InAppEnabled,
		&prefs.PushEnabled,
		&prefs.NotifyOnPhotoAdded,
		&prefs.NotifyOnStatusUpdated,
		&prefs.NotifyOnSeverityChanged,
		&prefs.NotifyOnCasualtyReported,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &prefs, nil
}

func (r *notificationPreferencesRepository) ensureDefaults(ctx context.Context, followerID uuid.UUID) error {
	pushEnabledDefault, err := r.hasActivePushSubscription(ctx, followerID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO notification_preferences (
			follower_id,
			push_enabled
		)
		VALUES ($1, $2)
		ON CONFLICT (follower_id) DO NOTHING
	`, followerID, pushEnabledDefault)
	return err
}

func (r *notificationPreferencesRepository) hasActivePushSubscription(ctx context.Context, followerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM push_subscriptions
			WHERE follower_id = $1
			  AND disabled_at IS NULL
		)
	`, followerID).Scan(&exists)
	return exists, err
}

func (r *notificationPreferencesRepository) findByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.NotificationPreferences, error) {
	var prefs domain.NotificationPreferences
	err := r.db.QueryRow(ctx, `
		SELECT follower_id,
			notifications_enabled,
			in_app_enabled,
			push_enabled,
			notify_on_photo_added,
			notify_on_status_updated,
			notify_on_severity_changed,
			notify_on_casualty_reported,
			created_at,
			updated_at
		FROM notification_preferences
		WHERE follower_id = $1
	`, followerID).Scan(
		&prefs.FollowerID,
		&prefs.NotificationsEnabled,
		&prefs.InAppEnabled,
		&prefs.PushEnabled,
		&prefs.NotifyOnPhotoAdded,
		&prefs.NotifyOnStatusUpdated,
		&prefs.NotifyOnSeverityChanged,
		&prefs.NotifyOnCasualtyReported,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &prefs, nil
}

func nullableBoolValue(value *bool) any {
	if value == nil {
		return nil
	}
	return *value
}

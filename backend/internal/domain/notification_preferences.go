package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationPreferences struct {
	FollowerID               uuid.UUID `json:"follower_id"`
	NotificationsEnabled     bool      `json:"notifications_enabled"`
	InAppEnabled             bool      `json:"in_app_enabled"`
	PushEnabled              bool      `json:"push_enabled"`
	NotifyOnPhotoAdded       bool      `json:"notify_on_photo_added"`
	NotifyOnStatusUpdated    bool      `json:"notify_on_status_updated"`
	NotifyOnSeverityChanged  bool      `json:"notify_on_severity_changed"`
	NotifyOnCasualtyReported bool      `json:"notify_on_casualty_reported"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type NotificationPreferencesPatch struct {
	NotificationsEnabled     *bool
	InAppEnabled             *bool
	PushEnabled              *bool
	NotifyOnPhotoAdded       *bool
	NotifyOnStatusUpdated    *bool
	NotifyOnSeverityChanged  *bool
	NotifyOnCasualtyReported *bool
}

func (p NotificationPreferencesPatch) IsEmpty() bool {
	return p.NotificationsEnabled == nil &&
		p.InAppEnabled == nil &&
		p.PushEnabled == nil &&
		p.NotifyOnPhotoAdded == nil &&
		p.NotifyOnStatusUpdated == nil &&
		p.NotifyOnSeverityChanged == nil &&
		p.NotifyOnCasualtyReported == nil
}

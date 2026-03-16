package domain

import (
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID         uuid.UUID  `json:"id"`
	FollowerID uuid.UUID  `json:"follower_id"`
	Endpoint   string     `json:"endpoint"`
	P256DH     string     `json:"p256dh"`
	Auth       string     `json:"auth"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DisabledAt *time.Time `json:"disabled_at,omitempty"`
}

type PushSubscriptionStatus struct {
	Enabled           bool   `json:"enabled"`
	Subscribed        bool   `json:"subscribed"`
	SubscriptionCount int    `json:"subscription_count"`
	VAPIDPublicKey    string `json:"vapid_public_key,omitempty"`
}

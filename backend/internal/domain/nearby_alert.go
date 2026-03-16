package domain

import (
	"time"

	"github.com/google/uuid"
)

type NearbyAlertSubscription struct {
	ID         uuid.UUID `json:"id"`
	FollowerID uuid.UUID `json:"follower_id"`
	Label      *string   `json:"label"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	RadiusM    int       `json:"radius_m"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type NearbyAlertSubscriptionPatch struct {
	Label     *string
	Latitude  *float64
	Longitude *float64
	RadiusM   *int
	Enabled   *bool
}

func (p NearbyAlertSubscriptionPatch) IsEmpty() bool {
	return p.Label == nil &&
		p.Latitude == nil &&
		p.Longitude == nil &&
		p.RadiusM == nil &&
		p.Enabled == nil
}
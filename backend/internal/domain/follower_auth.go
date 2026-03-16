package domain

import (
	"time"

	"github.com/google/uuid"
)

type FollowerAuthBinding struct {
	FollowerID      uuid.UUID `json:"follower_id"`
	DeviceTokenHash string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type FollowerAuthToken struct {
	FollowerID string    `json:"follower_id"`
	Token      string    `json:"follower_token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// Device maps to the devices table.
// AnonTokenHash is the SHA-256 hex of the raw token; raw token is never persisted.
type Device struct {
	ID            uuid.UUID `json:"id"`
	AnonTokenHash string    `json:"-"` // never expose
	TrustScore    int       `json:"trust_score"`
	IsBanned      bool      `json:"is_banned"`
	LastIP        *string   `json:"-"`
	LastUserAgent *string   `json:"-"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DeviceConsent maps to the device_consents table.
// id is BIGSERIAL, so it is not included here.
type DeviceConsent struct {
	DeviceID       uuid.UUID `json:"device_id"`
	TermsVersion   string    `json:"terms_version"`
	PrivacyVersion *string   `json:"privacy_version,omitempty"`
	IPAddress      *string   `json:"-"`
	UserAgent      *string   `json:"-"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// MediaItem is returned in issue detail responses.
// ObjectKey is the raw storage key; PublicURL is computed by the handler via the storage driver.
type MediaItem struct {
	ID        uuid.UUID `json:"id"`
	ObjectKey string    `json:"object_key"`
	PublicURL string    `json:"public_url"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int       `json:"size_bytes"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	Blurhash  *string   `json:"blurhash,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// SubmissionSummary is a lightweight submission view used in issue detail responses.
type SubmissionSummary struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	Severity    int       `json:"severity"`
	HasCasualty bool      `json:"has_casualty"`
	Note        *string   `json:"note,omitempty"`
	ReportedAt  time.Time `json:"reported_at"`
}

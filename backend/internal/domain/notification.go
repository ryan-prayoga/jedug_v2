package domain

import (
	"time"

	"github.com/google/uuid"
)

// Notification is a single update entry for a follower derived from an issue event.
// EventID is int64 because issue_events.id is BIGSERIAL in the production schema.
type Notification struct {
	ID        uuid.UUID  `json:"id"`
	IssueID   uuid.UUID  `json:"issue_id"`
	EventID   int64      `json:"event_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

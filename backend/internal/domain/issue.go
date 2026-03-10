package domain

import (
	"time"

	"github.com/google/uuid"
)

// Issue maps to the issues table.
// public_location is stored as GEOGRAPHY(POINT); returned as lon/lat floats.
type Issue struct {
	ID                 uuid.UUID `json:"id"`
	Status             string    `json:"status"`
	VerificationStatus string    `json:"verification_status"`
	SeverityCurrent    int       `json:"severity_current"`
	SeverityMax        int       `json:"severity_max"`
	Longitude          float64   `json:"longitude"`
	Latitude           float64   `json:"latitude"`
	RegionID           *int64    `json:"region_id,omitempty"`
	RegionName         *string   `json:"region_name,omitempty"`
	RoadName           *string   `json:"road_name,omitempty"`
	RoadType           *string   `json:"road_type,omitempty"`
	SubmissionCount    int       `json:"submission_count"`
	PhotoCount         int       `json:"photo_count"`
	CasualtyCount      int       `json:"casualty_count"`
	ReactionCount      int       `json:"reaction_count"`
	FlagCount          int       `json:"flag_count"`
	FirstSeenAt        time.Time `json:"first_seen_at"`
	LastSeenAt         time.Time `json:"last_seen_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// IssueDetail is returned by the GET /api/v1/issues/:id endpoint.
type IssueDetail struct {
	*Issue
	PrimaryMedia      *MediaItem           `json:"primary_media,omitempty"`
	PublicNote        *string              `json:"public_note,omitempty"`
	Media             []*MediaItem         `json:"media"`
	RecentSubmissions []*SubmissionSummary `json:"recent_submissions"`
}

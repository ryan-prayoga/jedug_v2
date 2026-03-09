package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminIssue is the admin view of an issue, including is_hidden.
type AdminIssue struct {
	ID                 uuid.UUID `json:"id"`
	Status             string    `json:"status"`
	VerificationStatus string    `json:"verification_status"`
	SeverityCurrent    int       `json:"severity_current"`
	SeverityMax        int       `json:"severity_max"`
	Longitude          float64   `json:"longitude"`
	Latitude           float64   `json:"latitude"`
	RegionID           *int64    `json:"region_id,omitempty"`
	RoadName           *string   `json:"road_name,omitempty"`
	RoadType           *string   `json:"road_type,omitempty"`
	SubmissionCount    int       `json:"submission_count"`
	PhotoCount         int       `json:"photo_count"`
	CasualtyCount      int       `json:"casualty_count"`
	ReactionCount      int       `json:"reaction_count"`
	FlagCount          int       `json:"flag_count"`
	IsHidden           bool      `json:"is_hidden"`
	FirstSeenAt        time.Time `json:"first_seen_at"`
	LastSeenAt         time.Time `json:"last_seen_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AdminIssueDetail is the admin view of issue detail.
type AdminIssueDetail struct {
	*AdminIssue
	Media         []*MediaItem              `json:"media"`
	Submissions   []*AdminSubmissionSummary  `json:"submissions"`
	ModerationLog []*ModerationAction        `json:"moderation_log"`
}

// AdminSubmissionSummary extends the public submission summary with device info.
type AdminSubmissionSummary struct {
	ID             uuid.UUID `json:"id"`
	DeviceID       uuid.UUID `json:"device_id"`
	DeviceIsBanned bool      `json:"device_is_banned"`
	Status         string    `json:"status"`
	Severity       int       `json:"severity"`
	HasCasualty    bool      `json:"has_casualty"`
	Note           *string   `json:"note,omitempty"`
	ReportedAt     time.Time `json:"reported_at"`
}

// ModerationAction is an audit log entry for moderation actions.
// ID is BIGSERIAL in the database.
type ModerationAction struct {
	ID            int64     `json:"id"`
	ActionType    string    `json:"action_type"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	AdminUsername *string   `json:"admin_username,omitempty"`
	Note          *string   `json:"note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type PublicStatsQuery struct {
	ProvinceID *int64
	RegencyID  *int64
}

// PublicStats is the response payload for GET /api/v1/stats.
// All fields are derived from public-safe issue data only.
type PublicStats struct {
	Global      PublicGlobalStats   `json:"global"`
	Status      PublicStatusStats   `json:"status"`
	Time        PublicTimeStats     `json:"time"`
	Filters     PublicStatsFilters  `json:"filters"`
	Regions     []*PublicRegionStat `json:"regions"`
	TopIssues   []*PublicTopIssue   `json:"top_issues"`
	GeneratedAt time.Time           `json:"generated_at"`
}

type PublicGlobalStats struct {
	TotalIssues         int64 `json:"total_issues"`
	TotalIssuesThisWeek int64 `json:"total_issues_this_week"`
	TotalCasualties     int64 `json:"total_casualties"`
	TotalPhotos         int64 `json:"total_photos"`
	TotalReports        int64 `json:"total_reports"`
}

type PublicStatusStats struct {
	Open     int64 `json:"open"`
	Fixed    int64 `json:"fixed"`
	Archived int64 `json:"archived"`
}

type PublicTimeStats struct {
	AverageIssueAgeDays    float64    `json:"average_issue_age_days"`
	OldestOpenIssueAgeDays int64      `json:"oldest_open_issue_age_days"`
	OldestOpenIssueID      *uuid.UUID `json:"oldest_open_issue_id,omitempty"`
	OldestOpenRoadName     *string    `json:"oldest_open_road_name,omitempty"`
	OldestOpenRegionName   *string    `json:"oldest_open_region_name,omitempty"`
	OldestOpenFirstSeenAt  *time.Time `json:"oldest_open_first_seen_at,omitempty"`
}

type PublicRegionStat struct {
	DistrictID    *int64 `json:"district_id,omitempty"`
	DistrictName  string `json:"district_name"`
	IssueCount    int64  `json:"issue_count"`
	CasualtyCount int64  `json:"casualty_count"`
	ReportCount   int64  `json:"report_count"`
}

type PublicStatsFilters struct {
	ProvinceOptions  []*PublicRegionOption `json:"province_options"`
	RegencyOptions   []*PublicRegionOption `json:"regency_options"`
	ActiveProvinceID *int64                `json:"active_province_id,omitempty"`
	ActiveProvince   *string               `json:"active_province,omitempty"`
	ActiveRegencyID  *int64                `json:"active_regency_id,omitempty"`
	ActiveRegency    *string               `json:"active_regency,omitempty"`
	ScopeLabel       *string               `json:"scope_label,omitempty"`
}

type PublicRegionOption struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	IssueCount  int64  `json:"issue_count"`
	ReportCount int64  `json:"report_count"`
}

type PublicTopIssue struct {
	Category        string    `json:"category"`
	Label           string    `json:"label"`
	MetricLabel     string    `json:"metric_label"`
	MetricValue     int64     `json:"metric_value"`
	IssueID         uuid.UUID `json:"issue_id"`
	Status          string    `json:"status"`
	RoadName        *string   `json:"road_name,omitempty"`
	RegionName      *string   `json:"region_name,omitempty"`
	DistrictName    *string   `json:"district_name,omitempty"`
	RegencyName     *string   `json:"regency_name,omitempty"`
	ProvinceName    *string   `json:"province_name,omitempty"`
	SubmissionCount int64     `json:"submission_count"`
	CasualtyCount   int64     `json:"casualty_count"`
	AgeDays         int64     `json:"age_days"`
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type StatsRepository interface {
	GetPublicStats(ctx context.Context, regionLimit int) (*domain.PublicStats, error)
}

type statsRepository struct {
	db *pgxpool.Pool
}

type topIssueRow struct {
	ID              uuid.UUID
	Status          string
	RoadName        *string
	RegionName      *string
	SubmissionCount int64
	CasualtyCount   int64
	AgeDays         int64
	FirstSeenAt     time.Time
}

func NewStatsRepository(db *pgxpool.Pool) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetPublicStats(ctx context.Context, regionLimit int) (*domain.PublicStats, error) {
	if regionLimit <= 0 || regionLimit > 50 {
		regionLimit = 10
	}

	stats := &domain.PublicStats{
		Regions:   make([]*domain.PublicRegionStat, 0),
		TopIssues: make([]*domain.PublicTopIssue, 0),
	}

	if err := r.querySummary(ctx, stats); err != nil {
		return nil, err
	}

	oldestOpen, err := r.queryOldestOpenIssue(ctx)
	if err != nil {
		return nil, err
	}
	if oldestOpen != nil {
		oldestID := oldestOpen.ID
		firstSeen := oldestOpen.FirstSeenAt
		stats.Time.OldestOpenIssueID = &oldestID
		stats.Time.OldestOpenIssueAgeDays = oldestOpen.AgeDays
		stats.Time.OldestOpenRoadName = oldestOpen.RoadName
		stats.Time.OldestOpenRegionName = oldestOpen.RegionName
		stats.Time.OldestOpenFirstSeenAt = &firstSeen
		stats.TopIssues = append(stats.TopIssues, buildTopIssue(
			"oldest_open",
			"Issue paling lama belum diperbaiki",
			"hari",
			oldestOpen.AgeDays,
			oldestOpen,
		))
	}

	topByReports, err := r.queryTopIssueByReports(ctx)
	if err != nil {
		return nil, err
	}
	if topByReports != nil {
		stats.TopIssues = append(stats.TopIssues, buildTopIssue(
			"most_reports",
			"Issue dengan laporan terbanyak",
			"laporan",
			topByReports.SubmissionCount,
			topByReports,
		))
	}

	topByCasualties, err := r.queryTopIssueByCasualties(ctx)
	if err != nil {
		return nil, err
	}
	if topByCasualties != nil {
		stats.TopIssues = append(stats.TopIssues, buildTopIssue(
			"most_casualties",
			"Issue dengan korban terbanyak",
			"korban",
			topByCasualties.CasualtyCount,
			topByCasualties,
		))
	}

	regions, err := r.queryRegionLeaderboard(ctx, regionLimit)
	if err != nil {
		return nil, err
	}
	stats.Regions = regions

	return stats, nil
}

func (r *statsRepository) querySummary(ctx context.Context, stats *domain.PublicStats) error {
	const summaryQuery = `
		SELECT
			COUNT(*)::bigint AS total_issues,
			COUNT(*) FILTER (
				WHERE i.created_at >= date_trunc('week', NOW())
			)::bigint AS total_issues_this_week,
			COALESCE(SUM(i.casualty_count), 0)::bigint AS total_casualties,
			COALESCE(SUM(i.photo_count), 0)::bigint AS total_photos,
			COALESCE(SUM(i.submission_count), 0)::bigint AS total_reports,
			COUNT(*) FILTER (
				WHERE i.status IN ('open', 'verified', 'in_progress')
			)::bigint AS open_issues,
			COUNT(*) FILTER (
				WHERE i.status = 'fixed'
			)::bigint AS fixed_issues,
			COUNT(*) FILTER (
				WHERE i.status = 'archived'
			)::bigint AS archived_issues,
			COALESCE(
				ROUND(
					AVG(
						GREATEST(
							EXTRACT(EPOCH FROM (NOW() - COALESCE(i.first_seen_at, i.created_at))),
							0
						) / 86400.0
					)::numeric,
					1
				),
				0
			)::double precision AS average_issue_age_days,
			COALESCE(
				MAX(
					FLOOR(
						GREATEST(
							EXTRACT(EPOCH FROM (NOW() - COALESCE(i.first_seen_at, i.created_at))),
							0
						) / 86400.0
					)
				) FILTER (WHERE i.status IN ('open', 'verified', 'in_progress')),
				0
			)::bigint AS oldest_open_issue_age_days
		FROM issues i
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
	`

	return r.db.QueryRow(ctx, summaryQuery).Scan(
		&stats.Global.TotalIssues,
		&stats.Global.TotalIssuesThisWeek,
		&stats.Global.TotalCasualties,
		&stats.Global.TotalPhotos,
		&stats.Global.TotalReports,
		&stats.Status.Open,
		&stats.Status.Fixed,
		&stats.Status.Archived,
		&stats.Time.AverageIssueAgeDays,
		&stats.Time.OldestOpenIssueAgeDays,
	)
}

func (r *statsRepository) queryOldestOpenIssue(ctx context.Context) (*topIssueRow, error) {
	const query = `
		SELECT
			i.id,
			i.status,
			i.road_name,
			r.name AS region_name,
			i.submission_count,
			i.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - COALESCE(i.first_seen_at, i.created_at))),
					0
				) / 86400.0
			)::bigint AS age_days,
			COALESCE(i.first_seen_at, i.created_at) AS first_seen_at
		FROM issues i
		LEFT JOIN regions r ON r.id = i.region_id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
		  AND i.status IN ('open', 'verified', 'in_progress')
		ORDER BY COALESCE(i.first_seen_at, i.created_at) ASC, i.created_at ASC
		LIMIT 1
	`

	return r.queryTopIssueRow(ctx, query)
}

func (r *statsRepository) queryTopIssueByReports(ctx context.Context) (*topIssueRow, error) {
	const query = `
		SELECT
			i.id,
			i.status,
			i.road_name,
			r.name AS region_name,
			i.submission_count,
			i.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - COALESCE(i.first_seen_at, i.created_at))),
					0
				) / 86400.0
			)::bigint AS age_days,
			COALESCE(i.first_seen_at, i.created_at) AS first_seen_at
		FROM issues i
		LEFT JOIN regions r ON r.id = i.region_id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
		ORDER BY i.submission_count DESC, i.casualty_count DESC, i.last_seen_at DESC
		LIMIT 1
	`

	return r.queryTopIssueRow(ctx, query)
}

func (r *statsRepository) queryTopIssueByCasualties(ctx context.Context) (*topIssueRow, error) {
	const query = `
		SELECT
			i.id,
			i.status,
			i.road_name,
			r.name AS region_name,
			i.submission_count,
			i.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - COALESCE(i.first_seen_at, i.created_at))),
					0
				) / 86400.0
			)::bigint AS age_days,
			COALESCE(i.first_seen_at, i.created_at) AS first_seen_at
		FROM issues i
		LEFT JOIN regions r ON r.id = i.region_id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
		ORDER BY i.casualty_count DESC, i.submission_count DESC, i.last_seen_at DESC
		LIMIT 1
	`

	return r.queryTopIssueRow(ctx, query)
}

func (r *statsRepository) queryTopIssueRow(ctx context.Context, query string, args ...any) (*topIssueRow, error) {
	var row topIssueRow
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&row.ID,
		&row.Status,
		&row.RoadName,
		&row.RegionName,
		&row.SubmissionCount,
		&row.CasualtyCount,
		&row.AgeDays,
		&row.FirstSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &row, nil
}

func (r *statsRepository) queryRegionLeaderboard(ctx context.Context, limit int) ([]*domain.PublicRegionStat, error) {
	const query = `
		SELECT
			COALESCE(NULLIF(TRIM(r.name), ''), 'Wilayah Tidak Diketahui') AS region_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(i.casualty_count), 0)::bigint AS casualty_count,
			COALESCE(SUM(i.submission_count), 0)::bigint AS report_count
		FROM issues i
		LEFT JOIN regions r ON r.id = i.region_id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
		GROUP BY 1
		ORDER BY issue_count DESC, casualty_count DESC, report_count DESC, region_name ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make([]*domain.PublicRegionStat, 0, limit)
	for rows.Next() {
		var item domain.PublicRegionStat
		if err := rows.Scan(
			&item.RegionName,
			&item.IssueCount,
			&item.CasualtyCount,
			&item.ReportCount,
		); err != nil {
			return nil, err
		}
		regions = append(regions, &item)
	}

	return regions, rows.Err()
}

func buildTopIssue(
	category, label, metricLabel string,
	metricValue int64,
	row *topIssueRow,
) *domain.PublicTopIssue {
	return &domain.PublicTopIssue{
		Category:        category,
		Label:           label,
		MetricLabel:     metricLabel,
		MetricValue:     metricValue,
		IssueID:         row.ID,
		Status:          row.Status,
		RoadName:        row.RoadName,
		RegionName:      row.RegionName,
		SubmissionCount: row.SubmissionCount,
		CasualtyCount:   row.CasualtyCount,
		AgeDays:         row.AgeDays,
	}
}

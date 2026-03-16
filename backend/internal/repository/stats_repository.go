package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type StatsRepository interface {
	GetPublicStats(ctx context.Context, query domain.PublicStatsQuery, regionLimit int) (*domain.PublicStats, error)
}

type statsRepository struct {
	db *pgxpool.Pool
}

type topIssueRow struct {
	ID              uuid.UUID
	Status          string
	RoadName        *string
	RegionName      *string
	DistrictName    *string
	RegencyName     *string
	ProvinceName    *string
	SubmissionCount int64
	CasualtyCount   int64
	AgeDays         int64
	FirstSeenAt     time.Time
}

const statsScopedIssuesCTE = `
	WITH latest_submission_regions AS (
		SELECT DISTINCT ON (s.issue_id)
			s.issue_id,
			s.region_id
		FROM issue_submissions s
		WHERE s.region_id IS NOT NULL
		ORDER BY s.issue_id, s.reported_at DESC
	), base_issues AS (
		SELECT
			i.id,
			i.status,
			i.road_name,
			i.submission_count,
			i.casualty_count,
			i.last_seen_at,
			COALESCE(i.first_seen_at, i.created_at) AS first_seen_at,
			COALESCE(i.region_id, latest.region_id) AS base_region_id
		FROM issues i
		LEFT JOIN latest_submission_regions latest ON latest.issue_id = i.id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
	), resolved_regions AS (
		SELECT
			base.*,
			r0.id AS region_id_0,
			NULLIF(TRIM(r0.name), '') AS region_name_0,
			r0.level AS region_level_0,
			r1.id AS region_id_1,
			NULLIF(TRIM(r1.name), '') AS region_name_1,
			r1.level AS region_level_1,
			r2.id AS region_id_2,
			NULLIF(TRIM(r2.name), '') AS region_name_2,
			r2.level AS region_level_2
		FROM base_issues base
		LEFT JOIN regions r0 ON r0.id = base.base_region_id
		LEFT JOIN regions r1 ON r1.id = r0.parent_id
		LEFT JOIN regions r2 ON r2.id = r1.parent_id
	), normalized AS (
		SELECT
			resolved.id,
			resolved.status,
			resolved.road_name,
			resolved.submission_count,
			resolved.casualty_count,
			resolved.last_seen_at,
			resolved.first_seen_at,
			resolved.region_name_0 AS raw_region_name,
			CASE
				WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_id_0
				WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_id_1
				WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_id_2
				ELSE NULL
			END AS district_id,
			CASE
				WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_name_0
				WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_name_1
				WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_name_2
				ELSE NULL
			END AS district_name,
			CASE
				WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_id_0
				WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_id_1
				WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_id_2
				ELSE NULL
			END AS regency_id,
			CASE
				WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_name_0
				WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_name_1
				WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_name_2
				ELSE NULL
			END AS regency_name,
			CASE
				WHEN resolved.region_level_0 = 'province' THEN resolved.region_id_0
				WHEN resolved.region_level_1 = 'province' THEN resolved.region_id_1
				WHEN resolved.region_level_2 = 'province' THEN resolved.region_id_2
				ELSE NULL
			END AS province_id,
			CASE
				WHEN resolved.region_level_0 = 'province' THEN resolved.region_name_0
				WHEN resolved.region_level_1 = 'province' THEN resolved.region_name_1
				WHEN resolved.region_level_2 = 'province' THEN resolved.region_name_2
				ELSE NULL
			END AS province_name
		FROM resolved_regions resolved
	)
`

func NewStatsRepository(db *pgxpool.Pool) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetPublicStats(ctx context.Context, query domain.PublicStatsQuery, regionLimit int) (*domain.PublicStats, error) {
	if regionLimit <= 0 || regionLimit > 50 {
		regionLimit = 10
	}

	stats := &domain.PublicStats{
		Regions:   make([]*domain.PublicRegionStat, 0),
		TopIssues: make([]*domain.PublicTopIssue, 0),
		Filters: domain.PublicStatsFilters{
			ProvinceOptions: make([]*domain.PublicRegionOption, 0),
			RegencyOptions:  make([]*domain.PublicRegionOption, 0),
		},
	}

	if err := r.querySummary(ctx, stats); err != nil {
		return nil, err
	}

	oldestOpenGlobal, err := r.queryOldestOpenIssue(ctx, domain.PublicStatsQuery{})
	if err != nil {
		return nil, err
	}
	if oldestOpenGlobal != nil {
		oldestID := oldestOpenGlobal.ID
		firstSeen := oldestOpenGlobal.FirstSeenAt
		stats.Time.OldestOpenIssueID = &oldestID
		stats.Time.OldestOpenIssueAgeDays = oldestOpenGlobal.AgeDays
		stats.Time.OldestOpenRoadName = oldestOpenGlobal.RoadName
		stats.Time.OldestOpenRegionName = oldestOpenGlobal.RegionName
		stats.Time.OldestOpenFirstSeenAt = &firstSeen
	}

	provinceOptions, err := r.queryProvinceOptions(ctx)
	if err != nil {
		return nil, err
	}
	stats.Filters.ProvinceOptions = provinceOptions

	activeProvinceID, activeProvinceName := chooseRegionOption(query.ProvinceID, provinceOptions)
	stats.Filters.ActiveProvinceID = activeProvinceID
	stats.Filters.ActiveProvince = activeProvinceName

	if activeProvinceID != nil {
		regencyOptions, regencyErr := r.queryRegencyOptions(ctx, *activeProvinceID)
		if regencyErr != nil {
			return nil, regencyErr
		}
		stats.Filters.RegencyOptions = regencyOptions
		activeRegencyID, activeRegencyName := chooseRegionOption(query.RegencyID, regencyOptions)
		stats.Filters.ActiveRegencyID = activeRegencyID
		stats.Filters.ActiveRegency = activeRegencyName
	}

	stats.Filters.ScopeLabel = buildScopeLabel(
		stats.Filters.ActiveProvince,
		stats.Filters.ActiveRegency,
	)

	scopeQuery := domain.PublicStatsQuery{
		ProvinceID: stats.Filters.ActiveProvinceID,
		RegencyID:  stats.Filters.ActiveRegencyID,
	}

	topOldestScoped, err := r.queryOldestOpenIssue(ctx, scopeQuery)
	if err != nil {
		return nil, err
	}
	if topOldestScoped != nil {
		stats.TopIssues = append(stats.TopIssues, buildTopIssue(
			"oldest_open",
			"Issue paling lama belum diperbaiki",
			"hari",
			topOldestScoped.AgeDays,
			topOldestScoped,
		))
	}

	topByReports, err := r.queryTopIssueByReports(ctx, scopeQuery)
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

	topByCasualties, err := r.queryTopIssueByCasualties(ctx, scopeQuery)
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

	regions, err := r.queryRegionLeaderboard(ctx, scopeQuery, regionLimit)
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

func (r *statsRepository) queryProvinceOptions(ctx context.Context) ([]*domain.PublicRegionOption, error) {
	query := statsScopedIssuesCTE + `
		SELECT
			normalized.province_id,
			normalized.province_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS report_count
		FROM normalized
		WHERE normalized.province_id IS NOT NULL
		  AND normalized.province_name IS NOT NULL
		GROUP BY normalized.province_id, normalized.province_name
		ORDER BY issue_count DESC, report_count DESC, normalized.province_name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.PublicRegionOption, 0)
	for rows.Next() {
		var item domain.PublicRegionOption
		if err := rows.Scan(&item.ID, &item.Name, &item.IssueCount, &item.ReportCount); err != nil {
			return nil, err
		}
		options = append(options, &item)
	}

	return options, rows.Err()
}

func (r *statsRepository) queryRegencyOptions(ctx context.Context, provinceID int64) ([]*domain.PublicRegionOption, error) {
	query := statsScopedIssuesCTE + `
		SELECT
			normalized.regency_id,
			normalized.regency_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS report_count
		FROM normalized
		WHERE normalized.province_id = $1
		  AND normalized.regency_id IS NOT NULL
		  AND normalized.regency_name IS NOT NULL
		GROUP BY normalized.regency_id, normalized.regency_name
		ORDER BY issue_count DESC, report_count DESC, normalized.regency_name ASC
	`

	rows, err := r.db.Query(ctx, query, provinceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.PublicRegionOption, 0)
	for rows.Next() {
		var item domain.PublicRegionOption
		if err := rows.Scan(&item.ID, &item.Name, &item.IssueCount, &item.ReportCount); err != nil {
			return nil, err
		}
		options = append(options, &item)
	}

	return options, rows.Err()
}

func (r *statsRepository) queryOldestOpenIssue(ctx context.Context, scope domain.PublicStatsQuery) (*topIssueRow, error) {
	query, args := buildScopedQuery(scope, `
		SELECT
			normalized.id,
			normalized.status,
			normalized.road_name,
			COALESCE(normalized.district_name, normalized.regency_name, normalized.province_name, normalized.raw_region_name) AS region_name,
			normalized.district_name,
			normalized.regency_name,
			normalized.province_name,
			normalized.submission_count,
			normalized.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - normalized.first_seen_at)),
					0
				) / 86400.0
			)::bigint AS age_days,
			normalized.first_seen_at
		FROM normalized
		WHERE normalized.status IN ('open', 'verified', 'in_progress')
	`, `
		ORDER BY normalized.first_seen_at ASC, normalized.last_seen_at ASC
		LIMIT 1
	`)

	return r.queryTopIssueRow(ctx, query, args...)
}

func (r *statsRepository) queryTopIssueByReports(ctx context.Context, scope domain.PublicStatsQuery) (*topIssueRow, error) {
	query, args := buildScopedQuery(scope, `
		SELECT
			normalized.id,
			normalized.status,
			normalized.road_name,
			COALESCE(normalized.district_name, normalized.regency_name, normalized.province_name, normalized.raw_region_name) AS region_name,
			normalized.district_name,
			normalized.regency_name,
			normalized.province_name,
			normalized.submission_count,
			normalized.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - normalized.first_seen_at)),
					0
				) / 86400.0
			)::bigint AS age_days,
			normalized.first_seen_at
		FROM normalized
	`, `
		ORDER BY normalized.submission_count DESC, normalized.casualty_count DESC, normalized.last_seen_at DESC
		LIMIT 1
	`)

	return r.queryTopIssueRow(ctx, query, args...)
}

func (r *statsRepository) queryTopIssueByCasualties(ctx context.Context, scope domain.PublicStatsQuery) (*topIssueRow, error) {
	query, args := buildScopedQuery(scope, `
		SELECT
			normalized.id,
			normalized.status,
			normalized.road_name,
			COALESCE(normalized.district_name, normalized.regency_name, normalized.province_name, normalized.raw_region_name) AS region_name,
			normalized.district_name,
			normalized.regency_name,
			normalized.province_name,
			normalized.submission_count,
			normalized.casualty_count,
			FLOOR(
				GREATEST(
					EXTRACT(EPOCH FROM (NOW() - normalized.first_seen_at)),
					0
				) / 86400.0
			)::bigint AS age_days,
			normalized.first_seen_at
		FROM normalized
	`, `
		ORDER BY normalized.casualty_count DESC, normalized.submission_count DESC, normalized.last_seen_at DESC
		LIMIT 1
	`)

	return r.queryTopIssueRow(ctx, query, args...)
}

func (r *statsRepository) queryTopIssueRow(ctx context.Context, query string, args ...any) (*topIssueRow, error) {
	var row topIssueRow
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&row.ID,
		&row.Status,
		&row.RoadName,
		&row.RegionName,
		&row.DistrictName,
		&row.RegencyName,
		&row.ProvinceName,
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

func (r *statsRepository) queryRegionLeaderboard(ctx context.Context, scope domain.PublicStatsQuery, limit int) ([]*domain.PublicRegionStat, error) {
	limitPlaceholder := 1
	if scope.ProvinceID != nil {
		limitPlaceholder++
	}
	if scope.RegencyID != nil {
		limitPlaceholder++
	}

	query, args := buildScopedQuery(scope, `
		SELECT
			MIN(normalized.district_id) AS district_id,
			COALESCE(
				normalized.district_name,
				normalized.regency_name,
				normalized.province_name,
				normalized.raw_region_name,
				'Wilayah administratif belum tersedia'
			) AS district_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(normalized.casualty_count), 0)::bigint AS casualty_count,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS report_count
		FROM normalized
		`, fmt.Sprintf(`
			GROUP BY 2
			ORDER BY issue_count DESC, report_count DESC, casualty_count DESC, district_name ASC
			LIMIT $%d
		`, limitPlaceholder))
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make([]*domain.PublicRegionStat, 0, limit)
	for rows.Next() {
		var item domain.PublicRegionStat
		if err := rows.Scan(
			&item.DistrictID,
			&item.DistrictName,
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

func buildScopedQuery(scope domain.PublicStatsQuery, selectSQL string, suffixSQL string) (string, []any) {
	query := statsScopedIssuesCTE + selectSQL
	args := make([]any, 0, 2)
	conditions := make([]string, 0, 2)

	if scope.ProvinceID != nil {
		args = append(args, *scope.ProvinceID)
		conditions = append(conditions, fmt.Sprintf("normalized.province_id = $%d", len(args)))
	}
	if scope.RegencyID != nil {
		args = append(args, *scope.RegencyID)
		conditions = append(conditions, fmt.Sprintf("normalized.regency_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		if strings.Contains(strings.ToUpper(selectSQL), " WHERE ") {
			query += " AND " + strings.Join(conditions, " AND ")
		} else {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
	}

	query += suffixSQL
	return query, args
}

func chooseRegionOption(requestedID *int64, options []*domain.PublicRegionOption) (*int64, *string) {
	if len(options) == 0 {
		return nil, nil
	}

	if requestedID != nil {
		for _, item := range options {
			if item.ID == *requestedID {
				id := item.ID
				name := item.Name
				return &id, &name
			}
		}
	}

	id := options[0].ID
	name := options[0].Name
	return &id, &name
}

func buildScopeLabel(provinceName, regencyName *string) *string {
	parts := make([]string, 0, 2)
	if regencyName != nil && strings.TrimSpace(*regencyName) != "" {
		parts = append(parts, strings.TrimSpace(*regencyName))
	}
	if provinceName != nil && strings.TrimSpace(*provinceName) != "" {
		parts = append(parts, strings.TrimSpace(*provinceName))
	}
	if len(parts) == 0 {
		return nil
	}

	label := strings.Join(parts, ", ")
	return &label
}

func buildTopIssue(
	category, label, metricLabel string,
	metricValue int64,
	row *topIssueRow,
) *domain.PublicTopIssue {
	regionName := row.RegionName
	if regionName == nil || strings.TrimSpace(*regionName) == "" {
		switch {
		case row.DistrictName != nil && strings.TrimSpace(*row.DistrictName) != "":
			regionName = row.DistrictName
		case row.RegencyName != nil && strings.TrimSpace(*row.RegencyName) != "":
			regionName = row.RegencyName
		case row.ProvinceName != nil && strings.TrimSpace(*row.ProvinceName) != "":
			regionName = row.ProvinceName
		case row.RoadName != nil && strings.TrimSpace(*row.RoadName) != "":
			fallback := "Sekitar " + strings.TrimSpace(*row.RoadName)
			regionName = &fallback
		default:
			fallback := "Wilayah administratif belum tersedia"
			regionName = &fallback
		}
	}

	return &domain.PublicTopIssue{
		Category:        category,
		Label:           label,
		MetricLabel:     metricLabel,
		MetricValue:     metricValue,
		IssueID:         row.ID,
		Status:          row.Status,
		RoadName:        row.RoadName,
		RegionName:      regionName,
		DistrictName:    row.DistrictName,
		RegencyName:     row.RegencyName,
		ProvinceName:    row.ProvinceName,
		SubmissionCount: row.SubmissionCount,
		CasualtyCount:   row.CasualtyCount,
		AgeDays:         row.AgeDays,
	}
}

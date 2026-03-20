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
	GetPublicRegionOptions(ctx context.Context) (*domain.PublicRegionOptions, error)
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

type regionScopeRow struct {
	ProvinceID   int64
	ProvinceName string
	RegencyID    *int64
	RegencyName  *string
	IssueCount   int64
	ReportCount  int64
}

type summarySnapshot struct {
	Summary domain.PublicGlobalStats
	Status  domain.PublicStatusStats
	Time    domain.PublicTimeStats
}

func regionLevelExpr(column string) string {
	return fmt.Sprintf(`
			CASE
				WHEN LOWER(COALESCE(%s, '')) IN ('province', 'provinsi') THEN 'province'
				WHEN LOWER(COALESCE(%s, '')) IN ('city', 'kota') THEN 'city'
				WHEN LOWER(COALESCE(%s, '')) IN ('regency', 'kabupaten') THEN 'regency'
				WHEN LOWER(COALESCE(%s, '')) IN ('district', 'kecamatan') THEN 'district'
				WHEN LOWER(COALESCE(%s, '')) IN ('subdistrict') THEN 'subdistrict'
				WHEN LOWER(COALESCE(%s, '')) IN ('village', 'kelurahan', 'desa') THEN 'village'
				ELSE LOWER(COALESCE(%s, ''))
			END
	`, column, column, column, column, column, column, column)
}

func regionPriorityExpr(column string) string {
	return fmt.Sprintf(`
			CASE
				WHEN LOWER(COALESCE(%s, '')) IN ('district', 'kecamatan') THEN 0
				WHEN LOWER(COALESCE(%s, '')) IN ('subdistrict') THEN 1
				WHEN LOWER(COALESCE(%s, '')) IN ('city', 'kota') THEN 2
				WHEN LOWER(COALESCE(%s, '')) IN ('regency', 'kabupaten') THEN 3
				WHEN LOWER(COALESCE(%s, '')) IN ('province', 'provinsi') THEN 4
				ELSE 5
			END
	`, column, column, column, column, column)
}

var statsScopedIssuesCTE = fmt.Sprintf(`
	WITH latest_submission_locations AS (
		SELECT DISTINCT ON (s.issue_id)
			s.issue_id,
			s.region_id,
			NULLIF(BTRIM(s.district_name), '') AS district_name,
			NULLIF(BTRIM(s.regency_name), '') AS regency_name,
			NULLIF(BTRIM(s.province_name), '') AS province_name
		FROM issue_submissions s
		WHERE s.region_id IS NOT NULL
		   OR NULLIF(BTRIM(s.district_name), '') IS NOT NULL
		   OR NULLIF(BTRIM(s.regency_name), '') IS NOT NULL
		   OR NULLIF(BTRIM(s.province_name), '') IS NOT NULL
		ORDER BY s.issue_id, s.reported_at DESC, s.created_at DESC
	), base_issues AS (
	SELECT
		i.id,
		i.status,
		i.road_name,
		i.photo_count,
		i.submission_count,
		i.casualty_count,
		i.created_at,
		i.last_seen_at,
		COALESCE(i.first_seen_at, i.created_at) AS first_seen_at,
		i.region_id AS issue_region_id,
		latest.region_id AS latest_region_id,
		latest.district_name AS latest_district_name,
		latest.regency_name AS latest_regency_name,
		latest.province_name AS latest_province_name,
		i.public_location::geometry AS public_location_geom
		FROM issues i
		LEFT JOIN latest_submission_locations latest ON latest.issue_id = i.id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
	), effective_regions AS (
		SELECT
			base.id,
			base.status,
			base.road_name,
			base.photo_count,
			base.submission_count,
			base.casualty_count,
			base.created_at,
			base.last_seen_at,
			base.first_seen_at,
			base.latest_district_name,
			base.latest_regency_name,
			base.latest_province_name,
			COALESCE(base.issue_region_id, base.latest_region_id, spatial.region_id) AS base_region_id
		FROM base_issues base
		LEFT JOIN LATERAL (
			SELECT reg.id AS region_id
			FROM regions reg
			WHERE ST_Covers(reg.geom, base.public_location_geom)
			ORDER BY
				%s,
				ST_Area(reg.geom::geography) ASC
			LIMIT 1
		) spatial
			ON base.issue_region_id IS NULL
		   AND base.latest_region_id IS NULL
		   AND base.public_location_geom IS NOT NULL
	), resolved_regions AS (
		SELECT
			base.id,
			base.status,
			base.road_name,
			base.photo_count,
			base.submission_count,
			base.casualty_count,
			base.created_at,
			base.last_seen_at,
			base.first_seen_at,
			base.latest_district_name,
			base.latest_regency_name,
			base.latest_province_name,
			r0.id AS region_id_0,
			NULLIF(TRIM(r0.name), '') AS region_name_0,
			%s AS region_level_0,
			r1.id AS region_id_1,
			NULLIF(TRIM(r1.name), '') AS region_name_1,
			%s AS region_level_1,
			r2.id AS region_id_2,
			NULLIF(TRIM(r2.name), '') AS region_name_2,
			%s AS region_level_2,
			r3.id AS region_id_3,
			NULLIF(TRIM(r3.name), '') AS region_name_3,
			%s AS region_level_3,
			r4.id AS region_id_4,
			NULLIF(TRIM(r4.name), '') AS region_name_4,
			%s AS region_level_4
		FROM effective_regions base
		LEFT JOIN regions r0 ON r0.id = base.base_region_id
		LEFT JOIN regions r1 ON r1.id = r0.parent_id
		LEFT JOIN regions r2 ON r2.id = r1.parent_id
		LEFT JOIN regions r3 ON r3.id = r2.parent_id
		LEFT JOIN regions r4 ON r4.id = r3.parent_id
	), normalized AS (
		SELECT
			resolved.id,
			resolved.status,
			resolved.road_name,
			resolved.photo_count,
			resolved.submission_count,
			resolved.casualty_count,
			resolved.created_at,
			resolved.last_seen_at,
			resolved.first_seen_at,
			resolved.region_name_0 AS raw_region_name,
			CASE
				WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_id_0
				WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_id_1
				WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_id_2
				WHEN resolved.region_level_3 IN ('district', 'subdistrict') THEN resolved.region_id_3
				WHEN resolved.region_level_4 IN ('district', 'subdistrict') THEN resolved.region_id_4
				ELSE NULL
			END AS district_id,
			COALESCE(
				resolved.latest_district_name,
				CASE
					WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('district', 'subdistrict') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('district', 'subdistrict') THEN resolved.region_name_4
					ELSE NULL
				END
			) AS district_name,
			CASE
				WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_id_0
				WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_id_1
				WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_id_2
				WHEN resolved.region_level_3 IN ('city', 'regency') THEN resolved.region_id_3
				WHEN resolved.region_level_4 IN ('city', 'regency') THEN resolved.region_id_4
				ELSE NULL
			END AS regency_id,
			COALESCE(
				resolved.latest_regency_name,
				CASE
					WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('city', 'regency') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('city', 'regency') THEN resolved.region_name_4
					ELSE NULL
				END
			) AS regency_name,
			CASE
				WHEN resolved.region_level_0 = 'province' THEN resolved.region_id_0
				WHEN resolved.region_level_1 = 'province' THEN resolved.region_id_1
				WHEN resolved.region_level_2 = 'province' THEN resolved.region_id_2
				WHEN resolved.region_level_3 = 'province' THEN resolved.region_id_3
				WHEN resolved.region_level_4 = 'province' THEN resolved.region_id_4
				ELSE NULL
			END AS province_id,
			COALESCE(
				resolved.latest_province_name,
				CASE
					WHEN resolved.region_level_0 = 'province' THEN resolved.region_name_0
					WHEN resolved.region_level_1 = 'province' THEN resolved.region_name_1
					WHEN resolved.region_level_2 = 'province' THEN resolved.region_name_2
					WHEN resolved.region_level_3 = 'province' THEN resolved.region_name_3
					WHEN resolved.region_level_4 = 'province' THEN resolved.region_name_4
					ELSE NULL
				END
			) AS province_name,
			COALESCE(
				NULLIF(CONCAT_WS(', ',
					COALESCE(
						resolved.latest_district_name,
						CASE
							WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_name_0
							WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_name_1
							WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_name_2
							WHEN resolved.region_level_3 IN ('district', 'subdistrict') THEN resolved.region_name_3
							WHEN resolved.region_level_4 IN ('district', 'subdistrict') THEN resolved.region_name_4
							ELSE NULL
						END
					),
					COALESCE(
						resolved.latest_regency_name,
						CASE
							WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_name_0
							WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_name_1
							WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_name_2
							WHEN resolved.region_level_3 IN ('city', 'regency') THEN resolved.region_name_3
							WHEN resolved.region_level_4 IN ('city', 'regency') THEN resolved.region_name_4
							ELSE NULL
						END
					),
					COALESCE(
						resolved.latest_province_name,
						CASE
							WHEN resolved.region_level_0 = 'province' THEN resolved.region_name_0
							WHEN resolved.region_level_1 = 'province' THEN resolved.region_name_1
							WHEN resolved.region_level_2 = 'province' THEN resolved.region_name_2
							WHEN resolved.region_level_3 = 'province' THEN resolved.region_name_3
							WHEN resolved.region_level_4 = 'province' THEN resolved.region_name_4
							ELSE NULL
						END
					)
				), ''),
				CASE
					WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('district', 'subdistrict') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('district', 'subdistrict') THEN resolved.region_name_4
					ELSE NULL
				END,
				CASE
					WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('city', 'regency') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('city', 'regency') THEN resolved.region_name_4
					ELSE NULL
				END,
				CASE
					WHEN resolved.region_level_0 = 'province' THEN resolved.region_name_0
					WHEN resolved.region_level_1 = 'province' THEN resolved.region_name_1
					WHEN resolved.region_level_2 = 'province' THEN resolved.region_name_2
					WHEN resolved.region_level_3 = 'province' THEN resolved.region_name_3
					WHEN resolved.region_level_4 = 'province' THEN resolved.region_name_4
					ELSE NULL
				END,
				resolved.region_name_0,
				resolved.region_name_1,
				resolved.region_name_2,
				resolved.region_name_3,
				resolved.region_name_4
			) AS admin_region_name
		FROM resolved_regions resolved
	)
`, regionPriorityExpr("reg.level"),
	regionLevelExpr("r0.level"),
	regionLevelExpr("r1.level"),
	regionLevelExpr("r2.level"),
	regionLevelExpr("r3.level"),
	regionLevelExpr("r4.level"),
)

func NewStatsRepository(db *pgxpool.Pool) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetPublicRegionOptions(ctx context.Context) (*domain.PublicRegionOptions, error) {
	provinceOptions, err := r.queryProvinceOptions(ctx)
	if err != nil {
		return nil, err
	}

	regencyRows, err := r.queryAllRegencyOptions(ctx)
	if err != nil {
		return nil, err
	}

	out := &domain.PublicRegionOptions{
		Provinces: make([]*domain.PublicProvinceOption, 0, len(provinceOptions)),
	}

	provinceIndex := make(map[int64]*domain.PublicProvinceOption, len(provinceOptions))
	for _, item := range provinceOptions {
		province := &domain.PublicProvinceOption{
			ID:          item.ID,
			Name:        item.Name,
			IssueCount:  item.IssueCount,
			ReportCount: item.ReportCount,
			Regencies:   make([]*domain.PublicRegionOption, 0),
		}
		provinceIndex[item.ID] = province
		out.Provinces = append(out.Provinces, province)
	}

	for _, row := range regencyRows {
		if row.RegencyID == nil || row.RegencyName == nil {
			continue
		}

		province := provinceIndex[row.ProvinceID]
		if province == nil {
			continue
		}

		province.Regencies = append(province.Regencies, &domain.PublicRegionOption{
			ID:          *row.RegencyID,
			Name:        *row.RegencyName,
			IssueCount:  row.IssueCount,
			ReportCount: row.ReportCount,
		})
	}

	return out, nil
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

	globalSummary, err := r.querySummary(ctx, domain.PublicStatsQuery{})
	if err != nil {
		return nil, err
	}
	stats.Global = globalSummary.Summary

	provinceOptions, err := r.queryProvinceOptions(ctx)
	if err != nil {
		return nil, err
	}
	stats.Filters.ProvinceOptions = provinceOptions

	useDefaultScope := query.ProvinceID == nil && query.RegencyID == nil
	activeProvinceID, activeProvinceName := chooseRegionOption(query.ProvinceID, provinceOptions, useDefaultScope)
	stats.Filters.ActiveProvinceID = activeProvinceID
	stats.Filters.ActiveProvince = activeProvinceName

	if activeProvinceID != nil {
		regencyOptions, regencyErr := r.queryRegencyOptions(ctx, *activeProvinceID)
		if regencyErr != nil {
			return nil, regencyErr
		}
		stats.Filters.RegencyOptions = regencyOptions
		activeRegencyID, activeRegencyName := chooseRegionOption(query.RegencyID, regencyOptions, useDefaultScope)
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
	stats.ActiveScope = buildStatsScope(scopeQuery, stats.Filters.ScopeLabel, useDefaultScope)

	scopedSummary, err := r.querySummary(ctx, scopeQuery)
	if err != nil {
		return nil, err
	}
	stats.Summary = scopedSummary.Summary
	stats.Status = scopedSummary.Status
	stats.Time = scopedSummary.Time

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

func (r *statsRepository) querySummary(ctx context.Context, scope domain.PublicStatsQuery) (*summarySnapshot, error) {
	query, args := buildScopedQuery(scope, `
		SELECT
			COUNT(*)::bigint AS total_issues,
			COUNT(*) FILTER (
				WHERE normalized.created_at >= date_trunc('week', NOW())
			)::bigint AS total_issues_this_week,
			COALESCE(SUM(normalized.casualty_count), 0)::bigint AS total_casualties,
			COALESCE(SUM(normalized.photo_count), 0)::bigint AS total_photos,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS total_reports,
			COUNT(*) FILTER (
				WHERE normalized.status IN ('open', 'verified', 'in_progress')
			)::bigint AS open_issues,
			COUNT(*) FILTER (
				WHERE normalized.status = 'fixed'
			)::bigint AS fixed_issues,
			COUNT(*) FILTER (
				WHERE normalized.status = 'archived'
			)::bigint AS archived_issues,
			COALESCE(
				ROUND(
					AVG(
						GREATEST(
							EXTRACT(EPOCH FROM (NOW() - normalized.first_seen_at)),
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
							EXTRACT(EPOCH FROM (NOW() - normalized.first_seen_at)),
							0
						) / 86400.0
					)
				) FILTER (WHERE normalized.status IN ('open', 'verified', 'in_progress')),
				0
			)::bigint AS oldest_open_issue_age_days
		FROM normalized
		WHERE TRUE
	`, ``)

	var snapshot summarySnapshot
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&snapshot.Summary.TotalIssues,
		&snapshot.Summary.TotalIssuesThisWeek,
		&snapshot.Summary.TotalCasualties,
		&snapshot.Summary.TotalPhotos,
		&snapshot.Summary.TotalReports,
		&snapshot.Status.Open,
		&snapshot.Status.Fixed,
		&snapshot.Status.Archived,
		&snapshot.Time.AverageIssueAgeDays,
		&snapshot.Time.OldestOpenIssueAgeDays,
	)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
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

func (r *statsRepository) queryAllRegencyOptions(ctx context.Context) ([]regionScopeRow, error) {
	query := statsScopedIssuesCTE + `
		SELECT
			normalized.province_id,
			normalized.province_name,
			normalized.regency_id,
			normalized.regency_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS report_count
		FROM normalized
		WHERE normalized.province_id IS NOT NULL
		  AND normalized.province_name IS NOT NULL
		  AND normalized.regency_id IS NOT NULL
		  AND normalized.regency_name IS NOT NULL
		GROUP BY
			normalized.province_id,
			normalized.province_name,
			normalized.regency_id,
			normalized.regency_name
		ORDER BY
			normalized.province_name ASC,
			issue_count DESC,
			report_count DESC,
			normalized.regency_name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]regionScopeRow, 0)
	for rows.Next() {
		var item regionScopeRow
		if err := rows.Scan(
			&item.ProvinceID,
			&item.ProvinceName,
			&item.RegencyID,
			&item.RegencyName,
			&item.IssueCount,
			&item.ReportCount,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *statsRepository) queryOldestOpenIssue(ctx context.Context, scope domain.PublicStatsQuery) (*topIssueRow, error) {
	query, args := buildScopedQuery(scope, `
		SELECT
			normalized.id,
			normalized.status,
			normalized.road_name,
			normalized.admin_region_name AS region_name,
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
			normalized.admin_region_name AS region_name,
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
			normalized.admin_region_name AS region_name,
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
			CASE
				WHEN normalized.district_id IS NOT NULL THEN normalized.district_id
				WHEN normalized.regency_id IS NOT NULL THEN normalized.regency_id
				WHEN normalized.province_id IS NOT NULL THEN normalized.province_id
				ELSE NULL
			END AS region_id,
			CASE
				WHEN normalized.district_id IS NOT NULL THEN 'district'
				WHEN normalized.regency_id IS NOT NULL THEN 'regency'
				WHEN normalized.province_id IS NOT NULL THEN 'province'
				ELSE 'unknown'
			END AS region_level,
			COALESCE(
				normalized.district_name,
				normalized.regency_name,
				normalized.province_name
			) AS region_name,
			CASE
				WHEN normalized.district_id IS NOT NULL THEN normalized.regency_name
				WHEN normalized.regency_id IS NOT NULL THEN normalized.province_name
				ELSE NULL
			END AS parent_region_name,
			normalized.district_id,
			COALESCE(
				normalized.district_name,
				normalized.regency_name,
				normalized.province_name
			) AS district_name,
			normalized.regency_name,
			normalized.province_name,
			COUNT(*)::bigint AS issue_count,
			COALESCE(SUM(normalized.casualty_count), 0)::bigint AS casualty_count,
			COALESCE(SUM(normalized.submission_count), 0)::bigint AS report_count
		FROM normalized
		WHERE (
			normalized.district_id IS NOT NULL
			OR normalized.regency_id IS NOT NULL
			OR normalized.province_id IS NOT NULL
		)
		`, fmt.Sprintf(`
			GROUP BY 1, 2, 3, 4, 5, 6, 7, 8
			ORDER BY issue_count DESC, report_count DESC, casualty_count DESC, region_name ASC
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
			&item.RegionID,
			&item.RegionLevel,
			&item.RegionName,
			&item.ParentRegionName,
			&item.DistrictID,
			&item.DistrictName,
			&item.RegencyName,
			&item.ProvinceName,
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

func chooseRegionOption(
	requestedID *int64,
	options []*domain.PublicRegionOption,
	allowFallback bool,
) (*int64, *string) {
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

	if !allowFallback {
		return nil, nil
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

func buildStatsScope(query domain.PublicStatsQuery, scopeLabel *string, isDefault bool) domain.PublicStatsScope {
	label := "Semua wilayah publik"
	kind := "global"

	switch {
	case query.RegencyID != nil:
		kind = "regency"
	case query.ProvinceID != nil:
		kind = "province"
	}

	if scopeLabel != nil && strings.TrimSpace(*scopeLabel) != "" {
		label = strings.TrimSpace(*scopeLabel)
	}

	return domain.PublicStatsScope{
		Kind:      kind,
		Label:     label,
		IsDefault: isDefault,
	}
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

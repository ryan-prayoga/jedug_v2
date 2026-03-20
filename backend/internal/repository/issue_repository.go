package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

const maxPublicNoteLength = 220

// BBoxFilter is an optional geographic bounding box filter for issue listing.
type BBoxFilter struct {
	MinLng, MinLat, MaxLng, MaxLat float64
}

type IssueRepository interface {
	List(ctx context.Context, limit, offset int, status *string, severity *int, bbox *BBoxFilter) ([]*domain.Issue, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	FindByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.IssueDetail, error)
	ListTimeline(ctx context.Context, id uuid.UUID, limit, offset int) ([]*domain.IssueTimelineEvent, error)
}

type issueRepository struct {
	db *pgxpool.Pool
}

func NewIssueRepository(db *pgxpool.Pool) IssueRepository {
	return &issueRepository{db: db}
}

var publicIssueScopedCTE = fmt.Sprintf(`
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
			i.verification_status,
			i.severity_current,
			i.severity_max,
			ST_X(i.public_location::geometry) AS longitude,
			ST_Y(i.public_location::geometry) AS latitude,
			i.region_id AS issue_region_id,
			latest.region_id AS latest_region_id,
			latest.district_name AS latest_district_name,
			latest.regency_name AS latest_regency_name,
			latest.province_name AS latest_province_name,
			i.public_location::geometry AS public_location_geom,
			i.road_name,
			i.road_type,
			i.submission_count,
			i.photo_count,
			i.casualty_count,
			i.reaction_count,
			i.flag_count,
			i.first_seen_at,
			i.last_seen_at,
			i.created_at,
			i.updated_at
		FROM issues i
		LEFT JOIN latest_submission_locations latest ON latest.issue_id = i.id
		WHERE i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
	), effective_regions AS (
		SELECT
			base.*,
			COALESCE(base.issue_region_id, base.latest_region_id, spatial.region_id) AS effective_region_id
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
			base.*,
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
		LEFT JOIN regions r0 ON r0.id = base.effective_region_id
		LEFT JOIN regions r1 ON r1.id = r0.parent_id
		LEFT JOIN regions r2 ON r2.id = r1.parent_id
		LEFT JOIN regions r3 ON r3.id = r2.parent_id
		LEFT JOIN regions r4 ON r4.id = r3.parent_id
	), normalized_regions AS (
		SELECT
			resolved.id,
			resolved.status,
			resolved.verification_status,
			resolved.severity_current,
			resolved.severity_max,
			resolved.longitude,
			resolved.latitude,
			resolved.public_location_geom,
			resolved.effective_region_id AS region_id,
			resolved.road_name,
			resolved.road_type,
			resolved.submission_count,
			resolved.photo_count,
			resolved.casualty_count,
			resolved.reaction_count,
			resolved.flag_count,
			resolved.first_seen_at,
			resolved.last_seen_at,
			resolved.created_at,
			resolved.updated_at,
			COALESCE(
				resolved.latest_district_name,
				CASE
					WHEN resolved.region_level_0 IN ('district', 'subdistrict') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('district', 'subdistrict') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('district', 'subdistrict') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('district', 'subdistrict') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('district', 'subdistrict') THEN resolved.region_name_4
					ELSE NULL
				END,
				NULL
			) AS district_name,
			COALESCE(
				resolved.latest_regency_name,
				CASE
					WHEN resolved.region_level_0 IN ('city', 'regency') THEN resolved.region_name_0
					WHEN resolved.region_level_1 IN ('city', 'regency') THEN resolved.region_name_1
					WHEN resolved.region_level_2 IN ('city', 'regency') THEN resolved.region_name_2
					WHEN resolved.region_level_3 IN ('city', 'regency') THEN resolved.region_name_3
					WHEN resolved.region_level_4 IN ('city', 'regency') THEN resolved.region_name_4
					ELSE NULL
				END,
				NULL
			) AS regency_name,
			COALESCE(
				resolved.latest_province_name,
				CASE
					WHEN resolved.region_level_0 = 'province' THEN resolved.region_name_0
					WHEN resolved.region_level_1 = 'province' THEN resolved.region_name_1
					WHEN resolved.region_level_2 = 'province' THEN resolved.region_name_2
					WHEN resolved.region_level_3 = 'province' THEN resolved.region_name_3
					WHEN resolved.region_level_4 = 'province' THEN resolved.region_name_4
					ELSE NULL
				END,
				NULL
			) AS province_name,
			COALESCE(
				resolved.region_name_0,
				resolved.region_name_1,
				resolved.region_name_2,
				resolved.region_name_3,
				resolved.region_name_4
			) AS raw_region_name
		FROM resolved_regions resolved
	), public_issues AS (
		SELECT
			normalized.id,
			normalized.status,
			normalized.verification_status,
			normalized.severity_current,
			normalized.severity_max,
			normalized.longitude,
			normalized.latitude,
			normalized.public_location_geom,
			normalized.region_id,
			COALESCE(
				NULLIF(CONCAT_WS(', ', normalized.district_name, normalized.regency_name, normalized.province_name), ''),
				normalized.raw_region_name
			) AS region_name,
			normalized.road_name,
			normalized.road_type,
			normalized.district_name,
			normalized.regency_name,
			normalized.province_name,
			normalized.submission_count,
			normalized.photo_count,
			normalized.casualty_count,
			normalized.reaction_count,
			normalized.flag_count,
			normalized.first_seen_at,
			normalized.last_seen_at,
			normalized.created_at,
			normalized.updated_at
		FROM normalized_regions normalized
	)
`, regionPriorityExpr("reg.level"),
	regionLevelExpr("r0.level"),
	regionLevelExpr("r1.level"),
	regionLevelExpr("r2.level"),
	regionLevelExpr("r3.level"),
	regionLevelExpr("r4.level"),
)

// issueCols is the shared SELECT column list.
const issueCols = `
	i.id,
	i.status,
	i.verification_status,
	i.severity_current,
	i.severity_max,
	i.longitude,
	i.latitude,
	i.region_id,
	i.region_name,
	i.road_name,
	i.road_type,
	i.district_name,
	i.regency_name,
	i.province_name,
	i.submission_count,
	i.photo_count,
	i.casualty_count,
	i.reaction_count,
	i.flag_count,
	i.first_seen_at,
	i.last_seen_at,
	i.created_at,
	i.updated_at
`

func scanIssueRow(row pgx.Row, i *domain.Issue) error {
	return row.Scan(
		&i.ID, &i.Status, &i.VerificationStatus, &i.SeverityCurrent, &i.SeverityMax,
		&i.Longitude, &i.Latitude,
		&i.RegionID, &i.RegionName, &i.RoadName, &i.RoadType,
		&i.DistrictName, &i.RegencyName, &i.ProvinceName,
		&i.SubmissionCount, &i.PhotoCount, &i.CasualtyCount, &i.ReactionCount, &i.FlagCount,
		&i.FirstSeenAt, &i.LastSeenAt, &i.CreatedAt, &i.UpdatedAt,
	)
}

func (r *issueRepository) List(ctx context.Context, limit, offset int, status *string, severity *int, bbox *BBoxFilter) ([]*domain.Issue, error) {
	conditions := []string{"1 = 1"}
	args := []any{}
	n := 1

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", n))
		args = append(args, *status)
		n++
	}

	if severity != nil {
		conditions = append(conditions, fmt.Sprintf("i.severity_current >= $%d", n))
		args = append(args, *severity)
		n++
	}

	if bbox != nil {
		conditions = append(conditions, fmt.Sprintf(
			"i.public_location_geom && ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326)",
			n, n+1, n+2, n+3,
		))
		args = append(args, bbox.MinLng, bbox.MinLat, bbox.MaxLng, bbox.MaxLat)
		n += 4
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(
		"%s SELECT %s FROM public_issues i WHERE %s ORDER BY i.last_seen_at DESC LIMIT $%d OFFSET $%d",
		publicIssueScopedCTE,
		issueCols,
		strings.Join(conditions, " AND "),
		n, n+1,
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*domain.Issue, 0)
	for rows.Next() {
		var i domain.Issue
		if err := rows.Scan(
			&i.ID, &i.Status, &i.VerificationStatus, &i.SeverityCurrent, &i.SeverityMax,
			&i.Longitude, &i.Latitude,
			&i.RegionID, &i.RegionName, &i.RoadName, &i.RoadType,
			&i.DistrictName, &i.RegencyName, &i.ProvinceName,
			&i.SubmissionCount, &i.PhotoCount, &i.CasualtyCount, &i.ReactionCount, &i.FlagCount,
			&i.FirstSeenAt, &i.LastSeenAt, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		issues = append(issues, &i)
	}
	return issues, rows.Err()
}

func (r *issueRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	query := publicIssueScopedCTE + ` SELECT ` + issueCols + `
		FROM public_issues i
		WHERE i.id = $1
		LIMIT 1`

	var i domain.Issue
	err := scanIssueRow(r.db.QueryRow(ctx, query, id), &i)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (r *issueRepository) FindByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.IssueDetail, error) {
	issue, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if issue == nil {
		return nil, nil
	}

	// Public gallery media, primary first.
	mediaRows, err := r.db.Query(ctx, `
		SELECT sm.id, sm.object_key, sm.mime_type, sm.size_bytes,
		       sm.width, sm.height, sm.blurhash, sm.is_primary, sm.created_at
		FROM submission_media sm
		JOIN issue_submissions s ON s.id = sm.submission_id
		WHERE s.issue_id = $1
		  AND s.status <> 'rejected'
		ORDER BY sm.is_primary DESC, sm.sort_order ASC, sm.created_at DESC
		LIMIT 20
	`, id)
	if err != nil {
		return nil, err
	}
	defer mediaRows.Close()

	media := make([]*domain.MediaItem, 0)
	for mediaRows.Next() {
		var m domain.MediaItem
		if err := mediaRows.Scan(
			&m.ID, &m.ObjectKey, &m.MimeType, &m.SizeBytes,
			&m.Width, &m.Height, &m.Blurhash, &m.IsPrimary, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		media = append(media, &m)
	}
	if err := mediaRows.Err(); err != nil {
		return nil, err
	}

	var primaryMedia *domain.MediaItem
	if len(media) > 0 {
		primaryMedia = media[0]
	}

	// Last 3 public-facing submissions (minimal public fields only).
	subRows, err := r.db.Query(ctx, `
		SELECT id, status, severity, has_casualty, casualty_count, note, reported_at
		FROM issue_submissions
		WHERE issue_id = $1
		  AND status <> 'rejected'
		ORDER BY reported_at DESC
		LIMIT 3
	`, id)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	subs := make([]*domain.SubmissionSummary, 0)
	for subRows.Next() {
		var s domain.SubmissionSummary
		if err := subRows.Scan(
			&s.ID, &s.Status, &s.Severity, &s.HasCasualty, &s.CasualtyCount, &s.Note, &s.ReportedAt,
		); err != nil {
			return nil, err
		}
		s.PublicNote = normalizePublicNote(s.Note, maxPublicNoteLength)
		subs = append(subs, &s)
	}
	if err := subRows.Err(); err != nil {
		return nil, err
	}

	var publicNote *string
	for _, submission := range subs {
		if submission.PublicNote != nil {
			publicNote = submission.PublicNote
			break
		}
	}

	return &domain.IssueDetail{
		Issue:             issue,
		PrimaryMedia:      primaryMedia,
		PublicNote:        publicNote,
		Media:             media,
		RecentSubmissions: subs,
	}, nil
}

func normalizePublicNote(note *string, maxLength int) *string {
	if note == nil {
		return nil
	}

	normalized := strings.Join(strings.Fields(strings.TrimSpace(*note)), " ")
	if normalized == "" {
		return nil
	}

	if maxLength > 0 && len([]rune(normalized)) > maxLength {
		runes := []rune(normalized)
		truncated := string(runes[:maxLength-1]) + "…"
		return &truncated
	}

	return &normalized
}

func (r *issueRepository) ListTimeline(ctx context.Context, id uuid.UUID, limit, offset int) ([]*domain.IssueTimelineEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ie.event_type, ie.created_at, ie.event_data
		FROM issue_events ie
		JOIN issues i ON i.id = ie.issue_id
		WHERE ie.issue_id = $1
		  AND i.is_hidden = FALSE
		  AND i.status NOT IN ('rejected', 'merged')
		ORDER BY ie.created_at DESC, ie.id DESC
		LIMIT $2 OFFSET $3
	`, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*domain.IssueTimelineEvent, 0)
	for rows.Next() {
		var (
			eventType string
			createdAt time.Time
			dataRaw   []byte
		)

		if err := rows.Scan(&eventType, &createdAt, &dataRaw); err != nil {
			return nil, err
		}

		data := map[string]any{}
		if len(dataRaw) > 0 {
			if err := json.Unmarshal(dataRaw, &data); err != nil {
				return nil, err
			}
		}

		events = append(events, &domain.IssueTimelineEvent{
			Type:      eventType,
			CreatedAt: createdAt,
			Data:      data,
		})
	}

	return events, rows.Err()
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

// BBoxFilter is an optional geographic bounding box filter for issue listing.
type BBoxFilter struct {
	MinLng, MinLat, MaxLng, MaxLat float64
}

type IssueRepository interface {
	List(ctx context.Context, limit, offset int, status *string, bbox *BBoxFilter) ([]*domain.Issue, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	FindByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.IssueDetail, error)
}

type issueRepository struct {
	db *pgxpool.Pool
}

func NewIssueRepository(db *pgxpool.Pool) IssueRepository {
	return &issueRepository{db: db}
}

// issueCols is the shared SELECT column list.
// ST_X/ST_Y extract lon/lat from the GEOGRAPHY(POINT) column.
const issueCols = `
	i.id,
	i.status,
	i.verification_status,
	i.severity_current,
	i.severity_max,
	ST_X(i.public_location::geometry) AS longitude,
	ST_Y(i.public_location::geometry) AS latitude,
	i.region_id,
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
`

func scanIssueRow(row pgx.Row, i *domain.Issue) error {
	return row.Scan(
		&i.ID, &i.Status, &i.VerificationStatus, &i.SeverityCurrent, &i.SeverityMax,
		&i.Longitude, &i.Latitude,
		&i.RegionID, &i.RoadName, &i.RoadType,
		&i.SubmissionCount, &i.PhotoCount, &i.CasualtyCount, &i.ReactionCount, &i.FlagCount,
		&i.FirstSeenAt, &i.LastSeenAt, &i.CreatedAt, &i.UpdatedAt,
	)
}

func (r *issueRepository) List(ctx context.Context, limit, offset int, status *string, bbox *BBoxFilter) ([]*domain.Issue, error) {
	conditions := []string{
		"i.is_hidden = FALSE",
		"i.status NOT IN ('rejected', 'merged')",
	}
	args := []any{}
	n := 1

	if status != nil {
		conditions = append(conditions, fmt.Sprintf("i.status = $%d", n))
		args = append(args, *status)
		n++
	}

	if bbox != nil {
		conditions = append(conditions, fmt.Sprintf(
			"i.public_location && ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326)::geography",
			n, n+1, n+2, n+3,
		))
		args = append(args, bbox.MinLng, bbox.MinLat, bbox.MaxLng, bbox.MaxLat)
		n += 4
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(
		"SELECT %s FROM issues i WHERE %s ORDER BY i.last_seen_at DESC LIMIT $%d OFFSET $%d",
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
			&i.RegionID, &i.RoadName, &i.RoadType,
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
	query := `SELECT ` + issueCols + `
		FROM issues i
		WHERE i.id = $1 AND i.is_hidden = FALSE
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

	// Top 5 media items, primary first
	mediaRows, err := r.db.Query(ctx, `
		SELECT sm.id, sm.object_key, sm.mime_type, sm.size_bytes,
		       sm.width, sm.height, sm.blurhash, sm.is_primary, sm.created_at
		FROM submission_media sm
		JOIN issue_submissions s ON s.id = sm.submission_id
		WHERE s.issue_id = $1
		ORDER BY sm.is_primary DESC, sm.created_at DESC
		LIMIT 5
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

	// Last 3 submissions (minimal public fields only)
	subRows, err := r.db.Query(ctx, `
		SELECT id, status, severity, has_casualty, note, reported_at
		FROM issue_submissions
		WHERE issue_id = $1
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
			&s.ID, &s.Status, &s.Severity, &s.HasCasualty, &s.Note, &s.ReportedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	if err := subRows.Err(); err != nil {
		return nil, err
	}

	return &domain.IssueDetail{
		Issue:             issue,
		Media:             media,
		RecentSubmissions: subs,
	}, nil
}


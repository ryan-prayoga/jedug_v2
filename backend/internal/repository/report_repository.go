package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmitMediaInput struct {
	ObjectKey string
	MimeType  string
	SizeBytes int
	Width     *int
	Height    *int
	SHA256    *string
	IsPrimary bool
	SortOrder int
}

type SubmitInput struct {
	DeviceID      uuid.UUID
	Longitude     float64
	Latitude      float64
	GPSAccuracyM  *float64
	CapturedAt    *time.Time
	Severity      int
	HasCasualty   bool
	CasualtyCount int
	Note          *string
	Media         []SubmitMediaInput
}

type SubmitResult struct {
	IssueID      uuid.UUID
	SubmissionID uuid.UUID
	IsNewIssue   bool
}

type ReportRepository interface {
	SubmitReport(ctx context.Context, input SubmitInput) (*SubmitResult, error)
}

type reportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) ReportRepository {
	return &reportRepository{db: db}
}

// SubmitReport runs the full submit flow inside a single transaction:
// resolve region → find nearest open issue → create/update issue → create submission → insert media.
func (r *reportRepository) SubmitReport(ctx context.Context, input SubmitInput) (*SubmitResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	regionID, err := resolveRegionID(ctx, tx, input.Longitude, input.Latitude)
	if err != nil {
		return nil, err
	}

	existingIssueID, err := findNearestOpenIssue(ctx, tx, input.Longitude, input.Latitude)
	if err != nil {
		return nil, err
	}

	var issueID uuid.UUID
	isNew := false

	if existingIssueID != nil {
		issueID = *existingIssueID
	} else {
		issueID = uuid.New()
		if err := createIssue(ctx, tx, issueID, regionID, input); err != nil {
			return nil, err
		}
		isNew = true
	}

	submissionID := uuid.New()
	clientRequestID := uuid.New()
	if err := createSubmission(ctx, tx, submissionID, clientRequestID, issueID, regionID, input); err != nil {
		return nil, err
	}

	if err := createSubmissionMedia(ctx, tx, submissionID, input.Media); err != nil {
		return nil, err
	}

	if !isNew {
		if err := updateIssueCounters(ctx, tx, issueID, input); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &SubmitResult{
		IssueID:      issueID,
		SubmissionID: submissionID,
		IsNewIssue:   isNew,
	}, nil
}

// resolveRegionID finds the district-level region that contains the given point.
// Returns nil (no error) if no matching region is found.
func resolveRegionID(ctx context.Context, tx pgx.Tx, lon, lat float64) (*int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		SELECT id FROM regions
		WHERE level = 'district'
		  AND ST_Covers(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326))
		LIMIT 1
	`, lon, lat).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

// findNearestOpenIssue looks for an existing open issue within 10 meters of the given point.
func findNearestOpenIssue(ctx context.Context, tx pgx.Tx, lon, lat float64) (*uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id FROM issues
		WHERE status = 'open'
		  AND is_hidden = FALSE
		  AND ST_DWithin(
		        public_location,
		        ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
		        10
		      )
		ORDER BY ST_Distance(
		    public_location,
		    ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
		)
		LIMIT 1
	`, lon, lat).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func createIssue(ctx context.Context, tx pgx.Tx, id uuid.UUID, regionID *int64, input SubmitInput) error {
	casualtyCount := 0
	if input.HasCasualty {
		casualtyCount = input.CasualtyCount
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO issues (
			id, status, severity_current, severity_max,
			public_location, region_id,
			submission_count, photo_count, casualty_count,
			first_seen_at, last_seen_at
		) VALUES (
			$1, 'open', $2, $2,
			ST_SetSRID(ST_MakePoint($3, $4), 4326)::geography,
			$5, 1, $6, $7,
			NOW(), NOW()
		)
	`, id, input.Severity, input.Longitude, input.Latitude,
		regionID, len(input.Media), casualtyCount,
	)
	return err
}

func createSubmission(ctx context.Context, tx pgx.Tx, id, clientReqID, issueID uuid.UUID, regionID *int64, input SubmitInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO issue_submissions (
			id, issue_id, client_request_id, device_id, status,
			location, region_id, gps_accuracy_m, captured_at, reported_at,
			severity, has_casualty, casualty_count, note, source
		) VALUES (
			$1, $2, $3, $4, 'pending',
			ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography,
			$7, $8, $9, NOW(),
			$10, $11, $12, $13, 'pwa'
		)
	`, id, issueID, clientReqID, input.DeviceID,
		input.Longitude, input.Latitude,
		regionID, input.GPSAccuracyM, input.CapturedAt,
		input.Severity, input.HasCasualty, input.CasualtyCount, input.Note,
	)
	return err
}

func createSubmissionMedia(ctx context.Context, tx pgx.Tx, submissionID uuid.UUID, media []SubmitMediaInput) error {
	for _, m := range media {
		_, err := tx.Exec(ctx, `
			INSERT INTO submission_media (
				id, submission_id, object_key, mime_type, size_bytes,
				width, height, sha256, sort_order, is_primary
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, uuid.New(), submissionID, m.ObjectKey, m.MimeType, m.SizeBytes,
			m.Width, m.Height, m.SHA256, m.SortOrder, m.IsPrimary,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateIssueCounters updates an existing issue when a new submission is added.
func updateIssueCounters(ctx context.Context, tx pgx.Tx, issueID uuid.UUID, input SubmitInput) error {
	casualtyCount := 0
	if input.HasCasualty {
		casualtyCount = input.CasualtyCount
	}
	_, err := tx.Exec(ctx, `
		UPDATE issues SET
			last_seen_at     = NOW(),
			submission_count = submission_count + 1,
			photo_count      = photo_count + $1,
			casualty_count   = casualty_count + $2,
			severity_current = GREATEST(severity_current, $3),
			severity_max     = GREATEST(severity_max, $3),
			updated_at       = NOW()
		WHERE id = $4
	`, len(input.Media), casualtyCount, input.Severity, issueID)
	return err
}

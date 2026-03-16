package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultDuplicateRadiusM    = 30.0
	duplicateCandidateQueryCap = 50
)

var (
	ErrSubmissionMediaPersistFailed = errors.New("submission media persist failed")

	duplicateStatusPriority = map[string]int{
		"open":        0,
		"verified":    1,
		"in_progress": 2,
	}
	duplicateVerificationPriority = map[string]int{
		"verified":   0,
		"pending":    1,
		"unverified": 2,
		"rejected":   3,
	}
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
	ClientRequestID uuid.UUID
	DeviceID        uuid.UUID
	ActorFollowerID *uuid.UUID
	Longitude       float64
	Latitude        float64
	GPSAccuracyM    *float64
	CapturedAt      *time.Time
	Severity        int
	HasCasualty     bool
	CasualtyCount   int
	Note            *string
	RoadName        *string
	Media           []SubmitMediaInput
}

type SubmitResult struct {
	IssueID      uuid.UUID
	SubmissionID uuid.UUID
	IsNewIssue   bool
}

type ReportRepositoryConfig struct {
	DuplicateRadiusM float64
	PushNotifier     NotificationPushNotifier
}

type ReportRepository interface {
	SubmitReport(ctx context.Context, input SubmitInput) (*SubmitResult, error)
	FindByClientRequestID(ctx context.Context, clientRequestID uuid.UUID) (*SubmitResult, error)
}

type reportRepository struct {
	db               *pgxpool.Pool
	duplicateRadiusM float64
	pushNotifier     NotificationPushNotifier
}

func NewReportRepository(db *pgxpool.Pool, cfg ReportRepositoryConfig) ReportRepository {
	radius := cfg.DuplicateRadiusM
	if radius <= 0 {
		radius = defaultDuplicateRadiusM
	}
	return &reportRepository{
		db:               db,
		duplicateRadiusM: radius,
		pushNotifier:     cfg.PushNotifier,
	}
}

// SubmitReport runs the full submit flow inside a single transaction:
// resolve region → find duplicate candidate issue → create/update issue → create submission → insert media.
func (r *reportRepository) SubmitReport(ctx context.Context, input SubmitInput) (*SubmitResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Printf("[REPORT] tx_begin_failed device=%s error=%v", input.DeviceID, err)
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	regionID, err := resolveBestRegionID(ctx, tx, input.Longitude, input.Latitude)
	if err != nil {
		log.Printf("[REPORT] resolve_region_failed device=%s lon=%.6f lat=%.6f error=%v",
			input.DeviceID, input.Longitude, input.Latitude, err)
		return nil, fmt.Errorf("resolve region: %w", err)
	}
	candidate, err := findDuplicateCandidate(ctx, tx, input.Longitude, input.Latitude, r.duplicateRadiusM)
	if err != nil {
		log.Printf("[REPORT] duplicate_lookup_failed device=%s error=%v", input.DeviceID, err)
		return nil, fmt.Errorf("duplicate lookup: %w", err)
	}

	var issueID uuid.UUID
	isNew := false
	var previousState *issueAggregateState

	if candidate != nil {
		issueID = candidate.IssueID
		previousState, err = getIssueAggregateState(ctx, tx, issueID)
		if err != nil {
			log.Printf("[REPORT] get_issue_state_failed device=%s issue=%s error=%v",
				input.DeviceID, issueID, err)
			return nil, fmt.Errorf("get issue state: %w", err)
		}
		log.Printf(
			"[REPORT] duplicate_merge_selected issue=%s distance_m=%.2f status=%s verification=%s",
			candidate.IssueID, candidate.DistanceM, candidate.Status, candidate.VerificationStatus,
		)
	} else {
		issueID = uuid.New()
		if err := createIssue(ctx, tx, issueID, regionID, input); err != nil {
			log.Printf("[REPORT] create_issue_failed device=%s issue=%s error=%v",
				input.DeviceID, issueID, err)
			return nil, fmt.Errorf("create issue: %w", err)
		}
		isNew = true
		log.Printf("[REPORT] duplicate_merge_miss_new_issue issue=%s", issueID)
	}

	submissionID := uuid.New()
	clientRequestID := input.ClientRequestID
	if err := createSubmission(ctx, tx, submissionID, clientRequestID, issueID, regionID, input); err != nil {
		log.Printf("[REPORT] create_submission_failed device=%s issue=%s submission=%s error=%v",
			input.DeviceID, issueID, submissionID, err)
		return nil, fmt.Errorf("create submission: %w", err)
	}

	if err := createSubmissionMedia(ctx, tx, submissionID, input.Media); err != nil {
		log.Printf("[REPORT] create_media_failed device=%s submission=%s media_count=%d error=%v",
			input.DeviceID, submissionID, len(input.Media), err)
		return nil, fmt.Errorf("create media: %w", err)
	}

	if !isNew {
		if err := updateIssueAggregatesOnMerge(ctx, tx, issueID, regionID, input); err != nil {
			log.Printf("[REPORT] update_aggregates_failed device=%s issue=%s error=%v",
				input.DeviceID, issueID, err)
			return nil, fmt.Errorf("update aggregates: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[REPORT] tx_commit_failed device=%s issue=%s error=%v",
			input.DeviceID, issueID, err)
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	log.Printf("[REPORT] submit_tx_committed device=%s issue=%s submission=%s is_new=%t",
		input.DeviceID, issueID, submissionID, isNew)

	// Timeline events are secondary audit data — inserted after the main commit
	// using the pool directly so a missing/failing issue_events table never aborts
	// the report submission. Errors are only logged.
	r.insertTimelineEvents(ctx, issueID, submissionID, isNew, previousState, input)

	return &SubmitResult{
		IssueID:      issueID,
		SubmissionID: submissionID,
		IsNewIssue:   isNew,
	}, nil
}

// insertTimelineEvents records audit events for a completed submission.
// Non-fatal: all errors are logged and swallowed so the caller's success path
// is never affected by a missing or temporarily broken issue_events table.
func (r *reportRepository) insertTimelineEvents(
	ctx context.Context,
	issueID uuid.UUID,
	submissionID uuid.UUID,
	isNew bool,
	previousState *issueAggregateState,
	input SubmitInput,
) {
	incomingPhotos := len(input.Media)
	incomingCasualty := incomingCasualtyCount(input)

	type event struct {
		eventType string
		data      map[string]any
	}

	var events []event

	if isNew {
		events = append(events, event{"issue_created", map[string]any{
			"submission_id":  submissionID,
			"severity":       input.Severity,
			"photo_count":    incomingPhotos,
			"casualty_count": incomingCasualty,
		}})
	}

	if incomingPhotos > 0 {
		events = append(events, event{"photo_added", map[string]any{
			"submission_id": submissionID,
			"photo_count":   incomingPhotos,
		}})
	}

	if !isNew && previousState != nil && input.Severity > previousState.SeverityCurrent {
		events = append(events, event{"severity_changed", map[string]any{
			"submission_id": submissionID,
			"from":          previousState.SeverityCurrent,
			"to":            input.Severity,
		}})
	}

	if incomingCasualty > 0 {
		shouldLog := isNew
		if !isNew && previousState != nil && incomingCasualty > previousState.CasualtyCount {
			shouldLog = true
		}
		if shouldLog {
			fromCasualty := 0
			if previousState != nil {
				fromCasualty = previousState.CasualtyCount
			}
			events = append(events, event{"casualty_reported", map[string]any{
				"submission_id": submissionID,
				"from":          fromCasualty,
				"to":            incomingCasualty,
			}})
		}
	}

	for _, ev := range events {
		payload, marshalErr := json.Marshal(ev.data)
		if marshalErr != nil {
			log.Printf("[REPORT] timeline_event_marshal_error issue=%s type=%s error=%v", issueID, ev.eventType, marshalErr)
			continue
		}
		// Use string(payload) so pgx v5 sends the JSON as text (OID 25).
		// Passing []byte would encode as bytea (OID 17), and PostgreSQL cannot
		// cast bytea to jsonb — causing the INSERT to fail.
		var eventID int64
		execErr := r.db.QueryRow(ctx, `
			INSERT INTO issue_events (issue_id, event_type, event_data)
			VALUES ($1, $2, $3::jsonb)
			RETURNING id
		`, issueID, ev.eventType, string(payload)).Scan(&eventID)
		if execErr != nil {
			log.Printf("[REPORT] timeline_event_insert_error issue=%s type=%s error=%v", issueID, ev.eventType, execErr)
			continue
		}
		if dispatchErr := DispatchNotificationsForEvent(ctx, r.db, r.pushNotifier, issueID, eventID, ev.eventType, input.ActorFollowerID); dispatchErr != nil {
			log.Printf("[REPORT] notification_dispatch_error issue=%s event=%d error=%v", issueID, eventID, dispatchErr)
		}
		if ev.eventType == "issue_created" {
			if nearbyErr := DispatchNearbyAlertsForIssueCreated(ctx, r.db, r.pushNotifier, issueID, eventID, input.ActorFollowerID); nearbyErr != nil {
				log.Printf("[REPORT] nearby_alert_dispatch_error issue=%s event=%d error=%v", issueID, eventID, nearbyErr)
			}
		}
	}
}

// resolveBestRegionID finds the most relevant internal region for the given point.
// Prefer district-level region; if unavailable, fallback to smallest covering region.
func resolveBestRegionID(ctx context.Context, tx pgx.Tx, lon, lat float64) (*int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		WITH input_point AS (
			SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326) AS geom
		)
		SELECT reg.id
		FROM regions reg
		CROSS JOIN input_point p
		WHERE ST_Covers(reg.geom, p.geom)
		ORDER BY
			CASE
				WHEN LOWER(COALESCE(reg.level, '')) IN ('district', 'kecamatan') THEN 0
				WHEN LOWER(COALESCE(reg.level, '')) IN ('subdistrict') THEN 1
				WHEN LOWER(COALESCE(reg.level, '')) IN ('city', 'kota') THEN 2
				WHEN LOWER(COALESCE(reg.level, '')) IN ('regency', 'kabupaten') THEN 3
				WHEN LOWER(COALESCE(reg.level, '')) IN ('province', 'provinsi') THEN 4
				ELSE 5
			END,
			ST_Area(reg.geom::geography) ASC
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

type duplicateCandidate struct {
	IssueID            uuid.UUID
	Status             string
	VerificationStatus string
	LastSeenAt         time.Time
	SeverityCurrent    int
	DistanceM          float64
}

func findDuplicateCandidate(ctx context.Context, tx pgx.Tx, lon, lat, radiusM float64) (*duplicateCandidate, error) {
	rows, err := tx.Query(ctx, `
		WITH input_point AS (
			SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography AS point
		)
		SELECT i.id,
		       i.status,
		       i.verification_status,
		       i.last_seen_at,
		       i.severity_current,
		       ST_Distance(i.public_location, input_point.point) AS distance_m
		FROM issues i
		CROSS JOIN input_point
		WHERE i.is_hidden = FALSE
		  AND i.status IN ('open', 'verified', 'in_progress')
		  AND ST_DWithin(i.public_location, input_point.point, $3)
		ORDER BY distance_m ASC, i.last_seen_at DESC
		LIMIT $4
	`, lon, lat, radiusM, duplicateCandidateQueryCap)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]duplicateCandidate, 0)
	for rows.Next() {
		var candidate duplicateCandidate
		if err := rows.Scan(
			&candidate.IssueID,
			&candidate.Status,
			&candidate.VerificationStatus,
			&candidate.LastSeenAt,
			&candidate.SeverityCurrent,
			&candidate.DistanceM,
		); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	best := pickBestDuplicateCandidate(candidates)

	log.Printf(
		"[REPORT] duplicate_candidate_found issue=%s candidates=%d distance_m=%.2f radius_m=%.1f",
		best.IssueID, len(candidates), best.DistanceM, radiusM,
	)
	return &best, nil
}

func createIssue(ctx context.Context, tx pgx.Tx, id uuid.UUID, regionID *int64, input SubmitInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO issues (
			id, status, severity_current, severity_max,
			public_location, region_id, road_name,
			submission_count, photo_count, casualty_count,
			first_seen_at, last_seen_at
		) VALUES (
			$1, 'open', $2, $2,
			ST_SetSRID(ST_MakePoint($3, $4), 4326)::geography,
			$5, $6, 1, $7, $8,
			NOW(), NOW()
		)
	`, id, input.Severity, input.Longitude, input.Latitude,
		regionID, input.RoadName, len(input.Media), incomingCasualtyCount(input),
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
		input.Severity, input.HasCasualty, incomingCasualtyCount(input), input.Note,
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
			return fmt.Errorf("%w: %v", ErrSubmissionMediaPersistFailed, err)
		}
	}
	return nil
}

// updateIssueAggregatesOnMerge updates issue aggregate fields after attaching a new submission.
func updateIssueAggregatesOnMerge(
	ctx context.Context,
	tx pgx.Tx,
	issueID uuid.UUID,
	regionID *int64,
	input SubmitInput,
) error {
	incomingPhotos := len(input.Media)
	incomingCasualty := incomingCasualtyCount(input)
	roadNameProvided := input.RoadName != nil
	regionProvided := regionID != nil

	_, err := tx.Exec(ctx, `
		UPDATE issues SET
			last_seen_at     = NOW(),
			submission_count = submission_count + 1,
			photo_count      = photo_count + $1::int,
			casualty_count   = GREATEST(casualty_count, $2::int),
			severity_current = GREATEST(severity_current, $3::int),
			severity_max     = GREATEST(severity_max, $3::int),
			road_name = CASE
				WHEN (road_name IS NULL OR BTRIM(road_name) = '')
				     AND $4::text IS NOT NULL AND BTRIM($4::text) <> ''
				THEN $4::text
				ELSE road_name
			END,
			region_id = CASE
				WHEN region_id IS NULL AND $5::bigint IS NOT NULL
				THEN $5::bigint
				ELSE region_id
			END,
			updated_at       = NOW()
		WHERE id = $6
	`, incomingPhotos, incomingCasualty, input.Severity, input.RoadName, regionID, issueID)
	if err != nil {
		log.Printf("[REPORT] update_aggregates_query_failed path=duplicate_merge issue=%s photo_inc=%d casualty_in=%d severity_in=%d road_name_provided=%t region_provided=%t error=%v",
			issueID, incomingPhotos, incomingCasualty, input.Severity, roadNameProvided, regionProvided, err)
	}
	return err
}

func incomingCasualtyCount(input SubmitInput) int {
	if !input.HasCasualty {
		return 0
	}
	return input.CasualtyCount
}

type issueAggregateState struct {
	SeverityCurrent int
	CasualtyCount   int
}

func getIssueAggregateState(ctx context.Context, tx pgx.Tx, issueID uuid.UUID) (*issueAggregateState, error) {
	var state issueAggregateState
	err := tx.QueryRow(ctx, `
		SELECT severity_current, casualty_count
		FROM issues
		WHERE id = $1
		FOR UPDATE
	`, issueID).Scan(&state.SeverityCurrent, &state.CasualtyCount)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func pickBestDuplicateCandidate(candidates []duplicateCandidate) duplicateCandidate {
	best := candidates[0]
	for i := 1; i < len(candidates); i++ {
		if isBetterDuplicateCandidate(candidates[i], best) {
			best = candidates[i]
		}
	}
	return best
}

func isBetterDuplicateCandidate(next, current duplicateCandidate) bool {
	if next.DistanceM != current.DistanceM {
		return next.DistanceM < current.DistanceM
	}
	nextStatusRank := rankDuplicateStatus(next.Status)
	currentStatusRank := rankDuplicateStatus(current.Status)
	if nextStatusRank != currentStatusRank {
		return nextStatusRank < currentStatusRank
	}
	nextVerificationRank := rankDuplicateVerification(next.VerificationStatus)
	currentVerificationRank := rankDuplicateVerification(current.VerificationStatus)
	if nextVerificationRank != currentVerificationRank {
		return nextVerificationRank < currentVerificationRank
	}
	if !next.LastSeenAt.Equal(current.LastSeenAt) {
		return next.LastSeenAt.After(current.LastSeenAt)
	}
	if next.SeverityCurrent != current.SeverityCurrent {
		return next.SeverityCurrent > current.SeverityCurrent
	}
	return next.IssueID.String() < current.IssueID.String()
}

func rankDuplicateStatus(status string) int {
	if rank, ok := duplicateStatusPriority[status]; ok {
		return rank
	}
	return 99
}

func rankDuplicateVerification(status string) int {
	if rank, ok := duplicateVerificationPriority[status]; ok {
		return rank
	}
	return 99
}

func (r *reportRepository) FindByClientRequestID(ctx context.Context, clientRequestID uuid.UUID) (*SubmitResult, error) {
	var issueID, submissionID uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT issue_id, id FROM issue_submissions WHERE client_request_id = $1 LIMIT 1
	`, clientRequestID).Scan(&issueID, &submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &SubmitResult{
		IssueID:      issueID,
		SubmissionID: submissionID,
		IsNewIssue:   false,
	}, nil
}

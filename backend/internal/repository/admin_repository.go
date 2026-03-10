package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type AdminRepository interface {
	ListIssues(ctx context.Context, limit, offset int, status *string) ([]*domain.AdminIssue, error)
	FindIssueByID(ctx context.Context, id uuid.UUID) (*domain.AdminIssue, error)
	FindIssueByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.AdminIssueDetail, error)
	UpdateIssueHidden(ctx context.Context, id uuid.UUID, isHidden bool, reason *string) error
	UpdateIssueStatus(ctx context.Context, id uuid.UUID, status string) error
	BanDevice(ctx context.Context, id uuid.UUID, reason *string) error
	CreateModerationAction(ctx context.Context, actionType, targetType string, targetID uuid.UUID, adminUsername string, note *string) error
	GetModerationLog(ctx context.Context, targetType string, targetID uuid.UUID) ([]*domain.ModerationAction, error)
	AdjustSubmitterTrustScores(ctx context.Context, issueID uuid.UUID, delta int) error
}

type adminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) AdminRepository {
	return &adminRepository{db: db}
}

const adminIssueCols = `
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
	i.is_hidden,
	i.first_seen_at,
	i.last_seen_at,
	i.created_at,
	i.updated_at
`

func scanAdminIssue(row pgx.Row, i *domain.AdminIssue) error {
	return row.Scan(
		&i.ID, &i.Status, &i.VerificationStatus, &i.SeverityCurrent, &i.SeverityMax,
		&i.Longitude, &i.Latitude,
		&i.RegionID, &i.RoadName, &i.RoadType,
		&i.SubmissionCount, &i.PhotoCount, &i.CasualtyCount, &i.ReactionCount, &i.FlagCount,
		&i.IsHidden,
		&i.FirstSeenAt, &i.LastSeenAt, &i.CreatedAt, &i.UpdatedAt,
	)
}

func (r *adminRepository) ListIssues(ctx context.Context, limit, offset int, status *string) ([]*domain.AdminIssue, error) {
	query := `SELECT ` + adminIssueCols + ` FROM issues i`
	args := []any{}
	n := 1

	if status != nil {
		query += ` WHERE i.status = $1`
		args = append(args, *status)
		n++
	}

	query += fmt.Sprintf(` ORDER BY i.last_seen_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := make([]*domain.AdminIssue, 0)
	for rows.Next() {
		var i domain.AdminIssue
		if err := rows.Scan(
			&i.ID, &i.Status, &i.VerificationStatus, &i.SeverityCurrent, &i.SeverityMax,
			&i.Longitude, &i.Latitude,
			&i.RegionID, &i.RoadName, &i.RoadType,
			&i.SubmissionCount, &i.PhotoCount, &i.CasualtyCount, &i.ReactionCount, &i.FlagCount,
			&i.IsHidden,
			&i.FirstSeenAt, &i.LastSeenAt, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		issues = append(issues, &i)
	}
	return issues, rows.Err()
}

func (r *adminRepository) FindIssueByID(ctx context.Context, id uuid.UUID) (*domain.AdminIssue, error) {
	query := `SELECT ` + adminIssueCols + ` FROM issues i WHERE i.id = $1 LIMIT 1`
	var i domain.AdminIssue
	err := scanAdminIssue(r.db.QueryRow(ctx, query, id), &i)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (r *adminRepository) FindIssueByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.AdminIssueDetail, error) {
	issue, err := r.FindIssueByID(ctx, id)
	if err != nil || issue == nil {
		return nil, err
	}

	// Media (up to 20)
	mediaRows, err := r.db.Query(ctx, `
		SELECT sm.id, sm.object_key, sm.mime_type, sm.size_bytes,
		       sm.width, sm.height, sm.blurhash, sm.is_primary, sm.created_at
		FROM submission_media sm
		JOIN issue_submissions s ON s.id = sm.submission_id
		WHERE s.issue_id = $1
		ORDER BY sm.is_primary DESC, sm.created_at DESC
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

	// Submissions with device info (up to 20)
	subRows, err := r.db.Query(ctx, `
		SELECT s.id, s.device_id, d.is_banned, s.status, s.severity, s.has_casualty, s.note, s.reported_at
		FROM issue_submissions s
		JOIN devices d ON d.id = s.device_id
		WHERE s.issue_id = $1
		ORDER BY s.reported_at DESC
		LIMIT 20
	`, id)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	subs := make([]*domain.AdminSubmissionSummary, 0)
	for subRows.Next() {
		var s domain.AdminSubmissionSummary
		if err := subRows.Scan(
			&s.ID, &s.DeviceID, &s.DeviceIsBanned, &s.Status, &s.Severity, &s.HasCasualty, &s.Note, &s.ReportedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	if err := subRows.Err(); err != nil {
		return nil, err
	}

	// Moderation log for this issue
	modLog, err := r.GetModerationLog(ctx, "issue", id)
	if err != nil {
		return nil, err
	}

	return &domain.AdminIssueDetail{
		AdminIssue:    issue,
		Media:         media,
		Submissions:   subs,
		ModerationLog: modLog,
	}, nil
}

func (r *adminRepository) UpdateIssueHidden(ctx context.Context, id uuid.UUID, isHidden bool, reason *string) error {
	_, err := r.db.Exec(ctx, `UPDATE issues SET is_hidden = $1, hidden_reason = $2 WHERE id = $3`, isHidden, reason, id)
	return err
}

func (r *adminRepository) UpdateIssueStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE issues SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *adminRepository) BanDevice(ctx context.Context, id uuid.UUID, reason *string) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET is_banned = TRUE, ban_reason = $1, trust_score = -100 WHERE id = $2`, reason, id)
	return err
}

func (r *adminRepository) CreateModerationAction(ctx context.Context, actionType, targetType string, targetID uuid.UUID, adminUsername string, note *string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO moderation_actions (action_type, target_type, target_id, admin_username, note)
		VALUES ($1, $2, $3, $4, $5)
	`, actionType, targetType, targetID, adminUsername, note)
	return err
}

func (r *adminRepository) GetModerationLog(ctx context.Context, targetType string, targetID uuid.UUID) ([]*domain.ModerationAction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, action_type, target_type, target_id, admin_username, note, created_at
		FROM moderation_actions
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at DESC
		LIMIT 20
	`, targetType, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]*domain.ModerationAction, 0)
	for rows.Next() {
		var a domain.ModerationAction
		if err := rows.Scan(
			&a.ID, &a.ActionType, &a.TargetType, &a.TargetID, &a.AdminUsername, &a.Note, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		actions = append(actions, &a)
	}
	return actions, rows.Err()
}

func (r *adminRepository) AdjustSubmitterTrustScores(ctx context.Context, issueID uuid.UUID, delta int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE devices SET trust_score = trust_score + $1
		WHERE id IN (
			SELECT DISTINCT device_id FROM issue_submissions WHERE issue_id = $2
		)
	`, delta, issueID)
	return err
}



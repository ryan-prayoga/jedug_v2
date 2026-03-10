package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDuplicateFlag is returned when a device tries to flag the same issue twice.
var ErrDuplicateFlag = errors.New("duplicate flag")

type FlagRepository interface {
	CreateIssueFlag(ctx context.Context, issueID, deviceID uuid.UUID, reason string, note *string) error
	CountUniqueIssueFlags(ctx context.Context, issueID uuid.UUID) (int, error)
}

type flagRepository struct {
	db *pgxpool.Pool
}

func NewFlagRepository(db *pgxpool.Pool) FlagRepository {
	return &flagRepository{db: db}
}

func (r *flagRepository) CreateIssueFlag(ctx context.Context, issueID, deviceID uuid.UUID, reason string, note *string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO issue_flags (issue_id, device_id, reason, note)
		VALUES ($1, $2, $3, $4)
	`, issueID, deviceID, reason, note)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateFlag
		}
		return err
	}

	// Keep flag_count in sync using unique device count (not raw insert count)
	_, err = r.db.Exec(ctx, `
		UPDATE issues SET flag_count = (
			SELECT COUNT(DISTINCT device_id) FROM issue_flags WHERE issue_id = $1
		) WHERE id = $1
	`, issueID)
	return err
}

func (r *flagRepository) CountUniqueIssueFlags(ctx context.Context, issueID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT device_id) FROM issue_flags WHERE issue_id = $1
	`, issueID).Scan(&count)
	return count, err
}

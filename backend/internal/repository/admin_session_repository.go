package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminSessionRecord is a persisted admin session row (auth-relevant fields only).
type AdminSessionRecord struct {
	Username  string
	ExpiresAt time.Time
}

// AdminSessionRepository persists admin sessions so they survive restarts.
// Tokens are never stored raw; callers pass the SHA-256 hash of the token.
type AdminSessionRepository interface {
	// Create atomically replaces any existing session for username (single
	// active session per admin) and inserts the new one.
	Create(ctx context.Context, tokenHash, username string, expiresAt time.Time) error
	// FindValid returns the session for tokenHash only if it is not revoked
	// and not expired at now. Returns (nil, nil) when no valid row exists.
	FindValid(ctx context.Context, tokenHash string, now time.Time) (*AdminSessionRecord, error)
	// Delete removes a single session by token hash (logout/revoke).
	Delete(ctx context.Context, tokenHash string) error
	// DeleteExpired removes sessions whose expiry is older than cutoff.
	DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
}

type adminSessionRepository struct {
	db *pgxpool.Pool
}

func NewAdminSessionRepository(db *pgxpool.Pool) AdminSessionRepository {
	return &adminSessionRepository{db: db}
}

func (r *adminSessionRepository) Create(ctx context.Context, tokenHash, username string, expiresAt time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if _, err := tx.Exec(ctx, `DELETE FROM admin_sessions WHERE username = $1`, username); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO admin_sessions (token_hash, username, expires_at)
		VALUES ($1, $2, $3)
	`, tokenHash, username, expiresAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *adminSessionRepository) FindValid(ctx context.Context, tokenHash string, now time.Time) (*AdminSessionRecord, error) {
	var rec AdminSessionRecord
	err := r.db.QueryRow(ctx, `
		SELECT username, expires_at
		FROM admin_sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`, tokenHash, now).Scan(&rec.Username, &rec.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *adminSessionRepository) Delete(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *adminSessionRepository) DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM admin_sessions WHERE expires_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

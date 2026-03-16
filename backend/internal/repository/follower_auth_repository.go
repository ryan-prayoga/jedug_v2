package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type FollowerAuthRepository interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.FollowerAuthBinding, error)
	Upsert(ctx context.Context, followerID uuid.UUID, deviceTokenHash string) error
	HasFootprint(ctx context.Context, followerID uuid.UUID) (bool, error)
}

type followerAuthRepository struct {
	db *pgxpool.Pool
}

func NewFollowerAuthRepository(db *pgxpool.Pool) FollowerAuthRepository {
	return &followerAuthRepository{db: db}
}

func (r *followerAuthRepository) GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.FollowerAuthBinding, error) {
	var binding domain.FollowerAuthBinding
	err := r.db.QueryRow(ctx, `
		SELECT follower_id, device_token_hash, created_at, updated_at
		FROM follower_auth_bindings
		WHERE follower_id = $1
	`, followerID).Scan(
		&binding.FollowerID,
		&binding.DeviceTokenHash,
		&binding.CreatedAt,
		&binding.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

func (r *followerAuthRepository) Upsert(ctx context.Context, followerID uuid.UUID, deviceTokenHash string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO follower_auth_bindings (follower_id, device_token_hash)
		VALUES ($1, $2)
		ON CONFLICT (follower_id) DO UPDATE
		SET device_token_hash = EXCLUDED.device_token_hash,
			updated_at = NOW()
	`, followerID, deviceTokenHash)
	return err
}

func (r *followerAuthRepository) HasFootprint(ctx context.Context, followerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM issue_followers WHERE follower_id = $1
			UNION ALL
			SELECT 1 FROM notifications WHERE follower_id = $1
			UNION ALL
			SELECT 1 FROM push_subscriptions WHERE follower_id = $1
			UNION ALL
			SELECT 1 FROM nearby_alert_subscriptions WHERE follower_id = $1
		)
	`, followerID).Scan(&exists)
	return exists, err
}

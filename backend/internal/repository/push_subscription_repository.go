package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"jedug_backend/internal/domain"
)

type PushSubscriptionUpsertInput struct {
	FollowerID uuid.UUID
	Endpoint   string
	P256DH     string
	Auth       string
	UserAgent  *string
}

type PushSubscriptionRepository interface {
	Upsert(ctx context.Context, input PushSubscriptionUpsertInput) (*domain.PushSubscription, error)
	Disable(ctx context.Context, followerID uuid.UUID, endpoint string) (bool, error)
	DisableByEndpoint(ctx context.Context, endpoint string) error
	CountActiveByFollowerID(ctx context.Context, followerID uuid.UUID) (int, error)
	GetActiveByFollowerIDs(ctx context.Context, followerIDs []uuid.UUID) (map[uuid.UUID][]*domain.PushSubscription, error)
}

type pushSubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewPushSubscriptionRepository(db *pgxpool.Pool) PushSubscriptionRepository {
	return &pushSubscriptionRepository{db: db}
}

func (r *pushSubscriptionRepository) Upsert(ctx context.Context, input PushSubscriptionUpsertInput) (*domain.PushSubscription, error) {
	var subscription domain.PushSubscription
	err := r.db.QueryRow(ctx, `
		INSERT INTO push_subscriptions (
			id,
			follower_id,
			endpoint,
			p256dh,
			auth,
			user_agent
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		ON CONFLICT (endpoint) DO UPDATE
		SET follower_id = EXCLUDED.follower_id,
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth,
			user_agent = EXCLUDED.user_agent,
			disabled_at = NULL,
			updated_at = NOW()
		RETURNING id, follower_id, endpoint, p256dh, auth, user_agent, created_at, updated_at, disabled_at
	`, input.FollowerID, input.Endpoint, input.P256DH, input.Auth, input.UserAgent).Scan(
		&subscription.ID,
		&subscription.FollowerID,
		&subscription.Endpoint,
		&subscription.P256DH,
		&subscription.Auth,
		&subscription.UserAgent,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
		&subscription.DisabledAt,
	)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *pushSubscriptionRepository) Disable(ctx context.Context, followerID uuid.UUID, endpoint string) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE push_subscriptions
		SET disabled_at = COALESCE(disabled_at, NOW()),
			updated_at = NOW()
		WHERE follower_id = $1
		  AND endpoint = $2
		  AND disabled_at IS NULL
	`, followerID, endpoint)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *pushSubscriptionRepository) DisableByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE push_subscriptions
		SET disabled_at = COALESCE(disabled_at, NOW()),
			updated_at = NOW()
		WHERE endpoint = $1
		  AND disabled_at IS NULL
	`, endpoint)
	return err
}

func (r *pushSubscriptionRepository) CountActiveByFollowerID(ctx context.Context, followerID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM push_subscriptions
		WHERE follower_id = $1
		  AND disabled_at IS NULL
	`, followerID).Scan(&count)
	return count, err
}

func (r *pushSubscriptionRepository) GetActiveByFollowerIDs(ctx context.Context, followerIDs []uuid.UUID) (map[uuid.UUID][]*domain.PushSubscription, error) {
	result := make(map[uuid.UUID][]*domain.PushSubscription, len(followerIDs))
	if len(followerIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, follower_id, endpoint, p256dh, auth, user_agent, created_at, updated_at, disabled_at
		FROM push_subscriptions
		WHERE follower_id = ANY($1::uuid[])
		  AND disabled_at IS NULL
		ORDER BY created_at DESC
	`, followerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var subscription domain.PushSubscription
		if err := rows.Scan(
			&subscription.ID,
			&subscription.FollowerID,
			&subscription.Endpoint,
			&subscription.P256DH,
			&subscription.Auth,
			&subscription.UserAgent,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.DisabledAt,
		); err != nil {
			return nil, err
		}
		result[subscription.FollowerID] = append(result[subscription.FollowerID], &subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, followerID := range followerIDs {
		if _, ok := result[followerID]; !ok {
			result[followerID] = nil
		}
	}

	return result, nil
}

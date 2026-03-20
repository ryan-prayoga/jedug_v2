package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushDeliveryJob struct {
	ID           uuid.UUID
	FollowerID   uuid.UUID
	IssueID      uuid.UUID
	EventID      int64
	Type         string
	Title        string
	Message      string
	AttemptCount int
	CreatedAt    time.Time
}

type PushDeliveryJobRepository interface {
	EnqueueBatch(ctx context.Context, deliveries []PushDelivery) error
	ClaimBatch(ctx context.Context, limit int, lockTimeout time.Duration) ([]*PushDeliveryJob, error)
	MarkDelivered(ctx context.Context, jobID uuid.UUID) error
	MarkRetry(ctx context.Context, jobID uuid.UUID, nextAttemptAt time.Time, lastError string) error
	MarkFailed(ctx context.Context, jobID uuid.UUID, lastError string) error
}

type pushDeliveryJobRepository struct {
	db *pgxpool.Pool
}

func NewPushDeliveryJobRepository(db *pgxpool.Pool) PushDeliveryJobRepository {
	return &pushDeliveryJobRepository{db: db}
}

func (r *pushDeliveryJobRepository) EnqueueBatch(ctx context.Context, deliveries []PushDelivery) error {
	if len(deliveries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, delivery := range deliveries {
		batch.Queue(`
			INSERT INTO push_delivery_jobs (
				id,
				follower_id,
				issue_id,
				event_id,
				type,
				title,
				message
			)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
			ON CONFLICT (event_id, follower_id) DO NOTHING
		`, delivery.FollowerID, delivery.IssueID, delivery.EventID, delivery.Type, delivery.Title, delivery.Message)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range deliveries {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *pushDeliveryJobRepository) ClaimBatch(ctx context.Context, limit int, lockTimeout time.Duration) ([]*PushDeliveryJob, error) {
	if limit <= 0 {
		limit = 20
	}
	if lockTimeout <= 0 {
		lockTimeout = 2 * time.Minute
	}

	rows, err := r.db.Query(ctx, `
		WITH candidates AS (
			SELECT id
			FROM push_delivery_jobs
			WHERE delivered_at IS NULL
			  AND failed_at IS NULL
			  AND next_attempt_at <= NOW()
			  AND (locked_at IS NULL OR locked_at <= NOW() - ($2::bigint * INTERVAL '1 second'))
			ORDER BY next_attempt_at ASC, created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE push_delivery_jobs j
		SET locked_at = NOW(),
			last_attempt_at = NOW(),
			attempt_count = j.attempt_count + 1,
			updated_at = NOW()
		FROM candidates
		WHERE j.id = candidates.id
		RETURNING j.id, j.follower_id, j.issue_id, j.event_id, j.type, j.title, j.message, j.attempt_count, j.created_at
	`, limit, int64(lockTimeout.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*PushDeliveryJob, 0)
	for rows.Next() {
		var job PushDeliveryJob
		if err := rows.Scan(
			&job.ID,
			&job.FollowerID,
			&job.IssueID,
			&job.EventID,
			&job.Type,
			&job.Title,
			&job.Message,
			&job.AttemptCount,
			&job.CreatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}

	return jobs, rows.Err()
}

func (r *pushDeliveryJobRepository) MarkDelivered(ctx context.Context, jobID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE push_delivery_jobs
		SET delivered_at = NOW(),
			locked_at = NULL,
			last_error = NULL,
			updated_at = NOW()
		WHERE id = $1
	`, jobID)
	return err
}

func (r *pushDeliveryJobRepository) MarkRetry(ctx context.Context, jobID uuid.UUID, nextAttemptAt time.Time, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE push_delivery_jobs
		SET locked_at = NULL,
			next_attempt_at = $2,
			last_error = NULLIF($3, ''),
			updated_at = NOW()
		WHERE id = $1
	`, jobID, nextAttemptAt, lastError)
	return err
}

func (r *pushDeliveryJobRepository) MarkFailed(ctx context.Context, jobID uuid.UUID, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE push_delivery_jobs
		SET failed_at = NOW(),
			locked_at = NULL,
			last_error = NULLIF($2, ''),
			updated_at = NOW()
		WHERE id = $1
	`, jobID, lastError)
	return err
}

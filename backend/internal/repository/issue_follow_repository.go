package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IssueFollowRepository interface {
	FollowIssue(ctx context.Context, issueID, followerID uuid.UUID) error
	UnfollowIssue(ctx context.Context, issueID, followerID uuid.UUID) error
	CountFollowers(ctx context.Context, issueID uuid.UUID) (int, error)
	IsFollowing(ctx context.Context, issueID, followerID uuid.UUID) (bool, error)
}

type issueFollowRepository struct {
	db *pgxpool.Pool
}

func NewIssueFollowRepository(db *pgxpool.Pool) IssueFollowRepository {
	return &issueFollowRepository{db: db}
}

func (r *issueFollowRepository) FollowIssue(ctx context.Context, issueID, followerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO issue_followers (id, issue_id, follower_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (issue_id, follower_id) DO NOTHING
	`, uuid.New(), issueID, followerID)
	return err
}

func (r *issueFollowRepository) UnfollowIssue(ctx context.Context, issueID, followerID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM issue_followers
		WHERE issue_id = $1 AND follower_id = $2
	`, issueID, followerID)
	return err
}

func (r *issueFollowRepository) CountFollowers(ctx context.Context, issueID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM issue_followers
		WHERE issue_id = $1
	`, issueID).Scan(&count)
	return count, err
}

func (r *issueFollowRepository) IsFollowing(ctx context.Context, issueID, followerID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM issue_followers
			WHERE issue_id = $1 AND follower_id = $2
		)
	`, issueID, followerID).Scan(&exists)
	return exists, err
}

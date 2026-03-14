package service

import (
	"context"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type IssueFollowService interface {
	Follow(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error)
	Unfollow(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error)
	GetStatus(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error)
	GetCount(ctx context.Context, issueID uuid.UUID) (*domain.IssueFollowersCount, error)
}

type issueFollowService struct {
	issueRepo  repository.IssueRepository
	followRepo repository.IssueFollowRepository
}

func NewIssueFollowService(issueRepo repository.IssueRepository, followRepo repository.IssueFollowRepository) IssueFollowService {
	return &issueFollowService{
		issueRepo:  issueRepo,
		followRepo: followRepo,
	}
}

func (s *issueFollowService) Follow(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error) {
	if err := s.ensurePublicIssueExists(ctx, issueID); err != nil {
		return nil, err
	}

	if err := s.followRepo.FollowIssue(ctx, issueID, followerID); err != nil {
		return nil, err
	}

	return s.buildState(ctx, issueID, true)
}

func (s *issueFollowService) Unfollow(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error) {
	if err := s.ensurePublicIssueExists(ctx, issueID); err != nil {
		return nil, err
	}

	if err := s.followRepo.UnfollowIssue(ctx, issueID, followerID); err != nil {
		return nil, err
	}

	return s.buildState(ctx, issueID, false)
}

func (s *issueFollowService) GetStatus(ctx context.Context, issueID, followerID uuid.UUID) (*domain.IssueFollowState, error) {
	if err := s.ensurePublicIssueExists(ctx, issueID); err != nil {
		return nil, err
	}

	following, err := s.followRepo.IsFollowing(ctx, issueID, followerID)
	if err != nil {
		return nil, err
	}

	return s.buildState(ctx, issueID, following)
}

func (s *issueFollowService) GetCount(ctx context.Context, issueID uuid.UUID) (*domain.IssueFollowersCount, error) {
	if err := s.ensurePublicIssueExists(ctx, issueID); err != nil {
		return nil, err
	}

	count, err := s.followRepo.CountFollowers(ctx, issueID)
	if err != nil {
		return nil, err
	}

	return &domain.IssueFollowersCount{FollowersCount: count}, nil
}

func (s *issueFollowService) ensurePublicIssueExists(ctx context.Context, issueID uuid.UUID) error {
	issue, err := s.issueRepo.FindByID(ctx, issueID)
	if err != nil {
		return err
	}
	if issue == nil {
		return ErrIssueNotFound
	}
	return nil
}

func (s *issueFollowService) buildState(ctx context.Context, issueID uuid.UUID, following bool) (*domain.IssueFollowState, error) {
	count, err := s.followRepo.CountFollowers(ctx, issueID)
	if err != nil {
		return nil, err
	}

	return &domain.IssueFollowState{
		Following:      following,
		FollowersCount: count,
	}, nil
}

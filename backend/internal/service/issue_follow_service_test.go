package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type issueRepoStub struct {
	issue *domain.Issue
	err   error
}

func (s *issueRepoStub) List(_ context.Context, _ int, _ int, _ *string, _ *int, _ *repository.BBoxFilter) ([]*domain.Issue, error) {
	return nil, nil
}

func (s *issueRepoStub) FindByID(_ context.Context, _ uuid.UUID) (*domain.Issue, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.issue, nil
}

func (s *issueRepoStub) FindByIDWithDetail(_ context.Context, _ uuid.UUID) (*domain.IssueDetail, error) {
	return nil, nil
}

func (s *issueRepoStub) ListTimeline(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*domain.IssueTimelineEvent, error) {
	return nil, nil
}

type issueFollowRepoStub struct {
	followCalls   int
	unfollowCalls int
	following     bool
	count         int
	err           error
}

func (s *issueFollowRepoStub) FollowIssue(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	s.followCalls++
	if s.err != nil {
		return s.err
	}
	s.following = true
	return nil
}

func (s *issueFollowRepoStub) UnfollowIssue(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	s.unfollowCalls++
	if s.err != nil {
		return s.err
	}
	s.following = false
	return nil
}

func (s *issueFollowRepoStub) CountFollowers(_ context.Context, _ uuid.UUID) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	return s.count, nil
}

func (s *issueFollowRepoStub) IsFollowing(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.following, nil
}

func TestIssueFollowServiceFollowReturnsState(t *testing.T) {
	issueRepo := &issueRepoStub{
		issue: &domain.Issue{ID: uuid.New(), CreatedAt: time.Now()},
	}
	followRepo := &issueFollowRepoStub{count: 12}
	svc := NewIssueFollowService(issueRepo, followRepo)

	state, err := svc.Follow(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("follow error: %v", err)
	}
	if !state.Following {
		t.Fatalf("expected following=true")
	}
	if state.FollowersCount != 12 {
		t.Fatalf("followers_count mismatch: got %d want 12", state.FollowersCount)
	}
	if followRepo.followCalls != 1 {
		t.Fatalf("expected follow to be called once, got %d", followRepo.followCalls)
	}
}

func TestIssueFollowServiceUnfollowReturnsState(t *testing.T) {
	issueRepo := &issueRepoStub{
		issue: &domain.Issue{ID: uuid.New(), CreatedAt: time.Now()},
	}
	followRepo := &issueFollowRepoStub{count: 4, following: true}
	svc := NewIssueFollowService(issueRepo, followRepo)

	state, err := svc.Unfollow(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unfollow error: %v", err)
	}
	if state.Following {
		t.Fatalf("expected following=false")
	}
	if state.FollowersCount != 4 {
		t.Fatalf("followers_count mismatch: got %d want 4", state.FollowersCount)
	}
	if followRepo.unfollowCalls != 1 {
		t.Fatalf("expected unfollow to be called once, got %d", followRepo.unfollowCalls)
	}
}

func TestIssueFollowServiceReturnsNotFoundForMissingIssue(t *testing.T) {
	issueRepo := &issueRepoStub{}
	followRepo := &issueFollowRepoStub{}
	svc := NewIssueFollowService(issueRepo, followRepo)

	_, err := svc.GetCount(context.Background(), uuid.New())
	if !errors.Is(err, ErrIssueNotFound) {
		t.Fatalf("expected ErrIssueNotFound, got %v", err)
	}
}

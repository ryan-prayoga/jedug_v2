package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
)

type followerAuthRepoStub struct {
	bindings     map[uuid.UUID]*domain.FollowerAuthBinding
	hasFootprint bool
}

func (s *followerAuthRepoStub) GetByFollowerID(_ context.Context, followerID uuid.UUID) (*domain.FollowerAuthBinding, error) {
	return s.bindings[followerID], nil
}

func (s *followerAuthRepoStub) Upsert(_ context.Context, followerID uuid.UUID, deviceTokenHash string) error {
	if s.bindings == nil {
		s.bindings = make(map[uuid.UUID]*domain.FollowerAuthBinding)
	}
	now := time.Now().UTC()
	s.bindings[followerID] = &domain.FollowerAuthBinding{
		FollowerID:      followerID,
		DeviceTokenHash: deviceTokenHash,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return nil
}

func (s *followerAuthRepoStub) HasFootprint(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.hasFootprint, nil
}

type issueFollowRepoForAuthStub struct {
	following bool
}

func (s *issueFollowRepoForAuthStub) FollowIssue(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *issueFollowRepoForAuthStub) UnfollowIssue(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (s *issueFollowRepoForAuthStub) CountFollowers(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (s *issueFollowRepoForAuthStub) IsFollowing(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return s.following, nil
}

func TestFollowerAuthServiceIssueAndAuthenticateToken(t *testing.T) {
	repo := &followerAuthRepoStub{}
	followRepo := &issueFollowRepoForAuthStub{}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret: []byte("01234567890123456789012345678901"),
		TTL:    time.Hour,
	}).(*followerAuthService)

	followerID := uuid.New()
	issueID := uuid.New()
	token, err := svc.IssueForFollowMutation(context.Background(), issueID, followerID, "device-token-123")
	if err != nil {
		t.Fatalf("IssueForFollowMutation returned error: %v", err)
	}
	if token.Token == "" {
		t.Fatalf("expected non-empty follower token")
	}

	authFollowerID, err := svc.Authenticate(context.Background(), token.Token)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if authFollowerID != followerID {
		t.Fatalf("Authenticate follower mismatch: got %s want %s", authFollowerID, followerID)
	}
}

func TestFollowerAuthServiceRejectsBindingMismatch(t *testing.T) {
	existingFollowerID := uuid.New()
	repo := &followerAuthRepoStub{
		bindings: map[uuid.UUID]*domain.FollowerAuthBinding{
			existingFollowerID: {
				FollowerID:      existingFollowerID,
				DeviceTokenHash: "other-hash",
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			},
		},
	}
	followRepo := &issueFollowRepoForAuthStub{}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret: []byte("01234567890123456789012345678901"),
		TTL:    time.Hour,
	})

	_, err := svc.IssueForNotificationAccess(context.Background(), existingFollowerID, "device-token-123")
	if err != ErrFollowerBindingMismatch {
		t.Fatalf("expected ErrFollowerBindingMismatch, got %v", err)
	}
}

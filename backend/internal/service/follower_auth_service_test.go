package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
)

type followerAuthRepoStub struct {
	bindings map[uuid.UUID]*domain.FollowerAuthBinding
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
	followRepo := &issueFollowRepoForAuthStub{following: true}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
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
	if token.StreamToken == "" {
		t.Fatalf("expected non-empty follower stream token")
	}
	if token.StreamExpiresAt == nil {
		t.Fatalf("expected stream token expiry")
	}

	authFollowerID, err := svc.AuthenticateNotificationAccess(context.Background(), token.Token, "device-token-123")
	if err != nil {
		t.Fatalf("AuthenticateNotificationAccess returned error: %v", err)
	}
	if authFollowerID != followerID {
		t.Fatalf("AuthenticateNotificationAccess follower mismatch: got %s want %s", authFollowerID, followerID)
	}

	streamFollowerID, err := svc.AuthenticateNotificationStream(context.Background(), token.StreamToken)
	if err != nil {
		t.Fatalf("AuthenticateNotificationStream returned error: %v", err)
	}
	if streamFollowerID != followerID {
		t.Fatalf("AuthenticateNotificationStream follower mismatch: got %s want %s", streamFollowerID, followerID)
	}
}

func TestFollowerAuthServiceAuthorizeFollowMutationRejectsBindingMismatch(t *testing.T) {
	followerID := uuid.New()
	repo := &followerAuthRepoStub{
		bindings: map[uuid.UUID]*domain.FollowerAuthBinding{
			followerID: {
				FollowerID:      followerID,
				DeviceTokenHash: "other-hash",
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			},
		},
	}
	followRepo := &issueFollowRepoForAuthStub{}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
	})

	if err := svc.AuthorizeFollowMutation(context.Background(), followerID, "device-token-123"); err != ErrFollowerBindingMismatch {
		t.Fatalf("expected ErrFollowerBindingMismatch, got %v", err)
	}
}

func TestFollowerAuthServiceRejectsFollowMutationClaimWithoutExistingFollow(t *testing.T) {
	repo := &followerAuthRepoStub{}
	followRepo := &issueFollowRepoForAuthStub{}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
	})

	_, err := svc.IssueForFollowMutation(context.Background(), uuid.New(), uuid.New(), "device-token-123")
	if err != ErrFollowerBindingNotFound {
		t.Fatalf("expected ErrFollowerBindingNotFound, got %v", err)
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
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
	})

	_, err := svc.IssueForNotificationAccess(context.Background(), existingFollowerID, "device-token-123")
	if err != ErrFollowerBindingMismatch {
		t.Fatalf("expected ErrFollowerBindingMismatch, got %v", err)
	}
}

func TestFollowerAuthServiceRejectsNotificationAccessWithoutMatchingDeviceToken(t *testing.T) {
	followerID := uuid.New()
	repo := &followerAuthRepoStub{}
	followRepo := &issueFollowRepoForAuthStub{following: true}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
	}).(*followerAuthService)

	token, err := svc.IssueForFollowMutation(context.Background(), uuid.New(), followerID, "device-token-123")
	if err != nil {
		t.Fatalf("IssueForFollowMutation returned error: %v", err)
	}

	_, err = svc.AuthenticateNotificationAccess(context.Background(), token.Token, "other-device-token")
	if err != ErrFollowerBindingMismatch {
		t.Fatalf("expected ErrFollowerBindingMismatch, got %v", err)
	}
}

func TestFollowerAuthServiceRejectsWrongTokenPurpose(t *testing.T) {
	followerID := uuid.New()
	repo := &followerAuthRepoStub{}
	followRepo := &issueFollowRepoForAuthStub{following: true}
	svc := NewFollowerAuthService(repo, followRepo, FollowerAuthServiceConfig{
		Secret:    []byte("01234567890123456789012345678901"),
		TTL:       time.Hour,
		StreamTTL: 5 * time.Minute,
	}).(*followerAuthService)

	token, err := svc.IssueForFollowMutation(context.Background(), uuid.New(), followerID, "device-token-123")
	if err != nil {
		t.Fatalf("IssueForFollowMutation returned error: %v", err)
	}

	if _, err := svc.AuthenticateNotificationAccess(context.Background(), token.StreamToken, "device-token-123"); err != ErrFollowerTokenInvalid {
		t.Fatalf("expected ErrFollowerTokenInvalid for stream token on notification access, got %v", err)
	}
	if _, err := svc.AuthenticateNotificationStream(context.Background(), token.Token); err != ErrFollowerTokenInvalid {
		t.Fatalf("expected ErrFollowerTokenInvalid for access token on stream auth, got %v", err)
	}
}

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

type nearbyAlertRepoStub struct {
	items      []*domain.NearbyAlertSubscription
	count      int
	createErr  error
	updateErr  error
	deleteErr  error
	deleted    bool
	updated    *domain.NearbyAlertSubscription
	created    *domain.NearbyAlertSubscription
	createSeen repository.NearbyAlertCreateInput
	updateSeen domain.NearbyAlertSubscriptionPatch
}

func (s *nearbyAlertRepoStub) ListByFollowerID(_ context.Context, _ uuid.UUID) ([]*domain.NearbyAlertSubscription, error) {
	return s.items, nil
}

func (s *nearbyAlertRepoStub) CountByFollowerID(_ context.Context, _ uuid.UUID) (int, error) {
	return s.count, nil
}

func (s *nearbyAlertRepoStub) Create(_ context.Context, input repository.NearbyAlertCreateInput) (*domain.NearbyAlertSubscription, error) {
	s.createSeen = input
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.created != nil {
		return s.created, nil
	}
	now := time.Now().UTC()
	return &domain.NearbyAlertSubscription{
		ID:         uuid.New(),
		FollowerID: input.FollowerID,
		Label:      input.Label,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		RadiusM:    input.RadiusM,
		Enabled:    input.Enabled,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (s *nearbyAlertRepoStub) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, patch domain.NearbyAlertSubscriptionPatch) (*domain.NearbyAlertSubscription, error) {
	s.updateSeen = patch
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return s.updated, nil
}

func (s *nearbyAlertRepoStub) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	if s.deleteErr != nil {
		return false, s.deleteErr
	}
	return s.deleted, nil
}

func TestNearbyAlertServiceCreateDefaultsEnabledAndTrimsLabel(t *testing.T) {
	repo := &nearbyAlertRepoStub{}
	svc := NewNearbyAlertService(repo)
	label := "  Rumah  "

	created, err := svc.Create(context.Background(), uuid.New(), NearbyAlertCreateInput{
		Label:     &label,
		Latitude:  -6.2,
		Longitude: 106.8,
		RadiusM:   500,
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if created == nil {
		t.Fatalf("expected created subscription")
	}
	if !repo.createSeen.Enabled {
		t.Fatalf("expected enabled default true")
	}
	if repo.createSeen.Label == nil || *repo.createSeen.Label != "Rumah" {
		t.Fatalf("expected trimmed label, got %#v", repo.createSeen.Label)
	}
}

func TestNearbyAlertServiceCreateRejectsLimitExceeded(t *testing.T) {
	repo := &nearbyAlertRepoStub{count: NearbyAlertMaxSubscriptions}
	svc := NewNearbyAlertService(repo)

	_, err := svc.Create(context.Background(), uuid.New(), NearbyAlertCreateInput{
		Latitude:  -6.2,
		Longitude: 106.8,
		RadiusM:   500,
	})
	if !errors.Is(err, ErrNearbyAlertLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestNearbyAlertServiceUpdateRequiresCoordinatePair(t *testing.T) {
	repo := &nearbyAlertRepoStub{}
	svc := NewNearbyAlertService(repo)
	lat := -6.2

	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), domain.NearbyAlertSubscriptionPatch{
		Latitude: &lat,
	})
	if !errors.Is(err, ErrNearbyAlertCoordinatePairRequired) {
		t.Fatalf("expected coordinate pair error, got %v", err)
	}
}

func TestNearbyAlertServiceDeleteReturnsNotFound(t *testing.T) {
	repo := &nearbyAlertRepoStub{deleted: false}
	svc := NewNearbyAlertService(repo)

	_, err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrNearbyAlertNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
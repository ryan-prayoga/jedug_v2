package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

const (
	NearbyAlertMinRadiusM          = 100
	NearbyAlertMaxRadiusM          = 5000
	NearbyAlertMaxSubscriptions    = 10
	nearbyAlertMaxLabelLength      = 80
)

var (
	ErrNearbyAlertLimitExceeded            = errors.New("nearby alert subscription limit exceeded")
	ErrNearbyAlertInvalidCoordinates       = errors.New("nearby alert coordinates are invalid")
	ErrNearbyAlertCoordinatePairRequired   = errors.New("latitude and longitude must be provided together")
	ErrNearbyAlertInvalidRadius            = errors.New("nearby alert radius is invalid")
	ErrNearbyAlertLabelTooLong             = errors.New("nearby alert label is too long")
	ErrNearbyAlertNotFound                 = errors.New("nearby alert subscription not found")
	ErrNearbyAlertPatchRequired            = errors.New("at least one nearby alert field must be provided")
)

type NearbyAlertService interface {
	List(ctx context.Context, followerID uuid.UUID) ([]*domain.NearbyAlertSubscription, error)
	Create(ctx context.Context, followerID uuid.UUID, input NearbyAlertCreateInput) (*domain.NearbyAlertSubscription, error)
	Update(ctx context.Context, followerID, subscriptionID uuid.UUID, patch domain.NearbyAlertSubscriptionPatch) (*domain.NearbyAlertSubscription, error)
	Delete(ctx context.Context, followerID, subscriptionID uuid.UUID) (bool, error)
}

type NearbyAlertCreateInput struct {
	Label     *string
	Latitude  float64
	Longitude float64
	RadiusM   int
	Enabled   *bool
}

type nearbyAlertService struct {
	repo repository.NearbyAlertRepository
}

func NewNearbyAlertService(repo repository.NearbyAlertRepository) NearbyAlertService {
	return &nearbyAlertService{repo: repo}
}

func (s *nearbyAlertService) List(ctx context.Context, followerID uuid.UUID) ([]*domain.NearbyAlertSubscription, error) {
	return s.repo.ListByFollowerID(ctx, followerID)
}

func (s *nearbyAlertService) Create(ctx context.Context, followerID uuid.UUID, input NearbyAlertCreateInput) (*domain.NearbyAlertSubscription, error) {
	if err := validateNearbyAlertCoordinates(input.Latitude, input.Longitude); err != nil {
		return nil, err
	}
	if err := validateNearbyAlertRadius(input.RadiusM); err != nil {
		return nil, err
	}
	label, err := normalizeNearbyAlertLabel(input.Label, false)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountByFollowerID(ctx, followerID)
	if err != nil {
		return nil, err
	}
	if count >= NearbyAlertMaxSubscriptions {
		return nil, ErrNearbyAlertLimitExceeded
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	return s.repo.Create(ctx, repository.NearbyAlertCreateInput{
		FollowerID: followerID,
		Label:      label,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		RadiusM:    input.RadiusM,
		Enabled:    enabled,
	})
}

func (s *nearbyAlertService) Update(ctx context.Context, followerID, subscriptionID uuid.UUID, patch domain.NearbyAlertSubscriptionPatch) (*domain.NearbyAlertSubscription, error) {
	if patch.IsEmpty() {
		return nil, ErrNearbyAlertPatchRequired
	}
	if (patch.Latitude == nil) != (patch.Longitude == nil) {
		return nil, ErrNearbyAlertCoordinatePairRequired
	}
	if patch.Latitude != nil && patch.Longitude != nil {
		if err := validateNearbyAlertCoordinates(*patch.Latitude, *patch.Longitude); err != nil {
			return nil, err
		}
	}
	if patch.RadiusM != nil {
		if err := validateNearbyAlertRadius(*patch.RadiusM); err != nil {
			return nil, err
		}
	}
	label, err := normalizeNearbyAlertLabel(patch.Label, true)
	if err != nil {
		return nil, err
	}
	patch.Label = label

	updated, err := s.repo.Update(ctx, followerID, subscriptionID, patch)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNearbyAlertNotFound
	}
	return updated, nil
}

func (s *nearbyAlertService) Delete(ctx context.Context, followerID, subscriptionID uuid.UUID) (bool, error) {
	deleted, err := s.repo.Delete(ctx, followerID, subscriptionID)
	if err != nil {
		return false, err
	}
	if !deleted {
		return false, ErrNearbyAlertNotFound
	}
	return true, nil
}

func validateNearbyAlertCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return ErrNearbyAlertInvalidCoordinates
	}
	return nil
}

func validateNearbyAlertRadius(radiusM int) error {
	if radiusM < NearbyAlertMinRadiusM || radiusM > NearbyAlertMaxRadiusM {
		return ErrNearbyAlertInvalidRadius
	}
	return nil
}

func normalizeNearbyAlertLabel(label *string, allowEmpty bool) (*string, error) {
	if label == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*label)
	if len(trimmed) > nearbyAlertMaxLabelLength {
		return nil, ErrNearbyAlertLabelTooLong
	}
	if trimmed == "" && !allowEmpty {
		return nil, nil
	}
	return &trimmed, nil
}
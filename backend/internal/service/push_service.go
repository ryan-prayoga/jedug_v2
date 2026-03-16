package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var ErrPushDisabled = errors.New("browser push notification is disabled")

type PushServiceConfig struct {
	Enabled        bool
	VAPIDPublicKey string
}

type PushSubscribeInput struct {
	FollowerID uuid.UUID
	Endpoint   string
	P256DH     string
	Auth       string
	UserAgent  *string
}

type PushUnsubscribeInput struct {
	FollowerID uuid.UUID
	Endpoint   string
}

type PushService interface {
	GetStatus(ctx context.Context, followerID uuid.UUID) (*domain.PushSubscriptionStatus, error)
	Subscribe(ctx context.Context, input PushSubscribeInput) (*domain.PushSubscriptionStatus, error)
	Unsubscribe(ctx context.Context, input PushUnsubscribeInput) (*domain.PushSubscriptionStatus, bool, error)
}

type pushService struct {
	repo repository.PushSubscriptionRepository
	cfg  PushServiceConfig
}

func NewPushService(repo repository.PushSubscriptionRepository, cfg PushServiceConfig) PushService {
	return &pushService{
		repo: repo,
		cfg: PushServiceConfig{
			Enabled:        cfg.Enabled,
			VAPIDPublicKey: strings.TrimSpace(cfg.VAPIDPublicKey),
		},
	}
}

func (s *pushService) GetStatus(ctx context.Context, followerID uuid.UUID) (*domain.PushSubscriptionStatus, error) {
	count, err := s.repo.CountActiveByFollowerID(ctx, followerID)
	if err != nil {
		return nil, err
	}

	return &domain.PushSubscriptionStatus{
		Enabled:           s.cfg.Enabled,
		Subscribed:        count > 0,
		SubscriptionCount: count,
		VAPIDPublicKey:    s.cfg.VAPIDPublicKey,
	}, nil
}

func (s *pushService) Subscribe(ctx context.Context, input PushSubscribeInput) (*domain.PushSubscriptionStatus, error) {
	if !s.cfg.Enabled {
		return nil, ErrPushDisabled
	}

	if strings.TrimSpace(input.Endpoint) == "" || strings.TrimSpace(input.P256DH) == "" || strings.TrimSpace(input.Auth) == "" {
		return nil, errors.New("push subscription endpoint and keys are required")
	}

	if _, err := s.repo.Upsert(ctx, repository.PushSubscriptionUpsertInput{
		FollowerID: input.FollowerID,
		Endpoint:   strings.TrimSpace(input.Endpoint),
		P256DH:     strings.TrimSpace(input.P256DH),
		Auth:       strings.TrimSpace(input.Auth),
		UserAgent:  trimOptionalString(input.UserAgent),
	}); err != nil {
		return nil, err
	}

	return s.GetStatus(ctx, input.FollowerID)
}

func (s *pushService) Unsubscribe(ctx context.Context, input PushUnsubscribeInput) (*domain.PushSubscriptionStatus, bool, error) {
	if strings.TrimSpace(input.Endpoint) == "" {
		return nil, false, errors.New("endpoint is required")
	}

	unsubscribed, err := s.repo.Disable(ctx, input.FollowerID, strings.TrimSpace(input.Endpoint))
	if err != nil {
		return nil, false, err
	}

	status, err := s.GetStatus(ctx, input.FollowerID)
	if err != nil {
		return nil, false, err
	}

	return status, unsubscribed, nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

package service

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var ErrPushDisabled = errors.New("browser push notification is disabled")
var ErrInvalidPushSubscription = errors.New("invalid push subscription")

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

	endpoint, err := validatePushEndpoint(strings.TrimSpace(input.Endpoint))
	if err != nil {
		return nil, err
	}
	if err := validatePushKey(strings.TrimSpace(input.P256DH), 48); err != nil {
		return nil, err
	}
	if err := validatePushKey(strings.TrimSpace(input.Auth), 16); err != nil {
		return nil, err
	}

	if _, err := s.repo.Upsert(ctx, repository.PushSubscriptionUpsertInput{
		FollowerID: input.FollowerID,
		Endpoint:   endpoint,
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

var allowedPushHosts = map[string][]string{
	"fcm.googleapis.com":                {"/fcm/send/"},
	"updates.push.services.mozilla.com": {"/wpush/"},
	"push.services.mozilla.com":         {"/wpush/"},
	"web.push.apple.com":                {"/"},
}

func validatePushEndpoint(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint is malformed"))
	}
	if parsed.Scheme != "https" {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint must use https"))
	}
	if parsed.User != nil {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint must not include credentials"))
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "" {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint host is required"))
	}
	if net.ParseIP(host) != nil {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint IP host is not allowed"))
	}

	allowedPrefixes, ok := allowedPushHosts[host]
	if !ok {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint host is not allowed"))
	}

	validPrefix := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(parsed.EscapedPath(), prefix) {
			validPrefix = true
			break
		}
	}
	if !validPrefix {
		return "", errors.Join(ErrInvalidPushSubscription, errors.New("push endpoint path is not recognized"))
	}

	return parsed.String(), nil
}

func validatePushKey(raw string, minBytes int) error {
	decoded, err := decodeBase64Loose(raw)
	if err != nil {
		return errors.Join(ErrInvalidPushSubscription, errors.New("push subscription keys are malformed"))
	}
	if len(decoded) < minBytes {
		return errors.Join(ErrInvalidPushSubscription, errors.New("push subscription keys are too short"))
	}
	return nil
}

func decodeBase64Loose(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("empty")
	}

	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.URLEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}
	return base64.RawURLEncoding.DecodeString(raw)
}

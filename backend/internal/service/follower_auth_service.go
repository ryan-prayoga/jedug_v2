package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var (
	ErrFollowerBindingNotFound = errors.New("follower binding not found")
	ErrFollowerBindingMismatch = errors.New("follower binding mismatch")
	ErrFollowerTokenRequired   = errors.New("follower token is required")
	ErrFollowerTokenInvalid    = errors.New("follower token is invalid")
	ErrFollowerTokenExpired    = errors.New("follower token is expired")
	ErrDeviceBootstrapRequired = errors.New("device bootstrap is required")
)

type FollowerAuthService interface {
	IssueForNotificationAccess(ctx context.Context, followerID uuid.UUID, deviceToken string) (*domain.FollowerAuthToken, error)
	IssueForFollowMutation(ctx context.Context, issueID, followerID uuid.UUID, deviceToken string) (*domain.FollowerAuthToken, error)
	AuthenticateNotificationAccess(ctx context.Context, rawToken, deviceToken string) (uuid.UUID, error)
	AuthenticateNotificationStream(ctx context.Context, rawToken string) (uuid.UUID, error)
}

type FollowerAuthServiceConfig struct {
	Secret    []byte
	TTL       time.Duration
	StreamTTL time.Duration
}

type followerAuthService struct {
	repo       repository.FollowerAuthRepository
	followRepo repository.IssueFollowRepository
	secret     []byte
	ttl        time.Duration
	streamTTL  time.Duration
	now        func() time.Time
}

type followerTokenClaims struct {
	Version         int    `json:"ver"`
	FollowerID      string `json:"follower_id"`
	DeviceTokenHash string `json:"device_token_hash"`
	Purpose         string `json:"purpose"`
	IssuedAt        int64  `json:"iat"`
	ExpiresAt       int64  `json:"exp"`
}

const (
	followerTokenPurposeAccess = "notification_access"
	followerTokenPurposeStream = "notification_stream"
)

func NewFollowerAuthService(
	repo repository.FollowerAuthRepository,
	followRepo repository.IssueFollowRepository,
	cfg FollowerAuthServiceConfig,
) FollowerAuthService {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}

	streamTTL := cfg.StreamTTL
	if streamTTL <= 0 {
		streamTTL = 10 * time.Minute
	}

	return &followerAuthService{
		repo:       repo,
		followRepo: followRepo,
		secret:     append([]byte(nil), cfg.Secret...),
		ttl:        ttl,
		streamTTL:  streamTTL,
		now:        time.Now,
	}
}

func (s *followerAuthService) IssueForNotificationAccess(ctx context.Context, followerID uuid.UUID, deviceToken string) (*domain.FollowerAuthToken, error) {
	deviceTokenHash, err := normalizeDeviceToken(deviceToken)
	if err != nil {
		return nil, err
	}

	if err := s.ensureBinding(ctx, followerID, deviceTokenHash, false, uuid.Nil); err != nil {
		return nil, err
	}

	return s.issueTokenPair(followerID, deviceTokenHash)
}

func (s *followerAuthService) IssueForFollowMutation(ctx context.Context, issueID, followerID uuid.UUID, deviceToken string) (*domain.FollowerAuthToken, error) {
	deviceTokenHash, err := normalizeDeviceToken(deviceToken)
	if err != nil {
		return nil, err
	}

	if err := s.ensureBinding(ctx, followerID, deviceTokenHash, true, issueID); err != nil {
		return nil, err
	}

	return s.issueTokenPair(followerID, deviceTokenHash)
}

func (s *followerAuthService) AuthenticateNotificationAccess(ctx context.Context, rawToken, deviceToken string) (uuid.UUID, error) {
	deviceTokenHash, err := normalizeDeviceToken(deviceToken)
	if err != nil {
		return uuid.Nil, err
	}

	return s.authenticate(ctx, rawToken, followerTokenPurposeAccess, &deviceTokenHash)
}

func (s *followerAuthService) AuthenticateNotificationStream(ctx context.Context, rawToken string) (uuid.UUID, error) {
	return s.authenticate(ctx, rawToken, followerTokenPurposeStream, nil)
}

func (s *followerAuthService) authenticate(
	ctx context.Context,
	rawToken string,
	expectedPurpose string,
	deviceTokenHash *string,
) (uuid.UUID, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return uuid.Nil, ErrFollowerTokenRequired
	}

	parts := strings.Split(rawToken, ".")
	if len(parts) != 2 {
		return uuid.Nil, ErrFollowerTokenInvalid
	}

	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(parts[0]))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, ErrFollowerTokenInvalid
	}
	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return uuid.Nil, ErrFollowerTokenInvalid
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uuid.Nil, ErrFollowerTokenInvalid
	}

	var claims followerTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil, ErrFollowerTokenInvalid
	}
	if claims.Version != 1 {
		return uuid.Nil, ErrFollowerTokenInvalid
	}
	if claims.Purpose != expectedPurpose {
		return uuid.Nil, ErrFollowerTokenInvalid
	}
	if claims.DeviceTokenHash == "" {
		return uuid.Nil, ErrFollowerTokenInvalid
	}

	followerID, err := uuid.Parse(claims.FollowerID)
	if err != nil || followerID == uuid.Nil {
		return uuid.Nil, ErrFollowerTokenInvalid
	}

	nowUnix := s.now().Unix()
	if claims.ExpiresAt <= nowUnix {
		return uuid.Nil, ErrFollowerTokenExpired
	}

	binding, err := s.repo.GetByFollowerID(ctx, followerID)
	if err != nil {
		return uuid.Nil, err
	}
	if binding == nil {
		return uuid.Nil, ErrFollowerBindingNotFound
	}
	if binding.DeviceTokenHash != claims.DeviceTokenHash {
		return uuid.Nil, ErrFollowerTokenInvalid
	}
	if deviceTokenHash != nil && *deviceTokenHash != claims.DeviceTokenHash {
		return uuid.Nil, ErrFollowerBindingMismatch
	}

	return followerID, nil
}

func (s *followerAuthService) ensureBinding(
	ctx context.Context,
	followerID uuid.UUID,
	deviceTokenHash string,
	allowIssueClaim bool,
	issueID uuid.UUID,
) error {
	binding, err := s.repo.GetByFollowerID(ctx, followerID)
	if err != nil {
		return err
	}

	if binding != nil {
		if binding.DeviceTokenHash != deviceTokenHash {
			return ErrFollowerBindingMismatch
		}
		return nil
	}

	hasFootprint, err := s.repo.HasFootprint(ctx, followerID)
	if err != nil {
		return err
	}

	if !hasFootprint {
		return s.repo.Upsert(ctx, followerID, deviceTokenHash)
	}

	if allowIssueClaim && issueID != uuid.Nil {
		following, err := s.followRepo.IsFollowing(ctx, issueID, followerID)
		if err != nil {
			return err
		}
		if following {
			return s.repo.Upsert(ctx, followerID, deviceTokenHash)
		}
	}

	return ErrFollowerBindingNotFound
}

func (s *followerAuthService) issueTokenPair(followerID uuid.UUID, deviceTokenHash string) (*domain.FollowerAuthToken, error) {
	accessToken, accessExpiresAt, err := s.signToken(followerID, deviceTokenHash, followerTokenPurposeAccess, s.ttl)
	if err != nil {
		return nil, err
	}

	streamToken, streamExpiresAt, err := s.signToken(followerID, deviceTokenHash, followerTokenPurposeStream, s.streamTTL)
	if err != nil {
		return nil, err
	}

	return &domain.FollowerAuthToken{
		FollowerID:      followerID.String(),
		Token:           accessToken,
		ExpiresAt:       accessExpiresAt,
		StreamToken:     streamToken,
		StreamExpiresAt: &streamExpiresAt,
	}, nil
}

func (s *followerAuthService) signToken(
	followerID uuid.UUID,
	deviceTokenHash string,
	purpose string,
	ttl time.Duration,
) (string, time.Time, error) {
	now := s.now().UTC()
	expiresAt := now.Add(ttl)

	payload, err := json.Marshal(followerTokenClaims{
		Version:         1,
		FollowerID:      followerID.String(),
		DeviceTokenHash: deviceTokenHash,
		Purpose:         purpose,
		IssuedAt:        now.Unix(),
		ExpiresAt:       expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payloadPart))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payloadPart + "." + signaturePart, expiresAt, nil
}

func normalizeDeviceToken(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrDeviceBootstrapRequired
	}

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}

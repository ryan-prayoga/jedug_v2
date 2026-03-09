package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

var ErrDeviceNotFound = errors.New("device not found")

type DeviceService interface {
	Bootstrap(ctx context.Context, req BootstrapRequest) (*BootstrapResult, error)
	RecordConsent(ctx context.Context, req ConsentRequest) error
}

type BootstrapRequest struct {
	AnonToken *string
	UserAgent *string
	IPAddress *string
}

type BootstrapResult struct {
	DeviceID  uuid.UUID `json:"device_id"`
	AnonToken string    `json:"anon_token"` // raw token, returned once to client
	IsNew     bool      `json:"is_new"`
}

type ConsentRequest struct {
	AnonToken      string
	TermsVersion   string
	PrivacyVersion string
	IPAddress      *string
	UserAgent      *string
}

type deviceService struct {
	repo repository.DeviceRepository
}

func NewDeviceService(repo repository.DeviceRepository) DeviceService {
	return &deviceService{repo: repo}
}

func (s *deviceService) Bootstrap(ctx context.Context, req BootstrapRequest) (*BootstrapResult, error) {
	// Try to find existing device by hashing the provided token
	if req.AnonToken != nil && *req.AnonToken != "" {
		tokenHash := hashToken(*req.AnonToken)
		device, err := s.repo.FindByTokenHash(ctx, tokenHash)
		if err != nil {
			return nil, err
		}
		if device != nil {
			if err := s.repo.UpdateLastSeen(ctx, device.ID); err != nil {
				return nil, err
			}
			return &BootstrapResult{
				DeviceID:  device.ID,
				AnonToken: *req.AnonToken, // echo the raw token back
				IsNew:     false,
			}, nil
		}
	}

	// No valid existing device — generate raw token, store its hash
	rawToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	device := &domain.Device{
		ID:            uuid.New(),
		AnonTokenHash: hashToken(rawToken),
		LastUserAgent: req.UserAgent,
		LastIP:        req.IPAddress,
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, err
	}

	return &BootstrapResult{
		DeviceID:  device.ID,
		AnonToken: rawToken, // only time the raw token is exposed
		IsNew:     true,
	}, nil
}

func (s *deviceService) RecordConsent(ctx context.Context, req ConsentRequest) error {
	tokenHash := hashToken(req.AnonToken)
	device, err := s.repo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	pv := req.PrivacyVersion
	consent := &domain.DeviceConsent{
		DeviceID:       device.ID,
		TermsVersion:   req.TermsVersion,
		PrivacyVersion: &pv,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	}

	return s.repo.CreateConsent(ctx, consent)
}

// hashToken returns the SHA-256 hex digest of a raw token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

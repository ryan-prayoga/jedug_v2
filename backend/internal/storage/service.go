package storage

import (
	"context"
	"strings"
)

type Service struct {
	active      Driver
	legacyLocal *LocalDriver
}

func NewService(active Driver, legacyLocal *LocalDriver) *Service {
	if legacyLocal == nil {
		if local, ok := active.(*LocalDriver); ok {
			legacyLocal = local
		}
	}
	return &Service{
		active:      active,
		legacyLocal: legacyLocal,
	}
}

func (s *Service) CreatePresign(ctx context.Context, req UploadRequest) (*PresignResult, error) {
	if err := ValidateUploadRequest(req); err != nil {
		return nil, err
	}

	objectKey, err := s.active.GenerateObjectKey(req)
	if err != nil {
		return nil, err
	}
	if err := ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	return s.active.CreatePresign(ctx, req, objectKey)
}

func (s *Service) ResolvePublicURL(objectKey string) string {
	raw := strings.TrimSpace(objectKey)
	if raw == "" {
		return ""
	}
	if isAbsoluteURL(raw) {
		return raw
	}

	key := NormalizeObjectKey(raw)
	if s.legacyLocal != nil && s.legacyLocal.Exists(key) {
		return s.legacyLocal.BuildPublicURL(key)
	}
	return s.active.BuildPublicURL(key)
}

func (s *Service) UploadMode() string {
	return s.active.Name()
}

func (s *Service) LegacyLocal() *LocalDriver {
	return s.legacyLocal
}

func (s *Service) Upload(ctx context.Context, objectKey, contentType string, body []byte) error {
	if err := ValidateSubmittedMedia(objectKey, contentType, len(body)); err != nil {
		return err
	}
	return s.active.Upload(ctx, objectKey, contentType, body)
}

func (s *Service) Stat(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	if err := ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}
	return s.active.Stat(ctx, objectKey)
}

func (s *Service) Delete(ctx context.Context, uploadMode, objectKey string) error {
	if err := ValidateObjectKey(objectKey); err != nil {
		return err
	}

	key := NormalizeObjectKey(objectKey)
	if uploadMode == "local" && s.legacyLocal != nil {
		return s.legacyLocal.Delete(ctx, key)
	}
	return s.active.Delete(ctx, key)
}

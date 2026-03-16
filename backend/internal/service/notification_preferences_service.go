package service

import (
	"context"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type NotificationPreferencesService interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.NotificationPreferences, error)
	Update(ctx context.Context, followerID uuid.UUID, patch domain.NotificationPreferencesPatch) (*domain.NotificationPreferences, error)
}

type notificationPreferencesService struct {
	repo repository.NotificationPreferencesRepository
}

func NewNotificationPreferencesService(repo repository.NotificationPreferencesRepository) NotificationPreferencesService {
	return &notificationPreferencesService{repo: repo}
}

func (s *notificationPreferencesService) GetByFollowerID(ctx context.Context, followerID uuid.UUID) (*domain.NotificationPreferences, error) {
	return s.repo.GetByFollowerID(ctx, followerID)
}

func (s *notificationPreferencesService) Update(ctx context.Context, followerID uuid.UUID, patch domain.NotificationPreferencesPatch) (*domain.NotificationPreferences, error) {
	return s.repo.Update(ctx, followerID, patch)
}

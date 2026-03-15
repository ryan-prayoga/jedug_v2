package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type NotificationService interface {
	GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) (*time.Time, bool, error)
	Delete(ctx context.Context, notificationID, followerID uuid.UUID) (bool, error)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) GetByFollowerID(ctx context.Context, followerID uuid.UUID, limit int) ([]*domain.Notification, error) {
	return s.repo.GetByFollowerID(ctx, followerID, limit)
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID, followerID uuid.UUID) (*time.Time, bool, error) {
	return s.repo.MarkAsRead(ctx, notificationID, followerID)
}

func (s *notificationService) Delete(ctx context.Context, notificationID, followerID uuid.UUID) (bool, error) {
	return s.repo.Delete(ctx, notificationID, followerID)
}

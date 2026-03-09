package service

import (
	"context"

	"github.com/google/uuid"
	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

type IssueService interface {
	List(ctx context.Context, limit, offset int, status *string, bbox *repository.BBoxFilter) ([]*domain.Issue, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	GetByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.IssueDetail, error)
}

type issueService struct {
	repo repository.IssueRepository
}

func NewIssueService(repo repository.IssueRepository) IssueService {
	return &issueService{repo: repo}
}

func (s *issueService) List(ctx context.Context, limit, offset int, status *string, bbox *repository.BBoxFilter) ([]*domain.Issue, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset, status, bbox)
}

func (s *issueService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *issueService) GetByIDWithDetail(ctx context.Context, id uuid.UUID) (*domain.IssueDetail, error) {
	return s.repo.FindByIDWithDetail(ctx, id)
}


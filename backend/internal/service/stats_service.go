package service

import (
	"context"
	"sync"
	"time"

	"jedug_backend/internal/domain"
	"jedug_backend/internal/repository"
)

const (
	defaultStatsCacheTTL          = 45 * time.Second
	defaultStatsRegionLeaderLimit = 10
)

type StatsService interface {
	GetPublicStats(ctx context.Context) (*domain.PublicStats, error)
}

type statsService struct {
	repo        repository.StatsRepository
	cacheTTL    time.Duration
	regionLimit int
	now         func() time.Time

	mu             sync.RWMutex
	cached         *domain.PublicStats
	cacheExpiresAt time.Time
}

func NewStatsService(repo repository.StatsRepository) StatsService {
	return newStatsService(repo, defaultStatsCacheTTL, defaultStatsRegionLeaderLimit, time.Now)
}

func newStatsService(
	repo repository.StatsRepository,
	cacheTTL time.Duration,
	regionLimit int,
	nowFn func() time.Time,
) *statsService {
	if cacheTTL <= 0 {
		cacheTTL = defaultStatsCacheTTL
	}
	if regionLimit <= 0 {
		regionLimit = defaultStatsRegionLeaderLimit
	}
	if nowFn == nil {
		nowFn = time.Now
	}

	return &statsService{
		repo:        repo,
		cacheTTL:    cacheTTL,
		regionLimit: regionLimit,
		now:         nowFn,
	}
}

func (s *statsService) GetPublicStats(ctx context.Context) (*domain.PublicStats, error) {
	now := s.now()

	s.mu.RLock()
	cached := s.cached
	cacheExpiresAt := s.cacheExpiresAt
	s.mu.RUnlock()

	if cached != nil && now.Before(cacheExpiresAt) {
		return cached, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now = s.now()
	if s.cached != nil && now.Before(s.cacheExpiresAt) {
		return s.cached, nil
	}

	stats, err := s.repo.GetPublicStats(ctx, s.regionLimit)
	if err != nil {
		// Fallback to stale cache when DB is briefly unavailable.
		if s.cached != nil {
			return s.cached, nil
		}
		return nil, err
	}

	stats.GeneratedAt = now.UTC()
	s.cached = stats
	s.cacheExpiresAt = now.Add(s.cacheTTL)

	return stats, nil
}

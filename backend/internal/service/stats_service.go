package service

import (
	"context"
	"fmt"
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
	GetPublicStats(ctx context.Context, query domain.PublicStatsQuery) (*domain.PublicStats, error)
}

type statsService struct {
	repo        repository.StatsRepository
	cacheTTL    time.Duration
	regionLimit int
	now         func() time.Time

	mu    sync.RWMutex
	cache map[string]cachedPublicStats
}

type cachedPublicStats struct {
	stats     *domain.PublicStats
	expiresAt time.Time
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
		cache:       make(map[string]cachedPublicStats),
	}
}

func (s *statsService) GetPublicStats(ctx context.Context, query domain.PublicStatsQuery) (*domain.PublicStats, error) {
	now := s.now()
	cacheKey := buildStatsCacheKey(query)

	s.mu.RLock()
	cached, ok := s.cache[cacheKey]
	s.mu.RUnlock()

	if ok && cached.stats != nil && now.Before(cached.expiresAt) {
		return cached.stats, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now = s.now()
	cached, ok = s.cache[cacheKey]
	if ok && cached.stats != nil && now.Before(cached.expiresAt) {
		return cached.stats, nil
	}

	stats, err := s.repo.GetPublicStats(ctx, query, s.regionLimit)
	if err != nil {
		// Fallback to stale cache when DB is briefly unavailable.
		if ok && cached.stats != nil {
			return cached.stats, nil
		}
		return nil, err
	}

	stats.GeneratedAt = now.UTC()
	s.cache[cacheKey] = cachedPublicStats{
		stats:     stats,
		expiresAt: now.Add(s.cacheTTL),
	}

	return stats, nil
}

func buildStatsCacheKey(query domain.PublicStatsQuery) string {
	provinceID := int64(0)
	regencyID := int64(0)
	if query.ProvinceID != nil {
		provinceID = *query.ProvinceID
	}
	if query.RegencyID != nil {
		regencyID = *query.RegencyID
	}

	return fmt.Sprintf("province:%d|regency:%d", provinceID, regencyID)
}

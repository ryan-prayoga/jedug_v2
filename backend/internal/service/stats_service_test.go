package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"jedug_backend/internal/domain"
)

type statsRepoMock struct {
	calls      int
	result     *domain.PublicStats
	err        error
	regionArgs []int
}

func (m *statsRepoMock) GetPublicStats(_ context.Context, regionLimit int) (*domain.PublicStats, error) {
	m.calls++
	m.regionArgs = append(m.regionArgs, regionLimit)
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestStatsServiceReturnsCachedDataWithinTTL(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)

	repo := &statsRepoMock{
		result: &domain.PublicStats{
			Global: domain.PublicGlobalStats{TotalIssues: 42},
		},
	}

	svc := newStatsService(repo, 60*time.Second, 8, func() time.Time { return now })

	first, err := svc.GetPublicStats(context.Background())
	if err != nil {
		t.Fatalf("first fetch error: %v", err)
	}

	second, err := svc.GetPublicStats(context.Background())
	if err != nil {
		t.Fatalf("second fetch error: %v", err)
	}

	if repo.calls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.calls)
	}
	if first != second {
		t.Fatalf("expected cached pointer to be reused")
	}
	if first.GeneratedAt.IsZero() {
		t.Fatalf("expected generated_at to be set")
	}
}

func TestStatsServiceFallsBackToStaleCacheOnError(t *testing.T) {
	current := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return current }

	repo := &statsRepoMock{
		result: &domain.PublicStats{
			Global: domain.PublicGlobalStats{TotalIssues: 10},
		},
	}

	svc := newStatsService(repo, 30*time.Second, 10, nowFn)

	cached, err := svc.GetPublicStats(context.Background())
	if err != nil {
		t.Fatalf("initial fetch error: %v", err)
	}

	current = current.Add(31 * time.Second)
	repo.err = errors.New("db down")

	got, err := svc.GetPublicStats(context.Background())
	if err != nil {
		t.Fatalf("expected stale cache fallback, got error: %v", err)
	}
	if got != cached {
		t.Fatalf("expected stale cached stats pointer")
	}
}

func TestStatsServiceReturnsErrorWhenNoCacheAndRepoFails(t *testing.T) {
	repo := &statsRepoMock{
		err: errors.New("query failed"),
	}

	svc := newStatsService(repo, 45*time.Second, 10, time.Now)

	_, err := svc.GetPublicStats(context.Background())
	if err == nil {
		t.Fatalf("expected error when repository fails without cache")
	}
}

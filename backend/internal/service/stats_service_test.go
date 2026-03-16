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
	queries    []domain.PublicStatsQuery

	optionCalls   int
	optionsResult *domain.PublicRegionOptions
	optionsErr    error
}

func (m *statsRepoMock) GetPublicStats(_ context.Context, query domain.PublicStatsQuery, regionLimit int) (*domain.PublicStats, error) {
	m.calls++
	m.regionArgs = append(m.regionArgs, regionLimit)
	m.queries = append(m.queries, query)
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *statsRepoMock) GetPublicRegionOptions(_ context.Context) (*domain.PublicRegionOptions, error) {
	m.optionCalls++
	if m.optionsErr != nil {
		return nil, m.optionsErr
	}
	return m.optionsResult, nil
}

func TestStatsServiceReturnsCachedDataWithinTTL(t *testing.T) {
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)

	repo := &statsRepoMock{
		result: &domain.PublicStats{
			Global: domain.PublicGlobalStats{TotalIssues: 42},
		},
	}

	svc := newStatsService(repo, 60*time.Second, 8, func() time.Time { return now })

	first, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{})
	if err != nil {
		t.Fatalf("first fetch error: %v", err)
	}

	second, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{})
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

	cached, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{})
	if err != nil {
		t.Fatalf("initial fetch error: %v", err)
	}

	current = current.Add(31 * time.Second)
	repo.err = errors.New("db down")

	got, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{})
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

	_, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{})
	if err == nil {
		t.Fatalf("expected error when repository fails without cache")
	}
}

func TestStatsServiceCachesPerScope(t *testing.T) {
	provinceA := int64(11)
	provinceB := int64(12)
	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)

	repo := &statsRepoMock{
		result: &domain.PublicStats{
			Global: domain.PublicGlobalStats{TotalIssues: 7},
		},
	}

	svc := newStatsService(repo, 60*time.Second, 8, func() time.Time { return now })

	if _, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{ProvinceID: &provinceA}); err != nil {
		t.Fatalf("first scoped fetch error: %v", err)
	}
	if _, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{ProvinceID: &provinceA}); err != nil {
		t.Fatalf("second scoped fetch error: %v", err)
	}
	if _, err := svc.GetPublicStats(context.Background(), domain.PublicStatsQuery{ProvinceID: &provinceB}); err != nil {
		t.Fatalf("third scoped fetch error: %v", err)
	}

	if repo.calls != 2 {
		t.Fatalf("expected repository to be called twice for two scopes, got %d", repo.calls)
	}
}

func TestStatsServiceCachesRegionOptionsWithinTTL(t *testing.T) {
	now := time.Date(2026, 3, 16, 9, 0, 0, 0, time.UTC)

	repo := &statsRepoMock{
		optionsResult: &domain.PublicRegionOptions{
			Provinces: []*domain.PublicProvinceOption{
				{ID: 31, Name: "DKI Jakarta"},
			},
		},
	}

	svc := newStatsService(repo, 60*time.Second, 8, func() time.Time { return now })

	first, err := svc.GetPublicRegionOptions(context.Background())
	if err != nil {
		t.Fatalf("first region options fetch error: %v", err)
	}
	second, err := svc.GetPublicRegionOptions(context.Background())
	if err != nil {
		t.Fatalf("second region options fetch error: %v", err)
	}

	if repo.optionCalls != 1 {
		t.Fatalf("expected region options repository to be called once, got %d", repo.optionCalls)
	}
	if first != second {
		t.Fatalf("expected cached region options pointer to be reused")
	}
}

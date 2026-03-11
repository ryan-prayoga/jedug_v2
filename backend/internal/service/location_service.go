package service

import (
	"context"
	"strings"

	"jedug_backend/internal/repository"
)

type LocationService interface {
	ResolveLabel(ctx context.Context, longitude, latitude float64) (*LocationLabelResult, error)
}

type LocationLabelResult struct {
	Label           *string `json:"label"`
	RegionID        *int64  `json:"region_id"`
	RegionName      *string `json:"region_name"`
	RegionLevel     *string `json:"region_level"`
	ParentName      *string `json:"parent_name"`
	GrandparentName *string `json:"grandparent_name"`
	Source          string  `json:"source"`
}

type locationService struct {
	repo repository.LocationRepository
}

func NewLocationService(repo repository.LocationRepository) LocationService {
	return &locationService{repo: repo}
}

func (s *locationService) ResolveLabel(ctx context.Context, longitude, latitude float64) (*LocationLabelResult, error) {
	region, err := s.repo.ResolveLabelByPoint(ctx, longitude, latitude)
	if err != nil {
		return nil, err
	}
	if region == nil {
		return &LocationLabelResult{
			Label:           nil,
			RegionID:        nil,
			RegionName:      nil,
			RegionLevel:     nil,
			ParentName:      nil,
			GrandparentName: nil,
			Source:          "internal_regions",
		}, nil
	}

	label := buildHumanLabel(region.RegionName, region.ParentName, region.GrandparentName)

	regionID := region.RegionID
	regionName := region.RegionName
	regionLevel := region.RegionLevel

	return &LocationLabelResult{
		Label:           label,
		RegionID:        &regionID,
		RegionName:      &regionName,
		RegionLevel:     &regionLevel,
		ParentName:      region.ParentName,
		GrandparentName: region.GrandparentName,
		Source:          "internal_regions",
	}, nil
}

func buildHumanLabel(primary string, parent, grandparent *string) *string {
	parts := make([]string, 0, 3)
	seen := make(map[string]struct{}, 3)

	pushPart := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		parts = append(parts, trimmed)
	}

	pushPart(primary)
	if parent != nil {
		pushPart(*parent)
	}
	if grandparent != nil {
		pushPart(*grandparent)
	}

	if len(parts) == 0 {
		return nil
	}

	label := strings.Join(parts, ", ")
	return &label
}

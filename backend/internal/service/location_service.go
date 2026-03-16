package service

import (
	"context"
	"log"
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
	DistrictName    *string `json:"district_name"`
	RegencyName     *string `json:"regency_name"`
	ProvinceName    *string `json:"province_name"`
	Source          string  `json:"source"`
}

type locationService struct {
	repo            repository.LocationRepository
	reverseGeocoder ReverseGeocoder
}

func NewLocationService(repo repository.LocationRepository, reverseGeocoder ReverseGeocoder) LocationService {
	return &locationService{
		repo:            repo,
		reverseGeocoder: reverseGeocoder,
	}
}

func (s *locationService) ResolveLabel(ctx context.Context, longitude, latitude float64) (*LocationLabelResult, error) {
	region, err := s.repo.ResolveLabelByPoint(ctx, longitude, latitude)
	if err != nil {
		log.Printf("[LOCATION_LABEL] resolve_internal_error lat=%.6f lon=%.6f err=%v", latitude, longitude, err)
		return nil, err
	}
	if region != nil {
		label := buildHumanLabel(region.RegionName, region.ParentName, region.GrandparentName)

		regionID := region.RegionID
		regionName := region.RegionName
		regionLevel := region.RegionLevel

		out := &LocationLabelResult{
			Label:           label,
			RegionID:        &regionID,
			RegionName:      &regionName,
			RegionLevel:     &regionLevel,
			ParentName:      region.ParentName,
			GrandparentName: region.GrandparentName,
			DistrictName:    region.DistrictName,
			RegencyName:     region.RegencyName,
			ProvinceName:    region.ProvinceName,
			Source:          "internal_regions",
		}
		return out, nil
	}

	if s.reverseGeocoder != nil {
		reverse, reverseErr := s.reverseGeocoder.ReverseLookup(ctx, longitude, latitude)
		if reverseErr != nil {
			log.Printf("[LOCATION_LABEL] resolve_reverse_error lat=%.6f lon=%.6f err=%v", latitude, longitude, reverseErr)
		} else if reverse != nil {
			label := buildHumanLabelFromPtrs(reverse.RoadName, reverse.RegionName, reverse.CityName)
			regionLevel := "fallback_reverse_geocode"
			out := &LocationLabelResult{
				Label:           label,
				RegionID:        nil,
				RegionName:      reverse.RegionName,
				RegionLevel:     &regionLevel,
				ParentName:      reverse.CityName,
				GrandparentName: nil,
				DistrictName:    reverse.RegionName,
				RegencyName:     reverse.CityName,
				ProvinceName:    nil,
				Source:          "reverse_geocode",
			}
			return out, nil
		}
	}

	out := &LocationLabelResult{
		Label:           nil,
		RegionID:        nil,
		RegionName:      nil,
		RegionLevel:     nil,
		ParentName:      nil,
		GrandparentName: nil,
		DistrictName:    nil,
		RegencyName:     nil,
		ProvinceName:    nil,
		Source:          "unresolved",
	}
	return out, nil
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

func buildHumanLabelFromPtrs(primary, parent, grandparent *string) *string {
	primaryValue := ""
	if primary != nil {
		primaryValue = *primary
	}
	return buildHumanLabel(primaryValue, parent, grandparent)
}

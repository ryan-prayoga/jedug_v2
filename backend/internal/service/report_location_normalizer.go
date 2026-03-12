package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"jedug_backend/internal/repository"
)

type ReportLocationNormalization struct {
	RegionID   *int64
	RoadName   *string
	RegionName *string
	CityName   *string
}

type ReportLocationNormalizer interface {
	NormalizeForReport(ctx context.Context, longitude, latitude float64) ReportLocationNormalization
}

type reportLocationNormalizer struct {
	locationRepo    repository.LocationRepository
	reverseGeocoder ReverseGeocoder
}

func NewReportLocationNormalizer(
	locationRepo repository.LocationRepository,
	reverseGeocoder ReverseGeocoder,
) ReportLocationNormalizer {
	return &reportLocationNormalizer{
		locationRepo:    locationRepo,
		reverseGeocoder: reverseGeocoder,
	}
}

func (n *reportLocationNormalizer) NormalizeForReport(
	ctx context.Context,
	longitude, latitude float64,
) ReportLocationNormalization {
	out := ReportLocationNormalization{}

	if n.locationRepo != nil {
		region, err := n.locationRepo.ResolveLabelByPoint(ctx, longitude, latitude)
		if err != nil {
			log.Printf("[LOCATION] internal_lookup_failed lat=%.6f lon=%.6f err=%v", latitude, longitude, err)
		} else if region != nil {
			regionID := region.RegionID
			out.RegionID = &regionID
			out.RegionName = nonEmptyPtr(region.RegionName)
			out.CityName = pickCityName(region.ParentName, region.GrandparentName)
		}
	}

	if out.RoadName == nil && n.reverseGeocoder != nil {
		reverse, err := n.reverseGeocoder.ReverseLookup(ctx, longitude, latitude)
		if err != nil {
			log.Printf("[LOCATION] reverse_geocode_failed lat=%.6f lon=%.6f err=%v", latitude, longitude, err)
		} else if reverse != nil {
			out.RoadName = nonEmptyPtrFromPtr(reverse.RoadName)
			if out.RegionName == nil {
				out.RegionName = nonEmptyPtrFromPtr(reverse.RegionName)
			}
			if out.CityName == nil {
				out.CityName = nonEmptyPtrFromPtr(reverse.CityName)
			}
		}
	}

	if out.RegionName == nil {
		fallback := fallbackCoordinateLabel(latitude, longitude)
		out.RegionName = &fallback
	}

	if out.RoadName == nil {
		fallback := fallbackCoordinateLabel(latitude, longitude)
		out.RoadName = &fallback
	}

	return out
}

func pickCityName(parentName, grandparentName *string) *string {
	if parentName != nil {
		if v := nonEmptyPtrFromPtr(parentName); v != nil {
			return v
		}
	}
	return nonEmptyPtrFromPtr(grandparentName)
}

func nonEmptyPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nonEmptyPtrFromPtr(value *string) *string {
	if value == nil {
		return nil
	}
	return nonEmptyPtr(*value)
}

func fallbackCoordinateLabel(latitude, longitude float64) string {
	return fmt.Sprintf("Kawasan sekitar %.5f,%.5f", latitude, longitude)
}

package service

import (
	"context"
	"errors"
	"testing"

	"jedug_backend/internal/repository"
)

type locationRepoStub struct {
	label *repository.LocationLabel
	err   error
}

func (s *locationRepoStub) ResolveLabelByPoint(_ context.Context, _ float64, _ float64) (*repository.LocationLabel, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.label, nil
}

type reverseGeocoderStub struct {
	result *ReverseGeocodeResult
	err    error
}

func (s reverseGeocoderStub) ReverseLookup(_ context.Context, _ float64, _ float64) (*ReverseGeocodeResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestReportLocationNormalizerInternalRegionPriority(t *testing.T) {
	normalizer := NewReportLocationNormalizer(
		&locationRepoStub{
			label: &repository.LocationLabel{
				RegionID:    11,
				RegionName:  "Kebon Melati",
				RegionLevel: "district",
				ParentName:  strPtr("Jakarta Pusat"),
			},
		},
		reverseGeocoderStub{
			result: &ReverseGeocodeResult{
				RoadName:   strPtr("Jl. KH. Mas Mansyur"),
				RegionName: strPtr("Tanah Abang"),
				CityName:   strPtr("Jakarta"),
			},
		},
	)

	got := normalizer.NormalizeForReport(context.Background(), 106.81, -6.20)

	if got.RegionID == nil || *got.RegionID != 11 {
		t.Fatalf("expected region id 11, got %+v", got.RegionID)
	}
	if got.RegionName == nil || *got.RegionName != "Kebon Melati" {
		t.Fatalf("expected internal region name, got %+v", got.RegionName)
	}
	if got.RoadName == nil || *got.RoadName != "Jl. KH. Mas Mansyur" {
		t.Fatalf("expected reverse geocoded road name, got %+v", got.RoadName)
	}
	if got.CityName == nil || *got.CityName != "Jakarta Pusat" {
		t.Fatalf("expected city from internal parent, got %+v", got.CityName)
	}
}

func TestReportLocationNormalizerFallbackLabel(t *testing.T) {
	normalizer := NewReportLocationNormalizer(
		&locationRepoStub{err: errors.New("lookup failed")},
		reverseGeocoderStub{err: errors.New("reverse failed")},
	)

	got := normalizer.NormalizeForReport(context.Background(), 106.816666, -6.200000)

	if got.RoadName == nil || got.RegionName == nil {
		t.Fatalf("expected fallback labels, got %#v", got)
	}

	want := "Kawasan sekitar -6.20000,106.81667"
	if *got.RoadName != want {
		t.Fatalf("unexpected road fallback: got %q want %q", *got.RoadName, want)
	}
	if *got.RegionName != want {
		t.Fatalf("unexpected region fallback: got %q want %q", *got.RegionName, want)
	}
}

func TestReportLocationNormalizerUsesHumanAreaWhenRoadMissing(t *testing.T) {
	normalizer := NewReportLocationNormalizer(
		&locationRepoStub{
			label: &repository.LocationLabel{
				RegionID:    77,
				RegionName:  "Blimbing Gede",
				RegionLevel: "village",
				ParentName:  strPtr("Bojonegoro"),
			},
		},
		reverseGeocoderStub{
			result: &ReverseGeocodeResult{
				RoadName:   nil,
				RegionName: strPtr("Blimbing Gede"),
				CityName:   strPtr("Bojonegoro"),
			},
		},
	)

	got := normalizer.NormalizeForReport(context.Background(), 111.55412, -7.25378)

	if got.RoadName == nil || *got.RoadName != "Blimbing Gede" {
		t.Fatalf("expected human area fallback for road name, got %+v", got.RoadName)
	}
	if got.RegionName == nil || *got.RegionName != "Blimbing Gede" {
		t.Fatalf("expected internal region name to stay available, got %+v", got.RegionName)
	}
}

package service

import (
	"context"
	"errors"
	"testing"

	"jedug_backend/internal/repository"
)

func TestBuildHumanLabel(t *testing.T) {
	tests := []struct {
		name        string
		primary     string
		parent      *string
		grandparent *string
		want        *string
	}{
		{
			name:        "joins primary and ancestors",
			primary:     "Kecamatan Tebet",
			parent:      strPtr("Jakarta Selatan"),
			grandparent: strPtr("DKI Jakarta"),
			want:        strPtr("Kecamatan Tebet, Jakarta Selatan, DKI Jakarta"),
		},
		{
			name:        "deduplicates repeated names",
			primary:     "Kecamatan Tebet",
			parent:      strPtr("Kecamatan Tebet"),
			grandparent: strPtr("DKI Jakarta"),
			want:        strPtr("Kecamatan Tebet, DKI Jakarta"),
		},
		{
			name:        "returns nil when all parts empty",
			primary:     "   ",
			parent:      strPtr(""),
			grandparent: nil,
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHumanLabel(tt.primary, tt.parent, tt.grandparent)
			if !equalStringPtr(got, tt.want) {
				t.Fatalf("buildHumanLabel() = %v, want %v", valueOf(got), valueOf(tt.want))
			}
		})
	}
}

func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func valueOf(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func strPtr(v string) *string {
	return &v
}

type locationRepoFake struct {
	result *repository.LocationLabel
	err    error
}

func (f locationRepoFake) ResolveLabelByPoint(_ context.Context, _ float64, _ float64) (*repository.LocationLabel, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

type reverseGeocoderFake struct {
	result *ReverseGeocodeResult
	err    error
}

func (f reverseGeocoderFake) ReverseLookup(_ context.Context, _ float64, _ float64) (*ReverseGeocodeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestLocationServiceResolveLabelFallsBackToReverseGeocode(t *testing.T) {
	svc := NewLocationService(
		locationRepoFake{result: nil},
		reverseGeocoderFake{
			result: &ReverseGeocodeResult{
				RoadName:   strPtr("Jl. KH. Mas Mansyur"),
				RegionName: strPtr("Kebon Melati"),
				CityName:   strPtr("Jakarta Pusat"),
			},
		},
	)

	got, err := svc.ResolveLabel(context.Background(), 106.816666, -6.200000)
	if err != nil {
		t.Fatalf("ResolveLabel error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil result")
	}
	if got.Source != "reverse_geocode" {
		t.Fatalf("unexpected source: %s", got.Source)
	}
	if got.Label == nil || *got.Label != "Jl. KH. Mas Mansyur, Kebon Melati, Jakarta Pusat" {
		t.Fatalf("unexpected label: %v", valueOf(got.Label))
	}
	if got.RegionName == nil || *got.RegionName != "Kebon Melati" {
		t.Fatalf("unexpected region_name: %v", valueOf(got.RegionName))
	}
}

func TestLocationServiceResolveLabelUnresolvedWhenAllLookupsFail(t *testing.T) {
	svc := NewLocationService(
		locationRepoFake{err: errors.New("db error")},
		reverseGeocoderFake{result: nil},
	)

	_, err := svc.ResolveLabel(context.Background(), 106.816666, -6.200000)
	if err == nil {
		t.Fatalf("expected error when internal lookup returns error")
	}
}

func TestLocationServiceResolveLabelUnresolvedWhenNoData(t *testing.T) {
	svc := NewLocationService(
		locationRepoFake{result: nil},
		reverseGeocoderFake{result: nil},
	)

	got, err := svc.ResolveLabel(context.Background(), 106.816666, -6.200000)
	if err != nil {
		t.Fatalf("ResolveLabel error: %v", err)
	}
	if got.Source != "unresolved" {
		t.Fatalf("unexpected source: %s", got.Source)
	}
	if got.Label != nil {
		t.Fatalf("expected nil label, got: %v", valueOf(got.Label))
	}
}

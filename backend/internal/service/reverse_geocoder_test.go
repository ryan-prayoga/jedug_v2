package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPReverseGeocoderMappingAndCache(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.URL.Query().Get("accept-language"); got != "id" {
			t.Fatalf("expected accept-language=id query param, got %q", got)
		}
		if got := r.Header.Get("Accept-Language"); got != "id" {
			t.Fatalf("expected Accept-Language=id header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"display_name": "Jl. MH Thamrin, Menteng, Jakarta Pusat, DKI Jakarta, Indonesia",
			"category": "highway",
			"type": "primary",
			"addresstype": "road",
			"place_rank": 26,
			"address": {
				"road": "Jl. MH Thamrin",
				"neighbourhood": "Menteng",
				"city": "Jakarta Pusat",
				"state": "Special Capital Region of Jakarta",
				"postcode": "10110",
				"country": "Indonesia",
				"country_code": "id"
			}
		}`))
	}))
	defer server.Close()

	geocoder := NewHTTPReverseGeocoder(true, server.URL, "jedug-test", time.Second, time.Minute)

	first, err := geocoder.ReverseLookup(context.Background(), 106.81666, -6.20000)
	if err != nil {
		t.Fatalf("first lookup err: %v", err)
	}
	if first == nil || first.RoadName == nil || *first.RoadName != "Jl. MH Thamrin" {
		t.Fatalf("unexpected first road name: %#v", first)
	}
	if first.RegionName == nil || *first.RegionName != "Menteng" {
		t.Fatalf("unexpected first region name: %#v", first)
	}
	if first.CityName == nil || *first.CityName != "Jakarta Pusat" {
		t.Fatalf("unexpected first city name: %#v", first)
	}
	if first.DistrictName == nil || *first.DistrictName != "Menteng" {
		t.Fatalf("unexpected first district name: %#v", first)
	}
	if first.RegencyName == nil || *first.RegencyName != "Jakarta Pusat" {
		t.Fatalf("unexpected first regency name: %#v", first)
	}
	if first.ProvinceName == nil || *first.ProvinceName != "Daerah Khusus Ibukota Jakarta" {
		t.Fatalf("unexpected first province name: %#v", first)
	}
	if first.DisplayName == nil || *first.DisplayName == "" {
		t.Fatalf("unexpected display name: %#v", first)
	}
	if first.Postcode == nil || *first.Postcode != "10110" {
		t.Fatalf("unexpected postcode: %#v", first)
	}
	if first.CountryName == nil || *first.CountryName != "Indonesia" {
		t.Fatalf("unexpected country name: %#v", first)
	}
	if first.CountryCode == nil || *first.CountryCode != "ID" {
		t.Fatalf("unexpected country code: %#v", first)
	}
	if first.Category == nil || *first.Category != "highway" {
		t.Fatalf("unexpected category: %#v", first)
	}
	if first.Type == nil || *first.Type != "primary" {
		t.Fatalf("unexpected type: %#v", first)
	}
	if first.AddressType == nil || *first.AddressType != "road" {
		t.Fatalf("unexpected addresstype: %#v", first)
	}
	if first.PlaceRank == nil || *first.PlaceRank != 26 {
		t.Fatalf("unexpected place rank: %#v", first)
	}

	second, err := geocoder.ReverseLookup(context.Background(), 106.81666, -6.20000)
	if err != nil {
		t.Fatalf("second lookup err: %v", err)
	}
	if second == nil || second.RoadName == nil || *second.RoadName != "Jl. MH Thamrin" {
		t.Fatalf("unexpected cached result: %#v", second)
	}
	if calls != 1 {
		t.Fatalf("expected single upstream call due cache, got %d", calls)
	}
}

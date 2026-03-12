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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"address": {
				"road": "Jl. MH Thamrin",
				"suburb": "Menteng",
				"city": "Jakarta Pusat"
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

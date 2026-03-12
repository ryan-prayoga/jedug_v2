package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ReverseGeocodeResult struct {
	RoadName   *string
	RegionName *string
	CityName   *string
}

type ReverseGeocoder interface {
	ReverseLookup(ctx context.Context, longitude, latitude float64) (*ReverseGeocodeResult, error)
}

type noopReverseGeocoder struct{}

func (noopReverseGeocoder) ReverseLookup(_ context.Context, _ float64, _ float64) (*ReverseGeocodeResult, error) {
	return nil, nil
}

type reverseGeocodeCacheEntry struct {
	result    *ReverseGeocodeResult
	expiresAt time.Time
}

type httpReverseGeocoder struct {
	baseURL   string
	userAgent string
	timeout   time.Duration
	cacheTTL  time.Duration
	client    *http.Client

	mu    sync.RWMutex
	cache map[string]reverseGeocodeCacheEntry
}

const bigDataCloudReverseURL = "https://api.bigdatacloud.net/data/reverse-geocode-client"

func NewHTTPReverseGeocoder(
	enabled bool,
	baseURL string,
	userAgent string,
	timeout time.Duration,
	cacheTTL time.Duration,
) ReverseGeocoder {
	if !enabled {
		return noopReverseGeocoder{}
	}

	trimmedURL := strings.TrimSpace(baseURL)
	if trimmedURL == "" {
		return noopReverseGeocoder{}
	}

	if strings.TrimSpace(userAgent) == "" {
		userAgent = "jedug-api/1.0"
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}

	return &httpReverseGeocoder{
		baseURL:   trimmedURL,
		userAgent: userAgent,
		timeout:   timeout,
		cacheTTL:  cacheTTL,
		client:    &http.Client{},
		cache:     make(map[string]reverseGeocodeCacheEntry),
	}
}

func (g *httpReverseGeocoder) ReverseLookup(ctx context.Context, longitude, latitude float64) (*ReverseGeocodeResult, error) {
	cacheKey := roundedCoordinateKey(longitude, latitude)
	now := time.Now()

	g.mu.RLock()
	cached, ok := g.cache[cacheKey]
	g.mu.RUnlock()
	if ok && now.Before(cached.expiresAt) {
		return cached.result, nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	var lookupErrs []error

	result, err := g.lookupNominatim(requestCtx, longitude, latitude)
	if err != nil {
		lookupErrs = append(lookupErrs, err)
	}

	if result == nil {
		secondary, secondaryErr := g.lookupBigDataCloud(requestCtx, longitude, latitude)
		if secondaryErr != nil {
			lookupErrs = append(lookupErrs, secondaryErr)
		}
		if secondary != nil {
			result = mergeReverseResults(result, secondary)
		}
	}

	if result != nil {
		// Keep consistent quality: normalize empty strings to nil.
		result.RoadName = firstNonEmptyStringPtr(ptrValueOrEmpty(result.RoadName))
		result.RegionName = firstNonEmptyStringPtr(ptrValueOrEmpty(result.RegionName))
		result.CityName = firstNonEmptyStringPtr(ptrValueOrEmpty(result.CityName))
	}

	if result == nil && len(lookupErrs) > 0 {
		return nil, errors.Join(lookupErrs...)
	}

	g.mu.Lock()
	g.cache[cacheKey] = reverseGeocodeCacheEntry{
		result:    result,
		expiresAt: now.Add(g.cacheTTL),
	}
	g.mu.Unlock()

	return result, nil
}

func (g *httpReverseGeocoder) lookupNominatim(
	ctx context.Context,
	longitude, latitude float64,
) (*ReverseGeocodeResult, error) {
	endpoint, err := url.Parse(g.baseURL)
	if err != nil {
		return nil, fmt.Errorf("nominatim parse url: %w", err)
	}

	query := endpoint.Query()
	query.Set("format", "jsonv2")
	query.Set("addressdetails", "1")
	query.Set("zoom", "18")
	query.Set("lat", strconv.FormatFloat(latitude, 'f', 6, 64))
	query.Set("lon", strconv.FormatFloat(longitude, 'f', 6, 64))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("nominatim build request: %w", err)
	}
	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "id")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("nominatim status: %d", resp.StatusCode)
	}

	var payload struct {
		Address struct {
			Road          string `json:"road"`
			Pedestrian    string `json:"pedestrian"`
			Residential   string `json:"residential"`
			Street        string `json:"street"`
			Suburb        string `json:"suburb"`
			Neighbourhood string `json:"neighbourhood"`
			Village       string `json:"village"`
			District      string `json:"district"`
			CityDistrict  string `json:"city_district"`
			City          string `json:"city"`
			Town          string `json:"town"`
			County        string `json:"county"`
			Regency       string `json:"regency"`
			StateDistrict string `json:"state_district"`
			Municipality  string `json:"municipality"`
		} `json:"address"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("nominatim decode response: %w", err)
	}

	result := &ReverseGeocodeResult{
		RoadName: firstNonEmptyStringPtr(
			payload.Address.Road,
			payload.Address.Pedestrian,
			payload.Address.Residential,
			payload.Address.Street,
		),
		RegionName: firstNonEmptyStringPtr(
			payload.Address.Suburb,
			payload.Address.Neighbourhood,
			payload.Address.Village,
			payload.Address.District,
			payload.Address.CityDistrict,
		),
		CityName: firstNonEmptyStringPtr(
			payload.Address.City,
			payload.Address.Town,
			payload.Address.Regency,
			payload.Address.County,
			payload.Address.StateDistrict,
			payload.Address.Municipality,
		),
	}

	if result.RoadName == nil && result.RegionName == nil && result.CityName == nil {
		return nil, nil
	}

	return result, nil
}

func (g *httpReverseGeocoder) lookupBigDataCloud(
	ctx context.Context,
	longitude, latitude float64,
) (*ReverseGeocodeResult, error) {
	endpoint, err := url.Parse(bigDataCloudReverseURL)
	if err != nil {
		return nil, fmt.Errorf("bigdatacloud parse url: %w", err)
	}

	query := endpoint.Query()
	query.Set("latitude", strconv.FormatFloat(latitude, 'f', 6, 64))
	query.Set("longitude", strconv.FormatFloat(longitude, 'f', 6, 64))
	query.Set("localityLanguage", "id")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("bigdatacloud build request: %w", err)
	}
	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bigdatacloud request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("bigdatacloud status: %d", resp.StatusCode)
	}

	var payload struct {
		City                 string `json:"city"`
		Locality             string `json:"locality"`
		PrincipalSubdivision string `json:"principalSubdivision"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("bigdatacloud decode response: %w", err)
	}

	result := &ReverseGeocodeResult{
		RoadName: nil,
		RegionName: firstNonEmptyStringPtr(
			payload.Locality,
		),
		CityName: firstNonEmptyStringPtr(
			payload.City,
			payload.PrincipalSubdivision,
		),
	}
	if result.RegionName == nil && result.CityName == nil {
		return nil, nil
	}
	return result, nil
}

func mergeReverseResults(primary, secondary *ReverseGeocodeResult) *ReverseGeocodeResult {
	if primary == nil {
		return secondary
	}
	if secondary == nil {
		return primary
	}

	return &ReverseGeocodeResult{
		RoadName:   firstNonEmptyStringPtr(ptrValueOrEmpty(primary.RoadName), ptrValueOrEmpty(secondary.RoadName)),
		RegionName: firstNonEmptyStringPtr(ptrValueOrEmpty(primary.RegionName), ptrValueOrEmpty(secondary.RegionName)),
		CityName:   firstNonEmptyStringPtr(ptrValueOrEmpty(primary.CityName), ptrValueOrEmpty(secondary.CityName)),
	}
}

func roundedCoordinateKey(longitude, latitude float64) string {
	return fmt.Sprintf("%.5f,%.5f", longitude, latitude)
}

func firstNonEmptyStringPtr(values ...string) *string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		return &trimmed
	}
	return nil
}

func ptrValueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

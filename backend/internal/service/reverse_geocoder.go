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
	RoadName     *string
	RegionName   *string
	CityName     *string
	DistrictName *string
	RegencyName  *string
	ProvinceName *string
	DisplayName  *string
	Postcode     *string
	CountryName  *string
	CountryCode  *string
	Category     *string
	Type         *string
	AddressType  *string
	PlaceRank    *int
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
		result.RegionName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.RegionName)))
		result.CityName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.CityName)))
		result.DistrictName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.DistrictName)))
		result.RegencyName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.RegencyName)))
		result.ProvinceName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.ProvinceName)))
		result.DisplayName = firstNonEmptyStringPtr(ptrValueOrEmpty(result.DisplayName))
		result.Postcode = firstNonEmptyStringPtr(ptrValueOrEmpty(result.Postcode))
		result.CountryName = normalizePlaceNamePtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.CountryName)))
		result.CountryCode = upperStringPtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.CountryCode)))
		result.Category = lowerStringPtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.Category)))
		result.Type = lowerStringPtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.Type)))
		result.AddressType = lowerStringPtr(firstNonEmptyStringPtr(ptrValueOrEmpty(result.AddressType)))
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
	query.Set("accept-language", "id")
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
		DisplayName string `json:"display_name"`
		Category    string `json:"category"`
		Type        string `json:"type"`
		AddressType string `json:"addresstype"`
		PlaceRank   any    `json:"place_rank"`
		Address     struct {
			Road          string `json:"road"`
			Pedestrian    string `json:"pedestrian"`
			Residential   string `json:"residential"`
			Street        string `json:"street"`
			Quarter       string `json:"quarter"`
			Suburb        string `json:"suburb"`
			Neighbourhood string `json:"neighbourhood"`
			Village       string `json:"village"`
			Hamlet        string `json:"hamlet"`
			District      string `json:"district"`
			CityDistrict  string `json:"city_district"`
			Region        string `json:"region"`
			City          string `json:"city"`
			Town          string `json:"town"`
			County        string `json:"county"`
			Regency       string `json:"regency"`
			StateDistrict string `json:"state_district"`
			State         string `json:"state"`
			Postcode      string `json:"postcode"`
			Country       string `json:"country"`
			CountryCode   string `json:"country_code"`
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
			payload.Address.Neighbourhood,
			payload.Address.Suburb,
			payload.Address.Village,
			payload.Address.Hamlet,
			payload.Address.District,
			payload.Address.CityDistrict,
			payload.Address.Quarter,
		),
		CityName: firstNonEmptyStringPtr(
			payload.Address.City,
			payload.Address.Town,
			payload.Address.Regency,
			payload.Address.County,
			payload.Address.StateDistrict,
			payload.Address.Municipality,
			payload.Address.State,
		),
		DistrictName: firstNonEmptyStringPtr(
			payload.Address.Neighbourhood,
			payload.Address.Suburb,
			payload.Address.Village,
			payload.Address.Hamlet,
			payload.Address.District,
			payload.Address.CityDistrict,
			payload.Address.Quarter,
		),
		RegencyName: firstNonEmptyStringPtr(
			payload.Address.City,
			payload.Address.Town,
			payload.Address.Regency,
			payload.Address.County,
			payload.Address.StateDistrict,
			payload.Address.Municipality,
		),
		ProvinceName: firstNonEmptyStringPtr(
			payload.Address.State,
			payload.Address.Region,
		),
		DisplayName: firstNonEmptyStringPtr(
			payload.DisplayName,
		),
		Postcode: firstNonEmptyStringPtr(
			payload.Address.Postcode,
		),
		CountryName: firstNonEmptyStringPtr(
			payload.Address.Country,
		),
		CountryCode: firstNonEmptyStringPtr(
			payload.Address.CountryCode,
		),
		Category: firstNonEmptyStringPtr(
			payload.Category,
		),
		Type: firstNonEmptyStringPtr(
			payload.Type,
		),
		AddressType: firstNonEmptyStringPtr(
			payload.AddressType,
		),
	}
	result.PlaceRank = parseOptionalInt(payload.PlaceRank)

	if result.RoadName == nil &&
		result.RegionName == nil &&
		result.CityName == nil &&
		result.DistrictName == nil &&
		result.RegencyName == nil &&
		result.ProvinceName == nil &&
		result.DisplayName == nil {
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
		DistrictName: firstNonEmptyStringPtr(
			payload.Locality,
		),
		RegencyName: firstNonEmptyStringPtr(
			payload.City,
		),
		ProvinceName: firstNonEmptyStringPtr(
			payload.PrincipalSubdivision,
		),
	}
	if result.RegionName == nil &&
		result.CityName == nil &&
		result.DistrictName == nil &&
		result.RegencyName == nil &&
		result.ProvinceName == nil {
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
		RoadName:     firstNonEmptyStringPtr(ptrValueOrEmpty(primary.RoadName), ptrValueOrEmpty(secondary.RoadName)),
		RegionName:   firstNonEmptyStringPtr(ptrValueOrEmpty(primary.RegionName), ptrValueOrEmpty(secondary.RegionName)),
		CityName:     firstNonEmptyStringPtr(ptrValueOrEmpty(primary.CityName), ptrValueOrEmpty(secondary.CityName)),
		DistrictName: firstNonEmptyStringPtr(ptrValueOrEmpty(primary.DistrictName), ptrValueOrEmpty(secondary.DistrictName)),
		RegencyName:  firstNonEmptyStringPtr(ptrValueOrEmpty(primary.RegencyName), ptrValueOrEmpty(secondary.RegencyName)),
		ProvinceName: firstNonEmptyStringPtr(ptrValueOrEmpty(primary.ProvinceName), ptrValueOrEmpty(secondary.ProvinceName)),
		DisplayName:  firstNonEmptyStringPtr(ptrValueOrEmpty(primary.DisplayName), ptrValueOrEmpty(secondary.DisplayName)),
		Postcode:     firstNonEmptyStringPtr(ptrValueOrEmpty(primary.Postcode), ptrValueOrEmpty(secondary.Postcode)),
		CountryName:  firstNonEmptyStringPtr(ptrValueOrEmpty(primary.CountryName), ptrValueOrEmpty(secondary.CountryName)),
		CountryCode:  firstNonEmptyStringPtr(ptrValueOrEmpty(primary.CountryCode), ptrValueOrEmpty(secondary.CountryCode)),
		Category:     firstNonEmptyStringPtr(ptrValueOrEmpty(primary.Category), ptrValueOrEmpty(secondary.Category)),
		Type:         firstNonEmptyStringPtr(ptrValueOrEmpty(primary.Type), ptrValueOrEmpty(secondary.Type)),
		AddressType:  firstNonEmptyStringPtr(ptrValueOrEmpty(primary.AddressType), ptrValueOrEmpty(secondary.AddressType)),
		PlaceRank:    firstNonNilInt(primary.PlaceRank, secondary.PlaceRank),
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

func normalizePlaceNamePtr(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := normalizeIndonesianPlaceName(*value)
	if strings.TrimSpace(normalized) == "" {
		return nil
	}
	return &normalized
}

func normalizeIndonesianPlaceName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	switch strings.ToLower(trimmed) {
	case "central jakarta":
		return "Jakarta Pusat"
	case "south jakarta":
		return "Jakarta Selatan"
	case "west jakarta":
		return "Jakarta Barat"
	case "east jakarta":
		return "Jakarta Timur"
	case "north jakarta":
		return "Jakarta Utara"
	case "special capital region of jakarta", "jakarta special capital region":
		return "Daerah Khusus Ibukota Jakarta"
	case "east java":
		return "Jawa Timur"
	case "central java":
		return "Jawa Tengah"
	case "west java":
		return "Jawa Barat"
	case "indonesia":
		return "Indonesia"
	default:
		return trimmed
	}
}

func upperStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.ToUpper(strings.TrimSpace(*value))
	if normalized == "" {
		return nil
	}
	return &normalized
}

func lowerStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" {
		return nil
	}
	return &normalized
}

func parseOptionalInt(value any) *int {
	switch v := value.(type) {
	case nil:
		return nil
	case float64:
		out := int(v)
		return &out
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

func firstNonNilInt(values ...*int) *int {
	for _, value := range values {
		if value != nil {
			out := *value
			return &out
		}
	}
	return nil
}

package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName                 string
	AppEnv                  string
	AppPort                 string
	DatabaseURL             string
	CORSAllowOrigins        string
	StorageDriver           string
	StoragePublicBaseURL    string
	UploadDir               string
	R2AccountID             string
	R2AccessKeyID           string
	R2SecretAccessKey       string
	R2Bucket                string
	R2Endpoint              string
	R2PublicBaseURL         string
	AdminUsername           string
	AdminPassword           string
	DuplicateRadiusM        float64
	ReverseGeocodeEnabled   bool
	ReverseGeocodeURL       string
	ReverseGeocodeUserAgent string
	ReverseGeocodeTimeout   time.Duration
	ReverseGeocodeCacheTTL  time.Duration
	WebPushVAPIDPublicKey   string
	WebPushVAPIDPrivateKey  string
	WebPushSubscriber       string
	WebPushSiteURL          string
	WebPushTTLSeconds       int
}

func Load() *Config {
	cfg := &Config{
		AppName:                 getEnv("APP_NAME", "jedug-api"),
		AppEnv:                  getEnv("APP_ENV", "development"),
		AppPort:                 getEnv("APP_PORT", "8080"),
		DatabaseURL:             mustGetEnv("DATABASE_URL"),
		CORSAllowOrigins:        getEnv("CORS_ALLOW_ORIGINS", "*"),
		StorageDriver:           strings.ToLower(getEnv("STORAGE_DRIVER", "local")),
		StoragePublicBaseURL:    getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:               getEnv("UPLOAD_DIR", "./uploads/gallery"),
		R2AccountID:             getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:           getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:       getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:                getEnv("R2_BUCKET", ""),
		R2Endpoint:              getEnv("R2_ENDPOINT", ""),
		R2PublicBaseURL:         getEnv("R2_PUBLIC_BASE_URL", ""),
		AdminUsername:           getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:           mustGetEnv("ADMIN_PASSWORD"),
		DuplicateRadiusM:        getEnvPositiveFloat64("DUPLICATE_RADIUS_M", 30),
		ReverseGeocodeEnabled:   getEnvBool("REVERSE_GEOCODE_ENABLED", true),
		ReverseGeocodeURL:       getEnv("REVERSE_GEOCODE_URL", "https://nominatim.openstreetmap.org/reverse"),
		ReverseGeocodeUserAgent: getEnv("REVERSE_GEOCODE_USER_AGENT", "jedug-api/1.0"),
		ReverseGeocodeTimeout:   getEnvPositiveDurationMS("REVERSE_GEOCODE_TIMEOUT_MS", 2000),
		ReverseGeocodeCacheTTL:  getEnvPositiveDurationSec("REVERSE_GEOCODE_CACHE_TTL_SEC", 300),
		WebPushVAPIDPublicKey:   strings.TrimSpace(getEnv("WEB_PUSH_VAPID_PUBLIC_KEY", "")),
		WebPushVAPIDPrivateKey:  strings.TrimSpace(getEnv("WEB_PUSH_VAPID_PRIVATE_KEY", "")),
		WebPushSubscriber:       strings.TrimSpace(getEnv("WEB_PUSH_SUBSCRIBER", "")),
		WebPushSiteURL:          strings.TrimRight(strings.TrimSpace(getEnv("WEB_PUSH_SITE_URL", "")), "/"),
		WebPushTTLSeconds:       getEnvPositiveInt("WEB_PUSH_TTL_SEC", 300),
	}

	validateWebPushConfig(cfg)
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvPositiveFloat64(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil || parsed <= 0 {
		panic("invalid positive float environment variable: " + key)
	}

	return parsed
}

func getEnvPositiveInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		panic("invalid positive integer environment variable: " + key)
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}

	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		panic("invalid boolean environment variable: " + key)
	}
}

func getEnvPositiveDurationMS(key string, fallbackMS int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(fallbackMS) * time.Millisecond
	}

	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		panic("invalid positive duration(ms) environment variable: " + key)
	}

	return time.Duration(parsed) * time.Millisecond
}

func getEnvPositiveDurationSec(key string, fallbackSec int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(fallbackSec) * time.Second
	}

	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		panic("invalid positive duration(sec) environment variable: " + key)
	}

	return time.Duration(parsed) * time.Second
}

// mustGetEnv panics at startup if a required env var is missing.
func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func validateWebPushConfig(cfg *Config) {
	values := map[string]string{
		"WEB_PUSH_VAPID_PUBLIC_KEY":  cfg.WebPushVAPIDPublicKey,
		"WEB_PUSH_VAPID_PRIVATE_KEY": cfg.WebPushVAPIDPrivateKey,
		"WEB_PUSH_SUBSCRIBER":        cfg.WebPushSubscriber,
		"WEB_PUSH_SITE_URL":          cfg.WebPushSiteURL,
	}

	setCount := 0
	for _, value := range values {
		if value != "" {
			setCount++
		}
	}

	if setCount == 0 {
		return
	}

	if setCount != len(values) {
		panic("web push env must be configured together: WEB_PUSH_VAPID_PUBLIC_KEY, WEB_PUSH_VAPID_PRIVATE_KEY, WEB_PUSH_SUBSCRIBER, WEB_PUSH_SITE_URL")
	}
}

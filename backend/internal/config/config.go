package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName                            string
	AppEnv                             string
	AppPort                            string
	DatabaseURL                        string
	CORSAllowOrigins                   string
	StorageDriver                      string
	StoragePublicBaseURL               string
	UploadDir                          string
	R2AccountID                        string
	R2AccessKeyID                      string
	R2SecretAccessKey                  string
	R2Bucket                           string
	R2Endpoint                         string
	R2PublicBaseURL                    string
	AdminUsername                      string
	AdminPassword                      string
	DuplicateRadiusM                   float64
	ReverseGeocodeEnabled              bool
	ReverseGeocodeURL                  string
	ReverseGeocodeUserAgent            string
	ReverseGeocodeTimeout              time.Duration
	ReverseGeocodeCacheTTL             time.Duration
	UploadTokenSecret                  string
	UploadTicketTTL                    time.Duration
	UploadPendingWindow                time.Duration
	UploadPendingLimit                 int
	FollowerTokenSecret                string
	FollowerTokenTTL                   time.Duration
	FollowerStreamTokenTTL             time.Duration
	WebPushVAPIDPublicKey              string
	WebPushVAPIDPrivateKey             string
	WebPushSubscriber                  string
	WebPushSiteURL                     string
	WebPushTTLSeconds                  int
	MaintenanceEnabled                 bool
	MaintenanceInterval                time.Duration
	NotificationsRetention             time.Duration
	PushSubscriptionsStaleAfter        time.Duration
	PushSubscriptionsDisabledRetention time.Duration
	PushDeliveryDeliveredRetention     time.Duration
	PushDeliveryFailedRetention        time.Duration
	UploadOrphanRetention              time.Duration
}

func Load() *Config {
	cfg := &Config{
		AppName:                            getEnv("APP_NAME", "jedug-api"),
		AppEnv:                             getEnv("APP_ENV", "development"),
		AppPort:                            getEnv("APP_PORT", "8080"),
		DatabaseURL:                        mustGetEnv("DATABASE_URL"),
		CORSAllowOrigins:                   getEnv("CORS_ALLOW_ORIGINS", "*"),
		StorageDriver:                      strings.ToLower(getEnv("STORAGE_DRIVER", "local")),
		StoragePublicBaseURL:               getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:                          getEnv("UPLOAD_DIR", "./uploads/gallery"),
		R2AccountID:                        getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:                      getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:                  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Bucket:                           getEnv("R2_BUCKET", ""),
		R2Endpoint:                         getEnv("R2_ENDPOINT", ""),
		R2PublicBaseURL:                    getEnv("R2_PUBLIC_BASE_URL", ""),
		AdminUsername:                      strings.TrimSpace(mustGetEnv("ADMIN_USERNAME")),
		AdminPassword:                      mustGetEnv("ADMIN_PASSWORD"),
		DuplicateRadiusM:                   getEnvPositiveFloat64("DUPLICATE_RADIUS_M", 30),
		ReverseGeocodeEnabled:              getEnvBool("REVERSE_GEOCODE_ENABLED", true),
		ReverseGeocodeURL:                  getEnv("REVERSE_GEOCODE_URL", "https://nominatim.openstreetmap.org/reverse"),
		ReverseGeocodeUserAgent:            getEnv("REVERSE_GEOCODE_USER_AGENT", "jedug-api/1.0"),
		ReverseGeocodeTimeout:              getEnvPositiveDurationMS("REVERSE_GEOCODE_TIMEOUT_MS", 2000),
		ReverseGeocodeCacheTTL:             getEnvPositiveDurationSec("REVERSE_GEOCODE_CACHE_TTL_SEC", 300),
		UploadTokenSecret:                  strings.TrimSpace(getEnv("UPLOAD_TOKEN_SECRET", "")),
		UploadTicketTTL:                    getEnvPositiveDurationSec("UPLOAD_TICKET_TTL_SEC", 10*60),
		UploadPendingWindow:                getEnvPositiveDurationSec("UPLOAD_PENDING_WINDOW_SEC", 30*60),
		UploadPendingLimit:                 getEnvPositiveInt("UPLOAD_PENDING_LIMIT", 4),
		FollowerTokenSecret:                strings.TrimSpace(mustGetEnv("FOLLOWER_TOKEN_SECRET")),
		FollowerTokenTTL:                   getEnvPositiveDurationSec("FOLLOWER_TOKEN_TTL_SEC", 12*60*60),
		FollowerStreamTokenTTL:             getEnvPositiveDurationSec("FOLLOWER_STREAM_TOKEN_TTL_SEC", 10*60),
		WebPushVAPIDPublicKey:              strings.TrimSpace(getEnv("WEB_PUSH_VAPID_PUBLIC_KEY", "")),
		WebPushVAPIDPrivateKey:             strings.TrimSpace(getEnv("WEB_PUSH_VAPID_PRIVATE_KEY", "")),
		WebPushSubscriber:                  strings.TrimSpace(getEnv("WEB_PUSH_SUBSCRIBER", "")),
		WebPushSiteURL:                     strings.TrimRight(strings.TrimSpace(getEnv("WEB_PUSH_SITE_URL", "")), "/"),
		WebPushTTLSeconds:                  getEnvPositiveInt("WEB_PUSH_TTL_SEC", 300),
		MaintenanceEnabled:                 getEnvBool("MAINTENANCE_ENABLED", true),
		MaintenanceInterval:                getEnvPositiveDurationSec("MAINTENANCE_INTERVAL_SEC", 6*60*60),
		NotificationsRetention:             getEnvPositiveDurationDays("NOTIFICATIONS_RETENTION_DAYS", 90),
		PushSubscriptionsStaleAfter:        getEnvPositiveDurationDays("PUSH_SUBSCRIPTIONS_STALE_DAYS", 180),
		PushSubscriptionsDisabledRetention: getEnvPositiveDurationDays("PUSH_SUBSCRIPTIONS_DISABLED_RETENTION_DAYS", 30),
		PushDeliveryDeliveredRetention:     getEnvPositiveDurationDays("PUSH_DELIVERY_DELIVERED_RETENTION_DAYS", 14),
		PushDeliveryFailedRetention:        getEnvPositiveDurationDays("PUSH_DELIVERY_FAILED_RETENTION_DAYS", 30),
		UploadOrphanRetention:              getEnvPositiveDurationSec("UPLOAD_ORPHAN_RETENTION_SEC", 12*60*60),
	}
	if cfg.UploadTokenSecret == "" {
		cfg.UploadTokenSecret = cfg.FollowerTokenSecret
	}

	validateWebPushConfig(cfg)
	validateFollowerTokenConfig(cfg)
	validateUploadTokenConfig(cfg)
	validateAdminConfig(cfg)
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

func getEnvPositiveDurationDays(key string, fallbackDays int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(fallbackDays) * 24 * time.Hour
	}

	parsed, err := strconv.Atoi(v)
	if err != nil || parsed <= 0 {
		panic("invalid positive duration(days) environment variable: " + key)
	}

	return time.Duration(parsed) * 24 * time.Hour
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

func validateFollowerTokenConfig(cfg *Config) {
	if len(cfg.FollowerTokenSecret) < 32 {
		panic("FOLLOWER_TOKEN_SECRET must be at least 32 characters")
	}
}

func validateUploadTokenConfig(cfg *Config) {
	if len(cfg.UploadTokenSecret) < 32 {
		panic("UPLOAD_TOKEN_SECRET must be at least 32 characters")
	}
}

func validateAdminConfig(cfg *Config) {
	if cfg.AdminUsername == "" {
		panic("ADMIN_USERNAME must not be empty")
	}
	if len(strings.TrimSpace(cfg.AdminPassword)) < 12 {
		panic("ADMIN_PASSWORD must be at least 12 characters")
	}
}

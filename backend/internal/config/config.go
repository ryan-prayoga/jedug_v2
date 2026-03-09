package config

import "os"

type Config struct {
	AppName              string
	AppEnv               string
	AppPort              string
	DatabaseURL          string
	CORSAllowOrigins     string
	StorageDriver        string
	StoragePublicBaseURL string
	UploadDir            string
	AdminUsername        string
	AdminPassword        string
}

func Load() *Config {
	return &Config{
		AppName:              getEnv("APP_NAME", "jedug-api"),
		AppEnv:               getEnv("APP_ENV", "development"),
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          mustGetEnv("DATABASE_URL"),
		CORSAllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "*"),
		StorageDriver:        getEnv("STORAGE_DRIVER", "local"),
		StoragePublicBaseURL: getEnv("STORAGE_PUBLIC_BASE_URL", "http://localhost:8080"),
		UploadDir:            getEnv("UPLOAD_DIR", "./uploads/gallery"),
		AdminUsername:        getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:        mustGetEnv("ADMIN_PASSWORD"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// mustGetEnv panics at startup if a required env var is missing.
func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultListenAddress  = ":8080"
	defaultAPIKey         = "parkiroid-dev-key"
	defaultJWTSecret      = "parkiroid-dev-jwt-secret"
	defaultTokenTTL       = 24 * time.Hour
	defaultDatabasePath   = "parkiroid.db"
	defaultFramesDir      = "frames"
	defaultRetentionDays  = 7
)

type Config struct {
	ListenAddress   string
	APIKey          string
	JWTSecret       string
	TokenTTL        time.Duration
	DatabasePath    string
	FramesDir       string
	RetentionPeriod time.Duration
}

func Load() Config {
	return Config{
		ListenAddress:   envOrDefault("PARKIROID_LISTEN_ADDRESS", defaultListenAddress),
		APIKey:          envOrDefault("PARKIROID_API_KEY", defaultAPIKey),
		JWTSecret:       envOrDefault("PARKIROID_JWT_SECRET", defaultJWTSecret),
		TokenTTL:        envDurationOrDefault("PARKIROID_TOKEN_TTL", defaultTokenTTL),
		DatabasePath:    envOrDefault("PARKIROID_DATABASE_PATH", defaultDatabasePath),
		FramesDir:       envOrDefault("PARKIROID_FRAMES_DIR", defaultFramesDir),
		RetentionPeriod: envDaysOrDefault("PARKIROID_RETENTION_DAYS", defaultRetentionDays),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func envDaysOrDefault(key string, fallbackDays int) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return time.Duration(fallbackDays) * 24 * time.Hour
	}

	days, err := strconv.Atoi(raw)
	if err != nil || days <= 0 {
		return time.Duration(fallbackDays) * 24 * time.Hour
	}

	return time.Duration(days) * 24 * time.Hour
}

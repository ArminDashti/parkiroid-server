package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultListenAddress = ":8080"
	defaultAPIKey        = "parkiroid-dev-key"
	defaultJWTSecret     = "parkiroid-dev-jwt-secret"
	defaultTokenTTL      = 24 * time.Hour
	defaultDatabasePath  = "parkiroid.db"
	defaultFramesDir     = "frames"
)

type Config struct {
	ListenAddress string
	APIKey        string
	JWTSecret     string
	TokenTTL      time.Duration
	DatabasePath  string
	FramesDir     string
}

func Load() Config {
	return Config{
		ListenAddress: envOrDefault("PARKIROID_LISTEN_ADDRESS", defaultListenAddress),
		APIKey:        envOrDefault("PARKIROID_API_KEY", defaultAPIKey),
		JWTSecret:     envOrDefault("PARKIROID_JWT_SECRET", defaultJWTSecret),
		TokenTTL:      envDurationOrDefault("PARKIROID_TOKEN_TTL", defaultTokenTTL),
		DatabasePath:  envOrDefault("PARKIROID_DATABASE_PATH", defaultDatabasePath),
		FramesDir:     envOrDefault("PARKIROID_FRAMES_DIR", defaultFramesDir),
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

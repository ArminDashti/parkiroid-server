package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultListenAddress     = ":8080"
	defaultJWTSecret         = "parkiroid-dev-jwt-secret"
	defaultTokenTTL          = 24 * time.Hour
	defaultDatabasePath      = "parkiroid.db"
	defaultFramesDir         = "frames"
	defaultRetentionDays     = 7
	defaultLiveKitURL        = "ws://localhost:7880"
	defaultLiveKitAPIKey     = "devkey"
	defaultLiveKitAPISecret  = "secret"
	defaultLiveKitTokenTTL   = time.Hour
)

type Config struct {
	ListenAddress   string
	JWTSecret       string
	TokenTTL        time.Duration
	DatabasePath    string
	FramesDir       string
	RetentionPeriod time.Duration
	LiveKitURL      string
	LiveKitAPIKey   string
	LiveKitAPISecret string
	LiveKitTokenTTL time.Duration
}

func Load() Config {
	return Config{
		ListenAddress:    envOrDefault("PARKIROID_LISTEN_ADDRESS", defaultListenAddress),
		JWTSecret:        envOrDefault("PARKIROID_JWT_SECRET", defaultJWTSecret),
		TokenTTL:         envDurationOrDefault("PARKIROID_TOKEN_TTL", defaultTokenTTL),
		DatabasePath:     envOrDefault("PARKIROID_DATABASE_PATH", defaultDatabasePath),
		FramesDir:        envOrDefault("PARKIROID_FRAMES_DIR", defaultFramesDir),
		RetentionPeriod:  envDaysOrDefault("PARKIROID_RETENTION_DAYS", defaultRetentionDays),
		LiveKitURL:       envOrDefault("PARKIROID_LIVEKIT_URL", defaultLiveKitURL),
		LiveKitAPIKey:    envOrDefault("PARKIROID_LIVEKIT_API_KEY", defaultLiveKitAPIKey),
		LiveKitAPISecret: envOrDefault("PARKIROID_LIVEKIT_API_SECRET", defaultLiveKitAPISecret),
		LiveKitTokenTTL:  envDurationOrDefault("PARKIROID_LIVEKIT_TOKEN_TTL", defaultLiveKitTokenTTL),
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

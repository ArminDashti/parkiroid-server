package auth

import (
	"crypto/subtle"
	"strings"
	"time"
)

const DefaultDeviceAPIKey = "dogan-dev-key"

func ValidateDeviceAPIKey(apiKey, embeddedAPIToken, configuredDeviceAPIKey string) bool {
	if apiKey == "" {
		return false
	}

	if embeddedAPIToken != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(embeddedAPIToken)) == 1 {
		return true
	}

	deviceAPIKey := configuredDeviceAPIKey
	if deviceAPIKey == "" {
		deviceAPIKey = DefaultDeviceAPIKey
	}

	return subtle.ConstantTimeCompare([]byte(apiKey), []byte(deviceAPIKey)) == 1
}

func DeviceTokenExpiry(validFor time.Duration) time.Time {
	if validFor <= 0 {
		validFor = 365 * 24 * time.Hour
	}
	return time.Now().UTC().Add(validFor)
}

func LoginIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if atIndex := strings.Index(value, "@"); atIndex > 0 {
		return value[:atIndex]
	}
	return value
}

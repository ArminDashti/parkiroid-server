package livekit

import (
	"strings"
	"unicode"
)

func RoomNameForDevice(deviceID string) string {
	sanitized := sanitizeIdentifier(deviceID)
	if sanitized == "" {
		return "device-unknown"
	}

	return "device-" + sanitized
}

func DefaultParticipantIdentity(role, deviceID string) string {
	sanitizedDeviceID := sanitizeIdentifier(deviceID)
	if sanitizedDeviceID == "" {
		sanitizedDeviceID = "unknown"
	}

	switch role {
	case RolePublisher:
		return "publisher-" + sanitizedDeviceID
	default:
		return "subscriber-" + sanitizedDeviceID
	}
}

func sanitizeIdentifier(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for _, character := range strings.TrimSpace(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || character == '-' || character == '_' {
			builder.WriteRune(character)
			continue
		}

		builder.WriteRune('-')
	}

	return strings.Trim(builder.String(), "-")
}

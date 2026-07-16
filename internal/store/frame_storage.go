package store

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PersistFrameImage(framesDir, deviceID string, capturedAt time.Time, imageData string) (string, error) {
	decodedImage, err := decodeImagePayload(imageData)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return "", fmt.Errorf("create frames directory: %w", err)
	}

	fileName := fmt.Sprintf("%s-%s.jpg", sanitizePathSegment(deviceID), capturedAt.Format("060102-150405"))
	framePath := filepath.Join(framesDir, fileName)

	if err := os.WriteFile(framePath, decodedImage, 0o644); err != nil {
		return "", fmt.Errorf("write frame image: %w", err)
	}

	return framePath, nil
}

func decodeImagePayload(imageData string) ([]byte, error) {
	payload := strings.TrimSpace(imageData)
	if payload == "" {
		return nil, fmt.Errorf("image_data is empty")
	}

	if commaIndex := strings.Index(payload, ","); strings.HasPrefix(payload, "data:") && commaIndex != -1 {
		payload = payload[commaIndex+1:]
	}

	decodedImage, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode image_data: %w", err)
	}

	return decodedImage, nil
}

func sanitizePathSegment(value string) string {
	replaced := strings.Map(func(character rune) rune {
		switch character {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|', ' ':
			return '-'
		default:
			return character
		}
	}, value)

	return strings.Trim(replaced, "-")
}

// SanitizeDeviceSegment is the exported form of sanitizePathSegment for handlers.
func SanitizeDeviceSegment(value string) string {
	return sanitizePathSegment(value)
}

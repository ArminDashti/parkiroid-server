package handlers

import (
	"strconv"
	"strings"

	"github.com/dogan/dogan-server/internal/models"
)

func settingsToFlatMap(settings []models.AppSettingRecord) map[string]any {
	flat := make(map[string]any, len(settings))
	for _, setting := range settings {
		flat[setting.Key] = coerceSettingValue(setting.Key, setting.Value)
	}
	return flat
}

func coerceSettingValue(key, value string) any {
	switch key {
	case "object_detection_on_device", "wifi_only_downloads":
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	case "capture_interval_ms", "telemetry_interval_ms":
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	case "screen_on_interval_min", "realtime_fps", "settings_sync_interval_sec":
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
		if err == nil {
			return int(parsed)
		}
	case "min_detection_confidence":
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return parsed
		}
	}

	return value
}

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
	case "object_detection_on_device", "wifi_only_downloads", "copilot_distance_control_enabled":
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	case "capture_interval_ms", "telemetry_interval_ms":
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	case "telemetry_interval_sec", "telemetry_retention_hours", "log_retention_days",
		"screen_on_interval_min", "realtime_fps", "settings_sync_interval_sec",
		"api_port", "stream_port", "copilot_video_chunk_minutes":
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

	if parsed, ok := coerceModePrefixedSetting(key, value); ok {
		return parsed
	}

	return value
}

func coerceModePrefixedSetting(key, value string) (any, bool) {
	switch {
	case strings.HasSuffix(key, "_fps"),
		strings.HasSuffix(key, "_image_retention_hours"),
		strings.HasSuffix(key, "_video_retention_hours"):
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
		if err == nil {
			return int(parsed), true
		}
	case strings.HasSuffix(key, "_min_confidence"):
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return parsed, true
		}
	case strings.HasSuffix(key, "_record_video"):
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			return parsed, true
		}
	}

	return nil, false
}

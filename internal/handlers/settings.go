package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsStore store.SettingsStore
}

func NewSettingsHandler(settingsStore store.SettingsStore) *SettingsHandler {
	return &SettingsHandler{settingsStore: settingsStore}
}

func (handler *SettingsHandler) GetSettings(context *gin.Context) {
	platform := context.Query("platform")
	deviceID := context.Query("device_id")
	if platform == "" && deviceID != "" {
		platform = "android"
	}
	if platform == "" {
		platform = "web"
	}

	settings, err := handler.settingsStore.GetSettings(platform)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve settings"})
		return
	}

	if platform == "web" && deviceID == "" && context.Query("platform") == "" {
		context.JSON(http.StatusOK, toWebSettingsResponse(settings))
		return
	}

	if deviceID != "" {
		flat := settingsToFlatMap(settings)
		flat["device_id"] = deviceID
		context.JSON(http.StatusOK, flat)
		return
	}

	if platform == "web" {
		context.JSON(http.StatusOK, toWebSettingsResponse(settings))
		return
	}

	context.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (handler *SettingsHandler) UpsertSetting(context *gin.Context) {
	var payload models.AppSettingPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	setting := models.AppSettingRecord{
		Platform:  payload.Platform,
		Key:       payload.Key,
		Value:     payload.Value,
		UpdatedAt: time.Now().UTC(),
	}

	if err := handler.settingsStore.UpsertSetting(setting); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save setting"})
		return
	}

	context.JSON(http.StatusOK, setting)
}

func (handler *SettingsHandler) PatchWebSettings(context *gin.Context) {
	var payload models.WebSettingsPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	now := time.Now().UTC()
	upserts := make([]models.AppSettingRecord, 0, 4)

	if payload.NotificationsEnabled != nil {
		upserts = append(upserts, models.AppSettingRecord{
			Platform:  "web",
			Key:       "notifications_enabled",
			Value:     strconv.FormatBool(*payload.NotificationsEnabled),
			UpdatedAt: now,
		})
	}
	if payload.TemperatureUnit != nil {
		upserts = append(upserts, models.AppSettingRecord{
			Platform:  "web",
			Key:       "temperature_unit",
			Value:     *payload.TemperatureUnit,
			UpdatedAt: now,
		})
	}
	if payload.NoiseAlertThresholdDb != nil {
		upserts = append(upserts, models.AppSettingRecord{
			Platform:  "web",
			Key:       "noise_alert_threshold_db",
			Value:     fmt.Sprintf("%v", *payload.NoiseAlertThresholdDb),
			UpdatedAt: now,
		})
	}
	if payload.DefaultDeviceID != nil {
		upserts = append(upserts, models.AppSettingRecord{
			Platform:  "web",
			Key:       "default_device_id",
			Value:     *payload.DefaultDeviceID,
			UpdatedAt: now,
		})
	}

	for _, setting := range upserts {
		if err := handler.settingsStore.UpsertSetting(setting); err != nil {
			context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save settings"})
			return
		}
	}

	settings, err := handler.settingsStore.GetSettings("web")
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve settings"})
		return
	}

	context.JSON(http.StatusOK, toWebSettingsResponse(settings))
}

func toWebSettingsResponse(settings []models.AppSettingRecord) models.WebSettingsResponse {
	response := models.WebSettingsResponse{
		NotificationsEnabled:  true,
		TemperatureUnit:       "celsius",
		NoiseAlertThresholdDb: 70,
	}

	for _, setting := range settings {
		switch setting.Key {
		case "notifications_enabled", "notificationsEnabled":
			parsed, err := strconv.ParseBool(setting.Value)
			if err == nil {
				response.NotificationsEnabled = parsed
			}
		case "temperature_unit", "temperatureUnit":
			if setting.Value != "" {
				response.TemperatureUnit = setting.Value
			}
		case "noise_alert_threshold_db", "noiseAlertThresholdDb":
			parsed, err := strconv.ParseFloat(setting.Value, 64)
			if err == nil {
				response.NoiseAlertThresholdDb = parsed
			}
		case "default_device_id", "defaultDeviceId":
			response.DefaultDeviceID = setting.Value
		}
	}

	return response
}

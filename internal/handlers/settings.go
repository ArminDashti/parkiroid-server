package handlers

import (
	"net/http"
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
	if platform == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "platform query parameter is required"})
		return
	}

	settings, err := handler.settingsStore.GetSettings(platform)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve settings"})
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

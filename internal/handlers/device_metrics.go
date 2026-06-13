package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/models"
	"github.com/parkiroid/parkiroid-server/internal/store"
)

type DeviceMetricsHandler struct {
	metricsStore store.MetricsStore
}

func NewDeviceMetricsHandler(metricsStore store.MetricsStore) *DeviceMetricsHandler {
	return &DeviceMetricsHandler{metricsStore: metricsStore}
}

func (handler *DeviceMetricsHandler) SubmitMetrics(context *gin.Context) {
	var payload models.DeviceMetricsPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	record := models.DeviceMetricsRecord{
		DeviceID:       payload.DeviceID,
		CPUUsage:       payload.CPUUsage,
		MemoryUsage:    payload.MemoryUsage,
		DiskUsage:      payload.DiskUsage,
		BatteryLevel:   payload.BatteryLevel,
		TemperatureC:   payload.TemperatureC,
		SignalStrength: payload.SignalStrength,
		RecordedAt:     store.NormalizeRecordedAt(payload.RecordedAt),
		ReceivedAt:     time.Now().UTC(),
	}

	if err := handler.metricsStore.SaveMetrics(record); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save metrics"})
		return
	}

	context.JSON(http.StatusCreated, record)
}

func (handler *DeviceMetricsHandler) GetLatestMetrics(context *gin.Context) {
	deviceID := context.Query("device-id")
	if deviceID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device-id query parameter is required"})
		return
	}

	metrics, err := handler.metricsStore.GetLatestMetrics(deviceID)
	if err != nil {
		if errors.Is(err, store.ErrMetricsNotFound) {
			context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "no metrics found for device"})
			return
		}
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve metrics"})
		return
	}

	context.JSON(http.StatusOK, metrics)
}

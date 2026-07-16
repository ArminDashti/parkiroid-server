package handlers

import (
	"net/http"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type TelemetryHandler struct {
	metricsStore store.MetricsStore
	frameStore   store.FrameStore
	framesDir    string
}

func NewTelemetryHandler(metricsStore store.MetricsStore, frameStore store.FrameStore, framesDir string) *TelemetryHandler {
	return &TelemetryHandler{
		metricsStore: metricsStore,
		frameStore:   frameStore,
		framesDir:    framesDir,
	}
}

func (handler *TelemetryHandler) SubmitTelemetry(context *gin.Context) {
	var payload models.TelemetryPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	recordedAt := store.NormalizeRecordedAt(payload.RecordedAt)
	batteryLevel := float64(payload.BatteryPercentage)
	tempC := payload.BatteryTemperatureC
	cabinNoise := payload.CabinNoiseRMS
	speed := payload.SpeedKmh
	ambient := payload.AmbientLightLux
	latency := payload.ServerLatencyMs
	signal := payload.NetworkSignalStrengthDbm

	var latitude *float64
	var longitude *float64
	if payload.GPSLocation != nil {
		lat := payload.GPSLocation.Latitude
		lon := payload.GPSLocation.Longitude
		latitude = &lat
		longitude = &lon
	}

	record := models.DeviceMetricsRecord{
		DeviceID:         payload.DeviceID,
		BatteryLevel:     &batteryLevel,
		SignalStrength:   &signal,
		NetworkType:      payload.NetworkType,
		TemperatureC:     &tempC,
		Latitude:         latitude,
		Longitude:        longitude,
		CabinNoiseRMS:    &cabinNoise,
		GPSSignalQuality: payload.GPSSignalQuality,
		SpeedKmh:         &speed,
		AmbientLightLux:  &ambient,
		ServerLatencyMs:  &latency,
		DeviceIPAddress:  payload.DeviceIPAddress,
		RecordedAt:       recordedAt,
		ReceivedAt:       time.Now().UTC(),
	}

	if err := handler.metricsStore.SaveMetrics(record); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save telemetry"})
		return
	}

	handler.persistOptionalFrame(payload.DeviceID, recordedAt, payload.RearCameraFrameBase64)
	handler.persistOptionalFrame(payload.DeviceID, recordedAt, payload.FrontCameraFrameBase64)

	context.Status(http.StatusNoContent)
}

func (handler *TelemetryHandler) persistOptionalFrame(deviceID string, capturedAt time.Time, imageData string) {
	if imageData == "" {
		return
	}

	framePath, err := store.PersistFrameImage(handler.framesDir, deviceID, capturedAt, imageData)
	if err != nil {
		return
	}

	_ = handler.frameStore.SaveFrame(models.FrameRecord{
		DeviceID:   deviceID,
		Path:       framePath,
		CapturedAt: capturedAt,
		ReceivedAt: time.Now().UTC(),
	})
}

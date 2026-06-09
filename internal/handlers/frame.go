package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/models"
	"github.com/parkiroid/parkiroid-server/internal/store"
)

type FrameHandler struct {
	frameStore store.FrameStore
	framesDir  string
}

func NewFrameHandler(frameStore store.FrameStore, framesDir string) *FrameHandler {
	return &FrameHandler{frameStore: frameStore, framesDir: framesDir}
}

func (handler *FrameHandler) SubmitFrame(context *gin.Context) {
	var payload models.FramePayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	capturedAt := store.NormalizeCapturedAt(payload.CapturedAt)
	framePath, err := store.PersistFrameImage(handler.framesDir, payload.DeviceID, capturedAt, payload.ImageData)
	if err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid image_data"})
		return
	}

	record := models.FrameRecord{
		DeviceID:   payload.DeviceID,
		Path:       framePath,
		CapturedAt: capturedAt,
		ReceivedAt: time.Now().UTC(),
	}

	if err := handler.frameStore.SaveFrame(record); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save frame"})
		return
	}

	context.JSON(http.StatusCreated, record)
}

func (handler *FrameHandler) GetLastFrame(context *gin.Context) {
	deviceID := context.Query("device-id")
	if deviceID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device-id query parameter is required"})
		return
	}

	frame, err := handler.frameStore.GetLastFrame(deviceID)
	if err != nil {
		if errors.Is(err, store.ErrFrameNotFound) {
			context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "no frame found for device"})
			return
		}
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to retrieve frame"})
		return
	}

	context.JSON(http.StatusOK, frame)
}

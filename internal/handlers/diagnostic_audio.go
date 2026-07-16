package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type DiagnosticAudioHandler struct {
	audioStore store.DiagnosticAudioStore
	audioDir   string
}

func NewDiagnosticAudioHandler(audioStore store.DiagnosticAudioStore, audioDir string) *DiagnosticAudioHandler {
	return &DiagnosticAudioHandler{audioStore: audioStore, audioDir: audioDir}
}

func (handler *DiagnosticAudioHandler) SubmitDiagnosticAudio(context *gin.Context) {
	metadataRaw := context.PostForm("metadata")
	if metadataRaw == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "metadata form field is required"})
		return
	}

	var metadata models.DiagnosticAudioMetadata
	if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid metadata json"})
		return
	}
	if metadata.DeviceID == "" || metadata.SegmentID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device_id and segment_id are required"})
		return
	}

	fileHeader, err := context.FormFile("audio")
	if err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "audio file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "failed to read audio file"})
		return
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "failed to read audio file"})
		return
	}

	deviceDir := filepath.Join(handler.audioDir, store.SanitizeDeviceSegment(metadata.DeviceID))
	if err := os.MkdirAll(deviceDir, 0o755); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store audio"})
		return
	}

	fileName := fmt.Sprintf("%s-%s.wav", store.SanitizeDeviceSegment(metadata.SegmentID), time.Now().UTC().Format("060102-150405"))
	audioPath := filepath.Join(deviceDir, fileName)
	if err := os.WriteFile(audioPath, contents, 0o644); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store audio"})
		return
	}

	if err := handler.audioStore.SaveDiagnosticAudio(
		metadata.DeviceID,
		metadata.SegmentID,
		audioPath,
		metadata.StartMs,
		metadata.EndMs,
		metadata.RMSPeak,
		metadata.LinkedAlertID,
		metadata.Mode,
	); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save audio metadata"})
		return
	}

	context.Status(http.StatusCreated)
}

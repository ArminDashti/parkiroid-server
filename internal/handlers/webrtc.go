package handlers

import (
	"net/http"
	"strconv"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type WebRTCHandler struct {
	webrtcStore store.WebRTCStore
}

func NewWebRTCHandler(webrtcStore store.WebRTCStore) *WebRTCHandler {
	return &WebRTCHandler{webrtcStore: webrtcStore}
}

func (handler *WebRTCHandler) ListConnections(context *gin.Context) {
	deviceID := context.Query("device-id")
	if deviceID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device-id query parameter is required"})
		return
	}

	limit := 50
	if limitRaw := context.Query("limit"); limitRaw != "" {
		parsedLimit, err := strconv.Atoi(limitRaw)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	connections, err := handler.webrtcStore.ListConnections(deviceID, limit)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list webrtc connections"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"connections": connections})
}

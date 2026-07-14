package handlers

import (
	"errors"
	"net/http"
	"time"

	livekitauth "github.com/dogan/dogan-server/internal/livekit"
	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type WebRTCSessionHandler struct {
	tokenIssuer *livekitauth.TokenIssuer
	webrtcStore store.WebRTCStore
}

func NewWebRTCSessionHandler(tokenIssuer *livekitauth.TokenIssuer, webrtcStore store.WebRTCStore) *WebRTCSessionHandler {
	return &WebRTCSessionHandler{
		tokenIssuer: tokenIssuer,
		webrtcStore: webrtcStore,
	}
}

func (handler *WebRTCSessionHandler) CreateSession(context *gin.Context) {
	if !handler.tokenIssuer.Enabled() {
		context.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "livekit is not configured"})
		return
	}

	var request models.WebRTCSessionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	tokenResponse, err := handler.tokenIssuer.IssueToken(livekitauth.TokenRequest{
		DeviceID: request.DeviceID,
		Role:     livekitauth.RolePublisher,
	})
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to issue livekit token"})
		return
	}

	_, _ = handler.webrtcStore.SaveConnection(models.WebRTCConnectionRecord{
		DeviceID:    request.DeviceID,
		Room:        tokenResponse.Room,
		Identity:    tokenResponse.Identity,
		Role:        livekitauth.RolePublisher,
		ConnectedAt: time.Now().UTC(),
		Status:      "active",
	})

	context.JSON(http.StatusOK, models.WebRTCSessionResponse{
		SessionID:  tokenResponse.Room,
		Token:      tokenResponse.Token,
		URL:        tokenResponse.URL,
		Room:       tokenResponse.Room,
		Identity:   tokenResponse.Identity,
		ExpiresAt:  tokenResponse.ExpiresAt,
		IceServers: livekitauth.DefaultIceServers(),
	})
}

type DevicesHandler struct {
	deviceStore store.DeviceStore
	tokenIssuer *livekitauth.TokenIssuer
	webrtcStore store.WebRTCStore
}

func NewDevicesHandler(deviceStore store.DeviceStore, tokenIssuer *livekitauth.TokenIssuer, webrtcStore store.WebRTCStore) *DevicesHandler {
	return &DevicesHandler{
		deviceStore: deviceStore,
		tokenIssuer: tokenIssuer,
		webrtcStore: webrtcStore,
	}
}

func (handler *DevicesHandler) ListDevices(context *gin.Context) {
	devices, err := handler.deviceStore.ListDevices()
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list devices"})
		return
	}

	context.JSON(http.StatusOK, devices)
}

func (handler *DevicesHandler) GetDeviceStream(context *gin.Context) {
	if !handler.tokenIssuer.Enabled() {
		context.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "livekit is not configured"})
		return
	}

	deviceID := context.Param("id")
	if deviceID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "device id is required"})
		return
	}

	tokenResponse, err := handler.tokenIssuer.IssueToken(livekitauth.TokenRequest{
		DeviceID: deviceID,
		Role:     livekitauth.RoleSubscriber,
	})
	if err != nil {
		switch {
		case errors.Is(err, livekitauth.ErrInvalidRole):
			context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		default:
			context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to issue livekit token"})
		}
		return
	}

	_, _ = handler.webrtcStore.SaveConnection(models.WebRTCConnectionRecord{
		DeviceID:    deviceID,
		Room:        tokenResponse.Room,
		Identity:    tokenResponse.Identity,
		Role:        livekitauth.RoleSubscriber,
		ConnectedAt: time.Now().UTC(),
		Status:      "active",
	})

	context.JSON(http.StatusOK, models.DeviceStreamResponse{
		DeviceID:  deviceID,
		Token:     tokenResponse.Token,
		URL:       tokenResponse.URL,
		Room:      tokenResponse.Room,
		Identity:  tokenResponse.Identity,
		ExpiresAt: tokenResponse.ExpiresAt,
	})
}

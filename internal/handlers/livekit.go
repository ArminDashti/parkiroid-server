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

type LiveKitHandler struct {
	tokenIssuer *livekitauth.TokenIssuer
	webrtcStore store.WebRTCStore
}

func NewLiveKitHandler(tokenIssuer *livekitauth.TokenIssuer, webrtcStore store.WebRTCStore) *LiveKitHandler {
	return &LiveKitHandler{
		tokenIssuer: tokenIssuer,
		webrtcStore: webrtcStore,
	}
}

func (handler *LiveKitHandler) IssueToken(context *gin.Context) {
	if !handler.tokenIssuer.Enabled() {
		context.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "livekit is not configured"})
		return
	}

	var request models.LiveKitTokenRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	tokenResponse, err := handler.tokenIssuer.IssueToken(livekitauth.TokenRequest{
		DeviceID: request.DeviceID,
		Identity: request.Identity,
		Role:     request.Role,
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
		DeviceID:    request.DeviceID,
		Room:        tokenResponse.Room,
		Identity:    tokenResponse.Identity,
		Role:        defaultRole(request.Role),
		ConnectedAt: time.Now().UTC(),
		Status:      "active",
	})

	context.JSON(http.StatusOK, models.LiveKitTokenResponse{
		Token:     tokenResponse.Token,
		URL:       tokenResponse.URL,
		Room:      tokenResponse.Room,
		Identity:  tokenResponse.Identity,
		ExpiresAt: tokenResponse.ExpiresAt,
	})
}

func defaultRole(role string) string {
	if role == "" {
		return livekitauth.RoleSubscriber
	}
	return role
}

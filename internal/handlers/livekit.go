package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	livekitauth "github.com/parkiroid/parkiroid-server/internal/livekit"
	"github.com/parkiroid/parkiroid-server/internal/models"
)

type LiveKitHandler struct {
	tokenIssuer *livekitauth.TokenIssuer
}

func NewLiveKitHandler(tokenIssuer *livekitauth.TokenIssuer) *LiveKitHandler {
	return &LiveKitHandler{tokenIssuer: tokenIssuer}
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

	context.JSON(http.StatusOK, models.LiveKitTokenResponse{
		Token:     tokenResponse.Token,
		URL:       tokenResponse.URL,
		Room:      tokenResponse.Room,
		Identity:  tokenResponse.Identity,
		ExpiresAt: tokenResponse.ExpiresAt,
	})
}

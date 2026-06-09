package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/auth"
	"github.com/parkiroid/parkiroid-server/internal/models"
)

type AuthHandler struct {
	apiKey       string
	tokenIssuer  *auth.TokenIssuer
}

func NewAuthHandler(apiKey string, tokenIssuer *auth.TokenIssuer) *AuthHandler {
	return &AuthHandler{
		apiKey:      apiKey,
		tokenIssuer: tokenIssuer,
	}
}

func (handler *AuthHandler) Authenticate(context *gin.Context) {
	var request models.AuthRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if request.APIKey != handler.apiKey {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid api key"})
		return
	}

	token, expiresAt, err := handler.tokenIssuer.IssueToken("parkiroid-client")
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to issue token"})
		return
	}

	context.JSON(http.StatusOK, models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

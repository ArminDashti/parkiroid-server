package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/auth"
	"github.com/parkiroid/parkiroid-server/internal/models"
)

type AuthHandler struct {
	tokenIssuer *auth.TokenIssuer
}

func NewAuthHandler(tokenIssuer *auth.TokenIssuer) *AuthHandler {
	return &AuthHandler{
		tokenIssuer: tokenIssuer,
	}
}

func (handler *AuthHandler) Authenticate(context *gin.Context) {
	var request models.AuthRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if !auth.VerifyAdminCredentials(request.Username, request.Password) {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid username or password"})
		return
	}

	token, expiresAt, err := handler.tokenIssuer.IssueToken(auth.AdminUsername)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to issue token"})
		return
	}

	context.JSON(http.StatusOK, models.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

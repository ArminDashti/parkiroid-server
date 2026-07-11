package handlers

import (
	"net/http"

	"github.com/dogan/dogan-server/internal/auth"
	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tokenIssuer   *auth.TokenIssuer
	loginLogStore store.LoginLogStore
}

func NewAuthHandler(tokenIssuer *auth.TokenIssuer, loginLogStore store.LoginLogStore) *AuthHandler {
	return &AuthHandler{
		tokenIssuer:   tokenIssuer,
		loginLogStore: loginLogStore,
	}
}

func (handler *AuthHandler) Authenticate(context *gin.Context) {
	var request models.AuthRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	clientIP := context.ClientIP()
	browserInfo := context.GetHeader("User-Agent")
	success := auth.VerifyAdminCredentials(request.Username, request.Password)

	if err := handler.loginLogStore.SaveLoginLog(clientIP, request.Username, browserInfo, success); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save login log"})
		return
	}

	if !success {
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

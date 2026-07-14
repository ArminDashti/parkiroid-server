package handlers

import (
	"net/http"
	"time"

	"github.com/dogan/dogan-server/internal/auth"
	"github.com/dogan/dogan-server/internal/middleware"
	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tokenIssuer      *auth.TokenIssuer
	loginLogStore    store.LoginLogStore
	embeddedAPIToken string
	deviceAPIKey     string
	tokenTTL         time.Duration
}

func NewAuthHandler(
	tokenIssuer *auth.TokenIssuer,
	loginLogStore store.LoginLogStore,
	embeddedAPIToken string,
	deviceAPIKey string,
	tokenTTL time.Duration,
) *AuthHandler {
	return &AuthHandler{
		tokenIssuer:      tokenIssuer,
		loginLogStore:    loginLogStore,
		embeddedAPIToken: embeddedAPIToken,
		deviceAPIKey:     deviceAPIKey,
		tokenTTL:         tokenTTL,
	}
}

func (handler *AuthHandler) Authenticate(context *gin.Context) {
	var request models.AuthRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if request.APIKey != "" {
		handler.authenticateDevice(context, request.APIKey)
		return
	}

	if request.Username == "" || request.Password == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "username and password or api_key required"})
		return
	}

	handler.authenticateAdmin(context, request.Username, request.Password)
}

func (handler *AuthHandler) authenticateDevice(context *gin.Context, apiKey string) {
	if !auth.ValidateDeviceAPIKey(apiKey, handler.embeddedAPIToken, handler.deviceAPIKey) {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid api key"})
		return
	}

	expiresAt := auth.DeviceTokenExpiry(handler.tokenTTL)
	context.JSON(http.StatusOK, models.AuthResponse{
		Token:     handler.embeddedAPIToken,
		ExpiresAt: expiresAt,
	})
}

func (handler *AuthHandler) authenticateAdmin(context *gin.Context, username, password string) {
	clientIP := context.ClientIP()
	browserInfo := context.GetHeader("User-Agent")
	success := auth.VerifyAdminCredentials(username, password)

	if err := handler.loginLogStore.SaveLoginLog(clientIP, username, browserInfo, success); err != nil {
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

func (handler *AuthHandler) WebLogin(context *gin.Context) {
	var request models.WebLoginRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	loginName := auth.LoginIdentifier(request.Email)
	handler.authenticateAdminForWeb(context, loginName, request.Email, request.Password)
}

func (handler *AuthHandler) authenticateAdminForWeb(context *gin.Context, username, loginLabel, password string) {
	clientIP := context.ClientIP()
	browserInfo := context.GetHeader("User-Agent")
	success := auth.VerifyAdminCredentials(username, password)

	if err := handler.loginLogStore.SaveLoginLog(clientIP, loginLabel, browserInfo, success); err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save login log"})
		return
	}

	if !success {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid email or password"})
		return
	}

	token, _, err := handler.tokenIssuer.IssueToken(auth.AdminUsername)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to issue token"})
		return
	}

	context.JSON(http.StatusOK, models.WebAuthResponse{
		Token: token,
		User: models.WebUser{
			ID:    auth.AdminUsername,
			Email: loginLabel,
			Name:  auth.AdminDisplayName,
		},
	})
}

func (handler *AuthHandler) CurrentUser(context *gin.Context) {
	subject, ok := context.Get(middleware.AuthSubjectKey)
	if !ok {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid or expired token"})
		return
	}

	username, ok := subject.(string)
	if !ok || username == "" {
		context.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid or expired token"})
		return
	}

	if username == "device" {
		context.JSON(http.StatusOK, models.WebUser{
			ID:    "device",
			Email: "device@dogan.local",
			Name:  "Device",
		})
		return
	}

	context.JSON(http.StatusOK, models.WebUser{
		ID:    username,
		Email: username + "@dogan.local",
		Name:  auth.AdminDisplayName,
	})
}

func (handler *AuthHandler) Logout(context *gin.Context) {
	context.Status(http.StatusNoContent)
}

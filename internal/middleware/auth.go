package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/dogan/dogan-server/internal/auth"
	"github.com/dogan/dogan-server/internal/models"
)

func RequireBearerToken(tokenIssuer *auth.TokenIssuer, embeddedAPIToken string) gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationHeader := context.GetHeader("Authorization")
		if authorizationHeader == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			context.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid authorization header format",
			})
			return
		}

		if err := auth.ValidateBearerToken(parts[1], tokenIssuer, embeddedAPIToken); err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
			})
			return
		}

		context.Next()
	}
}

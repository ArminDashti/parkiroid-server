package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/dogan/dogan-server/internal/auth"
	"github.com/dogan/dogan-server/internal/models"
	"github.com/gin-gonic/gin"
)

const AuthSubjectKey = "auth_subject"

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

		tokenString := parts[1]
		if embeddedAPIToken != "" && subtle.ConstantTimeCompare([]byte(tokenString), []byte(embeddedAPIToken)) == 1 {
			context.Set(AuthSubjectKey, "device")
			context.Next()
			return
		}

		claims, err := tokenIssuer.ParseToken(tokenString)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
			})
			return
		}

		subject := claims.Subject
		if subject == "" {
			subject = claims.RegisteredClaims.Subject
		}
		context.Set(AuthSubjectKey, subject)
		context.Next()
	}
}

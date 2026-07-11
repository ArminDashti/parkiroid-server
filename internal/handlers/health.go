package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dogan/dogan-server/internal/models"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (handler *HealthHandler) GetHealth(context *gin.Context) {
	context.JSON(http.StatusOK, models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

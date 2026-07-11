package handlers

import (
	"net/http"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type AIModelsHandler struct {
	aiModelStore store.AIModelStore
}

func NewAIModelsHandler(aiModelStore store.AIModelStore) *AIModelsHandler {
	return &AIModelsHandler{aiModelStore: aiModelStore}
}

func (handler *AIModelsHandler) ListAIModels(context *gin.Context) {
	modelsList, err := handler.aiModelStore.ListAIModels()
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list ai models"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"models": modelsList})
}

func (handler *AIModelsHandler) UpsertAIModel(context *gin.Context) {
	var payload models.AIModelPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	model := models.AIModelRecord{
		ModelName: payload.ModelName,
		Path:      payload.Path,
		Version:   payload.Version,
		UpdatedAt: time.Now().UTC(),
	}

	savedModel, err := handler.aiModelStore.UpsertAIModel(model)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save ai model"})
		return
	}

	context.JSON(http.StatusOK, savedModel)
}

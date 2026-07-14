package handlers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

type AIModelsHandler struct {
	aiModelStore store.AIModelStore
	modelsDir    string
}

func NewAIModelsHandler(aiModelStore store.AIModelStore, modelsDir string) *AIModelsHandler {
	return &AIModelsHandler{aiModelStore: aiModelStore, modelsDir: modelsDir}
}

func (handler *AIModelsHandler) ListAIModels(context *gin.Context) {
	modelsList, err := handler.aiModelStore.ListAIModels()
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list ai models"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"models": modelsList})
}

func (handler *AIModelsHandler) ListModelsManifest(context *gin.Context) {
	modelsList, err := handler.aiModelStore.ListAIModels()
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list ai models"})
		return
	}

	baseURL := requestBaseURL(context)
	manifest := make([]models.AIModelManifestEntry, 0, len(modelsList))
	for _, model := range modelsList {
		if !store.HasModelFiles(handler.modelsDir, model.ModelName) {
			continue
		}

		entry := models.AIModelManifestEntry{
			ID:          model.ModelName,
			ParamURL:    baseURL + apiBasePath + "/models/" + model.ModelName + "/param",
			BinURL:      baseURL + apiBasePath + "/models/" + model.ModelName + "/bin",
			ParamSHA256: model.ParamSHA256,
			BinSHA256:   model.BinSHA256,
			Format:      model.Format,
			Labels:      model.Labels,
		}
		if entry.Format == "" {
			entry.Format = "ncnn"
		}
		if entry.Labels == nil {
			entry.Labels = []string{}
		}
		manifest = append(manifest, entry)
	}

	context.JSON(http.StatusOK, gin.H{"models": manifest})
}

func (handler *AIModelsHandler) GetModelParam(context *gin.Context) {
	handler.serveModelFile(context, store.ModelParamFileName)
}

func (handler *AIModelsHandler) GetModelBin(context *gin.Context) {
	handler.serveModelFile(context, store.ModelBinFileName)
}

func (handler *AIModelsHandler) serveModelFile(context *gin.Context, fileName string) {
	modelID := strings.TrimSpace(context.Param("id"))
	if modelID == "" {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "model id is required"})
		return
	}

	_, err := handler.aiModelStore.GetAIModelByName(modelID)
	if err != nil {
		if errors.Is(err, store.ErrAIModelNotFound) {
			context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "model not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to load model"})
		return
	}

	var filePath string
	switch fileName {
	case store.ModelParamFileName:
		filePath = store.ModelParamPath(handler.modelsDir, modelID)
	case store.ModelBinFileName:
		filePath = store.ModelBinPath(handler.modelsDir, modelID)
	default:
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "invalid model file"})
		return
	}

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			context.JSON(http.StatusNotFound, models.ErrorResponse{Error: "model file not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read model file"})
		return
	}

	context.File(filePath)
}

func (handler *AIModelsHandler) UpsertAIModel(context *gin.Context) {
	var payload models.AIModelPayload
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	paramSHA256 := strings.TrimSpace(payload.ParamSHA256)
	binSHA256 := strings.TrimSpace(payload.BinSHA256)
	if paramSHA256 == "" || binSHA256 == "" {
		if !store.HasModelFiles(handler.modelsDir, payload.ModelName) {
			context.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "model files not found on server"})
			return
		}
		if paramSHA256 == "" {
			computed, err := store.ComputeFileSHA256(store.ModelParamPath(handler.modelsDir, payload.ModelName))
			if err != nil {
				context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash model param file"})
				return
			}
			paramSHA256 = computed
		}
		if binSHA256 == "" {
			computed, err := store.ComputeFileSHA256(store.ModelBinPath(handler.modelsDir, payload.ModelName))
			if err != nil {
				context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash model bin file"})
				return
			}
			binSHA256 = computed
		}
	}

	format := strings.TrimSpace(payload.Format)
	if format == "" {
		format = "ncnn"
	}

	labels := payload.Labels
	if labels == nil {
		labels = []string{}
	}

	model := models.AIModelRecord{
		ModelName:   payload.ModelName,
		ParamSHA256: paramSHA256,
		BinSHA256:   binSHA256,
		Labels:      labels,
		Format:      format,
		Version:     payload.Version,
		UpdatedAt:   time.Now().UTC(),
	}

	savedModel, err := handler.aiModelStore.UpsertAIModel(model)
	if err != nil {
		context.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save ai model"})
		return
	}

	context.JSON(http.StatusOK, savedModel)
}

func requestBaseURL(context *gin.Context) string {
	scheme := context.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if context.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := context.Request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = context.Request.Host
	}

	return scheme + "://" + host
}

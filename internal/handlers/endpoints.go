package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/models"
)

const apiBasePath = "/parkiroid/api/v1"

type EndpointsHandler struct{}

func NewEndpointsHandler() *EndpointsHandler {
	return &EndpointsHandler{}
}

func (handler *EndpointsHandler) ListEndpoints(context *gin.Context) {
	endpoints := []models.EndpointDescriptor{
		{
			Method:      http.MethodPost,
			Path:        apiBasePath + "/auth",
			Description: "Exchange API key for a bearer token",
			Auth:        false,
		},
		{
			Method:      http.MethodGet,
			Path:        apiBasePath + "/endpoints",
			Description: "List available API endpoints",
			Auth:        false,
		},
		{
			Method:      http.MethodGet,
			Path:        apiBasePath + "/health",
			Description: "Service health check",
			Auth:        false,
		},
		{
			Method:      http.MethodGet,
			Path:        apiBasePath + "/last-frame",
			Description: "Retrieve the most recent frame for a device",
			Auth:        true,
		},
		{
			Method:      http.MethodPost,
			Path:        apiBasePath + "/frame",
			Description: "Submit a camera frame from a device",
			Auth:        true,
		},
		{
			Method:      http.MethodGet,
			Path:        apiBasePath + "/device-metrics",
			Description: "Retrieve the latest metrics for a device",
			Auth:        true,
		},
		{
			Method:      http.MethodPost,
			Path:        apiBasePath + "/device-metrics",
			Description: "Submit device telemetry metrics",
			Auth:        true,
		},
	}

	context.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

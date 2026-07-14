package handlers

import (
	"net/http"

	"github.com/dogan/dogan-server/internal/models"
	"github.com/gin-gonic/gin"
)

const apiBasePath = "/dogan/api/v1"

type EndpointsHandler struct{}

func NewEndpointsHandler() *EndpointsHandler {
	return &EndpointsHandler{}
}

func (handler *EndpointsHandler) ListEndpoints(context *gin.Context) {
	endpoints := []models.EndpointDescriptor{
		{Method: http.MethodPost, Path: apiBasePath + "/auth", Description: "Login and get JWT token", Auth: false},
		{Method: http.MethodPost, Path: apiBasePath + "/auth/login", Description: "Web login and get JWT token with user", Auth: false},
		{Method: http.MethodGet, Path: apiBasePath + "/auth/me", Description: "Get current authenticated user", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/auth/logout", Description: "Log out current session", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/endpoints", Description: "List available API endpoints", Auth: false},
		{Method: http.MethodGet, Path: apiBasePath + "/health", Description: "Health check", Auth: false},
		{Method: http.MethodPost, Path: apiBasePath + "/frame", Description: "Upload camera frame from Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/last-frame", Description: "Get latest frame metadata", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/frame/image", Description: "Download latest frame JPEG", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/device-metrics", Description: "Submit Android telemetry", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/device-metrics", Description: "Get latest Android telemetry", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/actions", Description: "Queue action for Android phone", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/actions/pending", Description: "Poll pending actions for device", Auth: true},
		{Method: http.MethodPut, Path: apiBasePath + "/actions/:id/ack", Description: "Acknowledge action completion", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/settings", Description: "Get app settings by platform", Auth: true},
		{Method: http.MethodPut, Path: apiBasePath + "/settings", Description: "Upsert app setting", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/ai-models", Description: "List registered AI models (admin metadata)", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/ai-models", Description: "Register or update NCNN model metadata", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models", Description: "List downloadable NCNN models for Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models/:id/param", Description: "Download NCNN model.param file", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models/:id/bin", Description: "Download NCNN model.bin file", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/webrtc/connections", Description: "List recent WebRTC sessions", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/streaming/token", Description: "Issue LiveKit WebRTC token", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/webrtc/session", Description: "Create LiveKit publisher session for Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices", Description: "List registered devices", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices/:id/stream", Description: "Get LiveKit subscriber credentials for device stream", Auth: true},
	}

	context.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

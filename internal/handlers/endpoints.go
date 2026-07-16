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
		{Method: http.MethodPost, Path: apiBasePath + "/auth", Description: "Login and get JWT token (Android api_key or admin username/password)", Auth: false},
		{Method: http.MethodPost, Path: apiBasePath + "/auth/login", Description: "Web login and get JWT token with user", Auth: false},
		{Method: http.MethodGet, Path: apiBasePath + "/auth/me", Description: "Get current authenticated user", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/auth/logout", Description: "Log out current session", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/endpoints", Description: "List available API endpoints", Auth: false},
		{Method: http.MethodGet, Path: apiBasePath + "/health", Description: "Health check", Auth: false},
		{Method: http.MethodPost, Path: apiBasePath + "/telemetry", Description: "Submit Android unified telemetry (metrics + optional frames)", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/frame", Description: "Upload camera frame from Android (legacy)", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/last-frame", Description: "Get latest frame metadata", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/frame/image", Description: "Download latest frame JPEG", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/device-metrics", Description: "Submit Android telemetry (legacy)", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/device-metrics", Description: "Get latest Android telemetry (legacy)", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/actions", Description: "Queue action for Android phone", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/actions/pending", Description: "Poll pending actions for device", Auth: true},
		{Method: http.MethodPut, Path: apiBasePath + "/actions/:id/ack", Description: "Acknowledge action completion", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/settings", Description: "Get settings (Android flat map via device_id, web prefs by default)", Auth: true},
		{Method: http.MethodPut, Path: apiBasePath + "/settings", Description: "Upsert single Android/app setting key", Auth: true},
		{Method: http.MethodPatch, Path: apiBasePath + "/settings", Description: "Patch web preferences", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/ai-models", Description: "List registered AI models (admin metadata)", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/ai-models", Description: "Register or update NCNN model metadata", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models", Description: "List downloadable NCNN models for Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models/:id/param", Description: "Download NCNN model.param file", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/models/:id/bin", Description: "Download NCNN model.bin file", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/sounds", Description: "List downloadable alert sounds for Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/sounds/:id", Description: "Download alert sound file", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/diagnostic-audio", Description: "Upload cabin diagnostic WAV audio", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/webrtc/connections", Description: "List recent WebRTC sessions", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/streaming/token", Description: "Issue LiveKit WebRTC token", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/webrtc/session", Description: "Create LiveKit publisher session for Android", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices", Description: "List registered devices", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices/:id/stream", Description: "Get LiveKit subscriber credentials for device stream", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices/:id/telemetry", Description: "Get live device telemetry snapshot for web", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/devices/:id/metrics", Description: "Get device metrics history for web charts", Auth: true},
		{Method: http.MethodPost, Path: apiBasePath + "/devices/:id/capture", Description: "Queue capture and return latest frame", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/images", Description: "List gallery images", Auth: true},
		{Method: http.MethodGet, Path: apiBasePath + "/images/:id", Description: "Download gallery image by id", Auth: true},
	}

	context.JSON(http.StatusOK, gin.H{"endpoints": endpoints})
}

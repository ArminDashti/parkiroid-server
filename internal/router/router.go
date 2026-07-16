package router

import (
	"log"

	"github.com/dogan/dogan-server/internal/auth"
	"github.com/dogan/dogan-server/internal/config"
	"github.com/dogan/dogan-server/internal/handlers"
	livekitauth "github.com/dogan/dogan-server/internal/livekit"
	"github.com/dogan/dogan-server/internal/middleware"
	"github.com/dogan/dogan-server/internal/store"
	"github.com/gin-gonic/gin"
)

func New(applicationConfig config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	postgresStore, err := store.OpenPostgres(applicationConfig.DatabaseURL, applicationConfig.RetentionPeriod)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	retentionCleaner := store.NewRetentionCleaner(
		postgresStore,
		applicationConfig.FramesDir,
		applicationConfig.RetentionPeriod,
	)
	retentionCleaner.Start()

	tokenIssuer := auth.NewTokenIssuer(applicationConfig.JWTSecret, applicationConfig.TokenTTL)

	authHandler := handlers.NewAuthHandler(
		tokenIssuer,
		postgresStore,
		applicationConfig.EmbeddedAPIToken,
		applicationConfig.DeviceAPIKey,
		applicationConfig.TokenTTL,
	)
	healthHandler := handlers.NewHealthHandler()
	endpointsHandler := handlers.NewEndpointsHandler()
	frameHandler := handlers.NewFrameHandler(postgresStore, applicationConfig.FramesDir)
	deviceMetricsHandler := handlers.NewDeviceMetricsHandler(postgresStore)
	telemetryHandler := handlers.NewTelemetryHandler(postgresStore, postgresStore, applicationConfig.FramesDir)
	actionsHandler := handlers.NewActionsHandler(postgresStore)
	settingsHandler := handlers.NewSettingsHandler(postgresStore)
	aiModelsHandler := handlers.NewAIModelsHandler(postgresStore, applicationConfig.ModelsDir)
	soundsHandler := handlers.NewSoundsHandler(applicationConfig.SoundsDir)
	diagnosticAudioHandler := handlers.NewDiagnosticAudioHandler(postgresStore, applicationConfig.AudioDir)
	webrtcHandler := handlers.NewWebRTCHandler(postgresStore)
	liveKitTokenIssuer := livekitauth.NewTokenIssuer(livekitauth.Config{
		URL:       applicationConfig.LiveKitURL,
		PublicURL: applicationConfig.LiveKitPublicURL,
		APIKey:    applicationConfig.LiveKitAPIKey,
		APISecret: applicationConfig.LiveKitAPISecret,
		TokenTTL:  applicationConfig.LiveKitTokenTTL,
	})
	liveKitHandler := handlers.NewLiveKitHandler(liveKitTokenIssuer, postgresStore)
	webrtcSessionHandler := handlers.NewWebRTCSessionHandler(liveKitTokenIssuer, postgresStore)
	devicesHandler := handlers.NewDevicesHandler(postgresStore, liveKitTokenIssuer, postgresStore)
	devicesWebHandler := handlers.NewDevicesWebHandler(postgresStore, postgresStore, postgresStore, postgresStore)

	api := engine.Group("/dogan/api/v1")
	{
		api.POST("/auth", authHandler.Authenticate)
		api.POST("/auth/login", authHandler.WebLogin)
		api.GET("/endpoints", endpointsHandler.ListEndpoints)
		api.GET("/health", healthHandler.GetHealth)

		protected := api.Group("/")
		protected.Use(middleware.RequireBearerToken(tokenIssuer, applicationConfig.EmbeddedAPIToken))
		{
			protected.GET("/auth/me", authHandler.CurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/last-frame", frameHandler.GetLastFrame)
			protected.GET("/frame/image", frameHandler.GetFrameImage)
			protected.POST("/frame", frameHandler.SubmitFrame)
			protected.POST("/telemetry", telemetryHandler.SubmitTelemetry)
			protected.GET("/device-metrics", deviceMetricsHandler.GetLatestMetrics)
			protected.POST("/device-metrics", deviceMetricsHandler.SubmitMetrics)
			protected.POST("/actions", actionsHandler.CreateAction)
			protected.GET("/actions/pending", actionsHandler.GetPendingActions)
			protected.PUT("/actions/:id/ack", actionsHandler.AcknowledgeAction)
			protected.GET("/settings", settingsHandler.GetSettings)
			protected.PUT("/settings", settingsHandler.UpsertSetting)
			protected.PATCH("/settings", settingsHandler.PatchWebSettings)
			protected.GET("/ai-models", aiModelsHandler.ListAIModels)
			protected.POST("/ai-models", aiModelsHandler.UpsertAIModel)
			protected.GET("/models", aiModelsHandler.ListModelsManifest)
			protected.GET("/models/:id/param", aiModelsHandler.GetModelParam)
			protected.GET("/models/:id/bin", aiModelsHandler.GetModelBin)
			protected.GET("/sounds", soundsHandler.ListSounds)
			protected.GET("/sounds/:id", soundsHandler.GetSound)
			protected.POST("/diagnostic-audio", diagnosticAudioHandler.SubmitDiagnosticAudio)
			protected.GET("/webrtc/connections", webrtcHandler.ListConnections)
			protected.POST("/streaming/token", liveKitHandler.IssueToken)
			protected.POST("/webrtc/session", webrtcSessionHandler.CreateSession)
			protected.GET("/devices", devicesHandler.ListDevices)
			protected.GET("/devices/:id/stream", devicesHandler.GetDeviceStream)
			protected.GET("/devices/:id/telemetry", devicesWebHandler.GetDeviceTelemetry)
			protected.GET("/devices/:id/metrics", devicesWebHandler.GetDeviceMetrics)
			protected.POST("/devices/:id/capture", devicesWebHandler.CaptureDeviceFrame)
			protected.GET("/images", devicesWebHandler.ListImages)
			protected.GET("/images/:id", devicesWebHandler.GetImage)
		}
	}

	return engine
}

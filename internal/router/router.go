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
	actionsHandler := handlers.NewActionsHandler(postgresStore)
	settingsHandler := handlers.NewSettingsHandler(postgresStore)
	aiModelsHandler := handlers.NewAIModelsHandler(postgresStore, applicationConfig.ModelsDir)
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
			protected.GET("/device-metrics", deviceMetricsHandler.GetLatestMetrics)
			protected.POST("/device-metrics", deviceMetricsHandler.SubmitMetrics)
			protected.POST("/actions", actionsHandler.CreateAction)
			protected.GET("/actions/pending", actionsHandler.GetPendingActions)
			protected.PUT("/actions/:id/ack", actionsHandler.AcknowledgeAction)
			protected.GET("/settings", settingsHandler.GetSettings)
			protected.PUT("/settings", settingsHandler.UpsertSetting)
			protected.GET("/ai-models", aiModelsHandler.ListAIModels)
			protected.POST("/ai-models", aiModelsHandler.UpsertAIModel)
			protected.GET("/models", aiModelsHandler.ListModelsManifest)
			protected.GET("/models/:id/param", aiModelsHandler.GetModelParam)
			protected.GET("/models/:id/bin", aiModelsHandler.GetModelBin)
			protected.GET("/webrtc/connections", webrtcHandler.ListConnections)
			protected.POST("/streaming/token", liveKitHandler.IssueToken)
			protected.POST("/webrtc/session", webrtcSessionHandler.CreateSession)
			protected.GET("/devices", devicesHandler.ListDevices)
			protected.GET("/devices/:id/stream", devicesHandler.GetDeviceStream)
		}
	}

	return engine
}

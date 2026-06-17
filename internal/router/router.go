package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/parkiroid/parkiroid-server/internal/auth"
	"github.com/parkiroid/parkiroid-server/internal/config"
	"github.com/parkiroid/parkiroid-server/internal/handlers"
	livekitauth "github.com/parkiroid/parkiroid-server/internal/livekit"
	"github.com/parkiroid/parkiroid-server/internal/middleware"
	"github.com/parkiroid/parkiroid-server/internal/store"
)

func New(applicationConfig config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	sqliteStore, err := store.OpenSQLite(applicationConfig.DatabasePath, applicationConfig.RetentionPeriod)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	retentionCleaner := store.NewRetentionCleaner(
		sqliteStore,
		applicationConfig.FramesDir,
		applicationConfig.RetentionPeriod,
	)
	retentionCleaner.Start()

	tokenIssuer := auth.NewTokenIssuer(applicationConfig.JWTSecret, applicationConfig.TokenTTL)

	authHandler := handlers.NewAuthHandler(tokenIssuer)
	healthHandler := handlers.NewHealthHandler()
	endpointsHandler := handlers.NewEndpointsHandler()
	frameHandler := handlers.NewFrameHandler(sqliteStore, applicationConfig.FramesDir)
	deviceMetricsHandler := handlers.NewDeviceMetricsHandler(sqliteStore)
	liveKitTokenIssuer := livekitauth.NewTokenIssuer(livekitauth.Config{
		URL:       applicationConfig.LiveKitURL,
		APIKey:    applicationConfig.LiveKitAPIKey,
		APISecret: applicationConfig.LiveKitAPISecret,
		TokenTTL:  applicationConfig.LiveKitTokenTTL,
	})
	liveKitHandler := handlers.NewLiveKitHandler(liveKitTokenIssuer)

	api := engine.Group("/parkiroid/api/v1")
	{
		api.POST("/auth", authHandler.Authenticate)
		api.GET("/endpoints", endpointsHandler.ListEndpoints)
		api.GET("/health", healthHandler.GetHealth)

		protected := api.Group("/")
		protected.Use(middleware.RequireBearerToken(tokenIssuer, applicationConfig.EmbeddedAPIToken))
		{
			protected.GET("/last-frame", frameHandler.GetLastFrame)
			protected.POST("/frame", frameHandler.SubmitFrame)
			protected.GET("/device-metrics", deviceMetricsHandler.GetLatestMetrics)
			protected.POST("/device-metrics", deviceMetricsHandler.SubmitMetrics)
			protected.POST("/streaming/token", liveKitHandler.IssueToken)
		}
	}

	return engine
}

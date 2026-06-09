package main

import (
	"log"

	"github.com/parkiroid/parkiroid-server/internal/config"
	"github.com/parkiroid/parkiroid-server/internal/router"
)

func main() {
	applicationConfig := config.Load()
	engine := router.New(applicationConfig)

	log.Printf("parkiroid-server listening on %s", applicationConfig.ListenAddress)
	if err := engine.Run(applicationConfig.ListenAddress); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

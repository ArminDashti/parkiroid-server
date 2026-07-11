package main

import (
	"log"

	"github.com/dogan/dogan-server/internal/config"
	"github.com/dogan/dogan-server/internal/router"
)

func main() {
	applicationConfig := config.Load()
	engine := router.New(applicationConfig)

	log.Printf("dogan-server listening on %s", applicationConfig.ListenAddress)
	if err := engine.Run(applicationConfig.ListenAddress); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

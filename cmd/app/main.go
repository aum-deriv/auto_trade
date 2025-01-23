package main

import (
	"log"

	"github.com/aumbhatt/auto_trade/internal/config"
	"github.com/aumbhatt/auto_trade/internal/service"
)

func main() {
	log.Println("Starting application...")

	// Load configuration
	cfg := config.NewDefaultConfig()

	// Initialize service
	svc := service.NewService(cfg)

	// Run the service
	if err := svc.Run(); err != nil {
		log.Fatalf("Service error: %v", err)
	}

	log.Println("Application started successfully")
}

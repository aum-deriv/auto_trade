package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aumbhatt/auto_trade/internal/config"
	"github.com/aumbhatt/auto_trade/internal/service"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

func main() {
	log.Println("Starting application...")

	// Load configuration
	cfg := config.NewDefaultConfig()

	// Create and start WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize service with hub
	svc := service.NewService(cfg, hub)

	// Run the service
	go func() {
		if err := svc.Run(); err != nil {
			log.Fatalf("Service error: %v", err)
		}
	}()

	// Set up WebSocket route
	http.HandleFunc("/ws", websocket.HandleWebSocket(hub))

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

package main

import (
	"auto_trade/handlers"
	"auto_trade/websocket"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Initialize WebSocket hubs
	chatHub := websocket.NewHub()
	go chatHub.Run()

	ticksHub := websocket.NewTicksHub()
	go ticksHub.Run()

	// Configure routes
	http.HandleFunc("/health", handlers.HealthCheck)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(chatHub, w, r)
	})
	http.HandleFunc("/ws/ticks", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeTicksWs(ticksHub, w, r)
	})
	
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"auto_trade/handlers"
	"auto_trade/trading"
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

	// Initialize trade manager and handlers
	tradeManager := trading.NewTradeManager()
	tradeHandler := handlers.NewTradeHandler()
	tradeHandler.TradeManager = tradeManager

	// Initialize trade status hubs
	tradeStatusHub := websocket.NewTradeStatusHub(tradeManager)
	go tradeStatusHub.Run()
	go tradeStatusHub.BroadcastOpenTrades()

	singleTradeHub := websocket.NewSingleTradeHub(tradeManager)
	go singleTradeHub.Run()

	// Configure routes
	http.HandleFunc("/health", handlers.HealthCheck)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(chatHub, w, r)
	})
	http.HandleFunc("/ws/ticks", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeTicksWs(ticksHub, w, r)
	})
	http.HandleFunc("/api/trade/buy", tradeHandler.Buy)
	http.HandleFunc("/api/trade/sell", tradeHandler.Sell)
	http.HandleFunc("/ws/trades", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeAllTradesWs(tradeStatusHub, w, r)
	})
	http.HandleFunc("/ws/trade/", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeSingleTradeWs(singleTradeHub, w, r)
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

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

	// Initialize trade manager and handlers
	tradeManager := trading.NewTradeManager()
	tradeHandler := handlers.NewTradeHandler()
	tradeHandler.TradeManager = tradeManager
	
	// Initialize strategy handler and ticks handler
	strategyExecutor := handlers.InitStrategyHandler(tradeManager)
	strategyTicksHandler := websocket.NewStrategyTicksHandler(strategyExecutor)
	
	// Initialize ticks hub with strategy handler
	ticksHub := websocket.NewTicksHub(strategyTicksHandler)
	go ticksHub.Run()

	// Initialize strategy status hubs
	strategyStatusHub := websocket.NewStrategyStatusHub(strategyExecutor)
	go strategyStatusHub.Run()
	go strategyStatusHub.BroadcastActiveStrategies()

	singleStrategyHub := websocket.NewSingleStrategyHub(strategyExecutor)
	go singleStrategyHub.Run()

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
	http.HandleFunc("/api/strategy/start", handlers.StartStrategy)
	http.HandleFunc("/api/strategy/stop", handlers.StopStrategy)
	http.HandleFunc("/ws/trades", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeAllTradesWs(tradeStatusHub, w, r)
	})
	http.HandleFunc("/ws/trade/", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeSingleTradeWs(singleTradeHub, w, r)
	})
	http.HandleFunc("/ws/strategies", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeAllStrategiesWs(strategyStatusHub, w, r)
	})
	http.HandleFunc("/ws/strategy/", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeSingleStrategyWs(singleStrategyHub, w, r)
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

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aumbhatt/auto_trade/internal/config"
	"github.com/aumbhatt/auto_trade/internal/handler"
	"github.com/aumbhatt/auto_trade/internal/service"
	"github.com/aumbhatt/auto_trade/internal/source/mock"
	"github.com/aumbhatt/auto_trade/internal/store/memory"
	"github.com/aumbhatt/auto_trade/internal/strategy"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

func main() {
	log.Println("Starting application...")

	// Load configuration
	cfg := config.NewDefaultConfig()

	// Create registry and register handlers
	registry := handler.NewRegistry()

	// Create mock tick source
	mockSource := mock.NewMockTickSource()

	// Create and start WebSocket hub
	hub := websocket.NewHub(registry)
	go hub.Run()

	// Create handlers
	tradeStore := memory.NewInMemoryTradeStore()
	strategyStore := memory.NewInMemoryStrategyStore()
	strategyRunner := strategy.NewDefaultRunner(strategyStore, tradeStore)

	// Create tick handler
	tickHandler := handler.NewTickHandler(hub, mockSource)
	if err := registry.Register("ticks", tickHandler); err != nil {
		log.Fatal(err)
	}

	// Create trade handlers
	openPositionsHandler := handler.NewOpenPositionsHandler(tradeStore, hub)
	tradeHistoryHandler := handler.NewTradeHistoryHandler(tradeStore, hub)
	tradeHandler := handler.NewTradeHandler(tradeStore, hub, openPositionsHandler, tradeHistoryHandler)

	// Create strategy handlers
	activeStrategiesHandler := handler.NewActiveStrategiesHandler(strategyStore, hub)
	strategyHistoryHandler := handler.NewStrategyHistoryHandler(strategyStore, hub)
	strategyHandler := handler.NewStrategyHandler(strategyStore, strategyRunner, tickHandler, hub, activeStrategiesHandler, strategyHistoryHandler)

	// Register trade message handlers
	if err := registry.Register("open_positions", openPositionsHandler); err != nil {
		log.Fatal(err)
	}
	if err := registry.Register("trade_history", tradeHistoryHandler); err != nil {
		log.Fatal(err)
	}

	// Register strategy message handlers
	if err := registry.Register("active_strategies", activeStrategiesHandler); err != nil {
		log.Fatal(err)
	}
	if err := registry.Register("strategies_history", strategyHistoryHandler); err != nil {
		log.Fatal(err)
	}

	// Start all handlers
	if err := registry.StartAll(); err != nil {
		log.Fatal(err)
	}

	// Initialize service with hub
	svc := service.NewService(cfg, hub)

	// Set up routes
	http.HandleFunc("/api/trades/buy", tradeHandler.HandleBuy)
	http.HandleFunc("/api/trades/sell", tradeHandler.HandleSell)
	http.HandleFunc("/api/strategies/start", strategyHandler.HandleStart)
	http.HandleFunc("/api/strategies/stop", strategyHandler.HandleStop)

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

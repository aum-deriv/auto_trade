package websocket

import (
	"auto_trade/ticks"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type TicksHub struct {
	*Hub
	generator      *ticks.TickGenerator
	strategyHandler *StrategyTicksHandler
}

func NewTicksHub(strategyHandler *StrategyTicksHandler) *TicksHub {
	return &TicksHub{
		Hub:            NewHub(),
		generator:      ticks.NewTickGenerator(),
		strategyHandler: strategyHandler,
	}
}

func (h *TicksHub) Run() {
	// Start the base hub
	go h.Hub.Run()

	// Start tick generation with strategy processing
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			tick := h.generator.GenerateTick()
			
			// Broadcast to websocket clients
			h.broadcast <- tick

			// Process tick for strategies
			var tickData ticks.Tick
			if err := json.Unmarshal(tick, &tickData); err == nil {
				h.strategyHandler.HandleTick(&tickData)
			}
		}
	}()
}

// ServeTicksWs handles websocket requests for market data ticks
func ServeTicksWs(hub *TicksHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub.Hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	go client.writePump()
}

package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"auto_trade/trading/strategies"
)

// StrategyStatusHub maintains the set of active clients and broadcasts strategy updates
type StrategyStatusHub struct {
	*Hub
	executor *strategies.StrategyExecutor
}

// NewStrategyStatusHub creates a new strategy status hub
func NewStrategyStatusHub(executor *strategies.StrategyExecutor) *StrategyStatusHub {
	return &StrategyStatusHub{
		Hub:      NewHub(),
		executor: executor,
	}
}

// Run starts the strategy status hub
func (h *StrategyStatusHub) Run() {
	h.Hub.Run()
}

// BroadcastActiveStrategies periodically broadcasts all active strategies
func (h *StrategyStatusHub) BroadcastActiveStrategies() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		strategies := h.executor.GetActiveStrategies()
		if data, err := json.Marshal(strategies); err == nil {
			h.broadcast <- data
		}
	}
}

// ServeAllStrategiesWs handles websocket requests for all active strategies
func ServeAllStrategiesWs(hub *StrategyStatusHub, w http.ResponseWriter, r *http.Request) {
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

// SingleStrategyHub maintains the set of active clients and broadcasts updates for a single strategy
type SingleStrategyHub struct {
	*Hub
	executor   *strategies.StrategyExecutor
	strategyID string
}

// NewSingleStrategyHub creates a new single strategy hub
func NewSingleStrategyHub(executor *strategies.StrategyExecutor) *SingleStrategyHub {
	return &SingleStrategyHub{
		Hub:      NewHub(),
		executor: executor,
	}
}

// Run starts the single strategy hub
func (h *SingleStrategyHub) Run() {
	h.Hub.Run()
}

// BroadcastStrategy periodically broadcasts the strategy status
func (h *SingleStrategyHub) BroadcastStrategy() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if strategy := h.executor.GetStrategy(h.strategyID); strategy != nil {
			if data, err := json.Marshal(strategy); err == nil {
				h.broadcast <- data
			}
		}
	}
}

// ServeSingleStrategyWs handles websocket requests for a single strategy
func ServeSingleStrategyWs(hub *SingleStrategyHub, w http.ResponseWriter, r *http.Request) {
	strategyID := strings.TrimPrefix(r.URL.Path, "/ws/strategy/")
	if strategyID == "" {
		http.Error(w, "Strategy ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	hub.strategyID = strategyID

	client := &Client{
		hub:  hub.Hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	go client.writePump()
	go hub.BroadcastStrategy()
}

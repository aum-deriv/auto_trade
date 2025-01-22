package websocket

import (
	"auto_trade/trading"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type TradeStatusHub struct {
	*Hub
	tradeManager *trading.TradeManager
}

func NewTradeStatusHub(tradeManager *trading.TradeManager) *TradeStatusHub {
	return &TradeStatusHub{
		Hub:          NewHub(),
		tradeManager: tradeManager,
	}
}

// BroadcastOpenTrades sends updates for all open trades
func (h *TradeStatusHub) BroadcastOpenTrades() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		trades := h.tradeManager.GetOpenTrades()
		data, err := json.Marshal(trades)
		if err != nil {
			log.Printf("Error marshaling trades: %v", err)
			continue
		}
		h.broadcast <- data
	}
}

// ServeAllTradesWs handles websocket requests for all open trades
func ServeAllTradesWs(hub *TradeStatusHub, w http.ResponseWriter, r *http.Request) {
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

type SingleTradeHub struct {
	*Hub
	tradeManager *trading.TradeManager
	tradeID      string
}

func NewSingleTradeHub(tradeManager *trading.TradeManager) *SingleTradeHub {
	return &SingleTradeHub{
		Hub:          NewHub(),
		tradeManager: tradeManager,
	}
}

// broadcastSingleTrade sends updates for a specific trade
func (h *SingleTradeHub) broadcastSingleTrade() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		trade, err := h.tradeManager.GetTrade(h.tradeID)
		if err != nil {
			log.Printf("Error getting trade: %v", err)
			continue
		}
		data, err := json.Marshal(trade)
		if err != nil {
			log.Printf("Error marshaling trade: %v", err)
			continue
		}
		h.broadcast <- data
	}
}

// ServeSingleTradeWs handles websocket requests for a specific trade
func ServeSingleTradeWs(hub *SingleTradeHub, w http.ResponseWriter, r *http.Request) {
	// Extract trade ID from URL path
	tradeID := strings.TrimPrefix(r.URL.Path, "/ws/trade/")
	if tradeID == "" {
		http.Error(w, "Trade ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	hub.tradeID = tradeID
	client := &Client{
		hub:  hub.Hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	go client.writePump()
	go hub.broadcastSingleTrade()
}

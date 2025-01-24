package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/aumbhatt/auto_trade/internal/store"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

/*
Trade Handler Flow and Examples:

1. Components and Event Flow:
   ├── TradeHandler: Main HTTP handler
   ├── OpenPositionsHandler: WebSocket handler for open trades
   └── TradeHistoryHandler: WebSocket handler for trade history

   Event Processing:
   a. Trade Creation:
      1. HTTP request received
      2. Trade created via store
      3. Store emits TradeCreated event
      4. OpenPositionsHandler receives event
      5. Updates broadcast to WebSocket subscribers

   b. Trade Closure:
      1. HTTP request received
      2. Trade closed via store
      3. Store emits TradeClosed event
      4. Both handlers receive event
      5. Updates broadcast to WebSocket subscribers

2. REST Endpoints:
   a. Buy Trade (POST /api/trades/buy):
      Request:
      {
          "symbol": "AAPL",
          "entry_price": 150.25
      }

      Success Response: (200 OK)
      {
          "trade_id": "trade-abc123",
          "symbol": "AAPL",
          "entry_price": 150.25,
          "entry_time": "2025-01-23T14:23:38Z"
      }

      Error Response: (400 Bad Request)
      {
          "code": "INVALID_SYMBOL",
          "message": "Invalid trading symbol: XYZ"
      }

   b. Sell Trade (POST /api/trades/sell):
      Request:
      {
          "trade_id": "trade-abc123"
      }

      Success Response: (200 OK)
      {
          "trade_id": "trade-abc123",
          "symbol": "AAPL",
          "entry_price": 150.25,
          "exit_price": 151.50,
          "entry_time": "2025-01-23T14:23:38Z",
          "exit_time": "2025-01-23T14:30:00Z"
      }

      Error Response: (404 Not Found)
      {
          "code": "TRADE_NOT_FOUND",
          "message": "Trade not found: trade-abc123"
      }

3. WebSocket Messages:
   a. Subscribe to Open Positions:
      Request:
      {
          "type": "subscribe",
          "payload": {
              "type": "open_positions"
          }
      }

      Success Response:
      {
          "type": "open_positions",
          "subscribe_id": "sub-123",
          "payload": [
              {
                  "trade_id": "trade-abc123",
                  "symbol": "AAPL",
                  "entry_price": 150.25,
                  "entry_time": "2025-01-23T14:23:38Z"
              }
          ]
      }

      Error Response:
      {
          "type": "error",
          "payload": {
              "code": "NO_OPEN_POSITIONS",
              "message": "No open positions found"
          }
      }

   b. Subscribe to Trade History:
      Request:
      {
          "type": "subscribe",
          "payload": {
              "type": "trade_history"
          }
      }

      Success Response:
      {
          "type": "trade_history",
          "subscribe_id": "sub-456",
          "payload": [
              {
                  "trade_id": "trade-xyz789",
                  "symbol": "GOOGL",
                  "entry_price": 140.50,
                  "exit_price": 142.75,
                  "entry_time": "2025-01-23T13:00:00Z",
                  "exit_time": "2025-01-23T14:00:00Z"
              }
          ]
      }

      Error Response:
      {
          "type": "error",
          "payload": {
              "code": "NO_TRADE_HISTORY",
              "message": "No trade history found"
          }
      }
*/

// TradeHandler handles trade-related requests
type TradeHandler struct {
	store             store.TradeStore
	hub               *websocket.Hub
	openPosHandler    *OpenPositionsHandler
	tradeHistHandler  *TradeHistoryHandler
}

// NewTradeHandler creates a new TradeHandler instance
func NewTradeHandler(store store.TradeStore, hub *websocket.Hub, openPosHandler *OpenPositionsHandler, tradeHistHandler *TradeHistoryHandler) *TradeHandler {
	// Register handlers as trade event listeners
	store.AddListener(openPosHandler)
	store.AddListener(tradeHistHandler)

	return &TradeHandler{
		store:             store,
		hub:              hub,
		openPosHandler:    openPosHandler,
		tradeHistHandler:  tradeHistHandler,
	}
}

// HandleBuy handles trade creation requests
func (h *TradeHandler) HandleBuy(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trade, err := h.store.CreateTrade(req.Symbol, req.EntryPrice)
	if err != nil {
		if e, ok := err.(*models.TradeError); ok {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trade)
}

// HandleSell handles trade closing requests
func (h *TradeHandler) HandleSell(w http.ResponseWriter, r *http.Request) {
	var req models.CloseTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trade, err := h.store.CloseTrade(req.TradeID)
	if err != nil {
		if e, ok := err.(*models.TradeError); ok {
			switch e.Code {
			case models.ErrTradeNotFound:
				http.Error(w, e.Error(), http.StatusNotFound)
			default:
				http.Error(w, e.Error(), http.StatusBadRequest)
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(trade)
}

// OpenPositionsHandler handles open positions subscriptions
type OpenPositionsHandler struct {
	store store.TradeStore
	hub   *websocket.Hub
	// Track subscriptions
	subscriptions sync.Map // map[string]struct{} // subscribeID -> struct{}
	subMutex     sync.RWMutex // Protects subscription operations
}

// NewOpenPositionsHandler creates a new OpenPositionsHandler
func NewOpenPositionsHandler(store store.TradeStore, hub *websocket.Hub) *OpenPositionsHandler {
	return &OpenPositionsHandler{
		store: store,
		hub:   hub,
	}
}

// OnTradeEvent implements store.TradeEventListener
func (h *OpenPositionsHandler) OnTradeEvent(event store.TradeEvent) {
	// Get updated open trades list
	trades, err := h.store.GetOpenTrades()
	if err != nil {
		log.Printf("Error getting open trades: %v", err)
		return
	}

	// Broadcast update to all subscribers
	h.BroadcastUpdate(trades)
}

// HandleSubscribe handles subscription requests
func (h *OpenPositionsHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	h.subMutex.Lock()
	h.subscriptions.Store(subscribeID, struct{}{})
	h.subMutex.Unlock()

	trades, err := h.store.GetOpenTrades()
	if err != nil {
		// Return empty list instead of error
		trades = []*models.Trade{}
	}

	msg := websocket.Message{
		Type:        "open_positions",
		SubscribeID: subscribeID,
		Payload:     trades,
	}
	h.hub.Broadcast(msg)
	return nil
}

// HandleUnsubscribe handles unsubscribe requests
func (h *OpenPositionsHandler) HandleUnsubscribe(subscribeID string) error {
	h.subMutex.Lock()
	h.subscriptions.Delete(subscribeID)
	h.subMutex.Unlock()
	return nil
}

// BroadcastUpdate sends updates to all subscribers
func (h *OpenPositionsHandler) BroadcastUpdate(trades []*models.Trade) {
	// Collect subscribers under read lock
	h.subMutex.RLock()
	subscribers := make([]string, 0)
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribers = append(subscribers, key.(string))
		return true
	})
	h.subMutex.RUnlock()

	// Broadcast outside lock
	for _, subscribeID := range subscribers {
		h.hub.Broadcast(websocket.Message{
			Type:        "open_positions",
			SubscribeID: subscribeID,
			Payload:     trades,
		})
	}
}

// Start starts the handler
func (h *OpenPositionsHandler) Start() error {
	return nil // No startup needed
}

// Stop stops the handler
func (h *OpenPositionsHandler) Stop() error {
	return nil // No cleanup needed
}

// TradeHistoryHandler handles trade history subscriptions
type TradeHistoryHandler struct {
	store store.TradeStore
	hub   *websocket.Hub
	// Track subscriptions
	subscriptions sync.Map // map[string]struct{} // subscribeID -> struct{}
	subMutex     sync.RWMutex // Protects subscription operations
}

// NewTradeHistoryHandler creates a new TradeHistoryHandler
func NewTradeHistoryHandler(store store.TradeStore, hub *websocket.Hub) *TradeHistoryHandler {
	return &TradeHistoryHandler{
		store: store,
		hub:   hub,
	}
}

// OnTradeEvent implements store.TradeEventListener
func (h *TradeHistoryHandler) OnTradeEvent(event store.TradeEvent) {
	// Only process closed trades
	if event.Type != store.TradeClosed {
		return
	}

	// Get updated trade history
	trades, err := h.store.GetTradeHistory()
	if err != nil {
		log.Printf("Error getting trade history: %v", err)
		return
	}

	// Broadcast update to all subscribers
	h.BroadcastUpdate(trades)
}

// HandleSubscribe handles subscription requests
func (h *TradeHistoryHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	h.subMutex.Lock()
	h.subscriptions.Store(subscribeID, struct{}{})
	h.subMutex.Unlock()

	trades, err := h.store.GetTradeHistory()
	if err != nil {
		// Return empty list instead of error
		trades = []*models.Trade{}
	}

	msg := websocket.Message{
		Type:        "trade_history",
		SubscribeID: subscribeID,
		Payload:     trades,
	}
	h.hub.Broadcast(msg)
	return nil
}

// HandleUnsubscribe handles unsubscribe requests
func (h *TradeHistoryHandler) HandleUnsubscribe(subscribeID string) error {
	h.subMutex.Lock()
	h.subscriptions.Delete(subscribeID)
	h.subMutex.Unlock()
	return nil
}

// BroadcastUpdate sends updates to all subscribers
func (h *TradeHistoryHandler) BroadcastUpdate(trades []*models.Trade) {
	// Collect subscribers under read lock
	h.subMutex.RLock()
	subscribers := make([]string, 0)
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribers = append(subscribers, key.(string))
		return true
	})
	h.subMutex.RUnlock()

	// Broadcast outside lock
	for _, subscribeID := range subscribers {
		h.hub.Broadcast(websocket.Message{
			Type:        "trade_history",
			SubscribeID: subscribeID,
			Payload:     trades,
		})
	}
}

// Start starts the handler
func (h *TradeHistoryHandler) Start() error {
	return nil // No startup needed
}

// Stop stops the handler
func (h *TradeHistoryHandler) Stop() error {
	return nil // No cleanup needed
}

package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/aumbhatt/auto_trade/internal/store"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

/*
Trade Handler Flow and Examples:

1. REST Endpoints:

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

2. WebSocket Messages:

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

	// Broadcast update to open positions subscribers
	openTrades, err := h.store.GetOpenTrades()
	if err != nil {
		// Log error but continue with response
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast to open positions subscribers
	h.openPosHandler.BroadcastUpdate(openTrades)

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

	// Broadcast updates to subscribers
	openTrades, err := h.store.GetOpenTrades()
	if err != nil {
		// Log error but continue
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast to open positions subscribers
	h.openPosHandler.BroadcastUpdate(openTrades)

	tradeHistory, err := h.store.GetTradeHistory()
	if err != nil {
		// Log error but continue
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast to trade history subscribers
	h.tradeHistHandler.BroadcastUpdate(tradeHistory)

	json.NewEncoder(w).Encode(trade)
}

// OpenPositionsHandler handles open positions subscriptions
type OpenPositionsHandler struct {
	store store.TradeStore
	hub   *websocket.Hub
	// Track subscriptions
	subscriptions sync.Map // map[string]struct{} // subscribeID -> struct{}
}

// NewOpenPositionsHandler creates a new OpenPositionsHandler
func NewOpenPositionsHandler(store store.TradeStore, hub *websocket.Hub) *OpenPositionsHandler {
	return &OpenPositionsHandler{
		store: store,
		hub:   hub,
	}
}

// HandleSubscribe handles subscription requests
func (h *OpenPositionsHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	// Store subscription
	h.subscriptions.Store(subscribeID, struct{}{})

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
	h.subscriptions.Delete(subscribeID)
	return nil
}

// BroadcastUpdate sends updates to all subscribers
func (h *OpenPositionsHandler) BroadcastUpdate(trades []*models.Trade) {
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribeID := key.(string)
		h.hub.Broadcast(websocket.Message{
			Type:        "open_positions",
			SubscribeID: subscribeID,
			Payload:     trades,
		})
		return true
	})
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
}

// NewTradeHistoryHandler creates a new TradeHistoryHandler
func NewTradeHistoryHandler(store store.TradeStore, hub *websocket.Hub) *TradeHistoryHandler {
	return &TradeHistoryHandler{
		store: store,
		hub:   hub,
	}
}

// HandleSubscribe handles subscription requests
func (h *TradeHistoryHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	// Store subscription
	h.subscriptions.Store(subscribeID, struct{}{})

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
	h.subscriptions.Delete(subscribeID)
	return nil
}

// BroadcastUpdate sends updates to all subscribers
func (h *TradeHistoryHandler) BroadcastUpdate(trades []*models.Trade) {
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribeID := key.(string)
		h.hub.Broadcast(websocket.Message{
			Type:        "trade_history",
			SubscribeID: subscribeID,
			Payload:     trades,
		})
		return true
	})
}

// Start starts the handler
func (h *TradeHistoryHandler) Start() error {
	return nil // No startup needed
}

// Stop stops the handler
func (h *TradeHistoryHandler) Stop() error {
	return nil // No cleanup needed
}

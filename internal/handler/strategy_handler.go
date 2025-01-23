package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/aumbhatt/auto_trade/internal/source"
	"github.com/aumbhatt/auto_trade/internal/store"
	"github.com/aumbhatt/auto_trade/internal/strategy"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

/*
Strategy Handler Flow and Structure:

1. Components:
   StrategyHandler
   ├── store: StrategyStore          // Strategy storage
   ├── runner: Runner                // Strategy execution
   ├── tickSource: TickSource        // Price updates
   └── hub: *websocket.Hub           // WebSocket broadcasting

2. REST Endpoints:
   a. Start Strategy (POST /api/strategies/start):
      Request:
      {
          "name": "moving_average",
          "parameters": {
              "symbol": "AAPL",
              "period": 20,
              "threshold": 0.02
          }
      }

      Success Response: (200 OK)
      {
          "id": "moving_average-abc123",
          "start_time": "2025-01-23T14:23:38Z",
          "status": "active"
      }

      Error Response: (400 Bad Request)
      {
          "code": "INVALID_STRATEGY",
          "message": "Invalid strategy parameters"
      }

   b. Stop Strategy (POST /api/strategies/stop):
      Request:
      {
          "id": "moving_average-abc123"
      }

      Success Response: (200 OK)
      {
          "id": "moving_average-abc123",
          "start_time": "2025-01-23T14:23:38Z",
          "stop_time": "2025-01-23T14:30:00Z",
          "status": "stopped"
      }

      Error Response: (404 Not Found)
      {
          "code": "STRATEGY_NOT_FOUND",
          "message": "Strategy not found: moving_average-abc123"
      }

3. WebSocket Messages:

   a. Subscribe to Active Strategies:
      Request:
      {
          "type": "subscribe",
          "payload": {
              "type": "active_strategies"
          }
      }

      Response:
      {
          "type": "subscribe_response",
          "subscribe_id": "sub-123",
          "status": "success"
      }

      Updates:
      {
          "type": "active_strategies",
          "subscribe_id": "sub-123",
          "payload": [
              {
                  "id": "moving_average-abc123",
                  "name": "moving_average",
                  "parameters": {
                      "symbol": "AAPL",
                      "period": 20,
                      "threshold": 0.02
                  },
                  "start_time": "2025-01-23T14:23:38Z",
                  "status": "active"
              }
          ]
      }

   b. Subscribe to Strategy History:
      Request:
      {
          "type": "subscribe",
          "payload": {
              "type": "strategies_history"
          }
      }

      Response:
      {
          "type": "subscribe_response",
          "subscribe_id": "sub-456",
          "status": "success"
      }

      Updates:
      {
          "type": "strategies_history",
          "subscribe_id": "sub-456",
          "payload": [
              {
                  "id": "moving_average-xyz789",
                  "name": "moving_average",
                  "parameters": {
                      "symbol": "GOOGL",
                      "period": 50,
                      "threshold": 0.03
                  },
                  "start_time": "2025-01-23T13:00:00Z",
                  "stop_time": "2025-01-23T14:00:00Z",
                  "status": "stopped"
              }
          ]
      }

   c. Unsubscribe:
      Request:
      {
          "type": "unsubscribe",
          "payload": {
              "subscribe_id": "sub-123"
          }
      }

      Response:
      {
          "type": "unsubscribe_response",
          "subscribe_id": "sub-123",
          "status": "success"
      }
*/

// StrategyHandler handles strategy-related HTTP requests
type StrategyHandler struct {
	store                  store.StrategyStore
	runner                 strategy.Runner
	tickSource            source.TickSource
	hub                   *websocket.Hub
	activeStrategiesHandler  *ActiveStrategiesHandler
	strategyHistoryHandler   *StrategyHistoryHandler
}

// NewStrategyHandler creates a new StrategyHandler instance
func NewStrategyHandler(store store.StrategyStore, runner strategy.Runner, tickSource source.TickSource, hub *websocket.Hub, activeStrategiesHandler *ActiveStrategiesHandler, strategyHistoryHandler *StrategyHistoryHandler) *StrategyHandler {
	return &StrategyHandler{
		store:                  store,
		runner:                 runner,
		tickSource:            tickSource,
		hub:                   hub,
		activeStrategiesHandler:  activeStrategiesHandler,
		strategyHistoryHandler:   strategyHistoryHandler,
	}
}

// HandleStart handles strategy start requests
func (h *StrategyHandler) HandleStart(w http.ResponseWriter, r *http.Request) {
	var req models.StartStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create strategy
	strategy, err := h.store.CreateStrategy(req.Name, req.Parameters)
	if err != nil {
		if e, ok := err.(*models.StrategyError); ok {
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create tick channel for strategy
	tickChan := make(chan *models.Tick)

	// Start strategy
	if err := h.runner.Start(strategy, tickChan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast updates
	activeStrategies, _ := h.store.GetActiveStrategies()
	h.activeStrategiesHandler.BroadcastActiveStrategiesUpdate(activeStrategies)

	// Return response
	resp := models.StartStrategyResponse{
		ID:        strategy.ID,
		StartTime: strategy.StartTime,
		Status:    strategy.Status,
	}
	json.NewEncoder(w).Encode(resp)
}

// HandleStop handles strategy stop requests
func (h *StrategyHandler) HandleStop(w http.ResponseWriter, r *http.Request) {
	var req models.StopStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get strategy
	strategy, err := h.store.GetStrategyByID(req.ID)
	if err != nil {
		if e, ok := err.(*models.StrategyError); ok {
			switch e.Code {
			case models.ErrStrategyNotFound:
				http.Error(w, e.Error(), http.StatusNotFound)
			default:
				http.Error(w, e.Error(), http.StatusBadRequest)
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Stop strategy
	if err := h.runner.Stop(strategy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast updates
	activeStrategies, _ := h.store.GetActiveStrategies()
	historyStrategies, _ := h.store.GetStrategyHistory()
	h.activeStrategiesHandler.BroadcastActiveStrategiesUpdate(activeStrategies)
	h.strategyHistoryHandler.BroadcastStrategyHistoryUpdate(historyStrategies)

	// Return response
	resp := models.StopStrategyResponse{
		ID:        strategy.ID,
		StartTime: strategy.StartTime,
		StopTime:  *strategy.StopTime,
		Status:    strategy.Status,
	}
	json.NewEncoder(w).Encode(resp)
}

// ActiveStrategiesHandler handles active strategies subscriptions
type ActiveStrategiesHandler struct {
	store store.StrategyStore
	hub   *websocket.Hub
	// Track subscriptions
	subscriptions sync.Map // map[string]struct{} // subscribeID -> struct{}
}

// NewActiveStrategiesHandler creates a new ActiveStrategiesHandler
func NewActiveStrategiesHandler(store store.StrategyStore, hub *websocket.Hub) *ActiveStrategiesHandler {
	return &ActiveStrategiesHandler{
		store: store,
		hub:   hub,
	}
}

// HandleSubscribe handles subscription requests for active strategies
func (h *ActiveStrategiesHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	// Store subscription
	h.subscriptions.Store(subscribeID, struct{}{})

	strategies, err := h.store.GetActiveStrategies()
	if err != nil {
		// Return empty list instead of error
		strategies = []*models.Strategy{}
	}

	msg := websocket.Message{
		Type:        "active_strategies",
		SubscribeID: subscribeID,
		Payload:     strategies,
	}
	h.hub.Broadcast(msg)
	return nil
}

// HandleUnsubscribe handles unsubscribe requests for active strategies
func (h *ActiveStrategiesHandler) HandleUnsubscribe(subscribeID string) error {
	h.subscriptions.Delete(subscribeID)
	return nil
}

// BroadcastActiveStrategiesUpdate sends updates to all active strategies subscribers
func (h *ActiveStrategiesHandler) BroadcastActiveStrategiesUpdate(strategies []*models.Strategy) {
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribeID := key.(string)
		h.hub.Broadcast(websocket.Message{
			Type:        "active_strategies",
			SubscribeID: subscribeID,
			Payload:     strategies,
		})
		return true
	})
}

// Start starts the handler
func (h *ActiveStrategiesHandler) Start() error {
	return nil // No startup needed
}

// Stop stops the handler
func (h *ActiveStrategiesHandler) Stop() error {
	return nil // No cleanup needed
}

// StrategyHistoryHandler handles strategy history subscriptions
type StrategyHistoryHandler struct {
	store store.StrategyStore
	hub   *websocket.Hub
	// Track subscriptions
	subscriptions sync.Map // map[string]struct{} // subscribeID -> struct{}
}

// NewStrategyHistoryHandler creates a new StrategyHistoryHandler
func NewStrategyHistoryHandler(store store.StrategyStore, hub *websocket.Hub) *StrategyHistoryHandler {
	return &StrategyHistoryHandler{
		store: store,
		hub:   hub,
	}
}

// HandleSubscribe handles subscription requests for strategy history
func (h *StrategyHistoryHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	// Store subscription
	h.subscriptions.Store(subscribeID, struct{}{})

	strategies, err := h.store.GetStrategyHistory()
	if err != nil {
		// Return empty list instead of error
		strategies = []*models.Strategy{}
	}

	msg := websocket.Message{
		Type:        "strategies_history",
		SubscribeID: subscribeID,
		Payload:     strategies,
	}
	h.hub.Broadcast(msg)
	return nil
}

// HandleUnsubscribe handles unsubscribe requests for strategy history
func (h *StrategyHistoryHandler) HandleUnsubscribe(subscribeID string) error {
	h.subscriptions.Delete(subscribeID)
	return nil
}

// BroadcastStrategyHistoryUpdate sends updates to all strategy history subscribers
func (h *StrategyHistoryHandler) BroadcastStrategyHistoryUpdate(strategies []*models.Strategy) {
	h.subscriptions.Range(func(key, value interface{}) bool {
		subscribeID := key.(string)
		h.hub.Broadcast(websocket.Message{
			Type:        "strategies_history",
			SubscribeID: subscribeID,
			Payload:     strategies,
		})
		return true
	})
}

// Start starts the handler
func (h *StrategyHistoryHandler) Start() error {
	return nil // No startup needed
}

// Stop stops the handler
func (h *StrategyHistoryHandler) Stop() error {
	return nil // No cleanup needed
}

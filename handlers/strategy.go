package handlers

import (
	"encoding/json"
	"net/http"

	"auto_trade/trading"
	"auto_trade/trading/strategies"
)

// StartStrategyRequest represents the request body for starting a strategy
type StartStrategyRequest struct {
	Type      string  `json:"type"`      // Strategy type (e.g., "MARTINGALE")
	Symbol    string  `json:"symbol"`    // Trading pair
	BaseSize  float64 `json:"base_size"` // Initial position size
	MaxLosses int     `json:"max_losses"` // Maximum consecutive losses
}

// StrategyResponse represents the response for strategy operations
type StrategyResponse struct {
	StrategyID string `json:"strategy_id"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
}

// StopStrategyRequest represents the request body for stopping a strategy
type StopStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

var (
	// activeStrategies stores running strategy instances
	activeStrategies = make(map[string]interface{})
	// strategyExecutor manages strategy execution
	strategyExecutor *strategies.StrategyExecutor
)

// InitStrategyHandler initializes the strategy handler with required dependencies
func InitStrategyHandler(tradeManager *trading.TradeManager) *strategies.StrategyExecutor {
	strategyExecutor = strategies.NewStrategyExecutor(tradeManager)
	return strategyExecutor
}

// StartStrategy handles requests to start a new trading strategy
func StartStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Type == "" || req.Symbol == "" || req.BaseSize <= 0 || req.MaxLosses <= 0 {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}

	var strategy interface{}
	var strategyID string

	// Create strategy based on type
	switch req.Type {
	case "MARTINGALE":
		martStrategy := strategies.NewMartingaleStrategy(
			req.BaseSize,
			req.Symbol,
			req.MaxLosses,
		)
		strategy = martStrategy
		strategyID = martStrategy.StrategyID
	default:
		http.Error(w, "Unsupported strategy type", http.StatusBadRequest)
		return
	}

	// Store the strategy instance and start execution
	activeStrategies[strategyID] = strategy
	
	// Cast to Martingale strategy and start execution
	if martStrategy, ok := strategy.(*strategies.MartingaleStrategy); ok {
		if err := strategyExecutor.StartStrategy(martStrategy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Invalid strategy type", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := StrategyResponse{
		StrategyID: strategyID,
		Status:     "started",
		Message:    "Strategy started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopStrategy handles requests to stop a running strategy
func StopStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StopStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if strategy exists
	strategy, exists := activeStrategies[req.StrategyID]
	if !exists {
		http.Error(w, "Strategy not found", http.StatusNotFound)
		return
	}

	// If it's a Martingale strategy, reset it before stopping
	if martStrategy, ok := strategy.(*strategies.MartingaleStrategy); ok {
		martStrategy.Reset()
	}

	// Stop strategy execution
	if err := strategyExecutor.StopStrategy(req.StrategyID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove strategy from active strategies
	delete(activeStrategies, req.StrategyID)

	// Prepare response
	response := StrategyResponse{
		StrategyID: req.StrategyID,
		Status:     "stopped",
		Message:    "Strategy stopped successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

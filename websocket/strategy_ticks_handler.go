package websocket

import (
	"log"

	"auto_trade/ticks"
	"auto_trade/trading/strategies"
)

// StrategyTicksHandler processes market data ticks for strategies
type StrategyTicksHandler struct {
	executor *strategies.StrategyExecutor
}

// NewStrategyTicksHandler creates a new strategy ticks handler
func NewStrategyTicksHandler(executor *strategies.StrategyExecutor) *StrategyTicksHandler {
	return &StrategyTicksHandler{
		executor: executor,
	}
}

// HandleTick processes a market data tick
func (h *StrategyTicksHandler) HandleTick(tick *ticks.Tick) {
	log.Printf("Processing tick for strategies: %s %.2f", tick.Symbol, tick.Price)
	h.executor.ProcessTick(tick.Symbol, tick.Price)
}

package strategy

import "github.com/aumbhatt/auto_trade/internal/models"

/*
Strategy Executor Flow and Structure:

1. Interface:
   StrategyExecutor
   └── ProcessTick    // Process incoming tick data

2. Operation Flow:
   a. Runner receives tick
   b. Passes to executor
   c. Executor processes according to strategy logic
   d. Returns any errors

3. Implementation Requirements:
   - Must be thread-safe
   - Handle all trading logic
   - Return meaningful errors

4. Example Usage:
   executor := NewRepeatStrategy(runner, params)
   err := executor.ProcessTick(tick)
*/

// StrategyExecutor defines the interface for strategy implementations
type StrategyExecutor interface {
	// ProcessTick processes a single tick of market data
	// Returns error if the tick processing fails
	ProcessTick(tick *models.Tick) error
}

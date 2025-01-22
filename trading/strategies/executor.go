package strategies

import (
	"fmt"
	"log"
	"sync"

	"auto_trade/trading"
)

// StrategyExecutor manages running strategies and handles market data processing
type StrategyExecutor struct {
	tradeManager  *trading.TradeManager
	activeWorkers sync.Map    // Maps strategy ID to cancel channel
	strategies    sync.Map    // Maps strategy ID to strategy instance
	mutex         sync.RWMutex
}

// NewStrategyExecutor creates a new strategy executor
func NewStrategyExecutor(tradeManager *trading.TradeManager) *StrategyExecutor {
	return &StrategyExecutor{
		tradeManager: tradeManager,
	}
}

// StartStrategy begins executing a strategy
func (se *StrategyExecutor) StartStrategy(strategy *MartingaleStrategy) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	// Check if strategy is already running
	if _, exists := se.activeWorkers.Load(strategy.StrategyID); exists {
		return fmt.Errorf("strategy %s is already running", strategy.StrategyID)
	}

	// Store strategy instance
	se.strategies.Store(strategy.StrategyID, strategy)

	// Create cancel channel for this strategy
	done := make(chan struct{})
	se.activeWorkers.Store(strategy.StrategyID, done)

	// Start strategy worker
	go se.runStrategy(strategy, done)

	return nil
}

// StopStrategy stops a running strategy
func (se *StrategyExecutor) StopStrategy(strategyID string) error {
	// Get the done channel for this strategy
	value, exists := se.activeWorkers.Load(strategyID)
	if !exists {
		return fmt.Errorf("strategy %s is not running", strategyID)
	}

	// Send stop signal
	done := value.(chan struct{})
	close(done)

	// Remove from active workers
	se.activeWorkers.Delete(strategyID)

	return nil
}

// runStrategy executes the strategy logic in a separate goroutine
func (se *StrategyExecutor) runStrategy(strategy *MartingaleStrategy, done chan struct{}) {
	log.Printf("Starting strategy: %s", strategy.StrategyID)
	defer log.Printf("Stopping strategy: %s", strategy.StrategyID)

	// TODO: Subscribe to market data for the strategy's symbol
	// TODO: Implement trade execution logic
	// For now, just wait for stop signal
	<-done
}

// ProcessTick processes a new market data tick for all relevant strategies
func (se *StrategyExecutor) ProcessTick(symbol string, price float64) {
	// Iterate through active strategies
	se.activeWorkers.Range(func(key, value interface{}) bool {
		strategyID := key.(string)
		// TODO: Process tick for each relevant strategy
		log.Printf("Processing tick for strategy %s: %s %.2f", strategyID, symbol, price)
		return true
	})
}

// IsStrategyRunning checks if a strategy is currently running
func (se *StrategyExecutor) IsStrategyRunning(strategyID string) bool {
	_, exists := se.activeWorkers.Load(strategyID)
	return exists
}

// GetActiveStrategies returns information about all active strategies
func (se *StrategyExecutor) GetActiveStrategies() []map[string]interface{} {
	var strategies []map[string]interface{}
	
	se.strategies.Range(func(key, value interface{}) bool {
		if martStrat, ok := value.(*MartingaleStrategy); ok {
			strategies = append(strategies, map[string]interface{}{
				"strategy_id":        martStrat.StrategyID,
				"type":              martStrat.Type,
				"symbol":            martStrat.Symbol,
				"base_size":         martStrat.BaseSize,
				"max_losses":        martStrat.MaxLosses,
				"current_size":      martStrat.GetCurrentSize(),
				"consecutive_losses": martStrat.GetConsecutiveLosses(),
			})
		}
		return true
	})
	
	return strategies
}

// GetStrategy returns information about a specific strategy
func (se *StrategyExecutor) GetStrategy(strategyID string) map[string]interface{} {
	if value, ok := se.strategies.Load(strategyID); ok {
		if martStrat, ok := value.(*MartingaleStrategy); ok {
			return map[string]interface{}{
				"strategy_id":        martStrat.StrategyID,
				"type":              martStrat.Type,
				"symbol":            martStrat.Symbol,
				"base_size":         martStrat.BaseSize,
				"max_losses":        martStrat.MaxLosses,
				"current_size":      martStrat.GetCurrentSize(),
				"consecutive_losses": martStrat.GetConsecutiveLosses(),
			}
		}
	}
	return nil
}

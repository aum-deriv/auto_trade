package strategy

import (
	"fmt"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/aumbhatt/auto_trade/internal/store"
)

/*
Strategy Runner Flow and Structure:

1. Memory Structure:
   DefaultRunner
   ├── store: StrategyStore          // Strategy storage
   ├── tradeStore: TradeStore        // For executing trades
   ├── runningJobs: map[string]chan struct{}  // Strategy ID -> done channel
   └── mu: sync.RWMutex             // Protects runningJobs map

2. Operation Flow:
   a. Starting Strategy:
      1. Create done channel
      2. Store in runningJobs map
      3. Start goroutine for strategy
      4. Return success/error

   b. Running Strategy:
      1. Receive ticks from tickChan
      2. Process according to strategy logic
      3. Execute trades via tradeStore
      4. Continue until done channel closed

   c. Stopping Strategy:
      1. Close done channel
      2. Remove from runningJobs
      3. Update strategy status
      4. Return success/error

3. Concurrency:
   - Each strategy runs in separate goroutine
   - Done channel for graceful shutdown
   - Mutex protection for shared resources
   - Safe access to trade operations

4. Error Handling:
   - Strategy-specific errors
   - Runner operation errors
   - Trade execution errors

5. Example Usage:
   runner := NewDefaultRunner(strategyStore, tradeStore)

   // Start strategy
   err := runner.Start(strategy, tickChan)

   // Later...
   err = runner.Stop(strategy)
*/

// Runner defines the interface for strategy execution
type Runner interface {
	// Start begins executing a strategy with tick data
	Start(strategy *models.Strategy, tickChan <-chan *models.Tick) error

	// Stop gracefully stops a running strategy
	Stop(strategy *models.Strategy) error
}

// DefaultRunner implements the Runner interface
type DefaultRunner struct {
	store      store.StrategyStore
	tradeStore store.TradeStore
	runningJobs map[string]chan struct{} // strategy ID -> done channel
	mu         sync.RWMutex
}

// NewDefaultRunner creates a new DefaultRunner instance
func NewDefaultRunner(strategyStore store.StrategyStore, tradeStore store.TradeStore) *DefaultRunner {
	return &DefaultRunner{
		store:      strategyStore,
		tradeStore: tradeStore,
		runningJobs: make(map[string]chan struct{}),
	}
}

// Start begins executing a strategy
func (r *DefaultRunner) Start(strategy *models.Strategy, tickChan <-chan *models.Tick) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already running
	if _, exists := r.runningJobs[strategy.ID]; exists {
		return fmt.Errorf("strategy already running: %s", strategy.ID)
	}

	// Create done channel
	done := make(chan struct{})
	r.runningJobs[strategy.ID] = done

	// Start strategy in goroutine
	go r.runStrategy(strategy, tickChan, done)

	return nil
}

// Stop gracefully stops a running strategy
func (r *DefaultRunner) Stop(strategy *models.Strategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	done, exists := r.runningJobs[strategy.ID]
	if !exists {
		return fmt.Errorf("strategy not running: %s", strategy.ID)
	}

	// Signal strategy to stop
	close(done)
	delete(r.runningJobs, strategy.ID)

	// Update strategy status
	_, err := r.store.StopStrategy(strategy.ID)
	return err
}

// runStrategy executes the strategy logic
func (r *DefaultRunner) runStrategy(strategy *models.Strategy, tickChan <-chan *models.Tick, done chan struct{}) {
	// Create strategy executor
	executor, err := GetDefaultRegistry().Create(strategy.Name, r, strategy.Parameters)
	if err != nil {
		// Log error but don't block - strategy will effectively be stopped
		fmt.Printf("Failed to create strategy executor: %v\n", err)
		return
	}

	// Strategy runs until done channel is closed
	for {
		select {
		case tick := <-tickChan:
			if err := executor.ProcessTick(tick); err != nil {
				// Log error but continue running
				fmt.Printf("Strategy %s error: %v\n", strategy.ID, err)
			}
		case <-done:
			return
		}
	}
}

// Helper methods for strategy implementations to use
func (r *DefaultRunner) executeBuy(symbol string, price float64) (*models.Trade, error) {
	// Use trade store to create trade
	return r.tradeStore.CreateTrade(symbol, price)
}

func (r *DefaultRunner) executeSell(tradeID string) (*models.Trade, error) {
	// Use trade store to close trade
	return r.tradeStore.CloseTrade(tradeID)
}

package strategy

import (
	"context"
	"fmt"
	"log"
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
	store       store.StrategyStore
	tradeStore  store.TradeStore
	runningJobs map[string]*runningJob // strategy ID -> running job info
	mu          sync.RWMutex
}

// runningJob holds information about a running strategy
type runningJob struct {
	done    chan struct{}    // Signal to stop the strategy
	errChan chan error       // Channel for executor errors
	cancel  func()          // Cancel function for the context
}

// NewDefaultRunner creates a new DefaultRunner instance
func NewDefaultRunner(strategyStore store.StrategyStore, tradeStore store.TradeStore) *DefaultRunner {
	return &DefaultRunner{
		store:       strategyStore,
		tradeStore:  tradeStore,
		runningJobs: make(map[string]*runningJob),
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

	// Create running job with error channel
	job := &runningJob{
		done:    make(chan struct{}),
		errChan: make(chan error, 1), // Buffered to prevent blocking
	}

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	job.cancel = cancel

	r.runningJobs[strategy.ID] = job

	// Start strategy in goroutine
	go func() {
		r.runStrategy(ctx, strategy, tickChan, job)
	}()

	// Start error handler
	go func() {
		r.handleErrors(strategy.ID, job)
	}()

	return nil
}

// Stop gracefully stops a running strategy
func (r *DefaultRunner) Stop(strategy *models.Strategy) error {
	r.mu.Lock()
	job, exists := r.runningJobs[strategy.ID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("strategy not running: %s", strategy.ID)
	}

	// Cancel context and signal strategy to stop
	job.cancel()
	close(job.done)

	// Wait for error handler to finish
	close(job.errChan)

	r.mu.Lock()
	delete(r.runningJobs, strategy.ID)
	r.mu.Unlock()

	// Update strategy status
	_, err := r.store.StopStrategy(strategy.ID)
	return err
}

// handleErrors handles errors from the strategy executor
func (r *DefaultRunner) handleErrors(strategyID string, job *runningJob) {
	for err := range job.errChan {
		if err != nil {
			// Log error
			log.Printf("Strategy %s error: %v", strategyID, err)

			// Stop strategy on critical errors
			if isCriticalError(err) {
				log.Printf("Stopping strategy %s due to critical error", strategyID)
				r.mu.Lock()
				if _, exists := r.runningJobs[strategyID]; exists {
					job.cancel()
					close(job.done)
					delete(r.runningJobs, strategyID)
					// Update strategy status
					if _, err := r.store.StopStrategy(strategyID); err != nil {
						log.Printf("Error stopping strategy %s: %v", strategyID, err)
					}
				}
				r.mu.Unlock()
				return
			}
		}
	}
}

// isCriticalError determines if an error should stop the strategy
func isCriticalError(err error) bool {
	// Add logic to determine critical errors
	// For now, treat all errors as non-critical
	return false
}

// runStrategy executes the strategy logic
func (r *DefaultRunner) runStrategy(ctx context.Context, strategy *models.Strategy, tickChan <-chan *models.Tick, job *runningJob) {
	// Create strategy executor
	executor, err := GetDefaultRegistry().Create(strategy.Name, r, strategy.Parameters)
	if err != nil {
		job.errChan <- fmt.Errorf("failed to create strategy executor: %w", err)
		return
	}

	// Strategy runs until done channel is closed
	for {
		select {
		case tick := <-tickChan:
			if err := executor.ProcessTick(tick); err != nil {
				job.errChan <- err
			}
		case <-ctx.Done():
			return
		case <-job.done:
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

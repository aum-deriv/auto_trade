package store

import "github.com/aumbhatt/auto_trade/internal/models"

/*
Strategy Store Interface and Flow:

1. Interface Methods:
   StrategyStore
   ├── CreateStrategy        // Creates and stores new strategy
   ├── StopStrategy         // Stops a running strategy
   ├── GetActiveStrategies  // Lists all active strategies
   ├── GetStrategyHistory   // Lists all stopped strategies
   └── GetStrategyByID      // Retrieves specific strategy

2. Operation Flow:
   a. Creating Strategy:
      1. Receive name and parameters
      2. Create Strategy object
      3. Store in active strategies map
      4. Return new strategy

   b. Stopping Strategy:
      1. Find strategy by ID in active strategies
      2. Mark as stopped using strategy.Stop()
      3. Remove from active strategies map
      4. Add to strategy history map
      5. Return updated strategy

   c. Querying:
      - GetActiveStrategies returns strategies from active map
      - GetStrategyHistory returns strategies from history map
      - GetStrategyByID checks both maps

3. Data Organization:
   activeStrategies map[string]*Strategy
   ├── "moving_average-abc123" → Strategy{Status: "active"}
   └── "rsi-def456" → Strategy{Status: "active"}

   strategyHistory map[string]*Strategy
   └── "sma-xyz789" → Strategy{Status: "stopped"}

4. Error Handling:
   - Return StrategyError for known error cases
   - Include error codes for client handling
   - Wrap unexpected errors

5. Usage Example:
   store := NewInMemoryStrategyStore()

   // Create strategy (goes to active map)
   strategy, err := store.CreateStrategy("moving_average", params)

   // Get active strategies (from active map)
   active, err := store.GetActiveStrategies()

   // Stop strategy (moves from active to history map)
   stopped, err := store.StopStrategy("moving_average-abc123")

   // Get history (from history map)
   history, err := store.GetStrategyHistory()
*/

// StrategyStore defines the interface for strategy storage operations
type StrategyStore interface {
	// CreateStrategy creates a new strategy with given name and parameters
	// The new strategy is stored in the active strategies map
	CreateStrategy(name string, params map[string]interface{}) (*models.Strategy, error)

	// StopStrategy stops a running strategy
	// 1. Finds strategy in active strategies map
	// 2. Marks it as stopped
	// 3. Moves it from active to history map
	StopStrategy(id string) (*models.Strategy, error)

	// GetActiveStrategies returns all currently active strategies
	// Returns strategies from the active strategies map
	GetActiveStrategies() ([]*models.Strategy, error)

	// GetStrategyHistory returns all stopped strategies
	// Returns strategies from the strategy history map
	GetStrategyHistory() ([]*models.Strategy, error)

	// GetStrategyByID returns a strategy by its ID
	// Checks both active and history maps
	GetStrategyByID(id string) (*models.Strategy, error)
}

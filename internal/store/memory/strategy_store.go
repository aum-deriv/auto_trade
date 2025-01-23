package memory

import (
	"fmt"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
)

/*
In-Memory Strategy Store Flow and Structure:

1. Memory Structure:
   InMemoryStrategyStore
   ├── activeStrategies: map[string]*Strategy  // Currently running strategies
   ├── strategyHistory: map[string]*Strategy   // Stopped strategies
   └── mu: sync.RWMutex                       // Protects both maps

2. Concurrency Handling:
   a. Read Operations (GetActive, GetHistory, GetByID):
      1. Acquire read lock (RLock)
      2. Access map data
      3. Release read lock (RUnlock)
      4. Return copied data

   b. Write Operations (Create, Stop):
      1. Acquire write lock (Lock)
      2. Modify maps
      3. Release write lock (Unlock)
      4. Return updated data

3. Data Flow:
   a. Creating Strategy:
      1. Generate new Strategy object
      2. Acquire write lock
      3. Add to activeStrategies map
      4. Release lock
      5. Return strategy

   b. Stopping Strategy:
      1. Acquire write lock
      2. Find in activeStrategies
      3. Mark as stopped
      4. Move to strategyHistory
      5. Release lock
      6. Return updated strategy

4. Error Handling:
   - Not found errors when stopping/getting strategy
   - Already stopped errors
   - Thread-safe error returns
*/

// InMemoryStrategyStore implements StrategyStore interface with in-memory storage
type InMemoryStrategyStore struct {
	activeStrategies map[string]*models.Strategy
	strategyHistory  map[string]*models.Strategy
	mu              sync.RWMutex
}

// NewInMemoryStrategyStore creates a new instance of InMemoryStrategyStore
func NewInMemoryStrategyStore() *InMemoryStrategyStore {
	return &InMemoryStrategyStore{
		activeStrategies: make(map[string]*models.Strategy),
		strategyHistory:  make(map[string]*models.Strategy),
	}
}

// CreateStrategy creates a new strategy with given name and parameters
func (s *InMemoryStrategyStore) CreateStrategy(name string, params map[string]interface{}) (*models.Strategy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	strategy := models.NewStrategy(name, params)
	s.activeStrategies[strategy.ID] = strategy
	return strategy, nil
}

// StopStrategy stops a running strategy
func (s *InMemoryStrategyStore) StopStrategy(id string) (*models.Strategy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	strategy, exists := s.activeStrategies[id]
	if !exists {
		return nil, &models.StrategyError{
			Code:    models.ErrStrategyNotFound,
			Message: fmt.Sprintf("Strategy not found: %s", id),
		}
	}

	if strategy.Status == "stopped" {
		return nil, &models.StrategyError{
			Code:    models.ErrAlreadyStopped,
			Message: fmt.Sprintf("Strategy already stopped: %s", id),
		}
	}

	// Stop the strategy
	strategy.Stop()

	// Move from active to history
	delete(s.activeStrategies, id)
	s.strategyHistory[id] = strategy

	return strategy, nil
}

// GetActiveStrategies returns all currently active strategies
func (s *InMemoryStrategyStore) GetActiveStrategies() ([]*models.Strategy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make([]*models.Strategy, 0, len(s.activeStrategies))
	for _, strategy := range s.activeStrategies {
		strategies = append(strategies, strategy)
	}
	return strategies, nil
}

// GetStrategyHistory returns all stopped strategies
func (s *InMemoryStrategyStore) GetStrategyHistory() ([]*models.Strategy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make([]*models.Strategy, 0, len(s.strategyHistory))
	for _, strategy := range s.strategyHistory {
		strategies = append(strategies, strategy)
	}
	return strategies, nil
}

// GetStrategyByID returns a strategy by its ID
func (s *InMemoryStrategyStore) GetStrategyByID(id string) (*models.Strategy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if strategy, exists := s.activeStrategies[id]; exists {
		return strategy, nil
	}

	if strategy, exists := s.strategyHistory[id]; exists {
		return strategy, nil
	}

	return nil, &models.StrategyError{
		Code:    models.ErrStrategyNotFound,
		Message: fmt.Sprintf("Strategy not found: %s", id),
	}
}

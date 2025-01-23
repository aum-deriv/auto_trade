package strategy

import (
	"fmt"
	"sync"
)

/*
Strategy Registry Flow and Structure:

1. Memory Structure:
   Registry
   ├── factories: map[string]StrategyFactory  // Strategy name -> factory function
   └── mu: sync.RWMutex                      // Protects factories map

2. Operation Flow:
   a. Registration:
      1. Get factory function
      2. Add to factories map
      3. Available for creation

   b. Creation:
      1. Look up factory by name
      2. Create executor instance
      3. Return for use

3. Error Handling:
   - Unknown strategy errors
   - Invalid parameters
   - Factory errors

4. Example Usage:
   registry := NewRegistry()
   registry.Register("repeat", NewRepeatStrategy)
   executor := registry.Create("repeat", runner, params)
*/

// StrategyFactory is a function that creates a new strategy executor
type StrategyFactory func(runner *DefaultRunner, params map[string]interface{}) (StrategyExecutor, error)

// Registry manages strategy types and their creation
type Registry struct {
	factories map[string]StrategyFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new strategy registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]StrategyFactory),
	}
}

// Register adds a new strategy type to the registry
func (r *Registry) Register(name string, factory StrategyFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Create creates a new strategy executor instance
func (r *Registry) Create(name string, runner *DefaultRunner, params map[string]interface{}) (StrategyExecutor, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown strategy type: %s", name)
	}

	return factory(runner, params)
}

// GetAvailableStrategies returns a list of registered strategy names
func (r *Registry) GetAvailableStrategies() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	strategies := make([]string, 0, len(r.factories))
	for name := range r.factories {
		strategies = append(strategies, name)
	}
	return strategies
}

// Default registry instance
var defaultRegistry = NewRegistry()

// GetDefaultRegistry returns the default registry instance
func GetDefaultRegistry() *Registry {
	return defaultRegistry
}

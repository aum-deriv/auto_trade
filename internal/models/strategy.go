package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

/*
Strategy Model Flow and Structure:

1. Memory Structure:
   Strategy
   ├── ID: string                    // Unique identifier (<name>-<uuid>)
   ├── Name: string                  // Strategy name (e.g., "moving_average")
   ├── Parameters: map[string]any    // Strategy configuration
   │   ├── symbol: string           // Trading symbol
   │   ├── period: int             // Time period for calculations
   │   └── threshold: float64      // Trading threshold
   ├── StartTime: time.Time         // When strategy started
   ├── StopTime: *time.Time         // When strategy stopped (nil if active)
   └── Status: string               // "active" or "stopped"

2. Object Lifecycle:
   a. Creation:
      1. Client sends strategy name and parameters
      2. NewStrategy generates unique ID
      3. Sets start time and active status
      4. Returns new instance

   b. Operation:
      1. Runner uses Parameters for trading decisions
      2. Status indicates if strategy is running
      3. ID used for lookups and references

   c. Stopping:
      1. Client requests stop by ID
      2. Stop() sets stop time
      3. Updates status to stopped

3. Example Usage:
   strategy := NewStrategy("moving_average", map[string]interface{}{
       "symbol": "AAPL",
       "period": 20,
       "threshold": 0.02,
   })

   // Later...
   strategy.Stop()

4. Error Handling:
   - Custom StrategyError type
   - Predefined error codes
   - Human-readable messages
*/

// Strategy represents a trading strategy instance
type Strategy struct {
	ID         string                 `json:"id"`          // Unique identifier (format: <name>-<uuid>)
	Name       string                 `json:"name"`        // Strategy name
	Parameters map[string]interface{} `json:"parameters"`  // Strategy parameters
	StartTime  time.Time             `json:"start_time"`  // When strategy started
	StopTime   *time.Time            `json:"stop_time"`   // When strategy stopped (nil if active)
	Status     string                `json:"status"`      // "active" or "stopped"
}

// NewStrategy creates a new strategy instance
func NewStrategy(name string, params map[string]interface{}) *Strategy {
	return &Strategy{
		ID:         fmt.Sprintf("%s-%s", name, uuid.New().String()),
		Name:       name,
		Parameters: params,
		StartTime:  time.Now(),
		Status:     "active",
	}
}

// Stop marks the strategy as stopped
func (s *Strategy) Stop() {
	now := time.Now()
	s.StopTime = &now
	s.Status = "stopped"
}

// StrategyError represents strategy-related errors
type StrategyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *StrategyError) Error() string {
	return e.Message
}

// Error codes
const (
	ErrStrategyNotFound = "STRATEGY_NOT_FOUND"
	ErrAlreadyStopped  = "ALREADY_STOPPED"
	ErrInvalidStrategy = "INVALID_STRATEGY"
)

// Request/Response types
type StartStrategyRequest struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type StartStrategyResponse struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
}

type StopStrategyRequest struct {
	ID string `json:"id"`
}

type StopStrategyResponse struct {
	ID        string     `json:"id"`
	StartTime time.Time  `json:"start_time"`
	StopTime  time.Time  `json:"stop_time"`
	Status    string     `json:"status"`
}

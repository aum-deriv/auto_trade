package strategy

import (
	"fmt"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
)

/*
Repeat Strategy Flow and Structure:

1. Memory Structure:
   RepeatStrategy
   ├── runner: *DefaultRunner       // For executing trades
   ├── symbol: string              // Trading symbol
   ├── entryPrice: float64        // Buy when price <= this
   ├── exitPrice: float64         // Sell when price >= this
   ├── currentTrade: *models.Trade // Track current position
   └── mu: sync.Mutex             // Protects currentTrade

2. Operation Flow:
   a. No Position:
      IF price <= entryPrice
         Execute buy
         Store trade ID

   b. Has Position:
      IF price >= exitPrice
         Execute sell
         Clear trade ID
         Ready for next cycle

3. Parameters:
   {
       "symbol": "AAPL",
       "entry_price": 150.0,
       "exit_price": 155.0
   }

4. Error Handling:
   - Invalid parameters
   - Trade execution errors
   - Missing fields
*/

// RepeatStrategy implements a simple repeating buy/sell strategy
type RepeatStrategy struct {
	runner       *DefaultRunner
	symbol       string
	entryPrice   float64
	exitPrice    float64
	currentTrade *models.Trade
	mu           sync.Mutex
}

// NewRepeatStrategy creates a new repeat strategy instance
func NewRepeatStrategy(runner *DefaultRunner, params map[string]interface{}) (StrategyExecutor, error) {
	// Extract and validate symbol
	symbol, ok := params["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("invalid or missing symbol parameter")
	}

	// Extract and validate entry price
	entryPrice, ok := params["entry_price"].(float64)
	if !ok || entryPrice <= 0 {
		return nil, fmt.Errorf("invalid or missing entry_price parameter")
	}

	// Extract and validate exit price
	exitPrice, ok := params["exit_price"].(float64)
	if !ok || exitPrice <= entryPrice {
		return nil, fmt.Errorf("invalid or missing exit_price parameter (must be greater than entry_price)")
	}

	return &RepeatStrategy{
		runner:     runner,
		symbol:     symbol,
		entryPrice: entryPrice,
		exitPrice:  exitPrice,
	}, nil
}

// ProcessTick implements the StrategyExecutor interface
func (s *RepeatStrategy) ProcessTick(tick *models.Tick) error {
	// Ignore ticks for other symbols
	if tick.Symbol != s.symbol {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for buy condition
	if s.currentTrade == nil && tick.Price <= s.entryPrice {
		trade, err := s.runner.executeBuy(s.symbol, tick.Price)
		if err != nil {
			return fmt.Errorf("failed to execute buy: %w", err)
		}
		s.currentTrade = trade
		return nil
	}

	// Check for sell condition
	if s.currentTrade != nil && tick.Price >= s.exitPrice {
		_, err := s.runner.executeSell(s.currentTrade.ID)
		if err != nil {
			return fmt.Errorf("failed to execute sell: %w", err)
		}
		s.currentTrade = nil // Ready for next cycle
		return nil
	}

	return nil
}

// init registers the repeat strategy with the registry
func init() {
	defaultRegistry.Register("repeat", NewRepeatStrategy)
}

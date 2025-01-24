package strategy

import (
	"fmt"
	"log"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/models"
)

/*
Martingale Strategy Flow and Structure:

1. Memory Structure:
   MartingaleStrategy
   ├── runner: *DefaultRunner      // For executing trades
   ├── symbol: string             // Trading symbol
   ├── basePosition: float64      // Initial position size
   ├── takeProfit: float64        // Profit target percentage
   ├── maxPositions: int          // Max position increases
   ├── currentTrade: *models.Trade // Current position
   ├── positionCount: int         // Number of positions taken
   ├── currentSize: float64       // Current position size
   └── mu: sync.Mutex            // Protects shared state

2. Operation Flow:
   a. No Position:
      - Set position size (base or doubled)
      - Execute buy at market price
      - Store trade ID
      - Increment position count

   b. Has Position:
      IF price >= entry * (1 + takeProfit/100)
         - Execute sell
         - Reset position size to base
         - Reset position count
         - Clear trade ID
      ELSE IF price < entry
         - Execute sell (loss)
         - IF positionCount < maxPositions
            * Double position size
         ELSE
            * Reset to base position
            * Reset position count
         - Clear trade ID

3. Parameters:
   {
       "symbol": "AAPL",
       "base_position": 100.0,
       "take_profit": 1.0,
       "max_positions": 3
   }

4. Error Handling:
   - Invalid parameters
   - Trade execution errors
   - Missing fields
   - Position size validation
*/

// MartingaleStrategy implements the Martingale trading strategy
type MartingaleStrategy struct {
	runner       *DefaultRunner
	symbol       string
	basePosition float64
	takeProfit   float64
	maxPositions int
	currentTrade *models.Trade
	positionCount int
	currentSize  float64
	mu           sync.Mutex
}

// NewMartingaleStrategy creates a new Martingale strategy instance
func NewMartingaleStrategy(runner *DefaultRunner, params map[string]interface{}) (StrategyExecutor, error) {
	// Extract and validate symbol
	symbol, ok := params["symbol"].(string)
	if !ok || symbol == "" {
		return nil, fmt.Errorf("invalid or missing symbol parameter")
	}

	// Extract and validate base_position
	basePosition, ok := params["base_position"].(float64)
	if !ok || basePosition <= 0 {
		return nil, fmt.Errorf("invalid or missing base_position parameter")
	}

	// Extract and validate take_profit
	takeProfit, ok := params["take_profit"].(float64)
	if !ok || takeProfit <= 0 {
		return nil, fmt.Errorf("invalid or missing take_profit parameter")
	}

	// Extract and validate max_positions
	maxPositions, ok := params["max_positions"].(float64)
	if !ok || maxPositions < 1 {
		return nil, fmt.Errorf("invalid or missing max_positions parameter")
	}

	return &MartingaleStrategy{
		runner:       runner,
		symbol:       symbol,
		basePosition: basePosition,
		takeProfit:   takeProfit,
		maxPositions: int(maxPositions),
		currentSize:  basePosition,
		positionCount: 0,
	}, nil
}

// ProcessTick implements the StrategyExecutor interface
func (s *MartingaleStrategy) ProcessTick(tick *models.Tick) error {
	// Ignore ticks for other symbols
	if tick.Symbol != s.symbol {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Enter new position if none exists
	if s.currentTrade == nil {
		// Calculate quantity based on current position size
		quantity := s.currentSize / tick.Price
		trade, err := s.runner.executeBuy(s.symbol, tick.Price)
		if err != nil {
			return fmt.Errorf("failed to execute buy: %w", err)
		}
		s.currentTrade = trade
		s.positionCount++
		log.Printf("Opened position %d: Size=%.2f, Quantity=%.4f, Price=%.2f", 
			s.positionCount, s.currentSize, quantity, tick.Price)
		return nil
	}

	// Calculate take profit target
	entryPrice := s.currentTrade.EntryPrice
	targetPrice := entryPrice * (1 + s.takeProfit/100)

	// Check for take profit
	if tick.Price >= targetPrice {
		_, err := s.runner.executeSell(s.currentTrade.ID)
		if err != nil {
			return fmt.Errorf("failed to execute take profit sell: %w", err)
		}
		// Calculate profit/loss
		quantity := s.currentSize / s.currentTrade.EntryPrice
		profit := (tick.Price - s.currentTrade.EntryPrice) * quantity
		
		// Reset for next cycle
		s.currentTrade = nil
		s.currentSize = s.basePosition
		s.positionCount = 0
		log.Printf("Take profit: Profit=%.2f", profit)
		return nil
	}

	// Check for loss exit
	if tick.Price < entryPrice {
		_, err := s.runner.executeSell(s.currentTrade.ID)
		if err != nil {
			return fmt.Errorf("failed to execute loss sell: %w", err)
		}
		
		// Calculate loss
		quantity := s.currentSize / s.currentTrade.EntryPrice
		loss := (tick.Price - s.currentTrade.EntryPrice) * quantity
		
		// Prepare next position size
		if s.positionCount < s.maxPositions {
			s.currentSize *= 2
			log.Printf("Loss=%.2f, Doubling position size to %.2f", loss, s.currentSize)
		} else {
			s.currentSize = s.basePosition
			s.positionCount = 0
			log.Printf("Loss=%.2f, Max positions reached, resetting to base position %.2f", 
				loss, s.basePosition)
		}
		
		s.currentTrade = nil
		return nil
	}

	return nil
}

// Metadata for the Martingale strategy
var martingaleMetadata = models.StrategyMetadata{
	Name: "martingale",
	Parameters: []models.ParameterInfo{
		{
			Name:        "symbol",
			Type:        "string",
			Required:    true,
			Description: "Trading symbol (e.g. AAPL)",
		},
		{
			Name:        "base_position",
			Type:        "number",
			Required:    true,
			Description: "Initial position size in dollars",
		},
		{
			Name:        "take_profit",
			Type:        "number",
			Required:    true,
			Description: "Price increase percentage for taking profit (e.g. 1.0 for 1%)",
		},
		{
			Name:        "max_positions",
			Type:        "number",
			Required:    true,
			Description: "Maximum number of increasing positions allowed",
		},
	},
	Flow: []string{
		"1. Start with base_position size",
		"2. Enter long position at market price",
		"3. Set take profit target at entry_price * (1 + take_profit/100)",
		"4. If target hit: Take profit and reset position size to base_position",
		"5. If price drops: Exit at loss",
		"6. If under max_positions: Double position size and enter new position",
		"7. If at max_positions: Reset position size to base_position",
		"8. Repeat from step 1",
	},
}

// init registers the Martingale strategy with the registry
func init() {
	defaultRegistry.Register("martingale", NewMartingaleStrategy, martingaleMetadata)
}

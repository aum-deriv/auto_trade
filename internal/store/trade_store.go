package store

import "github.com/aumbhatt/auto_trade/internal/models"

/*
Trade Store Interface and Flow:

1. Interface Structure:
   TradeStore combines:
   ├── Basic trade operations
   │   ├── CreateTrade
   │   ├── CloseTrade
   │   ├── GetOpenTrades
   │   └── GetTradeHistory
   └── Event emission (from TradeEventEmitter)
       ├── AddListener
       └── RemoveListener

2. Usage Flow:
   a. Create Trade:
      symbol, price → CreateTrade() → Trade
      1. Generate trade ID
      2. Create trade object
      3. Store in open trades
      4. Emit trade created event
      5. Return trade

   b. Close Trade:
      id → CloseTrade() → Trade
      1. Find trade in open trades
      2. Add exit details
      3. Move to trade history
      4. Emit trade closed event
      5. Return updated trade

   c. Get Open Trades:
      GetOpenTrades() → []*Trade
      1. Return all open trades

   d. Get Trade History:
      GetTradeHistory() → []*Trade
      1. Return all closed trades

3. Future Extensions:
   - Add database persistence
   - Add filtering/pagination
   - Add trade updates
   - Add batch operations
*/

// BasicTradeStore defines the core trade operations
type BasicTradeStore interface {
	// CreateTrade creates a new trade with given symbol and entry price
	CreateTrade(symbol string, entryPrice float64) (*models.Trade, error)

	// CloseTrade closes an existing trade
	CloseTrade(id string) (*models.Trade, error)

	// GetOpenTrades returns all open trades
	GetOpenTrades() ([]*models.Trade, error)

	// GetTradeHistory returns all closed trades
	GetTradeHistory() ([]*models.Trade, error)
}

// TradeStore combines basic trade operations with event emission capabilities
type TradeStore interface {
	BasicTradeStore
	TradeEventEmitter
}

package models

import (
	"fmt"
	"time"
)

/*
Trade Model Flow and Memory Structure:

1. Memory Structure:
   Trade
   ├── ID: string        // Format: "trade-{uuid}"
   ├── Symbol: string    // Trading symbol (e.g., "AAPL")
   ├── EntryPrice: float64
   ├── ExitPrice: float64 (optional)
   ├── EntryTime: time.Time
   └── ExitTime: time.Time (optional)

2. Data Flow:
   a. Buy Trade:
      Request → Trade (with entry details) → Response
      Example:
      Request: {"symbol": "AAPL", "entry_price": 150.25}
      Internal: Trade{ID: "trade-abc", Symbol: "AAPL", ...}
      Response: {"trade_id": "trade-abc", ...}

   b. Sell Trade:
      Request → Update Trade (add exit details) → Response
      Example:
      Request: {"trade_id": "trade-abc"}
      Internal: Trade{ExitPrice: 151.00, ExitTime: now}
      Response: Complete trade details

3. Error Handling:
   - Invalid symbol
   - Invalid price
   - Trade not found
   - Trade already closed
*/

// Trade represents a trading position
type Trade struct {
	ID         string     `json:"trade_id"`
	Symbol     string     `json:"symbol"`
	EntryPrice float64    `json:"entry_price"`
	ExitPrice  float64    `json:"exit_price,omitempty"`
	EntryTime  time.Time  `json:"entry_time"`
	ExitTime   time.Time  `json:"exit_time,omitempty"`
}

// TradeError represents trading-related errors
type TradeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *TradeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	// Buy/Sell errors
	ErrInvalidSymbol      = "INVALID_SYMBOL"
	ErrInvalidEntryPrice  = "INVALID_ENTRY_PRICE"
	ErrTradeCreation      = "TRADE_CREATION_FAILED"
	ErrTradeNotFound      = "TRADE_NOT_FOUND"
	ErrTradeAlreadyClosed = "TRADE_ALREADY_CLOSED"
	ErrTradeClosing       = "TRADE_CLOSING_FAILED"

	// Open Positions errors
	ErrOpenPositionsFetch    = "OPEN_POSITIONS_FETCH_FAILED"
	ErrOpenPositionsEmpty    = "NO_OPEN_POSITIONS"
	ErrOpenPositionsInternal = "OPEN_POSITIONS_INTERNAL_ERROR"

	// Trade History errors
	ErrTradeHistoryFetch    = "TRADE_HISTORY_FETCH_FAILED"
	ErrTradeHistoryEmpty    = "NO_TRADE_HISTORY"
	ErrTradeHistoryInternal = "TRADE_HISTORY_INTERNAL_ERROR"
)

// CreateTradeRequest represents the request body for creating a trade
type CreateTradeRequest struct {
	Symbol     string  `json:"symbol"`
	EntryPrice float64 `json:"entry_price"`
}

// CloseTradeRequest represents the request body for closing a trade
type CloseTradeRequest struct {
	TradeID string `json:"trade_id"`
}

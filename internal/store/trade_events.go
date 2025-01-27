package store

import "github.com/aumbhatt/auto_trade/internal/models"

/*
Trade Events Flow and Structure:

1. Components:
   ├── TradeEvent: Event data structure
   ├── TradeEventType: Event type constants
   ├── TradeEventListener: Observer interface
   └── TradeEventEmitter: Event emitter interface

2. Event Flow:
   a. Trade Creation:
      1. Store creates/updates trade
      2. Store emits event
      3. Listeners receive event
      4. Listeners update their state
      5. Listeners notify their subscribers

3. Event Types:
   - TradeCreated: New trade opened
   - TradeClosed: Existing trade closed

4. Usage Example:
   store.AddListener(openPositionsHandler)
   store.CreateTrade(...) // Triggers event
   // Listener automatically updates and broadcasts
*/

// TradeEventType defines the type of trade event
type TradeEventType string

const (
	// TradeCreated indicates a new trade was opened
	TradeCreated TradeEventType = "created"
	
	// TradeClosed indicates an existing trade was closed
	TradeClosed TradeEventType = "closed"
)

// TradeEvent represents a trade-related event
type TradeEvent struct {
	Type  TradeEventType // Type of event
	Trade *models.Trade  // Associated trade
}

// TradeEventListener defines interface for objects that want to receive trade events
type TradeEventListener interface {
	// OnTradeEvent is called when a trade event occurs
	OnTradeEvent(event TradeEvent)
}

// TradeEventEmitter defines interface for objects that emit trade events
type TradeEventEmitter interface {
	// AddListener registers a new listener
	AddListener(listener TradeEventListener)
	
	// RemoveListener unregisters a trade event listener
	RemoveListener(listener TradeEventListener)
}

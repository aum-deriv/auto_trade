package memory

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/aumbhatt/auto_trade/internal/store"
	"github.com/google/uuid"
)

/*
In-Memory Trade Store Flow and Structure:

1. Memory Structure:
   InMemoryTradeStore
   ├── openTrades: map[string]*Trade    // Active trades
   ├── tradeHistory: map[string]*Trade  // Closed trades
   ├── listeners: []TradeEventListener  // Event observers
   └── mu: sync.RWMutex                // Protects maps and listeners

2. Data Organization:
   openTrades = {
     "trade-abc": Trade{ID: "trade-abc", Symbol: "AAPL", ...},
     "trade-def": Trade{ID: "trade-def", Symbol: "GOOGL", ...}
   }
   tradeHistory = {
     "trade-xyz": Trade{ID: "trade-xyz", Symbol: "MSFT", ...}
   }

3. Operation Flow:
   a. Create Trade:
      1. Generate UUID
      2. Create trade object
      3. Store in openTrades
      4. Emit TradeCreated event
      5. Return trade

   b. Close Trade:
      1. Find in openTrades
      2. Add exit details
      3. Move to tradeHistory
      4. Emit TradeClosed event
      5. Return updated trade

4. Event Handling:
   - AddListener registers new observers
   - RemoveListener unregisters observers
   - emitEvent notifies all observers
   - Events emitted after state changes
   - Listeners notified outside locks

5. Concurrency:
   - RWMutex for map access
   - Read operations use RLock
   - Write operations use Lock
   - Thread-safe event emission
*/

// InMemoryTradeStore implements store.TradeStore interface with in-memory storage
type InMemoryTradeStore struct {
	openTrades   map[string]*models.Trade
	tradeHistory map[string]*models.Trade
	listeners    []store.TradeEventListener
	mu           sync.RWMutex
}

// NewInMemoryTradeStore creates a new instance of InMemoryTradeStore
func NewInMemoryTradeStore() *InMemoryTradeStore {
	return &InMemoryTradeStore{
		openTrades:   make(map[string]*models.Trade),
		tradeHistory: make(map[string]*models.Trade),
		listeners:    make([]store.TradeEventListener, 0),
	}
}

// AddListener implements store.TradeEventEmitter
func (s *InMemoryTradeStore) AddListener(listener store.TradeEventListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

// RemoveListener implements store.TradeEventEmitter
func (s *InMemoryTradeStore) RemoveListener(listener store.TradeEventListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Find and remove the listener
	for i, l := range s.listeners {
		if l == listener {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
}

// emitEvent notifies all listeners of a trade event
func (s *InMemoryTradeStore) emitEvent(event store.TradeEvent) {
	s.mu.RLock()
	listeners := make([]store.TradeEventListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.RUnlock()

	// Notify listeners outside the lock to prevent deadlocks
	for _, listener := range listeners {
		listener.OnTradeEvent(event)
	}
}

// CreateTrade implements store.BasicTradeStore
func (s *InMemoryTradeStore) CreateTrade(symbol string, entryPrice float64) (*models.Trade, error) {
	s.mu.Lock()

	trade := &models.Trade{
		ID:         fmt.Sprintf("trade-%s", uuid.New().String()),
		Symbol:     symbol,
		EntryPrice: entryPrice,
		EntryTime:  time.Now(),
	}

	s.openTrades[trade.ID] = trade
	log.Printf("Trade opened: %s", trade.ID)
	
	// Make a copy of trade data for the event
	tradeCopy := *trade
	
	// Release lock before emitting event
	s.mu.Unlock()
	
	// Notify listeners with copied data
	s.emitEvent(store.TradeEvent{
		Type:  store.TradeCreated,
		Trade: &tradeCopy,
	})
	
	return trade, nil
}

// CloseTrade implements store.BasicTradeStore
func (s *InMemoryTradeStore) CloseTrade(id string) (*models.Trade, error) {
	s.mu.Lock()

	trade, exists := s.openTrades[id]
	if !exists {
		s.mu.Unlock()
		return nil, &models.TradeError{
			Code:    models.ErrTradeNotFound,
			Message: fmt.Sprintf("Trade not found: %s", id),
		}
	}

	// Check if already closed
	if !trade.ExitTime.IsZero() {
		s.mu.Unlock()
		return nil, &models.TradeError{
			Code:    models.ErrTradeAlreadyClosed,
			Message: fmt.Sprintf("Trade already closed: %s", id),
		}
	}

	// Close the trade
	trade.ExitTime = time.Now()
	trade.ExitPrice = trade.EntryPrice + 1 // Mock exit price for demo

	// Move to history
	delete(s.openTrades, id)
	s.tradeHistory[id] = trade

	log.Printf("Trade closed: %s", trade.ID)
	
	// Make a copy of trade data for the event
	tradeCopy := *trade
	
	// Release lock before emitting event
	s.mu.Unlock()
	
	// Notify listeners with copied data
	s.emitEvent(store.TradeEvent{
		Type:  store.TradeClosed,
		Trade: &tradeCopy,
	})
	
	return trade, nil
}

// GetOpenTrades implements store.BasicTradeStore
func (s *InMemoryTradeStore) GetOpenTrades() ([]*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trades := make([]*models.Trade, 0, len(s.openTrades))
	for _, trade := range s.openTrades {
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradeHistory implements store.BasicTradeStore
func (s *InMemoryTradeStore) GetTradeHistory() ([]*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trades := make([]*models.Trade, 0, len(s.tradeHistory))
	for _, trade := range s.tradeHistory {
		trades = append(trades, trade)
	}

	return trades, nil
}

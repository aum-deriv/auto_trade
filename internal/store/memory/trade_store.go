package memory

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aumbhatt/auto_trade/internal/models"
	"github.com/google/uuid"
)

/*
In-Memory Trade Store Flow and Structure:

1. Memory Structure:
   InMemoryTradeStore
   ├── openTrades: map[string]*Trade    // Active trades
   ├── tradeHistory: map[string]*Trade  // Closed trades
   └── mu: sync.RWMutex                // Protects both maps

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
      4. Return trade

   b. Close Trade:
      1. Find in openTrades
      2. Add exit details
      3. Move to tradeHistory
      4. Return updated trade

4. Concurrency:
   - RWMutex for map access
   - Read operations use RLock
   - Write operations use Lock
*/

// InMemoryTradeStore implements TradeStore interface with in-memory storage
type InMemoryTradeStore struct {
	openTrades   map[string]*models.Trade
	tradeHistory map[string]*models.Trade
	mu           sync.RWMutex
}

// NewInMemoryTradeStore creates a new instance of InMemoryTradeStore
func NewInMemoryTradeStore() *InMemoryTradeStore {
	return &InMemoryTradeStore{
		openTrades:   make(map[string]*models.Trade),
		tradeHistory: make(map[string]*models.Trade),
	}
}

// CreateTrade creates a new trade with given symbol and entry price
func (s *InMemoryTradeStore) CreateTrade(symbol string, entryPrice float64) (*models.Trade, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade := &models.Trade{
		ID:         fmt.Sprintf("trade-%s", uuid.New().String()),
		Symbol:     symbol,
		EntryPrice: entryPrice,
		EntryTime:  time.Now(),
	}

	s.openTrades[trade.ID] = trade
	log.Printf("Trade opened: %s", trade.ID)
	return trade, nil
}

// CloseTrade closes an existing trade
func (s *InMemoryTradeStore) CloseTrade(id string) (*models.Trade, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade, exists := s.openTrades[id]
	if !exists {
		return nil, &models.TradeError{
			Code:    models.ErrTradeNotFound,
			Message: fmt.Sprintf("Trade not found: %s", id),
		}
	}

	// Check if already closed
	if !trade.ExitTime.IsZero() {
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
	return trade, nil
}

// GetOpenTrades returns all open trades
func (s *InMemoryTradeStore) GetOpenTrades() ([]*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trades := make([]*models.Trade, 0, len(s.openTrades))
	for _, trade := range s.openTrades {
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradeHistory returns all closed trades
func (s *InMemoryTradeStore) GetTradeHistory() ([]*models.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trades := make([]*models.Trade, 0, len(s.tradeHistory))
	for _, trade := range s.tradeHistory {
		trades = append(trades, trade)
	}

	return trades, nil
}

package trading

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TradeType string

const (
	Buy  TradeType = "BUY"
	Sell TradeType = "SELL"
)

type Trade struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Type      TradeType `json:"type"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type TradeManager struct {
	trades     map[string]*Trade // All trades
	openTrades map[string]*Trade // Only open trades
	mutex      sync.RWMutex
}

func NewTradeManager() *TradeManager {
	return &TradeManager{
		trades:     make(map[string]*Trade),
		openTrades: make(map[string]*Trade),
	}
}

func (tm *TradeManager) PlaceBuyOrder(symbol string, price float64) (*Trade, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tradeID := fmt.Sprintf("TRADE-%s", uuid.New().String())

	trade := &Trade{
		ID:        tradeID,
		Symbol:    symbol,
		Type:      Buy,
		Price:     price,
		Status:    "OPEN",
		Timestamp: time.Now(),
	}

	tm.trades[tradeID] = trade
	tm.openTrades[tradeID] = trade
	return trade, nil
}

func (tm *TradeManager) CloseTrade(tradeID string, closePrice float64) (*Trade, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	trade, exists := tm.openTrades[tradeID]
	if !exists {
		return nil, fmt.Errorf("no open trade found with ID: %s", tradeID)
	}

	trade.Status = "CLOSED"
	delete(tm.openTrades, tradeID)

	return trade, nil
}

func (tm *TradeManager) GetTrade(id string) (*Trade, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	trade, exists := tm.trades[id]
	if !exists {
		return nil, fmt.Errorf("trade not found: %s", id)
	}

	return trade, nil
}

func (tm *TradeManager) GetOpenTrades() []*Trade {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	trades := make([]*Trade, 0, len(tm.openTrades))
	for _, trade := range tm.openTrades {
		trades = append(trades, trade)
	}
	return trades
}

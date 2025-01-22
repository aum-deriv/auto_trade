package strategies

import (
	"fmt"
	"math"

	"auto_trade/trading"

	"github.com/google/uuid"
)

// StrategyType represents the type of trading strategy
type StrategyType string

const (
	MartingaleType StrategyType = "MARTINGALE"
)

// MartingaleStrategy represents a Martingale trading strategy instance
type MartingaleStrategy struct {
	StrategyID  string       `json:"strategy_id"`  // Unique identifier for sharing and tracking
	Type        StrategyType `json:"type"`         // Type of strategy (MARTINGALE)
	BaseSize    float64      `json:"base_size"`    // Initial position size
	Symbol      string       `json:"symbol"`       // Trading pair symbol
	MaxLosses   int         `json:"max_losses"`   // Maximum consecutive losses allowed
	
	currentLosses int     // Current streak of losses
	currentSize   float64 // Current position size
}

// NewMartingaleStrategy creates a new instance of the Martingale strategy
func NewMartingaleStrategy(baseSize float64, symbol string, maxLosses int) *MartingaleStrategy {
	// Create a unique strategy ID
	strategyID := fmt.Sprintf("MART-%s", uuid.New().String())
	
	return &MartingaleStrategy{
		StrategyID:   strategyID,
		Type:         MartingaleType,
		BaseSize:     baseSize,
		Symbol:       symbol,
		MaxLosses:    maxLosses,
		currentSize:  baseSize,
		currentLosses: 0,
	}
}

// GetParameters returns the strategy parameters for sharing or persistence
func (ms *MartingaleStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"strategy_id": ms.StrategyID,
		"type":       ms.Type,
		"base_size":  ms.BaseSize,
		"symbol":     ms.Symbol,
		"max_losses": ms.MaxLosses,
	}
}

// ProcessTrade handles the outcome of a trade and calculates the next position size
func (ms *MartingaleStrategy) ProcessTrade(trade *trading.Trade, closePrice float64) (nextSize float64, shouldTrade bool, err error) {
	// Calculate if the trade was profitable
	var profitable bool
	if trade.Type == trading.Buy {
		profitable = closePrice > trade.Price
	} else {
		profitable = closePrice < trade.Price
	}

	if profitable {
		// Reset on win
		ms.currentLosses = 0
		ms.currentSize = ms.BaseSize
	} else {
		// Double size on loss
		ms.currentLosses++
		if ms.currentLosses >= ms.MaxLosses {
			return 0, false, fmt.Errorf("maximum consecutive losses (%d) reached", ms.MaxLosses)
		}
		ms.currentSize = ms.BaseSize * math.Pow(2, float64(ms.currentLosses))
	}

	return ms.currentSize, true, nil
}

// GetCurrentSize returns the current position size
func (ms *MartingaleStrategy) GetCurrentSize() float64 {
	return ms.currentSize
}

// GetConsecutiveLosses returns the current number of consecutive losses
func (ms *MartingaleStrategy) GetConsecutiveLosses() int {
	return ms.currentLosses
}

// Reset resets the strategy to its initial state
func (ms *MartingaleStrategy) Reset() {
	ms.currentLosses = 0
	ms.currentSize = ms.BaseSize
}

package mock

import (
	"math/rand"
	"time"

	"github.com/aumbhatt/auto_trade/internal/models"
)

// MockTickSource implements TickSource interface with mock data
type MockTickSource struct {
	symbols []string
}

// NewMockTickSource creates a new instance of MockTickSource
func NewMockTickSource() *MockTickSource {
	return &MockTickSource{
		symbols: []string{"AAPL", "GOOGL", "MSFT", "AMZN"},
	}
}

// GetTick generates and returns mock tick data
func (s *MockTickSource) GetTick() (*models.Tick, error) {
	symbol := s.symbols[rand.Intn(len(s.symbols))]
	
	return &models.Tick{
		Symbol:    symbol,
		Price:     100 + rand.Float64()*900, // Random price between 100 and 1000
		Volume:    rand.Int63n(10000),       // Random volume between 0 and 9999
		Timestamp: time.Now(),
	}, nil
}

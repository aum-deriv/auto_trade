package ticks

import (
	"encoding/json"
	"math/rand"
	"time"
)

type Tick struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type TickGenerator struct {
	symbols []string
	prices  map[string]float64
}

func NewTickGenerator() *TickGenerator {
	symbols := []string{"BTC/USD", "ETH/USD", "SOL/USD"}
	prices := map[string]float64{
		"BTC/USD": 40000.0,
		"ETH/USD": 2500.0,
		"SOL/USD": 100.0,
	}

	return &TickGenerator{
		symbols: symbols,
		prices:  prices,
	}
}

func (g *TickGenerator) GenerateTick() []byte {
	symbol := g.symbols[rand.Intn(len(g.symbols))]
	currentPrice := g.prices[symbol]

	// Generate random price movement (-0.5% to +0.5%)
	priceChange := currentPrice * (rand.Float64()*0.01 - 0.005)
	newPrice := currentPrice + priceChange
	g.prices[symbol] = newPrice

	tick := Tick{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    rand.Float64() * 100,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(tick)
	return data
}

func (g *TickGenerator) StartGeneration(broadcast chan<- []byte) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		broadcast <- g.GenerateTick()
	}
}

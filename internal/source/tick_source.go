package source

import "github.com/aumbhatt/auto_trade/internal/models"

// TickSource defines the interface for getting tick data
type TickSource interface {
	GetTick() (*models.Tick, error)
}

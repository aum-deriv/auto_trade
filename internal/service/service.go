package service

import (
	"github.com/aumbhatt/auto_trade/internal/config"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

// Service represents the main business logic of the application
type Service struct {
	config *config.Config
	hub    *websocket.Hub
}

// NewService creates a new instance of Service
func NewService(cfg *config.Config, hub *websocket.Hub) *Service {
	return &Service{
		config: cfg,
		hub:    hub,
	}
}

// Run starts the service
func (s *Service) Run() error {
	// Service is now ready to use the hub for broadcasting messages
	// You can add any service-specific initialization here
	return nil
}

// BroadcastMessage sends a message to all subscribed clients
func (s *Service) BroadcastMessage(msg websocket.Message) {
	s.hub.Broadcast(msg)
}

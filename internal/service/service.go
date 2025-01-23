package service

import "github.com/aumbhatt/auto_trade/internal/config"

// Service represents the main business logic of the application
type Service struct {
	config *config.Config
}

// NewService creates a new instance of Service
func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

// Run starts the service
func (s *Service) Run() error {
	// Basic implementation using config
	return nil
}

package config

import "time"

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig
	App    AppConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"logLevel"`
}

// NewDefaultConfig returns a Config instance with default values
func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  time.Second * 15,
			WriteTimeout: time.Second * 15,
		},
		App: AppConfig{
			Environment: "development",
			LogLevel:    "info",
		},
	}
}

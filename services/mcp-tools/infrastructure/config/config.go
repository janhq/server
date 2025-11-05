package config

import (
	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the MCP Tools service
type Config struct {
	HTTPPort     string `env:"HTTP_PORT" envDefault:"8091"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat    string `env:"LOG_FORMAT" envDefault:"json"` // json or console
	SerperAPIKey string `env:"SERPER_API_KEY"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the MCP Tools service
type Config struct {
	HTTPPort                     string   `env:"HTTP_PORT" envDefault:"8091"`
	LogLevel                     string   `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat                    string   `env:"LOG_FORMAT" envDefault:"json"` // json or console
	SerperAPIKey                 string   `env:"SERPER_API_KEY"`
	SearchEngine                 string   `env:"SEARCH_ENGINE" envDefault:"serper"`
	SearxngURL                   string   `env:"SEARXNG_URL" envDefault:"http://searxng:8080"`
	SerperDomainFilter           []string `env:"SERPER_DOMAIN_FILTER" envSeparator:","`
	SerperLocationHint           string   `env:"SERPER_LOCATION_HINT"`
	SerperOfflineMode            bool     `env:"SERPER_OFFLINE_MODE" envDefault:"false"`
	VectorStoreURL               string   `env:"VECTOR_STORE_URL" envDefault:"http://vector-store-mcp:3015"`
	SandboxFusionURL             string   `env:"SANDBOX_FUSION_URL" envDefault:"http://sandbox-fusion:8080"`
	SandboxFusionRequireApproval bool     `env:"SANDBOX_FUSION_REQUIRE_APPROVAL" envDefault:"false"`
	AuthEnabled                  bool     `env:"AUTH_ENABLED" envDefault:"false"`
	AuthIssuer                   string   `env:"AUTH_ISSUER"`
	AuthAudience                 string   `env:"AUTH_AUDIENCE"`
	AuthJWKSURL                  string   `env:"AUTH_JWKS_URL"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if cfg.AuthEnabled {
		if strings.TrimSpace(cfg.AuthIssuer) == "" {
			return nil, fmt.Errorf("AUTH_ISSUER is required when AUTH_ENABLED is true")
		}
		if strings.TrimSpace(cfg.AuthAudience) == "" {
			return nil, fmt.Errorf("AUTH_AUDIENCE is required when AUTH_ENABLED is true")
		}
		if strings.TrimSpace(cfg.AuthJWKSURL) == "" {
			return nil, fmt.Errorf("AUTH_JWKS_URL is required when AUTH_ENABLED is true")
		}
	}
	return cfg, nil
}

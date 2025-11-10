package config

import "github.com/caarlos0/env/v11"

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
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

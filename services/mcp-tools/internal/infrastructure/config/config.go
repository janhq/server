package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the MCP Tools service
type Config struct {
	// HTTP Server - using MCP_TOOLS_ prefix to avoid collisions
	HTTPPort  string `env:"MCP_TOOLS_HTTP_PORT" envDefault:"8091"`
	LogLevel  string `env:"MCP_TOOLS_LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"MCP_TOOLS_LOG_FORMAT" envDefault:"json"` // json or console

	// Search Configuration
	SerperAPIKey       string   `env:"SERPER_API_KEY"`
	SearchEngine       string   `env:"MCP_SEARCH_ENGINE" envDefault:"serper"`
	SearxngURL         string   `env:"SEARXNG_URL" envDefault:"http://searxng:8080"`
	SerperDomainFilter []string `env:"SERPER_DOMAIN_FILTER" envSeparator:","`
	SerperLocationHint string   `env:"SERPER_LOCATION_HINT"`
	SerperOfflineMode  bool     `env:"SERPER_OFFLINE_MODE" envDefault:"false"`

	// Circuit Breaker Configuration
	SerperCBFailureThreshold int `env:"SERPER_CB_FAILURE_THRESHOLD" envDefault:"15"`
	SerperCBSuccessThreshold int `env:"SERPER_CB_SUCCESS_THRESHOLD" envDefault:"5"`
	SerperCBTimeout          int `env:"SERPER_CB_TIMEOUT" envDefault:"45"`
	SerperCBMaxHalfOpen      int `env:"SERPER_CB_MAX_HALF_OPEN" envDefault:"10"`

	// HTTP Client Performance
	SerperHTTPTimeout     int `env:"SERPER_HTTP_TIMEOUT" envDefault:"15"`
	SerperScrapeTimeout   int `env:"SERPER_SCRAPE_TIMEOUT" envDefault:"30"` // Separate longer timeout for scrape operations
	SerperMaxConnsPerHost int `env:"SERPER_MAX_CONNS_PER_HOST" envDefault:"50"`
	SerperMaxIdleConns    int `env:"SERPER_MAX_IDLE_CONNS" envDefault:"100"`
	SerperIdleConnTimeout int `env:"SERPER_IDLE_CONN_TIMEOUT" envDefault:"90"`

	// Retry Configuration
	SerperRetryMaxAttempts   int     `env:"SERPER_RETRY_MAX_ATTEMPTS" envDefault:"5"`
	SerperRetryInitialDelay  int     `env:"SERPER_RETRY_INITIAL_DELAY" envDefault:"250"`
	SerperRetryMaxDelay      int     `env:"SERPER_RETRY_MAX_DELAY" envDefault:"5000"`
	SerperRetryBackoffFactor float64 `env:"SERPER_RETRY_BACKOFF_FACTOR" envDefault:"1.5"`

	// Tool Result Token Limits - Controls maximum output size for MCP tool results
	MaxSnippetChars       int `env:"MCP_MAX_SNIPPET_CHARS" envDefault:"5000"`        // Max chars for search result snippets
	MaxScrapePreviewChars int `env:"MCP_MAX_SCRAPE_PREVIEW_CHARS" envDefault:"5000"` // Max chars for scrape text preview
	MaxScrapeTextChars    int `env:"MCP_MAX_SCRAPE_TEXT_CHARS" envDefault:"50000"`   // Max chars for full scrape text (approx 12.5k tokens)

	// External Services
	VectorStoreURL   string `env:"VECTOR_STORE_URL" envDefault:"http://vector-store-mcp:3015"`
	SandboxFusionURL string `env:"SANDBOXFUSION_URL" envDefault:"http://sandbox-fusion:8080"`
	MemoryToolsURL   string `env:"MEMORY_TOOLS_URL" envDefault:"http://memory-tools:8090"`

	// LLM-API configuration for tool call tracking
	LLMAPIBaseURL      string `env:"LLM_API_BASE_URL" envDefault:"http://llm-api:8080"`
	MCPTrackingEnabled bool   `env:"MCP_TRACKING_ENABLED" envDefault:"true"`

	// Sandbox Configuration
	SandboxFusionRequireApproval bool `env:"MCP_SANDBOX_REQUIRE_APPROVAL" envDefault:"false"`
	EnablePythonExec             bool `env:"MCP_ENABLE_PYTHON_EXEC" envDefault:"true"`
	EnableMemoryRetrieve         bool `env:"MCP_ENABLE_MEMORY_RETRIEVE" envDefault:"true"`
	EnableFileSearch             bool `env:"MCP_ENABLE_FILE_SEARCH" envDefault:"false"`
	EnableImageGenerate          bool `env:"MCP_ENABLE_IMAGE_GENERATE" envDefault:"true"`

	// Authentication
	AuthEnabled bool   `env:"AUTH_ENABLED" envDefault:"false"`
	AuthIssuer  string `env:"AUTH_ISSUER"`
	Account     string `env:"ACCOUNT"`
	AuthJWKSURL string `env:"AUTH_JWKS_URL"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if strings.TrimSpace(os.Getenv("MCP_TOOLS_LOG_LEVEL")) == "" {
		if global := strings.TrimSpace(os.Getenv("LOG_LEVEL")); global != "" {
			cfg.LogLevel = global
		}
	}
	if strings.TrimSpace(os.Getenv("MCP_TOOLS_LOG_FORMAT")) == "" {
		if global := strings.TrimSpace(os.Getenv("LOG_FORMAT")); global != "" {
			cfg.LogFormat = global
		}
	}
	if cfg.AuthEnabled {
		if strings.TrimSpace(cfg.AuthIssuer) == "" {
			return nil, fmt.Errorf("AUTH_ISSUER is required when AUTH_ENABLED is true")
		}
		if strings.TrimSpace(cfg.Account) == "" {
			return nil, fmt.Errorf("ACCOUNT is required when AUTH_ENABLED is true")
		}
		if strings.TrimSpace(cfg.AuthJWKSURL) == "" {
			return nil, fmt.Errorf("AUTH_JWKS_URL is required when AUTH_ENABLED is true")
		}
	}
	return cfg, nil
}

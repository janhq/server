package configs

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

var global *Config

type Config struct {
	HTTPPort int `env:"MEMORY_TOOLS_PORT" envDefault:"8090"`

	DatabaseURL string `env:"DATABASE_URL,notEmpty"`

	EmbeddingServiceURL     string        `env:"EMBEDDING_SERVICE_URL" envDefault:"http://localhost:8091"`
	EmbeddingCacheType      string        `env:"EMBEDDING_CACHE_TYPE" envDefault:"memory"`
	EmbeddingCacheTTL       time.Duration `env:"EMBEDDING_CACHE_TTL" envDefault:"1h"`
	EmbeddingCacheMaxSize   int           `env:"EMBEDDING_CACHE_MAX_SIZE" envDefault:"10000"`
	EmbeddingCacheRedisURL  string        `env:"EMBEDDING_CACHE_REDIS_URL" envDefault:"redis://redis:6379/3"`
	EmbeddingCacheKeyPrefix string        `env:"EMBEDDING_CACHE_KEY_PREFIX" envDefault:"emb:"`

	ValidateEmbedding        bool          `env:"VALIDATE_EMBEDDING_ON_START" envDefault:"true"`
	ValidateEmbeddingTimeout time.Duration `env:"VALIDATE_EMBEDDING_TIMEOUT" envDefault:"10s"`

	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`

	APIKey string `env:"MEMORY_TOOLS_API_KEY"`

	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"console"`

	MigrationsDir string `env:"MIGRATIONS_DIR" envDefault:"migrations"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	cfg.LogLevel = strings.ToLower(strings.TrimSpace(cfg.LogLevel))
	cfg.LogFormat = strings.ToLower(strings.TrimSpace(cfg.LogFormat))

	global = cfg
	return cfg, nil
}

func GetGlobal() *Config {
	return global
}

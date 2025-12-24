package observability

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Config wraps monitoring settings from pkg/config
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string // dev, staging, production
	TracingEnabled bool
	MetricsEnabled bool
	OTLPEndpoint   string
	OTLPHeaders    map[string]string
	SamplingRate   float64 // 0.0 - 1.0
	PIILevel       string  // none|hashed|full
	MetricsPort    int

	// Advanced settings
	TraceBatchTimeout time.Duration
	MetricInterval    time.Duration
	ResourceAttrs     []attribute.KeyValue
}

// DefaultConfig returns sensible defaults
func DefaultConfig(serviceName string) Config {
	return Config{
		ServiceName:       serviceName,
		ServiceVersion:    "unknown",
		Environment:       "development",
		TracingEnabled:    true,
		MetricsEnabled:    true,
		OTLPEndpoint:      "http://otel-collector:4318",
		SamplingRate:      1.0,
		PIILevel:          "hashed",
		MetricsPort:       8080,
		TraceBatchTimeout: 5 * time.Second,
		MetricInterval:    15 * time.Second,
	}
}

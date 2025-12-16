package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// MCP-Tools Metrics - using explicit registration
var (
	// Request counters
	RequestsTotal *prometheus.CounterVec

	// Tool call counters
	ToolCallsTotal *prometheus.CounterVec

	// Tool token counters (approx payload tokens returned by tool)
	ToolTokensTotal *prometheus.CounterVec

	// Tool duration histogram
	ToolDuration *prometheus.HistogramVec

	// Circuit breaker state gauge
	CircuitBreakerState *prometheus.GaugeVec

	// External provider latency
	ExternalProviderLatency *prometheus.HistogramVec
)

// init creates and registers all metrics with the default registry
func init() {
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "requests_total",
			Help:      "Total number of MCP requests",
		},
		[]string{"method", "status"},
	)

	ToolCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "tool_calls_total",
			Help:      "Total tool invocations",
		},
		[]string{"tool_name", "provider", "status"},
	)

	ToolTokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "tool_tokens_total",
			Help:      "Total estimated tokens returned by tool payloads",
		},
		[]string{"tool_name", "provider"},
	)

	ToolDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "tool_duration_seconds",
			Help:      "Tool execution duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"tool_name", "provider"},
	)

	CircuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "circuit_breaker_state",
			Help:      "Circuit breaker state (0=closed, 0.5=half-open, 1=open)",
		},
		[]string{"provider"},
	)

	ExternalProviderLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "mcp",
			Name:      "external_provider_latency_seconds",
			Help:      "External provider response time in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"provider"},
	)

	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(ToolCallsTotal)
	prometheus.MustRegister(ToolTokensTotal)
	prometheus.MustRegister(ToolDuration)
	prometheus.MustRegister(CircuitBreakerState)
	prometheus.MustRegister(ExternalProviderLatency)
	log.Info().Msg("MCP metrics registered with Prometheus")
}

// RecordRequest records an MCP request
func RecordRequest(method, status string) {
	RequestsTotal.WithLabelValues(method, status).Inc()
}

// RecordToolCall records a tool invocation
func RecordToolCall(toolName, provider, status string, durationSec float64) {
	if provider == "" {
		provider = "unknown"
	}
	if status == "" {
		status = "unknown"
	}
	ToolCallsTotal.WithLabelValues(toolName, provider, status).Inc()
	ToolDuration.WithLabelValues(toolName, provider).Observe(durationSec)
}

// RecordToolTokens records estimated output tokens for a tool invocation
func RecordToolTokens(toolName, provider string, tokens float64) {
	if provider == "" {
		provider = "unknown"
	}
	if tokens < 0 {
		return
	}
	ToolTokensTotal.WithLabelValues(toolName, provider).Add(tokens)
}

// SetCircuitBreakerState sets the circuit breaker state
func SetCircuitBreakerState(provider string, state string) {
	var val float64
	switch state {
	case "closed":
		val = 0.0
	case "half-open":
		val = 0.5
	case "open":
		val = 1.0
	}
	CircuitBreakerState.WithLabelValues(provider).Set(val)
}

// RecordExternalProviderLatency records external provider response time
func RecordExternalProviderLatency(provider string, durationSec float64) {
	ExternalProviderLatency.WithLabelValues(provider).Observe(durationSec)
}

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strings"
)

// LLM-API Metrics
var (
	// Request counters
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status", "model", "stream"},
	)

	// Token counters
	TokensPromptTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "tokens_prompt_total",
			Help:      "Total prompt tokens consumed",
		},
		[]string{"model", "provider"},
	)

	TokensCompletionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "tokens_completion_total",
			Help:      "Total completion tokens generated",
		},
		[]string{"model", "provider"},
	)

	// Provider errors
	ProviderErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "provider_errors_total",
			Help:      "Total provider call failures",
		},
		[]string{"provider", "error_type"},
	)

	// Conversations
	ConversationsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "conversations_created_total",
			Help:      "Total conversations created",
		},
	)

	// Auth requests
	AuthRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "auth_requests_total",
			Help:      "Total authentication requests",
		},
		[]string{"auth_type", "status"},
	)

	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"method", "endpoint", "status"},
	)

	// LLM inference duration
	LLMDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "llm_duration_seconds",
			Help:      "LLM inference duration in seconds",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"model", "provider", "stream"},
	)

	// Time to first token (streaming)
	FirstTokenDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "first_token_seconds",
			Help:      "Time to first token for streaming requests",
			Buckets:   []float64{0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"model", "provider"},
	)

	// Tokens per request distribution
	TokensPerRequest = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "tokens_per_request",
			Help:      "Distribution of tokens per request",
			Buckets:   []float64{10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"model", "type"},
	)

	// Active streaming connections gauge
	ActiveStreams = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "active_streams",
			Help:      "Currently active streaming connections",
		},
		[]string{"model"},
	)

	// Provider health gauge
	ProviderHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "provider_health",
			Help:      "Provider health status (1=healthy, 0=unhealthy)",
		},
		[]string{"provider"},
	)

	// Sharing metrics
	SharesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "shares_total",
			Help:      "Share create/revoke attempts",
		},
		[]string{"scope", "status"},
	)

	PublicShareRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "public_share_requests_total",
			Help:      "Public share fetch/head requests",
		},
		[]string{"method", "status"},
	)

	// User agent metrics (normalized to keep low cardinality)
	UserAgentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "user_agents_total",
			Help:      "Requests by normalized user agent",
		},
		[]string{"user_agent"},
	)

	UserAgentFamilyTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "llm_api",
			Name:      "user_agent_family_total",
			Help:      "Requests by user agent family (browser/cli/sdk/unknown)",
		},
		[]string{"family"},
	)
)

// RecordRequest records an HTTP request with all relevant labels
func RecordRequest(method, endpoint, status, model string, stream bool, durationSec float64) {
	streamStr := "false"
	if stream {
		streamStr = "true"
	}
	RequestsTotal.WithLabelValues(method, endpoint, status, model, streamStr).Inc()
	RequestDuration.WithLabelValues(method, endpoint, status).Observe(durationSec)
}

// RecordTokens records token usage for a completion request
func RecordTokens(model, provider string, promptTokens, completionTokens int) {
	TokensPromptTotal.WithLabelValues(model, provider).Add(float64(promptTokens))
	TokensCompletionTotal.WithLabelValues(model, provider).Add(float64(completionTokens))
	TokensPerRequest.WithLabelValues(model, "prompt").Observe(float64(promptTokens))
	TokensPerRequest.WithLabelValues(model, "completion").Observe(float64(completionTokens))
}

// RecordLLMDuration records the duration of an LLM inference call
func RecordLLMDuration(model, provider string, stream bool, durationSec float64) {
	streamStr := "false"
	if stream {
		streamStr = "true"
	}
	LLMDuration.WithLabelValues(model, provider, streamStr).Observe(durationSec)
}

// RecordFirstToken records time to first token for streaming
func RecordFirstToken(model, provider string, durationSec float64) {
	FirstTokenDuration.WithLabelValues(model, provider).Observe(durationSec)
}

// RecordProviderError records a provider error
func RecordProviderError(provider, errorType string) {
	ProviderErrorsTotal.WithLabelValues(provider, errorType).Inc()
}

// SetProviderHealth sets the health status of a provider
func SetProviderHealth(provider string, healthy bool) {
	val := 0.0
	if healthy {
		val = 1.0
	}
	ProviderHealth.WithLabelValues(provider).Set(val)
}

// IncrementActiveStreams increments the active streams gauge
func IncrementActiveStreams(model string) {
	ActiveStreams.WithLabelValues(model).Inc()
}

// DecrementActiveStreams decrements the active streams gauge
func DecrementActiveStreams(model string) {
	ActiveStreams.WithLabelValues(model).Dec()
}

// RecordShare records a share create/revoke attempt
func RecordShare(scope, status string) {
	if scope == "" {
		scope = "unknown"
	}
	if status == "" {
		status = "unknown"
	}
	SharesTotal.WithLabelValues(scope, status).Inc()
}

// RecordPublicShareRequest records a public share GET/HEAD
func RecordPublicShareRequest(method, status string) {
	if method == "" {
		method = "UNKNOWN"
	}
	if status == "" {
		status = "unknown"
	}
	PublicShareRequestsTotal.WithLabelValues(method, status).Inc()
}

// RecordUserAgent records UA metrics with normalization and family bucketing
func RecordUserAgent(ua string) {
	norm := normalizeUserAgent(ua)
	family := userAgentFamily(norm)
	UserAgentsTotal.WithLabelValues(norm).Inc()
	UserAgentFamilyTotal.WithLabelValues(family).Inc()
}

func normalizeUserAgent(ua string) string {
	ua = strings.TrimSpace(strings.ToLower(ua))
	if ua == "" {
		return "unknown"
	}
	parts := strings.Fields(ua)
	norm := parts[0]
	if len(norm) > 60 {
		norm = norm[:60]
	}
	return norm
}

func userAgentFamily(normUA string) string {
	switch {
	case strings.Contains(normUA, "mozilla") || strings.Contains(normUA, "chrome") || strings.Contains(normUA, "safari") || strings.Contains(normUA, "firefox") || strings.Contains(normUA, "edge"):
		return "browser"
	case strings.Contains(normUA, "curl") || strings.Contains(normUA, "wget") || strings.Contains(normUA, "httpie"):
		return "cli"
	case strings.Contains(normUA, "postman") || strings.Contains(normUA, "insomnia"):
		return "api_client"
	case strings.Contains(normUA, "okhttp") || strings.Contains(normUA, "cfnetwork"):
		return "mobile"
	case strings.Contains(normUA, "axios") || strings.Contains(normUA, "fetch") || strings.Contains(normUA, "python-requests") || strings.Contains(normUA, "go-http-client") || strings.Contains(normUA, "java"):
		return "sdk"
	default:
		return "unknown"
	}
}

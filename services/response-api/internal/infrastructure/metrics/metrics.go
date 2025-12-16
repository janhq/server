package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Response-API Metrics
var (
	// Request counters
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"method", "endpoint"},
	)

	// Tool call counters
	ToolCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "tool_calls_total",
			Help:      "Total MCP tool invocations",
		},
		[]string{"tool_name", "status"},
	)

	// Tool duration histogram
	ToolDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "tool_duration_seconds",
			Help:      "Tool execution duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"tool_name"},
	)

	// Queue depth gauge
	QueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "queue_depth",
			Help:      "Background job queue depth",
		},
	)

	// Background jobs counter
	BackgroundJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "background_jobs_total",
			Help:      "Total background jobs processed",
		},
		[]string{"job_type", "status"},
	)

	// DB query duration
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "response_api",
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"query_type"},
	)
)

// RecordRequest records an HTTP request
func RecordRequest(method, endpoint, status string, durationSec float64) {
	RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(durationSec)
}

// RecordToolCall records an MCP tool invocation
func RecordToolCall(toolName, status string, durationSec float64) {
	ToolCallsTotal.WithLabelValues(toolName, status).Inc()
	ToolDuration.WithLabelValues(toolName).Observe(durationSec)
}

// SetQueueDepth sets the current queue depth
func SetQueueDepth(depth int) {
	QueueDepth.Set(float64(depth))
}

// RecordBackgroundJob records a background job execution
func RecordBackgroundJob(jobType, status string) {
	BackgroundJobsTotal.WithLabelValues(jobType, status).Inc()
}

// RecordDBQuery records a database query
func RecordDBQuery(queryType string, durationSec float64) {
	DBQueryDuration.WithLabelValues(queryType).Observe(durationSec)
}

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Memory-Tools Metrics
var (
	// Request counters
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// Memory load operations
	LoadTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "load_total",
			Help:      "Total memory load operations",
		},
		[]string{"status"},
	)

	// Memory observe operations
	ObserveTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "observe_total",
			Help:      "Total memory observe operations",
		},
		[]string{"status"},
	)

	// Embedding duration
	EmbeddingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "embedding_duration_seconds",
			Help:      "Embedding computation duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10},
		},
	)

	// Vector search duration
	VectorSearchDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "vector_search_duration_seconds",
			Help:      "Vector search duration in seconds",
			Buckets:   []float64{0.05, 0.1, 0.5, 1, 2},
		},
	)

	// Cache hits
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "cache_hits_total",
			Help:      "Total cache hits",
		},
		[]string{"cache_type"},
	)

	// Cache misses
	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "memory",
			Name:      "cache_misses_total",
			Help:      "Total cache misses",
		},
		[]string{"cache_type"},
	)
)

// Handler returns the Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordRequest records an HTTP request
func RecordRequest(method, endpoint, status string, durationSec float64) {
	RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(durationSec)
}

// RecordLoad records a memory load operation
func RecordLoad(status string) {
	LoadTotal.WithLabelValues(status).Inc()
}

// RecordObserve records a memory observe operation
func RecordObserve(status string) {
	ObserveTotal.WithLabelValues(status).Inc()
}

// RecordEmbedding records embedding computation time
func RecordEmbedding(durationSec float64) {
	EmbeddingDuration.Observe(durationSec)
}

// RecordVectorSearch records vector search time
func RecordVectorSearch(durationSec float64) {
	VectorSearchDuration.Observe(durationSec)
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	CacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	CacheMissesTotal.WithLabelValues(cacheType).Inc()
}

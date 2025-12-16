package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Media-API Metrics
var (
	// Request counters
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request duration histogram
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"method", "endpoint"},
	)

	// Upload counters
	UploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "uploads_total",
			Help:      "Total file uploads",
		},
		[]string{"content_type", "status"},
	)

	// Upload bytes counter
	UploadBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "upload_bytes_total",
			Help:      "Total bytes uploaded",
		},
		[]string{"content_type"},
	)

	// S3 operations counter
	S3OperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "s3_operations_total",
			Help:      "Total S3 operations",
		},
		[]string{"operation", "status"},
	)

	// S3 operation duration
	S3Duration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "s3_duration_seconds",
			Help:      "S3 operation duration in seconds",
			Buckets:   []float64{0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"operation"},
	)

	// Presign URL duration
	PresignDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "jan",
			Subsystem: "media_api",
			Name:      "presign_duration_seconds",
			Help:      "Presigned URL generation duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1},
		},
	)
)

// RecordRequest records an HTTP request
func RecordRequest(method, endpoint, status string, durationSec float64) {
	RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(durationSec)
}

// RecordUpload records a file upload
func RecordUpload(contentType, status string, bytes int64) {
	UploadsTotal.WithLabelValues(contentType, status).Inc()
	if status == "success" {
		UploadBytesTotal.WithLabelValues(contentType).Add(float64(bytes))
	}
}

// RecordS3Operation records an S3 operation
func RecordS3Operation(operation, status string, durationSec float64) {
	S3OperationsTotal.WithLabelValues(operation, status).Inc()
	S3Duration.WithLabelValues(operation).Observe(durationSec)
}

// RecordPresign records presigned URL generation
func RecordPresign(durationSec float64) {
	PresignDuration.Observe(durationSec)
}

// Package metrics provides Prometheus metrics for the realtime-api service.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveSessions tracks the number of active realtime sessions.
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "realtime_active_sessions",
			Help: "Number of currently active realtime sessions",
		},
	)

	// SessionsCreated tracks the total number of sessions created.
	SessionsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "realtime_sessions_created_total",
			Help: "Total number of realtime sessions created",
		},
	)

	// SessionsDeleted tracks the total number of sessions deleted.
	SessionsDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "realtime_sessions_deleted_total",
			Help: "Total number of realtime sessions deleted",
		},
	)

	// SessionStateTransitions tracks session state changes.
	SessionStateTransitions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "realtime_session_state_transitions_total",
			Help: "Total number of session state transitions",
		},
		[]string{"from_state", "to_state"},
	)

	// LiveKitSyncDuration tracks the duration of LiveKit sync operations.
	LiveKitSyncDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "realtime_livekit_sync_duration_seconds",
			Help:    "Duration of LiveKit room sync operations",
			Buckets: prometheus.DefBuckets,
		},
	)

	// LiveKitSyncErrors tracks errors during LiveKit sync.
	LiveKitSyncErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "realtime_livekit_sync_errors_total",
			Help: "Total number of errors during LiveKit sync",
		},
	)

	// TokenGenerationDuration tracks token generation time.
	TokenGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "realtime_token_generation_duration_seconds",
			Help:    "Duration of LiveKit token generation",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
	)
)

// RecordSessionCreated increments session creation metrics.
func RecordSessionCreated() {
	SessionsCreated.Inc()
	ActiveSessions.Inc()
}

// RecordSessionDeleted increments session deletion metrics.
func RecordSessionDeleted() {
	SessionsDeleted.Inc()
	ActiveSessions.Dec()
}

// RecordStateTransition records a session state change.
func RecordStateTransition(fromState, toState string) {
	SessionStateTransitions.WithLabelValues(fromState, toState).Inc()
}

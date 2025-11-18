package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware instruments HTTP handlers
func HTTPMiddleware(tracer trace.Tracer, meter metric.Meter, serviceName string) func(http.Handler) http.Handler {
	// Create metrics
	requestDuration, _ := meter.Float64Histogram(
		fmt.Sprintf("jan_%s_request_duration_seconds", serviceName),
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)

	requestsTotal, _ := meter.Int64Counter(
		fmt.Sprintf("jan_%s_requests_total", serviceName),
		metric.WithDescription("Total HTTP requests"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Start span
			ctx, span := tracer.Start(r.Context(), r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPMethod(r.Method),
					semconv.HTTPRoute(r.URL.Path),
					semconv.HTTPScheme(r.URL.Scheme),
				),
			)
			defer span.End()

			// Wrap response writer to capture status
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Record metrics
			duration := time.Since(start).Seconds()
			attrs := metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("route", r.URL.Path),
				attribute.Int("status", rw.statusCode),
			)

			requestDuration.Record(ctx, duration, attrs)
			requestsTotal.Add(ctx, 1, attrs)

			// Add status to span
			span.SetAttributes(semconv.HTTPStatusCode(rw.statusCode))
			if rw.statusCode >= 400 {
				span.RecordError(fmt.Errorf("HTTP %d", rw.statusCode))
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

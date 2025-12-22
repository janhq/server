package observability

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"jan-server/services/realtime-api/internal/config"
)

// Shutdown is a function that releases telemetry resources.
type Shutdown func(ctx context.Context) error

// Setup configures OpenTelemetry tracing and metrics if enabled.
func Setup(ctx context.Context, cfg *config.Config, log zerolog.Logger) (Shutdown, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	var (
		tracerProvider *sdktrace.TracerProvider
		meterProvider  *sdkmetric.MeterProvider
	)

	if cfg.EnableTracing && cfg.OTLPEndpoint != "" {
		// Normalize endpoint
		endpoint := cfg.OTLPEndpoint
		insecure := true
		if strings.HasPrefix(endpoint, "http://") {
			endpoint = strings.TrimPrefix(endpoint, "http://")
			insecure = true
		} else if strings.HasPrefix(endpoint, "https://") {
			endpoint = strings.TrimPrefix(endpoint, "https://")
			insecure = false
		}

		traceOpts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
		metricOpts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint)}
		if insecure {
			traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
			metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
		}

		traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
		if err != nil {
			return nil, err
		}

		meterExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
		if err != nil {
			return nil, err
		}

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)

		reader := sdkmetric.NewPeriodicReader(meterExporter, sdkmetric.WithInterval(30*time.Second))
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(reader),
			sdkmetric.WithResource(res),
		)

		log.Info().Str("endpoint", cfg.OTLPEndpoint).Msg("Tracing and metrics enabled")
	} else {
		tracerProvider = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		meterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithResource(res))
		log.Info().Msg("Tracing disabled, using noop providers")
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("shutdown meter provider")
			shutdownErr = err
		}
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("shutdown tracer provider")
			if shutdownErr == nil {
				shutdownErr = err
			}
		}
		return shutdownErr
	}

	return shutdown, nil
}

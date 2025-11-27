package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/janhq/jan-server/pkg/telemetry"
)

// Provider holds initialized OTEL components
type Provider struct {
	Tracer         trace.Tracer
	Meter          metric.Meter
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Sanitizer      *telemetry.Sanitizer

	shutdownFuncs []func(context.Context) error
}

// Init initializes OTEL for a service
func Init(ctx context.Context, cfg Config) (*Provider, error) {
	provider := &Provider{
		Sanitizer: telemetry.NewSanitizer(
			telemetry.PIILevel(cfg.PIILevel),
			cfg.ServiceName,
		),
	}

	// Create resource with service metadata
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
		resource.WithAttributes(cfg.ResourceAttrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracing if enabled
	if cfg.TracingEnabled {
		tp, err := initTracerProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to init tracer: %w", err)
		}
		provider.TracerProvider = tp
		provider.Tracer = tp.Tracer(cfg.ServiceName)
		provider.shutdownFuncs = append(provider.shutdownFuncs, tp.Shutdown)

		// Set global tracer provider
		otel.SetTracerProvider(tp)

		// Set propagator for trace context
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	// Initialize metrics if enabled
	if cfg.MetricsEnabled {
		mp, err := initMeterProvider(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to init meter: %w", err)
		}
		provider.MeterProvider = mp
		provider.Meter = mp.Meter(cfg.ServiceName)
		provider.shutdownFuncs = append(provider.shutdownFuncs, mp.Shutdown)

		// Set global meter provider
		otel.SetMeterProvider(mp)
	}

	return provider, nil
}

// Shutdown gracefully shuts down all providers
func (p *Provider) Shutdown(ctx context.Context) error {
	for _, shutdown := range p.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

func initTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithHeaders(cfg.OTLPHeaders),
		otlptracehttp.WithInsecure(), // TODO: Use TLS in production
	)
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SamplingRate),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.TraceBatchTimeout),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	return tp, nil
}

func initMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetrichttp.WithHeaders(cfg.OTLPHeaders),
		otlpmetrichttp.WithInsecure(), // TODO: Use TLS in production
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter,
				sdkmetric.WithInterval(cfg.MetricInterval),
			),
		),
		sdkmetric.WithResource(res),
	)

	return mp, nil
}

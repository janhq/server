# Observability Library

This package provides a shared observability library for Jan Server services, encapsulating OpenTelemetry (OTEL) setup and providing consistent instrumentation patterns.

## Features

- **Unified Configuration**: Single config structure for all OTEL settings
- **Automatic Instrumentation**: HTTP middleware, background worker tracking
- **PII Sanitization**: Built-in privacy controls with tenant-specific hashing
- **Standard Attributes**: Consistent span/metric attributes across services
- **Easy Integration**: Drop-in initialization for any Go service

## Quick Start

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/janhq/jan-server/pkg/config"
    "github.com/janhq/jan-server/pkg/observability"
    "github.com/janhq/jan-server/pkg/observability/middleware"
)

func main() {
    ctx := context.Background()

    // Load config
    cfg := config.Load()

    // Initialize observability
    obsCfg := observability.DefaultConfig("my-service")
    obsCfg.Environment = cfg.Environment
    obsCfg.TracingEnabled = cfg.Monitoring.OTEL.TracingEnabled
    obsCfg.PIILevel = cfg.Monitoring.OTEL.PIILevel

    provider, err := observability.Init(ctx, obsCfg)
    if err != nil {
        log.Fatalf("Failed to initialize observability: %v", err)
    }
    defer provider.Shutdown(ctx)

    // Setup HTTP server with middleware
    mux := http.NewServeMux()
    mux.HandleFunc("/health", handleHealth)

    handler := middleware.HTTPMiddleware(provider.Tracer, provider.Meter, "my-service")(mux)

    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

## Configuration

### Environment Variables

```bash
# Tracing
ENABLE_TRACING=true
OTEL_ENABLED=true
OTEL_SERVICE_NAME=my-service
OTEL_SERVICE_VERSION=1.0.0
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_TRACES_SAMPLER_ARG=1.0

# Privacy
TELEMETRY_PII_LEVEL=hashed  # none|hashed|full

# Metrics
OTEL_METRIC_EXPORT_INTERVAL=15s
```

### Programmatic Configuration

```go
cfg := observability.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    TracingEnabled: true,
    MetricsEnabled: true,
    OTLPEndpoint:   "http://otel-collector:4318",
    SamplingRate:   0.1,  // Sample 10% of traces
    PIILevel:       "hashed",
}
```

## Usage Patterns

### Adding Correlation Attributes

```go
import (
    "github.com/janhq/jan-server/pkg/observability"
    "go.opentelemetry.io/otel/trace"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    span := trace.SpanFromContext(ctx)

    // Add standard correlation attributes
    observability.AddConversationAttrsToSpan(
        span,
        conversationID,
        tenantID,
        userID,
        provider.Sanitizer,
    )

    // Add LLM-specific attributes
    span.SetAttributes(observability.WithLLMAttrs(
        "gpt-4",
        1500, // prompt tokens
        300,  // completion tokens
    )...)
}
```

### Instrumenting Background Workers

```go
import (
    "github.com/janhq/jan-server/pkg/observability/worker"
)

func main() {
    // ... init provider ...

    instrumenter, err := worker.NewWorkerInstrumenter(
        provider.Tracer,
        provider.Meter,
        "my-service",
    )

    // Use in worker pool
    err = instrumenter.InstrumentJob(ctx, "webhook", jobID, func(ctx context.Context) error {
        // Your job logic here
        return sendWebhook(ctx, payload)
    })
}
```

### PII Sanitization

```go
// Sanitize user prompts
sanitizedPrompt := provider.Sanitizer.SanitizePrompt(userPrompt)

// Sanitize user IDs
hashedUserID := provider.Sanitizer.SanitizeUserID(userID)

// Sanitize metadata
sanitizedMetadata := provider.Sanitizer.SanitizeMetadata(metadata)
```

## Standard Attributes

All services should use these standard attributes for correlation:

| Attribute               | Type   | Description                    |
| ----------------------- | ------ | ------------------------------ |
| `conversation_id`       | string | Unique conversation identifier |
| `tenant_id`             | string | Tenant identifier              |
| `user_id`               | string | Sanitized user identifier      |
| `request_id`            | string | Unique request identifier      |
| `llm.model`             | string | LLM model name                 |
| `llm.tokens.prompt`     | int64  | Prompt token count             |
| `llm.tokens.completion` | int64  | Completion token count         |
| `mcp.tool.name`         | string | MCP tool name                  |
| `prompt.category`       | string | Prompt category                |
| `prompt.persona`        | string | Prompt persona                 |
| `prompt.language`       | string | Prompt language                |

## Metrics Naming Convention

Follow the pattern: `jan_<service>_<metric>_<unit>`

Examples:

- `jan_llm_api_request_duration_seconds`
- `jan_response_api_queue_depth`
- `jan_media_api_s3_errors_total`

## Privacy Levels

### None (`PIILevel = "none"`)

- All user content redacted as `[REDACTED]`
- Maximum privacy, minimal debugging utility

### Hashed (`PIILevel = "hashed"`) - Default

- PII detected and replaced with tenant-specific hashes
- Emails: `[EMAIL:a1b2c3d4]`
- Phones: `[PHONE:e5f6g7h8]`
- SSNs/Credit Cards: `[SSN:REDACTED]`, `[CC:REDACTED]`
- User IDs: 8-character hash
- Balances privacy and debugging

### Full (`PIILevel = "full"`)

- No sanitization
- Use only in development/testing
- Never use in production

## Testing

```bash
# Run observability tests
cd pkg/observability
go test -v ./...
```

## Architecture

```
pkg/observability/
├── config.go           # Configuration structures
├── provider.go         # OTEL provider initialization
├── attributes.go       # Standard attribute helpers
├── middleware/
│   └── http.go        # HTTP instrumentation
└── worker/
    └── worker.go      # Background job instrumentation

pkg/telemetry/
├── sanitizer.go        # PII detection and hashing
└── sanitizer_test.go   # Comprehensive test suite
```

## Best Practices

1. **Always sanitize user content** before adding to spans/metrics
2. **Use standard attributes** for correlation across services
3. **Sample in production** - Set `SamplingRate < 1.0` for high-traffic services
4. **Include request IDs** for cross-service trace correlation
5. **Test with PII** - Verify sanitization catches real-world patterns
6. **Monitor overhead** - Keep instrumentation latency <100ms P95

## Troubleshooting

### Spans not appearing in Jaeger

1. Check OTEL Collector health: `curl http://otel-collector:13133/`
2. Verify sampling rate: `OTEL_TRACES_SAMPLER_ARG=1.0`
3. Check service logs for export errors
4. Verify trace propagation headers: `traceparent`, `tracestate`

### High memory usage

1. Reduce batch size in provider.go
2. Lower sampling rate
3. Increase metric export interval

### PII leaking to telemetry

1. Verify `PIILevel` is set to `hashed` or `none`
2. Check sanitizer is initialized correctly
3. Review custom span attributes for unsanitized data
4. Run `sanitizer_test.go` to validate patterns

## Related Documentation

- [Monitoring Guide](../../docs/guides/monitoring.md)
- [Monitoring Runbook](../../docs/runbooks/monitoring.md)
- [Observability Conventions](../../docs/conventions/observability.md)
- [Security Policy](../../docs/architecture/security.md)

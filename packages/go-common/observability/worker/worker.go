package worker

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// WorkerInstrumenter instruments background workers
type WorkerInstrumenter struct {
	tracer        trace.Tracer
	workersActive metric.Int64UpDownCounter
	workersIdle   metric.Int64UpDownCounter
	jobDuration   metric.Float64Histogram
	jobsTotal     metric.Int64Counter
}

// NewWorkerInstrumenter creates a new worker instrumenter
func NewWorkerInstrumenter(tracer trace.Tracer, meter metric.Meter, serviceName string) (*WorkerInstrumenter, error) {
	workersActive, err := meter.Int64UpDownCounter(
		fmt.Sprintf("jan_%s_workers_active", serviceName),
		metric.WithDescription("Number of active workers"),
	)
	if err != nil {
		return nil, err
	}

	workersIdle, err := meter.Int64UpDownCounter(
		fmt.Sprintf("jan_%s_workers_idle", serviceName),
		metric.WithDescription("Number of idle workers"),
	)
	if err != nil {
		return nil, err
	}

	jobDuration, err := meter.Float64Histogram(
		fmt.Sprintf("jan_%s_job_duration_seconds", serviceName),
		metric.WithDescription("Background job duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	jobsTotal, err := meter.Int64Counter(
		fmt.Sprintf("jan_%s_jobs_total", serviceName),
		metric.WithDescription("Total background jobs processed"),
	)
	if err != nil {
		return nil, err
	}

	return &WorkerInstrumenter{
		tracer:        tracer,
		workersActive: workersActive,
		workersIdle:   workersIdle,
		jobDuration:   jobDuration,
		jobsTotal:     jobsTotal,
	}, nil
}

// InstrumentJob wraps a job execution with observability
func (w *WorkerInstrumenter) InstrumentJob(ctx context.Context, jobType string, jobID string, fn func(context.Context) error) error {
	// Update worker status
	w.workersIdle.Add(ctx, -1)
	w.workersActive.Add(ctx, 1)
	defer func() {
		w.workersActive.Add(ctx, -1)
		w.workersIdle.Add(ctx, 1)
	}()

	// Start span
	ctx, span := w.tracer.Start(ctx, fmt.Sprintf("worker.%s", jobType),
		trace.WithAttributes(
			attribute.String("job.type", jobType),
			attribute.String("job.id", jobID),
		),
	)
	defer span.End()

	// Execute job
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
	}

	attrs := metric.WithAttributes(
		attribute.String("job.type", jobType),
		attribute.String("status", status),
	)

	w.jobDuration.Record(ctx, duration, attrs)
	w.jobsTotal.Add(ctx, 1, attrs)

	return err
}

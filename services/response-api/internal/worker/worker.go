package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"jan-server/services/response-api/internal/domain/response"
	"jan-server/services/response-api/internal/infrastructure/queue"
)

// Worker processes background tasks from the queue.
type Worker struct {
	id              int
	queue           queue.TaskQueue
	responseService response.Service
	taskTimeout     time.Duration
	log             zerolog.Logger
	stopChan        chan struct{}
}

// NewWorker creates a new background worker.
func NewWorker(
	id int,
	queue queue.TaskQueue,
	responseService response.Service,
	taskTimeout time.Duration,
	log zerolog.Logger,
) *Worker {
	return &Worker{
		id:              id,
		queue:           queue,
		responseService: responseService,
		taskTimeout:     taskTimeout,
		log:             log.With().Int("worker_id", id).Str("component", "worker").Logger(),
		stopChan:        make(chan struct{}),
	}
}

// Start begins processing tasks from the queue.
func (w *Worker) Start(ctx context.Context) {
	w.log.Info().Msg("worker started")

	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info().Msg("worker stopped by context")
			return
		case <-w.stopChan:
			w.log.Info().Msg("worker stopped")
			return
		case <-ticker.C:
			w.processNextTask(ctx)
		}
	}
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) processNextTask(ctx context.Context) {
	// Dequeue next task
	task, err := w.queue.Dequeue(ctx)
	if err != nil {
		w.log.Error().Err(err).Msg("failed to dequeue task")
		return
	}

	if task == nil {
		// No tasks available
		return
	}

	w.log.Info().
		Str("response_id", task.PublicID).
		Str("user_id", task.UserID).
		Str("model", task.Model).
		Msg("processing background task")

	// Mark as processing
	if err := w.queue.MarkProcessing(ctx, task.PublicID); err != nil {
		w.log.Error().Err(err).Str("response_id", task.PublicID).Msg("failed to mark processing")
		return
	}

	// Execute task with timeout
	taskCtx, cancel := context.WithTimeout(ctx, w.taskTimeout)
	defer cancel()

	// Execute the background task using the service's ExecuteBackground method
	if err := w.executeTask(taskCtx, task.PublicID); err != nil {
		w.log.Error().Err(err).Str("response_id", task.PublicID).Msg("task execution failed")
		if markErr := w.queue.MarkFailed(ctx, task.PublicID, err); markErr != nil {
			w.log.Error().Err(markErr).Str("response_id", task.PublicID).Msg("failed to mark task as failed")
		}
		return
	}

	w.log.Info().Str("response_id", task.PublicID).Msg("task completed successfully")
}

func (w *Worker) executeTask(ctx context.Context, publicID string) error {
	// Check if service has ExecuteBackground method
	type backgroundExecutor interface {
		ExecuteBackground(ctx context.Context, publicID string) error
	}

	if executor, ok := w.responseService.(backgroundExecutor); ok {
		return executor.ExecuteBackground(ctx, publicID)
	}

	return fmt.Errorf("response service does not support background execution")
}

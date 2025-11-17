package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"jan-server/services/response-api/internal/domain/response"
	"jan-server/services/response-api/internal/infrastructure/queue"
)

// Pool manages multiple background workers.
type Pool struct {
	workers         []*Worker
	queue           queue.TaskQueue
	responseService response.Service
	workerCount     int
	taskTimeout     time.Duration
	log             zerolog.Logger
	wg              sync.WaitGroup
	stopChan        chan struct{}
}

// Config contains worker pool configuration.
type Config struct {
	WorkerCount int
	TaskTimeout time.Duration
}

// NewPool creates a new worker pool.
func NewPool(
	queue queue.TaskQueue,
	responseService response.Service,
	cfg Config,
	log zerolog.Logger,
) *Pool {
	return &Pool{
		queue:           queue,
		responseService: responseService,
		workerCount:     cfg.WorkerCount,
		taskTimeout:     cfg.TaskTimeout,
		log:             log.With().Str("component", "worker-pool").Logger(),
		stopChan:        make(chan struct{}),
	}
}

// Start initializes and starts all workers.
func (p *Pool) Start(ctx context.Context) error {
	p.log.Info().Int("worker_count", p.workerCount).Msg("starting worker pool")

	p.workers = make([]*Worker, p.workerCount)
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(
			i+1,
			p.queue,
			p.responseService,
			p.taskTimeout,
			p.log,
		)
		p.workers[i] = worker

		p.wg.Add(1)
		go func(w *Worker) {
			defer p.wg.Done()
			w.Start(ctx)
		}(worker)
	}

	p.log.Info().Msg("worker pool started")

	return nil
}

// Stop gracefully shuts down all workers.
func (p *Pool) Stop() {
	p.log.Info().Msg("stopping worker pool")

	// Signal all workers to stop
	for _, worker := range p.workers {
		worker.Stop()
	}

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		p.log.Info().Msg("all workers stopped gracefully")
	case <-time.After(30 * time.Second):
		p.log.Warn().Msg("worker pool shutdown timed out")
	}
}

// GetQueueDepth returns the current queue depth.
func (p *Pool) GetQueueDepth(ctx context.Context) (int64, error) {
	return p.queue.GetQueueDepth(ctx)
}

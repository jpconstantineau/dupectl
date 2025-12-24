package worker

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/jpconstantineau/dupectl/pkg/logger"
)

// WorkItem represents a unit of work
type WorkItem interface {
	Process(ctx context.Context) error
	ID() string
}

// PoolMetrics tracks worker pool performance
type PoolMetrics struct {
	ItemsProcessed int64
	ItemsFailed    int64
	ItemsQueued    int64
	WorkersActive  int32
}

// WorkerPool manages concurrent work processing
type WorkerPool struct {
	workers   int
	workQueue chan WorkItem
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	errors    chan error
	metrics   *PoolMetrics
	mu        sync.Mutex
}

// NewWorkerPool creates a worker pool with specified worker count
func NewWorkerPool(ctx context.Context, workers int) (*WorkerPool, error) {
	if workers < 1 {
		workers = 1
	}

	poolCtx, cancel := context.WithCancel(ctx)

	pool := &WorkerPool{
		workers:   workers,
		workQueue: make(chan WorkItem, workers*10),
		ctx:       poolCtx,
		cancel:    cancel,
		errors:    make(chan error, workers*10),
		metrics:   &PoolMetrics{},
	}

	pool.Start()
	return pool, nil
}

// Start begins processing work items
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker processes work items from the queue
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case item, ok := <-wp.workQueue:
			if !ok {
				return
			}

			atomic.AddInt32(&wp.metrics.WorkersActive, 1)

			err := wp.processItem(item)
			if err != nil {
				atomic.AddInt64(&wp.metrics.ItemsFailed, 1)
				select {
				case wp.errors <- err:
				default:
					logger.Error("Error channel full, dropping error: %v", err)
				}
			} else {
				atomic.AddInt64(&wp.metrics.ItemsProcessed, 1)
			}

			atomic.AddInt32(&wp.metrics.WorkersActive, -1)
		}
	}
}

// processItem safely executes a work item with panic recovery
func (wp *WorkerPool) processItem(item WorkItem) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic processing item %s: %v", item.ID(), r)
			err = nil // Don't propagate panic
		}
	}()

	return item.Process(wp.ctx)
}

// Submit adds a work item to the queue
func (wp *WorkerPool) Submit(item WorkItem) error {
	select {
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	case wp.workQueue <- item:
		atomic.AddInt64(&wp.metrics.ItemsQueued, 1)
		return nil
	}
}

// Stop stops accepting new work and cancels context
func (wp *WorkerPool) Stop() {
	wp.cancel()
}

// Wait blocks until all workers complete and returns errors
func (wp *WorkerPool) Wait() []error {
	close(wp.workQueue)
	wp.wg.Wait()
	close(wp.errors)

	var errs []error
	for err := range wp.errors {
		errs = append(errs, err)
	}
	return errs
}

// Metrics returns current pool performance statistics
func (wp *WorkerPool) Metrics() PoolMetrics {
	return PoolMetrics{
		ItemsProcessed: atomic.LoadInt64(&wp.metrics.ItemsProcessed),
		ItemsFailed:    atomic.LoadInt64(&wp.metrics.ItemsFailed),
		ItemsQueued:    atomic.LoadInt64(&wp.metrics.ItemsQueued),
		WorkersActive:  atomic.LoadInt32(&wp.metrics.WorkersActive),
	}
}

package workerpool

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	inCh          chan Request
	workerManager *WorkerManager
	scaler        *Scaler
	reqHandler    map[int]RequestHandler
}

func NewDispatcher(
	bufferSize int,
	wg *sync.WaitGroup,
	maxWorkers int,
	reqHandler map[int]RequestHandler,
) WorkerPoolManager {
	inCh := make(chan Request, bufferSize)
	stopCh := make(chan struct{}, maxWorkers)
	workerManager := NewWorkerManager(wg, inCh, stopCh, reqHandler)
	scaler := NewScaler(workerManager, inCh, DefaultMinWorkers, DefaultMaxWorkers, DefaultLoadThreshold)

	return &Dispatcher{
		inCh:          inCh,
		workerManager: workerManager,
		scaler:        scaler,
		reqHandler:    reqHandler,
	}
}

func (d *Dispatcher) AddWorker(w *Worker) {
	d.workerManager.AddWorker(w)
}

func (d *Dispatcher) RemoveWorker(minWorkers int) {
	if d.workerManager.WorkerCount() > minWorkers {
		d.workerManager.RemoveWorker()
	}
}

func (d *Dispatcher) ScaleWorkers(ctx context.Context, minWorkers, maxWorkers, loadThreshold int) {
	d.scaler.Start(ctx)
}

func (d *Dispatcher) MakeRequest(r Request) {
	select {
	case d.inCh <- r:
	default:
		log.Info().Msg("Request channel is full. Dropping request.")
	}
}

func (d *Dispatcher) Stop(ctx context.Context) {
	log.Info().Msg("Graceful shutdown initiated")

	// First, stop receiving new requests
	close(d.inCh)

	// Check the number of pending requests
	pendingRequests := len(d.inCh)
	log.Info().Msgf("Pending requests: %d", pendingRequests)

	// Stop all workers
	d.workerManager.StopAllWorkers()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		d.workerManager.WaitForAllWorkers()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		log.Info().Msg("All workers stopped gracefully")
	case <-ctx.Done():
		log.Info().Msg("Timeout reached, some requests may not have been processed")
	}

	// Report remaining requests
	remainingRequests := len(d.inCh)
	log.Info().Msgf("Unprocessed requests: %d", remainingRequests)

	// Optional: Log remaining requests to a file or another system
	if remainingRequests > 0 {
		d.logRemainingRequests()
	}

	log.Info().Msg("Shutdown complete")
}

func (d *Dispatcher) logRemainingRequests() {
	for req := range d.inCh {
		log.Info().Msgf("Unprocessed request: %v", req)
		// You can log remaining requests to a file or another system here
	}
}

package workerpool

import (
	"context"
	"fmt"
	"sync"
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
		fmt.Println("Request channel is full. Dropping request.")
	}
}

func (d *Dispatcher) Stop(ctx context.Context) {
	fmt.Println("\nGraceful shutdown initiated")

	// First, stop receiving new requests
	close(d.inCh)

	// Check the number of pending requests
	pendingRequests := len(d.inCh)
	fmt.Printf("Pending requests: %d\n", pendingRequests)

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
		fmt.Println("All workers stopped gracefully")
	case <-ctx.Done():
		fmt.Println("Timeout reached, some requests may not have been processed")
	}

	// Report remaining requests
	remainingRequests := len(d.inCh)
	fmt.Printf("Unprocessed requests: %d\n", remainingRequests)

	// Optional: Log remaining requests to a file or another system
	if remainingRequests > 0 {
		d.logRemainingRequests()
	}

	fmt.Println("Shutdown complete")
}

func (d *Dispatcher) logRemainingRequests() {
	for req := range d.inCh {
		fmt.Printf("Unprocessed request: %v\n", req)
		// You can log remaining requests to a file or another system here
	}
}

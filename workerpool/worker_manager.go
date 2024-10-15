package workerpool

import (
	"context"
	"sync"
)

// Dispatcher is an interface for managing the worker pool.
type WorkerPoolManager interface {
	AddWorker(w *Worker)
	RemoveWorker(minWorkers int)
	ScaleWorkers(ctx context.Context, minWorkers, maxWorkers, loadThreshold int)
	MakeRequest(Request)
	Stop(ctx context.Context)
}

type WorkerManager struct {
	workers     map[int]*Worker
	workerCount int
	mu          sync.Mutex
	wg          *sync.WaitGroup
	inCh        chan Request
	stopCh      chan struct{}
	reqHandler  map[int]RequestHandler
}

func NewWorkerManager(
	wg *sync.WaitGroup,
	inCh chan Request,
	stopCh chan struct{},
	reqHandler map[int]RequestHandler,
) *WorkerManager {
	return &WorkerManager{
		workers:    make(map[int]*Worker),
		wg:         wg,
		inCh:       inCh,
		stopCh:     stopCh,
		reqHandler: reqHandler,
	}
}

func (wm *WorkerManager) AddWorker(w *Worker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workers[wm.workerCount] = w
	wm.workerCount++
	wm.wg.Add(1)
	w.LaunchWorker(wm.inCh, wm.stopCh)
}

func (wm *WorkerManager) RemoveWorker() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if wm.workerCount > 0 {
		wm.workerCount--
		wm.stopCh <- struct{}{}
	}
}

func (wm *WorkerManager) WorkerCount() int {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.workerCount
}

func (wm *WorkerManager) StopAllWorkers() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for i := 0; i < wm.workerCount; i++ {
		wm.stopCh <- struct{}{}
	}
	wm.workerCount = 0
}

func (wm *WorkerManager) WaitForAllWorkers() {
	wm.wg.Wait()
}

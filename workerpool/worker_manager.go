package workerpool

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Dispatcher is an interface for managing the worker pool.
type WorkerPoolManager interface {
	AddWorker(w *Worker)
	RemoveWorker(minWorkers int)
	ScaleWorkers(ctx context.Context)
	MakeRequest(Request)
	Stop(ctx context.Context)
}

type WorkerManager struct {
	workers       map[int]*Worker
	workerCount   int
	minWorkers    int
	maxWorkers    int
	loadThreshold int
	mu            sync.Mutex
	wg            *sync.WaitGroup
	inCh          chan Request
	stopCh        chan struct{}
	reqHandler    map[int]RequestHandler
}

func NewWorkerManager(
	wg *sync.WaitGroup,
	inCh chan Request,
	stopCh chan struct{},
	reqHandler map[int]RequestHandler,
	minWorker, maxWorker, loadThreshold int,
) *WorkerManager {
	return &WorkerManager{
		workers:       make(map[int]*Worker),
		wg:            wg,
		inCh:          inCh,
		stopCh:        stopCh,
		reqHandler:    reqHandler,
		minWorkers:    minWorker,
		maxWorkers:    maxWorker,
		loadThreshold: loadThreshold,
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

func (wm *WorkerManager) ScaleWorkers(ctx context.Context) {
	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Scaler: Context cancelled, stopping scaling")
			return
		case <-ticker.C:
			wm.scale()
		}
	}
}

func (wm *WorkerManager) scale() {
	load := len(wm.inCh)
	currentWorkers := wm.WorkerCount()

	if load > wm.loadThreshold && currentWorkers < wm.maxWorkers {
		log.Info().Msg("Scaling Up")
		newWorker := &Worker{
			Wg:         wm.wg,
			Id:         currentWorkers,
			ReqHandler: wm.reqHandler,
		}
		wm.AddWorker(newWorker)
	} else if float64(load) < float64(LoadThresholdScaleDownRatio)*float64(wm.loadThreshold) && currentWorkers > wm.minWorkers {
		log.Info().Msg("Scaling Down")
		wm.RemoveWorker()
	}
}

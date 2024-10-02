package workerpool

import (
	"context"
	"fmt"
	"time"
)

type Scaler struct {
	workerManager *WorkerManager
	inCh          chan Request
	minWorkers    int
	maxWorkers    int
	loadThreshold int
}

func NewScaler(wm *WorkerManager, inCh chan Request, min, max, threshold int) *Scaler {
	return &Scaler{
		workerManager: wm,
		inCh:          inCh,
		minWorkers:    min,
		maxWorkers:    max,
		loadThreshold: threshold,
	}
}

func (s *Scaler) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scaler: Context cancelled, stopping scaling")
			return
		case <-ticker.C:
			s.scale()
		}
	}
}

func (s *Scaler) scale() {
	load := len(s.inCh)
	currentWorkers := s.workerManager.WorkerCount()

	if load > s.loadThreshold && currentWorkers < s.maxWorkers {
		fmt.Println("Scaling Up")
		newWorker := &Worker{
			Wg:         s.workerManager.wg,
			Id:         currentWorkers,
			ReqHandler: s.workerManager.reqHandler,
		}
		s.workerManager.AddWorker(newWorker)
	} else if float64(load) < float64(LoadThresholdScaleDownRatio)*float64(s.loadThreshold) && currentWorkers > s.minWorkers {
		fmt.Println("Scaling Down")
		s.workerManager.RemoveWorker()
	}
}

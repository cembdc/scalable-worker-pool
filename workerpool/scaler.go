package workerpool

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
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
			log.Info().Msg("Scaler: Context cancelled, stopping scaling")
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
		log.Info().Msg("Scaling Up")
		newWorker := &Worker{
			Wg:         s.workerManager.wg,
			Id:         currentWorkers,
			ReqHandler: s.workerManager.reqHandler,
		}
		s.workerManager.AddWorker(newWorker)
	} else if float64(load) < float64(LoadThresholdScaleDownRatio)*float64(s.loadThreshold) && currentWorkers > s.minWorkers {
		log.Info().Msg("Scaling Down")
		s.workerManager.RemoveWorker()
	}
}

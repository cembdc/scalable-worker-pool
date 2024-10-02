package workerpool

import "context"

// WorkerLauncher is an interface for launching workers.
type WorkerLauncher interface {
	LaunchWorker(in chan Request, stopCh chan struct{})
}

// Dispatcher is an interface for managing the worker pool.
type WorkerPoolManager interface {
	AddWorker(w *Worker)
	RemoveWorker(minWorkers int)
	ScaleWorkers(ctx context.Context, minWorkers, maxWorkers, loadThreshold int)
	MakeRequest(Request)
	Stop(ctx context.Context)
}

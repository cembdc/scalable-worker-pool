package main

import (
	"context"
	"fmt"
	"runtime"
	wp "scalable-worker-pool/workerpool"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setMaxProcs()

	bufferSize := 50000
	requests := 50000
	var wg sync.WaitGroup

	reqHandler := createRequestHandler()
	dispatcher := wp.NewDispatcher(bufferSize, &wg, wp.DefaultMaxWorkers, reqHandler)

	startWorkers(dispatcher, &wg, reqHandler, wp.DefaultMinWorkers)

	go dispatcher.ScaleWorkers(ctx, wp.DefaultMinWorkers, wp.DefaultMaxWorkers, wp.DefaultLoadThreshold)

	sendRequests(dispatcher, requests)

	gracefulShutdown(dispatcher, ctx)
}

func setMaxProcs() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("Running with %d CPUs\n", numCPU)
}

func createRequestHandler() map[int]wp.RequestHandler {
	return map[int]wp.RequestHandler{
		1: func(data interface{}) error {

			return nil
		},
	}
}

func startWorkers(dispatcher wp.WorkerPoolManager, wg *sync.WaitGroup, reqHandler map[int]wp.RequestHandler, minWorkers int) {
	for i := 0; i < minWorkers; i++ {
		fmt.Printf("Starting worker with id %d\n", i)
		w := wp.NewWorker(i, wg, reqHandler)
		dispatcher.AddWorker(w)
	}
}

func sendRequests(dispatcher wp.WorkerPoolManager, requestCount int) {
	for i := 0; i < requestCount; i++ {
		req := wp.Request{
			Data:    fmt.Sprintf("(Msg_id: %d) -> Hello", i),
			Handler: func(result interface{}) error { return nil },
			Type:    1,
			Timeout: 5 * time.Second,
		}
		dispatcher.MakeRequest(req)
	}
}

func gracefulShutdown(dispatcher wp.WorkerPoolManager, ctx context.Context) {
	stopCtx, stopCancel := context.WithTimeout(ctx, 30*time.Second)
	defer stopCancel()

	dispatcher.Stop(stopCtx)
	fmt.Println("Exiting main!")
}

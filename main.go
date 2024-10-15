package main

import (
	"context"
	"fmt"
	"runtime"
	wp "scalable-worker-pool/workerpool"

	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	// SetLogger()

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

	// done := make(chan bool)
	// <-done
	// time.Sleep(5 * time.Second)
	gracefulShutdown(dispatcher, ctx)
}

func setMaxProcs() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	log.Info().Msgf("Running with %d CPUs", numCPU)
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
		log.Info().Msgf("Starting worker with id %d", i)
		w := wp.NewWorker(i, wg, reqHandler)
		dispatcher.AddWorker(w)
	}
}

func sendRequests(dispatcher wp.WorkerPoolManager, requestCount int) {
	for i := 0; i < requestCount; i++ {
		req := wp.Request{
			Data:    fmt.Sprintf("Hello MsgId: %d", i),
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
	log.Info().Msg("Exiting main!")
}

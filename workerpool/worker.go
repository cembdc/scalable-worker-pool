package workerpool

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// WorkerLauncher is an interface for launching workers.
type WorkerLauncher interface {
	LaunchWorker(in chan Request, stopCh chan struct{})
}

// Worker represents a worker that processes requests.
type Worker struct {
	Id         int
	Wg         *sync.WaitGroup
	ReqHandler map[int]RequestHandler
}

func NewWorker(id int, wg *sync.WaitGroup, reqHandler map[int]RequestHandler) *Worker {
	return &Worker{
		Id:         id,
		Wg:         wg,
		ReqHandler: reqHandler,
	}
}

// LaunchWorker launches the worker to process incoming requests.
// It runs in a separate goroutine, continuously listening for incoming requests on the input channel.
// The worker gracefully stops when either the input channel is closed or it receives a stop signal.
func (w *Worker) LaunchWorker(in chan Request, stopCh chan struct{}) {
	go func() {
		defer w.Wg.Done()
		for {
			select {
			case msg, open := <-in:
				if !open {
					// If the channel is closed, stop processing and return
					// if we skip close channel check then after closing channel,
					// worker keep reading empty values from closed channel.
					log.Info().Msgf("Stopping worker %d", w.Id)
					return
				}
				w.processRequest(msg)
				time.Sleep(1 * time.Microsecond) // Small delay to prevent tight loop
			case <-stopCh:
				log.Info().Msgf("Stopping worker %d", w.Id)
				return
			}
		}
	}()
}

// processRequest processes a single request.
func (w *Worker) processRequest(msg Request) {
	log.Info().Msgf("Worker %d processing request: %v", w.Id, msg)
	var handler RequestHandler
	var ok bool
	if handler, ok = w.ReqHandler[msg.Type]; !ok {
		log.Info().Msgf("Handler not implemented: workerID: %d", w.Id)
	} else {
		if msg.Timeout == 0 {
			msg.Timeout = time.Duration(DefaultWorkerTimeout)
		}
		for attempt := 0; attempt <= msg.MaxRetries; attempt++ {
			var err error
			done := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), msg.Timeout)
			defer cancel()

			go func() {
				err = handler(msg.Data)
				close(done)
			}()

			select {
			case <-done:
				if err == nil {
					return // Successfully processed
				}
				log.Error().Msgf("Worker %d: Error processing request: %v", w.Id, err)
			case <-ctx.Done():
				log.Info().Msgf("Worker %d: Timeout processing request: %v", w.Id, msg.Data)
			}
			log.Info().Msgf("Worker %d: Retry %d for request %v", w.Id, attempt, msg.Data)
		}
		log.Error().Msgf("Worker %d: Failed to process request %v after %d retries", w.Id, msg.Data, msg.MaxRetries)
	}
}

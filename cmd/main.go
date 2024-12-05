package main

import (
	"context"
	"fmt"
	"runtime"
	"scalable-worker-pool/internal/plugin"
	"scalable-worker-pool/internal/plugin/mqtt"
	"scalable-worker-pool/pkg/config"
	wp "scalable-worker-pool/pkg/workerpool"

	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize plugins
	pluginManager, err := initializePlugins(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize plugins")
	}

	// Start plugins
	if err := pluginManager.StartPlugins(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start plugins")
	}

	if err := pluginManager.StartRouting(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start routing")
	}

	// setMaxProcs()

	// bufferSize := 50000
	// requests := 50000
	// var wg sync.WaitGroup

	// reqHandler := createRequestHandler()
	// dispatcher := wp.NewDispatcher(bufferSize, &wg, wp.DefaultMaxWorkers, reqHandler)

	// startWorkers(dispatcher, &wg, reqHandler, wp.DefaultMinWorkers)

	// go dispatcher.ScaleWorkers(ctx)

	// sendRequests(dispatcher, requests)
	log.Info().Msg("Waiting for requests to complete")

	done := make(chan bool)
	<-done
	// // time.Sleep(5 * time.Second)
	// gracefulShutdown(dispatcher, ctx)
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

func registerPlugins(registry *plugin.Registry) {
	// Source plugins
	registry.RegisterSource("mqtt", mqtt.NewMQTTSource)
	// registry.RegisterSource("kafka", kafka.NewKafkaSource)
	// registry.RegisterSource("rabbitmq", rabbitmq.NewRabbitMQSource)

	// Target plugins
	registry.RegisterTarget("mqtt", mqtt.NewMQTTTarget)
	// registry.RegisterTarget("amqp", amqp.NewAMQPTarget)
	// registry.RegisterTarget("websocket", websocket.NewWebSocketTarget)
}

func initializePlugins(ctx context.Context) (*plugin.Manager, error) {
	// Create registry and register plugins
	registry := plugin.NewRegistry()
	registerPlugins(registry)

	// Load configuration
	configLoader := config.NewConfigLoader("./config.json")
	cfg, err := configLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create and initialize plugin manager
	manager := plugin.NewManager(ctx, registry)
	if err := manager.LoadPlugins(cfg); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}

	return manager, nil
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"scalable-worker-pool/internal/config"
	"scalable-worker-pool/internal/plugin"
	"scalable-worker-pool/internal/plugin/mqtt"

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
	log.Info().Msg("Waiting for requests to complete")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	gracefulShutdown(pluginManager, ctx)

	log.Info().Msg("Server stopped gracefully")
}

func setMaxProcs() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	log.Info().Msgf("Running with %d CPUs", numCPU)
}

func gracefulShutdown(pluginManager *plugin.Manager, ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pluginManager.StopRouting(shutdownCtx)
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

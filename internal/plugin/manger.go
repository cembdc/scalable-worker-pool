package plugin

import (
	"context"
	"fmt"
	"scalable-worker-pool/pkg/config"
	"scalable-worker-pool/pkg/workerpool"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	registry *Registry
	sources  map[string]SourcePlugin
	targets  map[string]TargetPlugin
	ctx      context.Context
	wp       workerpool.WorkerPoolManager
}

func NewManager(ctx context.Context, registry *Registry) *Manager {
	return &Manager{
		registry: registry,
		sources:  make(map[string]SourcePlugin),
		targets:  make(map[string]TargetPlugin),
		ctx:      ctx,
	}
}

func (m *Manager) LoadPlugins(cfg *config.Config) error {
	// Load source plugins
	for _, sourceCfg := range cfg.Sources {
		source, exists := m.registry.GetSource(sourceCfg.Type)
		if !exists {
			return fmt.Errorf("unknown source plugin type: %s", sourceCfg.Type)
		}

		if err := source.Initialize(m.ctx, sourceCfg.Config); err != nil {
			return fmt.Errorf("failed to initialize source %s: %w", sourceCfg.Type, err)
		}

		m.sources[sourceCfg.Type] = source
	}

	// Load target plugins
	for _, targetCfg := range cfg.Targets {
		target, exists := m.registry.GetTarget(targetCfg.Type)
		if !exists {
			return fmt.Errorf("unknown target plugin type: %s", targetCfg.Type)
		}

		if err := target.Initialize(m.ctx, targetCfg.Config); err != nil {
			return fmt.Errorf("failed to initialize target %s: %w", targetCfg.Type, err)
		}

		m.targets[targetCfg.Type] = target
	}
	return nil
}

func (m *Manager) StartPlugins(ctx context.Context) error {
	// Connect source plugins
	for _, source := range m.sources {
		if err := source.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect source %s: %w", source.Type(), err)
		}
	}

	// Connect target plugins
	for _, target := range m.targets {
		if err := target.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect target %s: %w", target.Type(), err)
		}
	}

	return nil
}

func (m *Manager) StopPlugins(ctx context.Context) error {
	var errs []error

	// Disconnect sources
	for _, source := range m.sources {
		if err := source.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect source %s: %w", source.Type(), err))
		}
	}

	// Disconnect targets
	for _, target := range m.targets {
		if err := target.Disconnect(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect target %s: %w", target.Type(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during plugin shutdown: %v", errs)
	}
	return nil
}

func (m *Manager) StartRouting(ctx context.Context) error {
	// Worker Pool setup
	var wg sync.WaitGroup
	bufferSize := 50000
	reqHandler := map[int]workerpool.RequestHandler{
		1: m.handleMessage,
	}

	m.wp = workerpool.NewDispatcher(bufferSize, &wg, workerpool.DefaultMaxWorkers, reqHandler)

	// Start workers
	for i := 0; i < workerpool.DefaultMinWorkers; i++ {
		worker := workerpool.NewWorker(i, &wg, reqHandler)
		m.wp.AddWorker(worker)
	}

	// Start scaling
	go m.wp.ScaleWorkers(ctx)

	// Start reading from each source
	for _, source := range m.sources {
		go m.readFromSource(ctx, source)
	}

	return nil
}

func (m *Manager) readFromSource(ctx context.Context, source SourcePlugin) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := source.Read(ctx)
			if err != nil {
				log.Error().Err(err).Str("source", source.Type()).Msg("Failed to read message")
				continue
			}

			// Create worker pool request
			req := workerpool.Request{
				Data:       msg,
				Type:       1,
				Timeout:    5 * time.Second,
				MaxRetries: 3,
			}

			m.wp.MakeRequest(req)
		}
	}
}

func (m *Manager) handleMessage(data interface{}) error {
	msg, ok := data.(*Message)
	if !ok {
		return fmt.Errorf("invalid message type")
	}

	var errs []error
	// Send to all targets
	for _, target := range m.targets {
		if err := target.Write(context.Background(), msg); err != nil {
			errs = append(errs, fmt.Errorf("failed to write to target %s: %w", target.Type(), err))
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors writing to targets: %v", errs)
	}
	return nil
}

func (m *Manager) StopRouting(ctx context.Context) error {
	// Graceful shutdown of worker pool
	m.wp.Stop(ctx)
	log.Info().Msg("Worker pool stopped")
	return nil
}

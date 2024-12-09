package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type ExchangeType string

const (
	DirectExchange  ExchangeType = "direct"
	TopicExchange   ExchangeType = "topic"
	FanoutExchange  ExchangeType = "fanout"
	HeadersExchange ExchangeType = "headers"
)

type AMQPConfig struct {
	URL          string            `json:"url"`
	Queue        string            `json:"queue"`
	Exchange     string            `json:"exchange"`
	ExchangeType ExchangeType      `json:"exchange_type"`
	RoutingKey   string            `json:"routing_key"`
	Properties   map[string]string `json:"properties"`
}

type AMQPType int

const (
	SourceType AMQPType = iota
	TargetType
)

type AMQPBase struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	config   AMQPConfig
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	amqpType AMQPType // Add type field
}

func (b *AMQPBase) Initialize(ctx context.Context, rawConfig json.RawMessage) error {
	if err := json.Unmarshal(rawConfig, &b.config); err != nil {
		return fmt.Errorf("amqp: config parse error: %w", err)
	}

	b.ctx, b.cancel = context.WithCancel(ctx)
	return b.validate()
}

func (b *AMQPBase) validate() error {
	// Common validations
	if b.config.URL == "" {
		return fmt.Errorf("amqp: URL required")
	}

	// Type specific validations
	switch b.amqpType {
	case SourceType:
		if b.config.Queue == "" {
			return fmt.Errorf("amqp: queue name required for source")
		}
	case TargetType:
		if b.config.Exchange == "" {
			return fmt.Errorf("amqp: exchange required for target")
		}
		if b.config.ExchangeType == "" {
			b.config.ExchangeType = DirectExchange // Default
		}
		// RoutingKey validation based on exchange type
		if (b.config.ExchangeType == DirectExchange || b.config.ExchangeType == TopicExchange) && b.config.RoutingKey == "" {
			return fmt.Errorf("amqp: routing key required for %s exchange", b.config.ExchangeType)
		}
	}

	return nil
}

func (b *AMQPBase) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	conn, err := amqp091.Dial(b.config.URL)
	if err != nil {
		return fmt.Errorf("amqp: connection failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("amqp: failed to open channel: %w", err)
	}

	b.conn = conn
	b.channel = ch

	return nil
}

func (b *AMQPBase) Disconnect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.channel != nil {
		if err := b.channel.Close(); err != nil {
			return fmt.Errorf("amqp: failed to close channel: %w", err)
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			return fmt.Errorf("amqp: failed to close connection: %w", err)
		}
	}
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *AMQPBase) HealthCheck(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.conn == nil || b.conn.IsClosed() {
		return fmt.Errorf("amqp: connection not established")
	}
	return nil
}

func (b *AMQPBase) Type() string {
	return "amqp"
}

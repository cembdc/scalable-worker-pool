package amqp

import (
	"context"
	"fmt"
	"scalable-worker-pool/internal/plugin"

	"github.com/rabbitmq/amqp091-go"
)

type AMQPTarget struct {
	AMQPBase
}

func NewAMQPTarget() plugin.TargetPlugin {
	return &AMQPTarget{
		AMQPBase: AMQPBase{
			amqpType: TargetType,
		},
	}
}

func (t *AMQPTarget) Write(ctx context.Context, msg *plugin.Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.channel == nil {
		return fmt.Errorf("amqp: channel not connected")
	}

	err := t.channel.PublishWithContext(ctx,
		t.config.Exchange,
		t.config.RoutingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg.Data,
		},
	)
	if err != nil {
		return fmt.Errorf("amqp: failed to publish message: %w", err)
	}

	return nil
}

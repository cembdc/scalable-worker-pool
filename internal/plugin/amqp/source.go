package amqp

import (
	"context"
	"fmt"
	"scalable-worker-pool/internal/plugin"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type AMQPSource struct {
	AMQPBase
	msgChan chan *plugin.Message
}

func NewAMQPSource() plugin.SourcePlugin {
	return &AMQPSource{
		AMQPBase: AMQPBase{
			amqpType: SourceType,
		},
		msgChan: make(chan *plugin.Message, 100),
	}
}

func (s *AMQPSource) Connect(ctx context.Context) error {
	if err := s.AMQPBase.Connect(ctx); err != nil {
		return err
	}

	msgs, err := s.channel.Consume(
		s.config.Queue,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("amqp: failed to register consumer: %w", err)
	}

	go s.handleDeliveries(msgs)
	return nil
}

func (s *AMQPSource) handleDeliveries(deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			s.msgChan <- &plugin.Message{
				Data: d.Body,
				Metadata: map[string]interface{}{
					"routing_key": d.RoutingKey,
					"exchange":    d.Exchange,
				},
				Timestamp: time.Now(),
			}
		}
	}
}

func (s *AMQPSource) Read(ctx context.Context) (*plugin.Message, error) {
	select {
	case msg := <-s.msgChan:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

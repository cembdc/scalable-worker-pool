package mqtt

import (
	"context"
	"fmt"
	"scalable-worker-pool/internal/plugin"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTSource struct {
	MQTTBase
	msgChan chan *plugin.Message
}

func NewMQTTSource() plugin.SourcePlugin {
	return &MQTTSource{
		msgChan: make(chan *plugin.Message, 100),
	}
}

func (s *MQTTSource) Connect(ctx context.Context) error {
	if err := s.MQTTBase.Connect(ctx); err != nil {
		return err
	}

	for _, topic := range s.config.Topics {
		if token := s.client.Subscribe(topic, byte(s.config.QoS), s.messageHandler); token.Wait() && token.Error() != nil {
			return fmt.Errorf("mqtt: subscribe failed for %s: %w", topic, token.Error())
		}
	}
	return nil
}

func (s *MQTTSource) messageHandler(client mqtt.Client, msg mqtt.Message) {
	select {
	case s.msgChan <- &plugin.Message{
		Data: msg.Payload(),
		Metadata: map[string]interface{}{
			"topic": msg.Topic(),
			"qos":   msg.Qos(),
		},
		Timestamp: time.Now(),
	}:
	case <-s.ctx.Done():
		return
	}
}

func (s *MQTTSource) Read(ctx context.Context) (*plugin.Message, error) {
	select {
	case msg := <-s.msgChan:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

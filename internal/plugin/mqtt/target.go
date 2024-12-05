package mqtt

import (
	"context"
	"fmt"
	"scalable-worker-pool/internal/plugin"
)

type MQTTTarget struct {
	MQTTBase
}

func NewMQTTTarget() plugin.TargetPlugin {
	return &MQTTTarget{}
}

func (t *MQTTTarget) Write(ctx context.Context, msg *plugin.Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.client.IsConnected() {
		return fmt.Errorf("mqtt: client not connected")
	}

	topic := t.config.Topics[0]

	token := t.client.Publish(topic, byte(t.config.QoS), false, msg.Data)
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("mqtt: publish failed: %w", token.Error())
	}

	return nil
}

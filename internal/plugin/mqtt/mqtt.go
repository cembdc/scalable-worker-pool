package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTConfig struct {
	Broker     string            `json:"broker"`
	Topics     []string          `json:"topics"`
	ClientID   string            `json:"client_id"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	QoS        int               `json:"qos"`
	Properties map[string]string `json:"properties"`
}

type MQTTBase struct {
	client mqtt.Client
	config MQTTConfig
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func (b *MQTTBase) Initialize(ctx context.Context, rawConfig json.RawMessage) error {
	if err := json.Unmarshal(rawConfig, &b.config); err != nil {
		return fmt.Errorf("mqtt: config parse error: %w", err)
	}

	b.ctx, b.cancel = context.WithCancel(ctx)
	return b.validate()
}

func (b *MQTTBase) validate() error {
	if b.config.Broker == "" {
		return fmt.Errorf("mqtt: broker URL required")
	}
	if len(b.config.Topics) == 0 {
		return fmt.Errorf("mqtt: at least one topic required")
	}
	return nil
}

func (b *MQTTBase) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	opts := mqtt.NewClientOptions().
		AddBroker(b.config.Broker).
		SetClientID(b.config.ClientID).
		SetUsername(b.config.Username).
		SetPassword(b.config.Password).
		SetAutoReconnect(true).
		SetConnectionLostHandler(b.onConnectionLost).
		SetOnConnectHandler(b.onConnect)

	b.client = mqtt.NewClient(opts)
	if token := b.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt: connection failed: %w", token.Error())
	}

	return nil
}

func (b *MQTTBase) Disconnect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.client != nil && b.client.IsConnected() {
		b.client.Disconnect(1000)
	}
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *MQTTBase) HealthCheck(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.client == nil || !b.client.IsConnected() {
		return fmt.Errorf("mqtt: client not connected")
	}
	return nil
}

func (b *MQTTBase) Type() string {
	return "mqtt"
}

func (b *MQTTBase) onConnectionLost(client mqtt.Client, err error) {
	// Log connection lost
}

func (b *MQTTBase) onConnect(client mqtt.Client) {
	// Log connection established
}

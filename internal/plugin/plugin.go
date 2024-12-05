package plugin

import (
	"context"
	"encoding/json"
	"time"
)

type Message struct {
	Data      []byte
	Metadata  map[string]interface{}
	Timestamp time.Time
}

type Plugin interface {
	Initialize(ctx context.Context, config json.RawMessage) error
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	HealthCheck(ctx context.Context) error
	Type() string
}

type SourcePlugin interface {
	Plugin
	Read(ctx context.Context) (*Message, error)
}

type TargetPlugin interface {
	Plugin
	Write(ctx context.Context, msg *Message) error
}

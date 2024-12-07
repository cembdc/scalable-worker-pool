package config

import (
	"encoding/json"
)

type Config struct {
	Sources []SourceConfig `json:"sources"`
	Targets []TargetConfig `json:"targets"`
}

type SourceConfig struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

type TargetConfig struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

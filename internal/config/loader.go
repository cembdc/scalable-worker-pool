package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigLoader struct {
	configPath string
}

func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
	}
}

func (l *ConfigLoader) Load() (*Config, error) {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("config file read error: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	if err := l.validate(&config); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return &config, nil
}

func (l *ConfigLoader) validate(cfg *Config) error {
	if len(cfg.Sources) == 0 && len(cfg.Targets) == 0 {
		return fmt.Errorf("at least one source or target required")
	}

	for _, source := range cfg.Sources {
		if source.Type == "" {
			return fmt.Errorf("source type cannot be empty")
		}
		if source.Config == nil {
			return fmt.Errorf("source config cannot be empty for type: %s", source.Type)
		}
	}

	for _, target := range cfg.Targets {
		if target.Type == "" {
			return fmt.Errorf("target type cannot be empty")
		}
		if target.Config == nil {
			return fmt.Errorf("target config cannot be empty for type: %s", target.Type)
		}
	}

	return nil
}

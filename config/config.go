package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ManagerConfig struct {
	BufferSize      int
	MinWorkers      int
	MaxWorkers      int
	DefaultTimeout  int
	MaxRetries      int
	MessageChanSize int
}

func LoadManagerConfig() (*ManagerConfig, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	return &ManagerConfig{
		BufferSize:      getEnvAsInt("BUFFER_SIZE", 50000),
		MinWorkers:      getEnvAsInt("MIN_WORKERS", 3),
		MaxWorkers:      getEnvAsInt("MAX_WORKERS", 20),
		DefaultTimeout:  getEnvAsInt("DEFAULT_TIMEOUT", 5),
		MaxRetries:      getEnvAsInt("MAX_RETRIES", 3),
		MessageChanSize: getEnvAsInt("MESSAGE_CHANNEL_SIZE", 100),
	}, nil
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

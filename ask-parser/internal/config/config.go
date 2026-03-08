package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultInputQueue    = "askorium.render.output"
	defaultOutputQueue   = "askorium.parser.output"
	defaultPrefetchCount = "10"
	defaultLogLevel      = "info"
)

type Config struct {
	AmqpURL       string
	InputQueue    string
	OutputQueue   string
	PrefetchCount int
	LogLevel      string
}

func Load() (*Config, error) {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		return nil, fmt.Errorf("AMQP_URL environment variable not set")
	}

	inputQueue := getEnvOrDefault("INPUT_QUEUE", defaultInputQueue)
	outputQueue := getEnvOrDefault("OUTPUT_QUEUE", defaultOutputQueue)
	logLevel := strings.ToLower(getEnvOrDefault("LOG_LEVEL", defaultLogLevel))

	prefetchStr := getEnvOrDefault("PREFETCH_COUNT", defaultPrefetchCount)
	prefetchCount, err := strconv.Atoi(prefetchStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PREFETCH_COUNT %q: %w", prefetchStr, err)
	}
	return &Config{
		AmqpURL:       amqpURL,
		InputQueue:    inputQueue,
		OutputQueue:   outputQueue,
		PrefetchCount: prefetchCount,
		LogLevel:      logLevel,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultRequestQueue  = "scrape.request"
	defaultResponseQueue = "scrape.response"
	defaultPrefetchCount = "5"
	defaultLogLevel      = "info"
)

type Config struct {
	RendererURL   string
	AmqpURL       string
	RequestQueue  string
	ResponseQueue string
	PrefetchCount int
	LogLevel      string
}

func Load() (*Config, error) {
	rendererURL := os.Getenv("RENDERER_URL")
	if rendererURL == "" {
		return nil, fmt.Errorf("RENDERER_URL environment variable not set")
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		return nil, fmt.Errorf("AMQP_URL environment variable not set")
	}

	requestQueue := getEnvOrDefault("AMQP_REQUEST_QUEUE", defaultRequestQueue)
	responseQueue := getEnvOrDefault("AMQP_RESPONSE_QUEUE", defaultResponseQueue)
	logLevel := strings.ToLower(getEnvOrDefault("LOG_LEVEL", defaultLogLevel))

	prefetchStr := getEnvOrDefault("PREFETCH_COUNT", defaultPrefetchCount)
	prefetchCount, err := strconv.Atoi(prefetchStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PREFETCH_COUNT %q: %w", prefetchStr, err)
	}
	return &Config{
		RendererURL:   rendererURL,
		AmqpURL:       amqpURL,
		RequestQueue:  requestQueue,
		ResponseQueue: responseQueue,
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

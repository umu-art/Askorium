package config

import (
	"github.com/spf13/viper"
	"os"
	"regexp"
)

type Config struct {
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Scrapper ScrapperConfig `mapstructure:"scrapper"`
	Server   ServerConfig   `mapstructure:"server"`
}

type RabbitMQConfig struct {
	URL             string `mapstructure:"url"`
	RequestQueue    string `mapstructure:"request-queue"`
	RetryQueue      string `mapstructure:"retry-queue"`
	DeadLetterQueue string `mapstructure:"dlq"`
	ResponseQueue   string `mapstructure:"response-queue"`
	PrefetchCount   int    `mapstructure:"prefetch-count"`
	MaxRetries      int    `mapstructure:"max-retries"`
	RetryDelayMs    int    `mapstructure:"retry-delay-ms"`
}

type ScrapperConfig struct {
	TimeoutMs int `mapstructure:"timeout-ms"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./src/internal/config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Заменяем ${ENV_VAR} на значения из переменных окружения
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		resolved := resolveEnvVars(val)
		viper.Set(key, resolved)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// resolveEnvVars заменяет ${VAR} и ${VAR:default} на значения переменных окружения
func resolveEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{(\w+)(?::([^}]*))?}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		parts := re.FindStringSubmatch(match)
		envKey := parts[1]
		defaultVal := parts[2]

		if val, ok := os.LookupEnv(envKey); ok {
			return val
		}
		if defaultVal != "" {
			return defaultVal
		}
		return match
	})
}

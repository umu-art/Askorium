package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ask-parser/internal/amqp"
	amqpimpl "ask-parser/internal/amqp/impl"
	"ask-parser/internal/config"
	"ask-parser/internal/handler"
	"ask-parser/internal/parser"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)
	logger.Info("starting ask-parser")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn := amqp.NewConnection(cfg.AmqpURL, logger)
	if err := conn.Connect(ctx); err != nil {
		logger.Error("failed to connect to amqp", "error", err)
		os.Exit(1)
	}

	broker := amqp.NewBroker(conn, logger)

	publisher := broker.RegisterPublisher(amqpimpl.PublisherConfig{
		Exchange:          "",
		DefaultRoutingKey: cfg.OutputQueue,
		Durable:           true,
		Queue:             cfg.OutputQueue,
		DLXExchange:       cfg.OutputQueue + ".dlx",
		DLQQueue:          cfg.OutputQueue + ".dlq",
	})

	askParser := parser.NewParser()
	h := handler.NewHandler(askParser, publisher, logger)

	broker.RegisterConsumer(amqpimpl.ConsumerConfig{
		Queue:       cfg.InputQueue,
		Prefetch:    cfg.PrefetchCount,
		Durable:     true,
		DLXExchange: cfg.InputQueue + ".dlx",
		DLQQueue:    cfg.InputQueue + ".dlq",
	}, h)

	if err := broker.Start(ctx); err != nil {
		logger.Error("broker stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("shutting down")
}

func setupLogger(level string) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(logger)
	return logger
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ask-crawler/src/app"
	infra_amqp "ask-crawler/src/infra/amqp"
	amqpimpl "ask-crawler/src/infra/amqp/impl"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	amqpURL := envOr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn := infra_amqp.NewConnection(amqpURL, logger)
	if err := conn.Connect(ctx); err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}

	broker := infra_amqp.NewBroker(conn, logger)

	renderPublisher := broker.RegisterPublisher(amqpimpl.PublisherConfig{
		DefaultRoutingKey: "askorium.render.input",
		Durable:           true,
	})

	eventPublisher := broker.RegisterPublisher(amqpimpl.PublisherConfig{
		Exchange:     "askorium.crawler.output",
		ExchangeKind: "topic",
		Durable:      false,
	})

	crawlerService := app.NewCrawlerService(renderPublisher, eventPublisher, logger)

	broker.RegisterConsumer(amqpimpl.ConsumerConfig{
		Queue:   "askorium.crawler.input",
		Durable: true,
	}, crawlerService.HandleTask)

	broker.RegisterConsumer(amqpimpl.ConsumerConfig{
		Queue:   "askorium.parser.output",
		Durable: true,
	}, crawlerService.HandleScrapeResult)

	logger.Info("crawler started")
	if err := broker.Start(ctx); err != nil {
		logger.Error("broker error", "error", err)
		os.Exit(1)
	}

	logger.Info("crawler stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

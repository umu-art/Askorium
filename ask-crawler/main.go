package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ask-crawler/src/app"
	applib "ask-crawler/src/app/lib"
	infra_amqp "ask-crawler/src/infra/amqp"
	amqpimpl "ask-crawler/src/infra/amqp/impl"
	redisinfra "ask-crawler/src/infra/redis"
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
		DefaultRoutingKey: "askorium.crawler.output",
		Durable:           true,
	})

	factory := buildFrontierFactory(ctx, logger)

	batchMode := envOr("CRAWLER_BATCH_MODE", "false") == "true"
	crawlerService := app.NewCrawlerService(renderPublisher, eventPublisher, logger, batchMode, factory)

	broker.RegisterConsumer(amqpimpl.ConsumerConfig{
		Queue:   "askorium.crawler.input",
		Durable: true,
	}, crawlerService.HandleTask)

	broker.RegisterConsumer(amqpimpl.ConsumerConfig{
		Queue:       "askorium.parser.output",
		Durable:     true,
		DLXExchange: "askorium.parser.output.dlx",
		DLQQueue:    "askorium.parser.output.dlq",
	}, crawlerService.HandleScrapeResult)

	go crawlerService.WatchTTL(ctx)
	go crawlerService.WatchBatch(ctx)

	logger.Info("crawler started")
	if err := broker.Start(ctx); err != nil {
		logger.Error("broker error", "error", err)
		os.Exit(1)
	}

	logger.Info("crawler stopped")
}

// buildFrontierFactory выбирает Redis или in-memory реализацию по REDIS_URL.
func buildFrontierFactory(ctx context.Context, logger *slog.Logger) applib.FrontierFactory {
	redisURL := envOr("REDIS_URL", "")
	if redisURL == "" {
		logger.Info("frontier: using in-memory store")
		return applib.NewMemFrontierFactory()
	}

	rc, err := redisinfra.NewClient(redisURL, logger)
	if err != nil {
		logger.Error("frontier: invalid Redis URL, falling back to in-memory", "error", err)
		return applib.NewMemFrontierFactory()
	}
	if err := rc.Ping(ctx); err != nil {
		logger.Error("frontier: Redis unreachable, falling back to in-memory", "error", err)
		return applib.NewMemFrontierFactory()
	}

	logger.Info("frontier: using Redis store", "url", redisURL)
	return redisinfra.NewFrontierFactory(rc)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

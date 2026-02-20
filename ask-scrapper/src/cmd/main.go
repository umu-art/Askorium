package main

import (
	"ask-scrapper/internal/handler"
	"ask-scrapper/internal/parser"
	"ask-scrapper/internal/pipeline"
	"ask-scrapper/internal/renderer"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ask-scrapper/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel)
	slog.Info("starting ask-scrapper")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	askRenderer := renderer.NewRenderer(cfg.RendererURL)

	askParser := parser.NewParser()
	scrapePipeline := pipeline.NewPipeline(askRenderer, askParser)

	consumer := handler.NewConsumer(cfg, scrapePipeline)
	defer consumer.Shutdown()
	if err := consumer.Start(ctx); err != nil {
		slog.Error("consumer stopped with error", "error", err)
		os.Exit(1)
	}

	slog.Info("shutting down")
}

func setupLogger(level string) {
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})))
}

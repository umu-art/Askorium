package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"ask-scrapper/internal/config"
	"ask-scrapper/internal/pipeline"

	scrappermodel "github.com/omo-ri/askorium/go-scrapper-api"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	cfg      *config.Config
	pipeline pipeline.Pipeline
	conn     *amqp.Connection
	channel  *amqp.Channel
}

func NewConsumer(cfg *config.Config, p pipeline.Pipeline) *Consumer {
	return &Consumer{cfg: cfg, pipeline: p}
}

func (c *Consumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.cfg.AmqpURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	c.channel = ch

	_, err = ch.QueueDeclare(c.cfg.RequestQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare request queue: %w", err)
	}
	_, err = ch.QueueDeclare(c.cfg.ResponseQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare response queue: %w", err)
	}

	err = ch.Qos(c.cfg.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := ch.Consume(
		c.cfg.RequestQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	slog.Info("consumer started",
		"request_queue", c.cfg.RequestQueue,
		"response_queue", c.cfg.ResponseQueue,
		"prefetch", c.cfg.PrefetchCount,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("consumer stopping: context cancelled")
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed unexpectedly")
			}
			c.handleDelivery(ctx, d)
		}
	}
}

func (c *Consumer) Shutdown() {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			slog.Error("failed to close channel", "error", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			slog.Error("failed to close connection", "error", err)
		}
	}
	slog.Info("consumer shutdown complete")
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	var req scrappermodel.ScrapeRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		slog.Error("failed to unmarshal request", "error", err, "body", string(d.Body))
		_ = d.Nack(false, false)
		return
	}

	slog.Info("received task", "task_id", req.GetTaskId(), "url", req.GetUrl())

	resp := c.pipeline.Process(ctx, &req)

	body, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal response", "task_id", req.GetTaskId(), "error", err)
		_ = d.Nack(false, false)
		return
	}

	err = c.channel.PublishWithContext(ctx,
		"",
		c.cfg.ResponseQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		slog.Error("failed to publish response", "task_id", req.GetTaskId(), "error", err)
		_ = d.Nack(false, false)
		return
	}

	if err := d.Ack(false); err != nil {
		slog.Error("failed to ack", "task_id", req.GetTaskId(), "error", err)
	}

	slog.Info("task completed", "task_id", req.GetTaskId(), "success", resp.GetSuccess())
}

package amqpimpl

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckAction int

const (
	// Ack — сообщение обработано успешно, подтвердить.
	Ack AckAction = iota
	// NackRequeue — временная ошибка (publish не прошёл и т.п.), вернуть в очередь.
	NackRequeue
	// NackDiscard — сообщение невалидно (битый JSON и т.п.), отклонить без requeue → DLQ.
	NackDiscard
)

type Handler func(ctx context.Context, msg []byte) AckAction

type ConsumerConfig struct {
	Queue        string
	Exchange     string
	ExchangeKind string
	RoutingKey   string
	Prefetch     int
	Durable      bool

	// DLX — если непусто, для очереди Queue создаётся Dead Letter Exchange + Dead Letter Queue.
	// Сообщения, отклонённые через NackDiscard, попадают в DLQQueue через DLXExchange.
	DLXExchange string
	DLQQueue    string
}

type Consumer struct {
	config  ConsumerConfig
	handler Handler
	logger  *slog.Logger
}

func NewConsumer(config ConsumerConfig, handler Handler, logger *slog.Logger) *Consumer {
	return &Consumer{config: config, handler: handler, logger: logger}
}

func (c *Consumer) Queue() string {
	return c.config.Queue
}

func (c *Consumer) Start(ctx context.Context, openChannel func() (*amqp.Channel, error)) error {
	ch, err := openChannel()
	if err != nil {
		return fmt.Errorf("amqp consumer: open channel: %w", err)
	}

	if err := c.declareTopology(ch); err != nil {
		ch.Close()
		return err
	}

	prefetch := c.config.Prefetch
	if prefetch == 0 {
		prefetch = 1
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		ch.Close()
		return fmt.Errorf("amqp consumer: set qos: %w", err)
	}

	msgs, err := ch.Consume(
		c.config.Queue,
		"",    // consumer tag
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("amqp consumer: start consume %q: %w", c.config.Queue, err)
	}

	go c.loop(ctx, ch, msgs)
	return nil
}

func (c *Consumer) declareTopology(ch *amqp.Channel) error {
	durable := c.config.Durable

	// Dead letter exchange + queue (если настроено).
	if c.config.DLXExchange != "" {
		if err := ch.ExchangeDeclare(c.config.DLXExchange, "fanout", durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: declare DLX %q: %w", c.config.DLXExchange, err)
		}

		dlqName := c.config.DLQQueue
		if dlqName == "" {
			dlqName = c.config.Queue + ".dlq"
		}
		if _, err := ch.QueueDeclare(dlqName, durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: declare DLQ %q: %w", dlqName, err)
		}

		if err := ch.QueueBind(dlqName, "", c.config.DLXExchange, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: bind DLQ %q to %q: %w", dlqName, c.config.DLXExchange, err)
		}
	}

	// Основной exchange (если настроен).
	if c.config.Exchange != "" {
		kind := c.config.ExchangeKind
		if kind == "" {
			kind = "direct"
		}
		if err := ch.ExchangeDeclare(c.config.Exchange, kind, durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: declare exchange %q: %w", c.config.Exchange, err)
		}
	}

	// Основная очередь, с привязкой к DLX если настроено.
	queueArgs := amqp.Table{}
	if c.config.DLXExchange != "" {
		queueArgs["x-dead-letter-exchange"] = c.config.DLXExchange
	}
	if _, err := ch.QueueDeclare(c.config.Queue, durable, false, false, false, queueArgs); err != nil {
		return fmt.Errorf("amqp consumer: declare queue %q: %w", c.config.Queue, err)
	}

	// Bind основной очереди к exchange (если настроен).
	if c.config.Exchange != "" {
		if err := ch.QueueBind(c.config.Queue, c.config.RoutingKey, c.config.Exchange, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: bind queue %q: %w", c.config.Queue, err)
		}
	}

	return nil
}

func (c *Consumer) loop(ctx context.Context, ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	defer ch.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			action := c.handler(ctx, msg.Body)
			switch action {
			case Ack:
				if err := msg.Ack(false); err != nil {
					c.logger.Error("amqp consumer: ack failed", "queue", c.config.Queue, "error", err)
				}
			case NackRequeue:
				if err := msg.Nack(false, true); err != nil {
					c.logger.Error("amqp consumer: nack(requeue) failed", "queue", c.config.Queue, "error", err)
				}
			case NackDiscard:
				if err := msg.Nack(false, false); err != nil {
					c.logger.Error("amqp consumer: nack(discard) failed", "queue", c.config.Queue, "error", err)
				}
			}
		}
	}
}

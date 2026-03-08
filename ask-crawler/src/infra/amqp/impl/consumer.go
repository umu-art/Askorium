package amqpimpl

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, msg []byte) error

type ConsumerConfig struct {
	Queue string
	// Exchange — имя exchange. Если пусто — используется default exchange, RoutingKey = имя очереди.
	Exchange string
	// ExchangeKind — тип exchange: "direct", "topic", "fanout". По умолчанию "direct".
	ExchangeKind string
	RoutingKey   string
	Prefetch     int
	Durable      bool
	DLXExchange  string
	DLQQueue     string
}

type Consumer struct {
	config  ConsumerConfig
	handler Handler
}

func NewConsumer(config ConsumerConfig, handler Handler) *Consumer {
	return &Consumer{config: config, handler: handler}
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

	if c.config.Exchange != "" {
		kind := c.config.ExchangeKind
		if kind == "" {
			kind = "direct"
		}
		if err := ch.ExchangeDeclare(c.config.Exchange, kind, durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: declare exchange %q: %w", c.config.Exchange, err)
		}
	}

	var queueArgs amqp.Table
	if c.config.DLXExchange != "" {
		queueArgs = amqp.Table{"x-dead-letter-exchange": c.config.DLXExchange}
		if err := ch.ExchangeDeclare(c.config.DLXExchange, "fanout", true, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp consumer: declare dlx %q: %w", c.config.DLXExchange, err)
		}
		if c.config.DLQQueue != "" {
			if _, err := ch.QueueDeclare(c.config.DLQQueue, true, false, false, false, nil); err != nil {
				return fmt.Errorf("amqp consumer: declare dlq %q: %w", c.config.DLQQueue, err)
			}
			if err := ch.QueueBind(c.config.DLQQueue, "", c.config.DLXExchange, false, nil); err != nil {
				return fmt.Errorf("amqp consumer: bind dlq %q: %w", c.config.DLQQueue, err)
			}
		}
	}

	if _, err := ch.QueueDeclare(c.config.Queue, durable, false, false, false, queueArgs); err != nil {
		return fmt.Errorf("amqp consumer: declare queue %q: %w", c.config.Queue, err)
	}

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
			err := c.handler(ctx, msg.Body)
			if err != nil {
				_ = msg.Nack(false, true)
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}

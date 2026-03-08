package amqpimpl

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	Exchange          string
	ExchangeKind      string
	DefaultRoutingKey string
	Durable           bool

	Queue       string
	DLXExchange string
	DLQQueue    string
}

type Publisher struct {
	config      PublisherConfig
	openChannel func() (*amqp.Channel, error)
	ch          *amqp.Channel
	mu          sync.Mutex
	logger      *slog.Logger
}

func NewPublisher(config PublisherConfig, openChannel func() (*amqp.Channel, error), logger *slog.Logger) *Publisher {
	return &Publisher{config: config, openChannel: openChannel, logger: logger}
}

func (p *Publisher) Init() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.init()
}

func (p *Publisher) init() error {
	ch, err := p.openChannel()
	if err != nil {
		return fmt.Errorf("amqp publisher: open channel: %w", err)
	}

	if p.config.Exchange != "" {
		kind := p.config.ExchangeKind
		if kind == "" {
			kind = "direct"
		}
		if err := ch.ExchangeDeclare(p.config.Exchange, kind, p.config.Durable, false, false, false, nil); err != nil {
			ch.Close()
			return fmt.Errorf("amqp publisher: declare exchange %q: %w", p.config.Exchange, err)
		}
	}

	if p.config.Queue != "" {
		if err := p.declareQueueTopology(ch); err != nil {
			ch.Close()
			return err
		}
	}

	p.ch = ch
	return nil
}

func (p *Publisher) declareQueueTopology(ch *amqp.Channel) error {
	durable := p.config.Durable

	if p.config.DLXExchange != "" {
		if err := ch.ExchangeDeclare(p.config.DLXExchange, "fanout", durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp publisher: declare DLX %q: %w", p.config.DLXExchange, err)
		}

		dlqName := p.config.DLQQueue
		if dlqName == "" {
			dlqName = p.config.Queue + ".dlq"
		}
		if _, err := ch.QueueDeclare(dlqName, durable, false, false, false, nil); err != nil {
			return fmt.Errorf("amqp publisher: declare DLQ %q: %w", dlqName, err)
		}

		if err := ch.QueueBind(dlqName, "", p.config.DLXExchange, false, nil); err != nil {
			return fmt.Errorf("amqp publisher: bind DLQ %q to %q: %w", dlqName, p.config.DLXExchange, err)
		}
	}

	queueArgs := amqp.Table{}
	if p.config.DLXExchange != "" {
		queueArgs["x-dead-letter-exchange"] = p.config.DLXExchange
	}
	if _, err := ch.QueueDeclare(p.config.Queue, durable, false, false, false, queueArgs); err != nil {
		return fmt.Errorf("amqp publisher: declare queue %q: %w", p.config.Queue, err)
	}

	if p.config.Exchange != "" {
		if err := ch.QueueBind(p.config.Queue, p.config.DefaultRoutingKey, p.config.Exchange, false, nil); err != nil {
			return fmt.Errorf("amqp publisher: bind queue %q: %w", p.config.Queue, err)
		}
	}

	return nil
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	return p.PublishWithKey(ctx, p.config.DefaultRoutingKey, body)
}

func (p *Publisher) PublishWithKey(ctx context.Context, routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil {
		return fmt.Errorf("amqp publisher: channel is not initialized")
	}

	err := p.ch.PublishWithContext(ctx,
		p.config.Exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("amqp publisher: publish to %q/%q: %w", p.config.Exchange, routingKey, err)
	}
	return nil
}

func (p *Publisher) Reinit() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		p.ch.Close()
		p.ch = nil
	}
	return p.init()
}

package amqpimpl

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	// Exchange — имя exchange. Если пусто — используется default exchange, RoutingKey = имя очереди.
	Exchange string
	// ExchangeKind — тип exchange: "direct", "topic", "fanout". По умолчанию "direct".
	ExchangeKind string
	DefaultRoutingKey string
	Durable bool
}

type Publisher struct {
	config        PublisherConfig
	openChannel   func() (*amqp.Channel, error)
	ch            *amqp.Channel
	mu            sync.Mutex
}

func NewPublisher(config PublisherConfig, openChannel func() (*amqp.Channel, error)) *Publisher {
	return &Publisher{config: config, openChannel: openChannel}
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

	p.ch = ch
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

func (p *Publisher) Reinit() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		p.ch.Close()
		p.ch = nil
	}
	_ = p.init()
}

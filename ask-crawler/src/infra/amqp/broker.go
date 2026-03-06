package amqp

import (
	"context"
	"fmt"
	"log/slog"

	amqpimpl "ask-crawler/src/infra/amqp/impl"
)

type Broker struct {
	conn       *Connection
	consumers  []*amqpimpl.Consumer
	publishers []*amqpimpl.Publisher
	logger     *slog.Logger
}

func NewBroker(conn *Connection, logger *slog.Logger) *Broker {
	return &Broker{conn: conn, logger: logger}
}

func (b *Broker) RegisterConsumer(config amqpimpl.ConsumerConfig, handler amqpimpl.Handler) {
	b.consumers = append(b.consumers, amqpimpl.NewConsumer(config, handler))
}

func (b *Broker) RegisterPublisher(config amqpimpl.PublisherConfig) MessagePublisher {
	p := amqpimpl.NewPublisher(config, b.conn.Channel)
	b.publishers = append(b.publishers, p)
	return p
}

func (b *Broker) Start(ctx context.Context) error {
	if err := b.initAll(ctx); err != nil {
		return err
	}

	b.conn.OnReconnect(func() {
		b.logger.Info("amqp: reinitializing after reconnect")
		if err := b.initAll(ctx); err != nil {
			b.logger.Error("amqp: reinit failed after reconnect", "error", err)
		}
	})

	b.logger.Info("amqp: broker started",
		"consumers", len(b.consumers),
		"publishers", len(b.publishers),
	)

	<-ctx.Done()
	b.logger.Info("amqp: broker shutting down")
	return nil
}

func (b *Broker) initAll(ctx context.Context) error {
	for _, p := range b.publishers {
		if err := p.Init(); err != nil {
			return fmt.Errorf("amqp broker: init publisher: %w", err)
		}
	}
	for _, c := range b.consumers {
		if err := c.Start(ctx, b.conn.Channel); err != nil {
			return fmt.Errorf("amqp broker: start consumer %q: %w", c.Queue(), err)
		}
	}
	return nil
}

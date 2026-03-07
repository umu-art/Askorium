package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reconnectDelay    = 3 * time.Second
	reconnectMaxDelay = 30 * time.Second
)

type Connection struct {
	url  string
	conn *amqp.Connection
	mu   sync.RWMutex

	onReconnect []func()
	logger      *slog.Logger
}

func NewConnection(url string, logger *slog.Logger) *Connection {
	return &Connection{
		url:    url,
		logger: logger,
	}
}

func (c *Connection) Connect(ctx context.Context) error {
	conn, err := c.dialWithRetry(ctx)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	go c.watchAndReconnect(ctx)
	return nil
}

func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil || c.conn.IsClosed() {
		return nil, fmt.Errorf("amqp: connection is not open")
	}
	return c.conn.Channel()
}

func (c *Connection) OnReconnect(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReconnect = append(c.onReconnect, fn)
}

func (c *Connection) watchAndReconnect(ctx context.Context) {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		closeErr := <-conn.NotifyClose(make(chan *amqp.Error, 1))
		if closeErr == nil {
			return
		}

		c.logger.Warn("amqp: connection closed, reconnecting",
			"reason", closeErr.Reason,
			"code", closeErr.Code,
		)

		newConn, err := c.dialWithRetry(ctx)
		if err != nil {
			return
		}

		c.mu.Lock()
		c.conn = newConn
		callbacks := make([]func(), len(c.onReconnect))
		copy(callbacks, c.onReconnect)
		c.mu.Unlock()

		c.logger.Info("amqp: reconnected successfully")

		for _, fn := range callbacks {
			fn()
		}
	}
}

func (c *Connection) dialWithRetry(ctx context.Context) (*amqp.Connection, error) {
	delay := reconnectDelay
	for {
		c.logger.Info("amqp: connecting", "url", c.url)
		conn, err := amqp.Dial(c.url)
		if err == nil {
			return conn, nil
		}

		c.logger.Warn("amqp: connection failed, retrying",
			"error", err,
			"delay", delay,
		)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("amqp: context cancelled while connecting: %w", ctx.Err())
		case <-time.After(delay):
		}

		delay = min(delay*2, reconnectMaxDelay)
	}
}

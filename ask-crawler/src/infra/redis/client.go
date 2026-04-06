package redisinfra

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb    *redis.Client
	logger *slog.Logger
}

// NewClient принимает Redis URL формата redis://host:port или rediss://...
func NewClient(redisURL string, logger *slog.Logger) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse URL %q: %w", redisURL, err)
	}
	return &Client{rdb: redis.NewClient(opts), logger: logger}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

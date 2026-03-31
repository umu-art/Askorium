package redisinfra

import (
	"context"
	"time"
)

type RedisVisitedStore struct {
	client *Client
	key    string
	ttl    time.Duration
}

func (s *RedisVisitedStore) Contains(url string) bool {
	exists, err := s.client.rdb.SIsMember(context.Background(), s.key, url).Result()
	if err != nil {
		s.client.logger.Error("redis: SIsMember visited failed", "key", s.key, "error", err)
		return false
	}
	return exists
}

func (s *RedisVisitedStore) Add(url string) {
	ctx := context.Background()
	if err := s.client.rdb.SAdd(ctx, s.key, url).Err(); err != nil {
		s.client.logger.Error("redis: SAdd visited failed", "key", s.key, "error", err)
		return
	}
	if err := s.client.rdb.Expire(ctx, s.key, s.ttl).Err(); err != nil {
		s.client.logger.Error("redis: Expire visited failed", "key", s.key, "error", err)
	}
}

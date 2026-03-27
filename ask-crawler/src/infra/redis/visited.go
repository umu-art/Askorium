package redisinfra

import "context"

type RedisVisitedStore struct {
	client *Client
	key    string
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
	if err := s.client.rdb.SAdd(context.Background(), s.key, url).Err(); err != nil {
		s.client.logger.Error("redis: SAdd visited failed", "key", s.key, "error", err)
	}
}

package redisinfra

import (
	"context"
	"time"

	applib "ask-crawler/src/app/lib"
)

// RedisFrontierFactory реализует FrontierFactory поверх Redis.
// Ключи задачи: crawler:{taskID}:frontier (ZSET или LIST), crawler:{taskID}:visited (SET).
// ttl — sliding TTL: сбрасывается при каждой записи; страховка от осиротевших ключей при краше краулера.
type RedisFrontierFactory struct {
	client *Client
	ttl    time.Duration
}

func NewFrontierFactory(client *Client, ttl time.Duration) *RedisFrontierFactory {
	return &RedisFrontierFactory{client: client, ttl: ttl}
}

func (f *RedisFrontierFactory) NewFrontier(taskID string, priority bool) applib.Frontier {
	key := "crawler:" + taskID + ":frontier"
	if priority {
		return &RedisPriorityFrontier{client: f.client, key: key, ttl: f.ttl}
	}
	return &RedisFIFOFrontier{client: f.client, key: key, ttl: f.ttl}
}

func (f *RedisFrontierFactory) NewVisitedStore(taskID string) applib.VisitedStore {
	return &RedisVisitedStore{
		client: f.client,
		key:    "crawler:" + taskID + ":visited",
		ttl:    f.ttl,
	}
}

func (f *RedisFrontierFactory) Cleanup(ctx context.Context, taskID string) {
	if err := f.client.rdb.Del(ctx,
		"crawler:"+taskID+":frontier",
		"crawler:"+taskID+":visited",
	).Err(); err != nil {
		f.client.logger.Error("redis: cleanup failed", "task_id", taskID, "error", err)
	}
}

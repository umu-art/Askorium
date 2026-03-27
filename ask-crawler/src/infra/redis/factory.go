package redisinfra

import (
	"context"

	applib "ask-crawler/src/app/lib"
)

// RedisFrontierFactory реализует FrontierFactory поверх Redis.
// Ключи задачи: crawler:{taskID}:frontier (ZSET или LIST), crawler:{taskID}:visited (SET).
type RedisFrontierFactory struct {
	client *Client
}

func NewFrontierFactory(client *Client) *RedisFrontierFactory {
	return &RedisFrontierFactory{client: client}
}

func (f *RedisFrontierFactory) NewFrontier(taskID string, priority bool) applib.Frontier {
	key := "crawler:" + taskID + ":frontier"
	if priority {
		return &RedisPriorityFrontier{client: f.client, key: key}
	}
	return &RedisFIFOFrontier{client: f.client, key: key}
}

func (f *RedisFrontierFactory) NewVisitedStore(taskID string) applib.VisitedStore {
	return &RedisVisitedStore{
		client: f.client,
		key:    "crawler:" + taskID + ":visited",
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

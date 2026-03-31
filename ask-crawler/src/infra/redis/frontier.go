package redisinfra

import (
	"context"
	"encoding/json"
	"time"

	applib "ask-crawler/src/app/lib"

	"github.com/redis/go-redis/v9"
)

// RedisPriorityFrontier — priority queue на Redis ZSET (max-heap по Score).
type RedisPriorityFrontier struct {
	client *Client
	key    string
	ttl    time.Duration
}

func (f *RedisPriorityFrontier) Push(c applib.URLCandidate) {
	data, _ := json.Marshal(c)
	ctx := context.Background()
	if err := f.client.rdb.ZAdd(ctx, f.key, redis.Z{
		Score:  c.Score,
		Member: string(data),
	}).Err(); err != nil {
		f.client.logger.Error("redis: ZAdd frontier failed", "key", f.key, "error", err)
		return
	}
	if err := f.client.rdb.Expire(ctx, f.key, f.ttl).Err(); err != nil {
		f.client.logger.Error("redis: Expire frontier failed", "key", f.key, "error", err)
	}
}

func (f *RedisPriorityFrontier) Pop() (applib.URLCandidate, bool) {
	results, err := f.client.rdb.ZPopMax(context.Background(), f.key, 1).Result()
	if err != nil {
		f.client.logger.Error("redis: ZPopMax frontier failed", "key", f.key, "error", err)
		return applib.URLCandidate{}, false
	}
	if len(results) == 0 {
		return applib.URLCandidate{}, false
	}
	var c applib.URLCandidate
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &c); err != nil {
		f.client.logger.Error("redis: unmarshal URLCandidate failed", "error", err)
		return applib.URLCandidate{}, false
	}
	return c, true
}

func (f *RedisPriorityFrontier) Len() int {
	n, err := f.client.rdb.ZCard(context.Background(), f.key).Result()
	if err != nil {
		f.client.logger.Error("redis: ZCard frontier failed", "key", f.key, "error", err)
		return 0
	}
	return int(n)
}

// RedisFIFOFrontier — FIFO очередь на Redis LIST.
type RedisFIFOFrontier struct {
	client *Client
	key    string
	ttl    time.Duration
}

func (f *RedisFIFOFrontier) Push(c applib.URLCandidate) {
	data, _ := json.Marshal(c)
	ctx := context.Background()
	if err := f.client.rdb.RPush(ctx, f.key, string(data)).Err(); err != nil {
		f.client.logger.Error("redis: RPush frontier failed", "key", f.key, "error", err)
		return
	}
	if err := f.client.rdb.Expire(ctx, f.key, f.ttl).Err(); err != nil {
		f.client.logger.Error("redis: Expire frontier failed", "key", f.key, "error", err)
	}
}

func (f *RedisFIFOFrontier) Pop() (applib.URLCandidate, bool) {
	val, err := f.client.rdb.LPop(context.Background(), f.key).Result()
	if err == redis.Nil {
		return applib.URLCandidate{}, false
	}
	if err != nil {
		f.client.logger.Error("redis: LPop frontier failed", "key", f.key, "error", err)
		return applib.URLCandidate{}, false
	}
	var c applib.URLCandidate
	if err := json.Unmarshal([]byte(val), &c); err != nil {
		f.client.logger.Error("redis: unmarshal URLCandidate failed", "error", err)
		return applib.URLCandidate{}, false
	}
	return c, true
}

func (f *RedisFIFOFrontier) Len() int {
	n, err := f.client.rdb.LLen(context.Background(), f.key).Result()
	if err != nil {
		f.client.logger.Error("redis: LLen frontier failed", "key", f.key, "error", err)
		return 0
	}
	return int(n)
}

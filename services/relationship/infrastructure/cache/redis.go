package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter 是好友申请限流（FR-014，账号 + IP 双维度，防骚扰刷申请）。
type RateLimiter struct {
	client *redis.Client
	window time.Duration
	limit  int
}

func NewRateLimiter(client *redis.Client, window time.Duration, limit int) *RateLimiter {
	if window <= 0 {
		window = time.Minute
	}
	if limit <= 0 {
		limit = 10
	}
	return &RateLimiter{client: client, window: window, limit: limit}
}

// Allow 判断该维度（账号或 IP）是否未超限。超限返回 false。
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := "relationship:ratelimit:request:" + key
	n, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		if err := r.client.Expire(ctx, fullKey, r.window).Err(); err != nil {
			return false, err
		}
	}
	return n <= int64(r.limit), nil
}

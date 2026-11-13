package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"isletalk/identity/domain/valueobject"
	"isletalk/identity/infrastructure/config"
)

// RedisCodeStore 实现 VerificationCodeStore（T010/R-5）。
// 验证码一次性由"校验后删除"保证；TTL 默认 5 分钟。
type RedisCodeStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCodeStore(client *redis.Client, ttl time.Duration) *RedisCodeStore {
	if ttl <= 0 {
		ttl = config.Duration("IDENTITY_VERIFY_CODE_TTL", 5*time.Minute)
	}
	return &RedisCodeStore{client: client, ttl: ttl}
}

func (s *RedisCodeStore) key(phone valueobject.PhoneHash) string {
	return "identity:verify:" + phone.String()
}

func (s *RedisCodeStore) Save(ctx context.Context, phone valueobject.PhoneHash, code string) error {
	return s.client.Set(ctx, s.key(phone), code, s.ttl).Err()
}

func (s *RedisCodeStore) Verify(ctx context.Context, phone valueobject.PhoneHash, code string) (bool, error) {
	key := s.key(phone)
	got, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // 不存在或已过期
	}
	if err != nil {
		return false, err
	}
	if got != code {
		return false, nil
	}
	// 一次性：校验后删除
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return false, err
	}
	return true, nil
}

// RateLimiter 是登录/注册风控限流（T045，账号 + IP 双维度，宪法 E 组）。
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
	fullKey := "identity:ratelimit:" + key
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

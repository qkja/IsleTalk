package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"isletalk/identity/infrastructure/cache"
)

func redisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 不可用，跳过集成测试: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestRedisCodeStoreOneTime(t *testing.T) {
	client := redisClient(t)
	ctx := context.Background()
	store := cache.NewRedisCodeStore(client, time.Minute)

	phone, _ := newPhone("phone:138")

	if err := store.Save(ctx, phone, "123456"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ok, err := store.Verify(ctx, phone, "123456")
	if err != nil || !ok {
		t.Fatalf("first Verify = %v, %v", ok, err)
	}

	// 一次性：第二次必须失败
	ok, _ = store.Verify(ctx, phone, "123456")
	if ok {
		t.Fatal("verification code must be one-time")
	}
}

func TestRedisCodeStoreWrongCode(t *testing.T) {
	client := redisClient(t)
	ctx := context.Background()
	store := cache.NewRedisCodeStore(client, time.Minute)

	phone, _ := newPhone("phone:138")
	_ = store.Save(ctx, phone, "123456")

	ok, _ := store.Verify(ctx, phone, "wrong")
	if ok {
		t.Fatal("wrong code should not verify")
	}
}

func TestRedisRateLimiter(t *testing.T) {
	client := redisClient(t)
	ctx := context.Background()
	limiter := cache.NewRateLimiter(client, time.Minute, 3)

	// 唯一 key，避免跨测试运行残留计数
	key := "login:test-" + time.Now().Format("20060102150405.000000000")
	for i := 0; i < 3; i++ {
		if ok, err := limiter.Allow(ctx, key); err != nil || !ok {
			t.Fatalf("Allow #%d = %v, %v", i+1, ok, err)
		}
	}
	// 第 4 次超限
	if ok, _ := limiter.Allow(ctx, key); ok {
		t.Fatal("4th attempt should be rate-limited")
	}
}

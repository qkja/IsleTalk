package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"isletalk/message/domain/entity"
)

// 覆盖真实 Redis 分支（本地/CI 有 Redis 时运行，否则 skip）。

func newRedis(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 不可用: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestUnreadCacheWithRedis(t *testing.T) {
	c := NewUnreadCache(newRedis(t))
	ctx := context.Background()
	c.Set(ctx, "c1", 5)
	n, ok := c.Get(ctx, "c1")
	if !ok || n != 5 {
		t.Fatalf("Get = %d, %v; want 5, true", n, ok)
	}
	c.Del(ctx, "c1")
	if _, ok := c.Get(ctx, "c1"); ok {
		t.Fatal("after Del, Get should be not-ok")
	}
}

func TestConversationCacheWithRedis(t *testing.T) {
	c := NewConversationCache(newRedis(t))
	ctx := context.Background()
	c.Set(ctx, "u1", []entity.Conversation{sampleConversation()})
	got, ok := c.Get(ctx, "u1")
	if !ok || len(got) != 1 || got[0].ConversationID() != "direct:a:b" {
		t.Fatalf("Get = %v, %v; want one conversation", got, ok)
	}
	c.Del(ctx, "u1")
	if _, ok := c.Get(ctx, "u1"); ok {
		t.Fatal("after Del, Get should be not-ok")
	}
}

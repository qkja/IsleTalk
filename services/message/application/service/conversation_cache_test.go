package service

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"

	"isletalk/message/domain/entity"
	"isletalk/message/infrastructure/cache"
)

// 覆盖会话列表 cache-aside 读路径（命中读缓存返回，未命中回源并回填）。
func TestConversationListCacheAside(t *testing.T) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis 不可用: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	convs := newMemConversationRepo()
	seedConversation(t, convs, "direct:a:b")
	svc := NewConversationService(convs, nil, cache.NewConversationCache(client))

	// 第一次：miss → 回源 MongoDB → 回填缓存
	if _, err := svc.List(ctx, uid(t, "a"), 0); err != nil {
		t.Fatal(err)
	}
	// 清空源，第二次应命中缓存仍返回 1 条
	convs.byID = map[string]entity.Conversation{}
	list, err := svc.List(ctx, uid(t, "a"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("cache hit should return 1 conversation, got %d", len(list))
	}
}

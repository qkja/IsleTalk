package service

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/cache"
)

// 覆盖 cache-aside 写后失效（T031）：发送/撤回后双方缓存均被失效。

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis 不可用: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestSendInvalidatesBothCaches(t *testing.T) {
	ctx := context.Background()
	client := newTestRedis(t)
	convCache := cache.NewConversationCache(client)
	unread := cache.NewUnreadCache(client)

	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	svc := NewSendMessageService(msgs, convs, &mockRelationship{allowed: true}, unread, convCache)

	a, b := uid(t, "a"), uid(t, "b")
	convID := entity.DirectConversationID(a, b)
	c := entity.NewConversation(convID, []valueobject.UserId{a, b})
	convCache.Set(ctx, a.String(), []entity.Conversation{c})
	convCache.Set(ctx, b.String(), []entity.Conversation{c})

	if _, err := svc.Send(ctx, mid(t, "m1"), convID, a, entity.MessageText, content(t, "hi")); err != nil {
		t.Fatal(err)
	}
	if _, ok := convCache.Get(ctx, a.String()); ok {
		t.Fatal("sender cache should be invalidated")
	}
	if _, ok := convCache.Get(ctx, b.String()); ok {
		t.Fatal("recipient cache should be invalidated")
	}
}

func TestRecallInvalidatesBothCaches(t *testing.T) {
	ctx := context.Background()
	convCache := cache.NewConversationCache(newTestRedis(t))

	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	svc := NewRecallService(msgs, convs, convCache)

	a, b := uid(t, "a"), uid(t, "b")
	convID := entity.DirectConversationID(a, b)
	convs.byID[convID] = entity.NewConversation(convID, []valueobject.UserId{a, b})
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), convID, a, entity.MessageText, content(t, "secret"), 1)

	c := entity.NewConversation(convID, []valueobject.UserId{a, b})
	convCache.Set(ctx, a.String(), []entity.Conversation{c})
	convCache.Set(ctx, b.String(), []entity.Conversation{c})

	if _, err := svc.Recall(ctx, mid(t, "m1"), a); err != nil {
		t.Fatal(err)
	}
	if _, ok := convCache.Get(ctx, a.String()); ok {
		t.Fatal("participant a cache should be invalidated")
	}
	if _, ok := convCache.Get(ctx, b.String()); ok {
		t.Fatal("participant b cache should be invalidated")
	}
}

package cache

import (
	"context"
	"testing"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

func sampleConversation() entity.Conversation {
	a, _ := valueobject.NewUserId("a")
	b, _ := valueobject.NewUserId("b")
	return entity.NewConversation("direct:a:b", []valueobject.UserId{a, b})
}

// 覆盖 nil-safe 降级分支（Redis 非唯一事实来源，故障只降级不 panic）。
func TestUnreadCacheNilSafe(t *testing.T) {
	ctx := context.Background()
	u := NewUnreadCache(nil)
	if _, ok := u.Get(ctx, "c1"); ok {
		t.Fatal("nil client Get should return not-ok")
	}
	u.Set(ctx, "c1", 3) // 不 panic
	u.Del(ctx, "c1")    // 不 panic

	var nilCache *UnreadCache
	if _, ok := nilCache.Get(ctx, "c1"); ok {
		t.Fatal("nil receiver Get should return not-ok")
	}
	nilCache.Set(ctx, "c1", 3)
	nilCache.Del(ctx, "c1")
}

func TestConversationCacheNilSafe(t *testing.T) {
	ctx := context.Background()
	conv := sampleConversation()
	c := NewConversationCache(nil)
	if _, ok := c.Get(ctx, "u1"); ok {
		t.Fatal("nil client Get should return not-ok")
	}
	c.Set(ctx, "u1", []entity.Conversation{conv})
	c.Del(ctx, "u1")

	var nilCache *ConversationCache
	if _, ok := nilCache.Get(ctx, "u1"); ok {
		t.Fatal("nil receiver Get should return not-ok")
	}
	nilCache.Set(ctx, "u1", []entity.Conversation{conv})
	nilCache.Del(ctx, "u1")
}

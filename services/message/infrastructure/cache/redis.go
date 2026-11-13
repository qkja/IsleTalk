package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

// 未读计数与会话列表的 cache-aside 缓存（T007/T030）。
// 铁律（宪法 A 组）：Redis 非唯一事实来源——消息本体存 MongoDB，Redis 故障只降级不丢消息。

// UnreadCache 是未读计数缓存（plan：message:unread:{conversation_id}）。
type UnreadCache struct {
	client *redis.Client
}

func NewUnreadCache(client *redis.Client) *UnreadCache {
	return &UnreadCache{client: client}
}

func unreadKey(conversationID string) string { return "message:unread:" + conversationID }

// Get 读未读计数缓存；未命中或故障返回 (0, false)。
func (c *UnreadCache) Get(ctx context.Context, conversationID string) (int, bool) {
	if c == nil || c.client == nil {
		return 0, false
	}
	n, err := c.client.Get(ctx, unreadKey(conversationID)).Int()
	if err != nil {
		return 0, false
	}
	return n, true
}

// Set 写未读计数缓存。
func (c *UnreadCache) Set(ctx context.Context, conversationID string, n int) {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Set(ctx, unreadKey(conversationID), n, 0).Err()
}

// Del 失效未读计数缓存（写后失效）。
func (c *UnreadCache) Del(ctx context.Context, conversationID string) {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Del(ctx, unreadKey(conversationID)).Err()
}

// ConversationCache 是会话列表缓存（plan：message:conversations:{user_id}，TTL 60s，cache-aside）。
type ConversationCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewConversationCache(client *redis.Client) *ConversationCache {
	return &ConversationCache{client: client, ttl: 60 * time.Second}
}

func conversationKey(userID string) string { return "message:conversations:" + userID }

// conversationSnapshot 是会话的可序列化快照（领域模型与缓存解耦，宪法原则 V）。
type conversationSnapshot struct {
	ConversationID     string   `json:"conversation_id"`
	Type               string   `json:"type"`
	Participants       []string `json:"participants"`
	LastMessagePreview string   `json:"last_message_preview"`
	UnreadCount        int      `json:"unread_count"`
	Pinned             bool     `json:"pinned"`
	Muted              bool     `json:"muted"`
	Deleted            bool     `json:"deleted"`
	LastActiveAt       int64    `json:"last_active_at"`
}

func toSnapshot(c entity.Conversation) conversationSnapshot {
	parts := make([]string, 0, len(c.Participants()))
	for _, p := range c.Participants() {
		parts = append(parts, p.String())
	}
	return conversationSnapshot{
		ConversationID:     c.ConversationID(),
		Type:               string(c.Type()),
		Participants:       parts,
		LastMessagePreview: c.LastMessagePreview(),
		UnreadCount:        c.UnreadCount(),
		Pinned:             c.Pinned(),
		Muted:              c.Muted(),
		Deleted:            c.Deleted(),
		LastActiveAt:       c.LastActiveAt().UnixMilli(),
	}
}

func (s conversationSnapshot) toEntity() entity.Conversation {
	parts := make([]valueobject.UserId, 0, len(s.Participants))
	for _, p := range s.Participants {
		if u, err := valueobject.NewUserId(p); err == nil {
			parts = append(parts, u)
		}
	}
	return entity.ReconstructConversation(
		s.ConversationID, entity.ConversationType(s.Type), parts, s.LastMessagePreview,
		s.UnreadCount, s.Pinned, s.Muted, s.Deleted, time.UnixMilli(s.LastActiveAt),
	)
}

// Get 读会话列表缓存（cache-aside 读）；未命中或故障返回 (nil, false)。
func (c *ConversationCache) Get(ctx context.Context, userID string) ([]entity.Conversation, bool) {
	if c == nil || c.client == nil {
		return nil, false
	}
	v, err := c.client.Get(ctx, conversationKey(userID)).Result()
	if err != nil {
		return nil, false
	}
	var snaps []conversationSnapshot
	if err := json.Unmarshal([]byte(v), &snaps); err != nil {
		return nil, false
	}
	out := make([]entity.Conversation, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, s.toEntity())
	}
	return out, true
}

// Set 写会话列表缓存（cache-aside 回填）。
func (c *ConversationCache) Set(ctx context.Context, userID string, convs []entity.Conversation) {
	if c == nil || c.client == nil {
		return
	}
	snaps := make([]conversationSnapshot, 0, len(convs))
	for _, conv := range convs {
		snaps = append(snaps, toSnapshot(conv))
	}
	b, err := json.Marshal(snaps)
	if err != nil {
		return
	}
	_ = c.client.Set(ctx, conversationKey(userID), b, c.ttl).Err()
}

// Del 失效会话列表缓存（写后失效）。
func (c *ConversationCache) Del(ctx context.Context, userID string) {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Del(ctx, conversationKey(userID)).Err()
}

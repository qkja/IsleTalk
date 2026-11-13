package entity

import (
	"strings"
	"time"

	"isletalk/message/domain/valueobject"
)

// ConversationType 是会话类型（首段仅 direct；group 由后续 feature 接入，FR-013）。
type ConversationType string

const (
	ConversationDirect ConversationType = "direct"
	ConversationGroup  ConversationType = "group"
)

// directPrefix 是单聊会话 ID 的前缀（conversation_id 由好友关系派生，FR-001）。
const directPrefix = "direct:"

// DirectConversationID 由两个用户 ID 派生确定性单聊会话 ID（排序后拼接）。
// 首段契约 SendMessageRequest 不携带接收方，接收方由 conversation_id 反解得到（FR-001「由好友关系派生」）。
func DirectConversationID(a, b valueobject.UserId) string {
	as, bs := a.String(), b.String()
	if as < bs {
		return directPrefix + as + ":" + bs
	}
	return directPrefix + bs + ":" + as
}

// ParseDirectConversationID 解析单聊会话 ID 得到两个参与者（排序保证确定性）。
// 非 direct 会话或格式非法返回 ok=false。
func ParseDirectConversationID(id string) (a, b valueobject.UserId, ok bool) {
	if !strings.HasPrefix(id, directPrefix) {
		return valueobject.UserId{}, valueobject.UserId{}, false
	}
	rest := strings.TrimPrefix(id, directPrefix)
	i := strings.IndexByte(rest, ':')
	if i <= 0 || i == len(rest)-1 {
		return valueobject.UserId{}, valueobject.UserId{}, false
	}
	first, second := rest[:i], rest[i+1:]
	af, err := valueobject.NewUserId(first)
	if err != nil {
		return valueobject.UserId{}, valueobject.UserId{}, false
	}
	bf, err := valueobject.NewUserId(second)
	if err != nil {
		return valueobject.UserId{}, valueobject.UserId{}, false
	}
	return af, bf, true
}

// Conversation 是会话聚合根（data-model：message.conversations）。
type Conversation struct {
	conversationID     string
	conversationType   ConversationType
	participants       []valueobject.UserId
	lastMessagePreview string
	unreadCount        int
	pinned             bool
	muted              bool
	deleted            bool
	lastActiveAt       time.Time
}

func NewConversation(conversationID string, participants []valueobject.UserId) Conversation {
	return Conversation{
		conversationID:   conversationID,
		conversationType: ConversationDirect,
		participants:     participants,
		lastActiveAt:     time.Now().UTC(),
	}
}

// ReconstructConversation 从持久化状态重建会话（仅基础设施仓储使用，勿用于新会话创建）。
func ReconstructConversation(
	conversationID string,
	conversationType ConversationType,
	participants []valueobject.UserId,
	lastMessagePreview string,
	unreadCount int,
	pinned, muted, deleted bool,
	lastActiveAt time.Time,
) Conversation {
	return Conversation{
		conversationID:     conversationID,
		conversationType:   conversationType,
		participants:       participants,
		lastMessagePreview: lastMessagePreview,
		unreadCount:        unreadCount,
		pinned:             pinned,
		muted:              muted,
		deleted:            deleted,
		lastActiveAt:       lastActiveAt,
	}
}

func (c Conversation) ConversationID() string             { return c.conversationID }
func (c Conversation) Type() ConversationType             { return c.conversationType }
func (c Conversation) Participants() []valueobject.UserId { return c.participants }
func (c Conversation) LastMessagePreview() string         { return c.lastMessagePreview }
func (c Conversation) UnreadCount() int                   { return c.unreadCount }
func (c Conversation) Pinned() bool                       { return c.pinned }
func (c Conversation) Muted() bool                        { return c.muted }
func (c Conversation) Deleted() bool                      { return c.deleted }
func (c Conversation) LastActiveAt() time.Time            { return c.lastActiveAt }

// HasParticipant 判断用户是否参与该会话（FR-011 授权边界）。
func (c Conversation) HasParticipant(uid valueobject.UserId) bool {
	for _, p := range c.participants {
		if p.Equal(uid) {
			return true
		}
	}
	return false
}

// SetPreview 更新最新消息摘要与最近活跃时间（FR-008）。
func (c *Conversation) SetPreview(preview string) {
	c.lastMessagePreview = preview
	c.lastActiveAt = time.Now().UTC()
}

// IncrementUnread 未读 +1（FR-008；免打扰仍计入未读 FR-023）。
func (c *Conversation) IncrementUnread() { c.unreadCount++ }

// ResetUnread 未读清零（FR-019 已读上报）。
func (c *Conversation) ResetUnread() { c.unreadCount = 0 }

// Pin / Unpin 置顶与取消置顶（FR-024）。
func (c *Conversation) Pin()   { c.pinned = true }
func (c *Conversation) Unpin() { c.pinned = false }

// Mute / Unmute 免打扰与取消免打扰（FR-024）。
func (c *Conversation) Mute()   { c.muted = true }
func (c *Conversation) Unmute() { c.muted = false }

// MarkDeleted 标记删除（FR-024；对方发新消息时重新出现，见应用层 TouchOnSend）。
func (c *Conversation) MarkDeleted() { c.deleted = true }

// Restore 删除后收到新消息重新出现（FR-024）。
func (c *Conversation) Restore() { c.deleted = false }

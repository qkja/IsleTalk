package port

import (
	"context"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

// ConversationRepository 是会话的持久化接口（T005）。
type ConversationRepository interface {
	// Save upsert 会话（conversation_id 唯一，FR-001）。
	Save(ctx context.Context, c entity.Conversation) (entity.Conversation, error)
	// FindByID 按会话 ID 查询。
	FindByID(ctx context.Context, conversationID string) (entity.Conversation, error)
	// ListByUser 列出某用户的会话（置顶优先，按最近活跃倒序，FR-008）。
	ListByUser(ctx context.Context, userID valueobject.UserId, limit int) ([]entity.Conversation, error)
	// NextSeq 原子分配会话内下一个 seq（单调递增，FR-031）。
	NextSeq(ctx context.Context, conversationID string) (uint64, error)
	// TouchOnSend 消息发送后原子更新摘要/未读/活跃时间，并恢复已删除会话（FR-008/024）。
	TouchOnSend(ctx context.Context, conversationID, preview string) error
	// ResetUnread 未读清零（FR-019 已读上报）。
	ResetUnread(ctx context.Context, conversationID string) error
	// TouchOnRecall 撤回后更新会话「最新消息摘要」为占位（FR-006，不增未读）。
	TouchOnRecall(ctx context.Context, conversationID, preview string) error
}

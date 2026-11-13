package port

import (
	"context"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

// MessageRepository 是消息的持久化接口（T005）。
type MessageRepository interface {
	// Save 插入消息（message_id 唯一索引兜底幂等，FR-007）。
	Save(ctx context.Context, m entity.Message) (entity.Message, error)
	// FindByID 按幂等键查询（幂等去重 FR-007）。
	FindByID(ctx context.Context, id valueobject.MessageId) (entity.Message, error)
	// ListByConversation 按会话分页拉取（seq 倒序，beforeSeq 游标，FR-012）。
	ListByConversation(ctx context.Context, conversationID string, beforeSeq uint64, limit int) ([]entity.Message, error)
	// MarkRead 将 seq ≤ upToSeq 且非 readerID 发送的消息标记已读，返回更新条数（FR-019）。
	MarkRead(ctx context.Context, conversationID string, readerID valueobject.UserId, upToSeq uint64) (int64, error)
	// Recall 标记消息已撤回（FR-020）。
	Recall(ctx context.Context, id valueobject.MessageId) error
}

// ErrNotFound 表示实体不存在。
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "entity not found" }

// IsNotFound 判断错误是否为「实体不存在」。
func IsNotFound(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

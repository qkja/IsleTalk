package service

import (
	"context"
	"log/slog"

	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/cache"
)

// MarkReadService 是已读上报用例（T020，FR-019）。
type MarkReadService struct {
	messages      port.MessageRepository
	conversations port.ConversationRepository
	unread        *cache.UnreadCache
	convCache     *cache.ConversationCache
}

func NewMarkReadService(
	messages port.MessageRepository,
	conversations port.ConversationRepository,
	unread *cache.UnreadCache,
	convCache *cache.ConversationCache,
) *MarkReadService {
	return &MarkReadService{messages: messages, conversations: conversations, unread: unread, convCache: convCache}
}

// MarkRead 上报已读（FR-019）：更新消息状态 delivered→read + 会话未读清零。
func (s *MarkReadService) MarkRead(ctx context.Context, conversationID string, userID valueobject.UserId, lastReadSeq uint64) error {
	slog.Info("mark read", "conversation_id", conversationID, "user", userID.String(), "last_read_seq", lastReadSeq)
	if _, err := s.messages.MarkRead(ctx, conversationID, userID, lastReadSeq); err != nil {
		slog.Error("mark read messages failed", "err", err)
		return err
	}
	if err := s.conversations.ResetUnread(ctx, conversationID); err != nil {
		slog.Error("mark read reset unread failed", "err", err)
		return err
	}
	// 写后失效缓存（Redis 非唯一事实来源）
	s.unread.Del(ctx, conversationID)
	s.convCache.Del(ctx, userID.String())
	return nil
}

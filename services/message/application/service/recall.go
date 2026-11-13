package service

import (
	"context"
	"log/slog"
	"time"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/cache"
)

// RecallWindow 是撤回时间窗（默认 2 分钟，可调；由服务端时钟判定，FR-006/020）。
const RecallWindow = 2 * time.Minute

// RecallPreview 是撤回后会话「最新消息摘要」的占位（FR-006 撤回后的摘要展示，MUST NOT 显示原文）。
const RecallPreview = "撤回了一条消息"

// RecallService 是撤回用例（T019/T029/T031，FR-006/020）。
type RecallService struct {
	messages      port.MessageRepository
	conversations port.ConversationRepository
	convCache     *cache.ConversationCache
	window        time.Duration
	now           func() time.Time
}

func NewRecallService(messages port.MessageRepository, conversations port.ConversationRepository, convCache *cache.ConversationCache) *RecallService {
	return &RecallService{messages: messages, conversations: conversations, convCache: convCache, window: RecallWindow, now: time.Now}
}

// Recall 撤回消息（FR-006/020）：
// 时间窗内（服务端时钟判定）→ 校验发送者身份 → recalled 标记；超窗或非本人拒绝。
// 若被撤回消息为会话最新消息，同步更新会话「最新消息摘要」为占位并失效双方列表缓存。
func (s *RecallService) Recall(ctx context.Context, messageID valueobject.MessageId, senderID valueobject.UserId) (bool, error) {
	slog.Info("recall message", "message_id", messageID.String(), "sender", senderID.String())
	m, err := s.messages.FindByID(ctx, messageID)
	if err != nil {
		if port.IsNotFound(err) {
			return false, ErrMessageNotFound
		}
		slog.Error("recall find failed", "err", err)
		return false, err
	}

	// 非本人撤回拒绝（FR-020）
	if !m.IsSentBy(senderID) {
		slog.Warn("recall not sender", "message_id", messageID.String())
		return false, ErrNotSender
	}

	// 时间窗（服务端时钟判定，MUST NOT 依赖客户端时间）
	if s.now().Sub(m.SentAt()) > s.window {
		slog.Warn("recall window expired", "message_id", messageID.String())
		return false, ErrRecallWindowExpired
	}

	if err := s.messages.Recall(ctx, messageID); err != nil {
		slog.Error("recall failed", "err", err)
		return false, err
	}

	// 若为会话最新消息，更新摘要占位并失效双方列表缓存（FR-006/CHK014/T031）
	if latest, err := s.messages.ListByConversation(ctx, m.ConversationID(), 0, 1); err == nil && len(latest) == 1 && latest[0].Seq() == m.Seq() {
		if err := s.conversations.TouchOnRecall(ctx, m.ConversationID(), RecallPreview); err != nil {
			slog.Error("recall update preview failed", "err", err)
		} else if a, b, ok := entity.ParseDirectConversationID(m.ConversationID()); ok {
			s.convCache.Del(ctx, a.String())
			s.convCache.Del(ctx, b.String())
		}
	}

	slog.Info("recall success", "message_id", messageID.String())
	return true, nil
}

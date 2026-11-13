package service

import (
	"context"
	"log/slog"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/cache"
)

// SendMessageService 是发消息用例（T015，FR-002/003/004/005/007/010/011/031）。
type SendMessageService struct {
	messages      port.MessageRepository
	conversations port.ConversationRepository
	relationship  port.RelationshipClient
	unread        *cache.UnreadCache
	convCache     *cache.ConversationCache
}

func NewSendMessageService(
	messages port.MessageRepository,
	conversations port.ConversationRepository,
	relationship port.RelationshipClient,
	unread *cache.UnreadCache,
	convCache *cache.ConversationCache,
) *SendMessageService {
	return &SendMessageService{
		messages:      messages,
		conversations: conversations,
		relationship:  relationship,
		unread:        unread,
		convCache:     convCache,
	}
}

// Send 发送一条消息：
// 授权（CanMessage）→ message_id 幂等 → 建会话（如需）→ seq 分配 → 建消息标记 delivered → 更新会话摘要/未读。
func (s *SendMessageService) Send(
	ctx context.Context,
	messageID valueobject.MessageId,
	conversationID string,
	senderID valueobject.UserId,
	messageType entity.MessageType,
	content valueobject.MessageContent,
) (entity.Message, error) {
	slog.Info("send message", "message_id", messageID.String(), "conversation_id", conversationID, "sender", senderID.String(), "type", string(messageType))

	// 消息类型校验（FR-016/017/018）
	if messageType != entity.MessageText && messageType != entity.MessageImage && messageType != entity.MessageFile {
		slog.Warn("send message invalid type", "type", string(messageType))
		return entity.Message{}, ErrInvalidMessageType
	}

	// 幂等去重（FR-007）：message_id 已存在则返回已有消息
	if existing, err := s.messages.FindByID(ctx, messageID); err == nil {
		slog.Info("send message idempotent hit", "message_id", messageID.String())
		return existing, nil
	} else if !port.IsNotFound(err) {
		slog.Error("send message find failed", "err", err)
		return entity.Message{}, err
	}

	// 会话 ID 反解参与者（首段契约 SendMessageRequest 不携带接收方，由 conversation_id 派生）
	participantA, participantB, ok := entity.ParseDirectConversationID(conversationID)
	if !ok {
		return entity.Message{}, ErrInvalidConversationID
	}
	var recipient valueobject.UserId
	switch {
	case participantA.Equal(senderID):
		recipient = participantB
	case participantB.Equal(senderID):
		recipient = participantA
	default:
		slog.Warn("send message sender not participant", "sender", senderID.String())
		return entity.Message{}, ErrNotParticipant
	}

	// 授权（FR-011）：互为好友且未拉黑
	allowed, blocked, err := s.relationship.CanMessage(ctx, senderID, recipient)
	if err != nil {
		slog.Error("send message authorize failed", "err", err)
		return entity.Message{}, err
	}
	if !allowed {
		if blocked {
			slog.Warn("send message blocked", "sender", senderID.String())
			return entity.Message{}, ErrBlocked
		}
		slog.Warn("send message not friends", "sender", senderID.String())
		return entity.Message{}, ErrNotFriends
	}

	// 建会话（FR-001：首次发送时隐式建立，无需对方同意）
	if _, err := s.conversations.FindByID(ctx, conversationID); err != nil {
		if port.IsNotFound(err) {
			c := entity.NewConversation(conversationID, []valueobject.UserId{participantA, participantB})
			if _, err := s.conversations.Save(ctx, c); err != nil {
				slog.Error("send message create conversation failed", "err", err)
				return entity.Message{}, err
			}
		} else {
			slog.Error("send message find conversation failed", "err", err)
			return entity.Message{}, err
		}
	}

	// seq 分配（FR-031 单调递增，原子）
	seq, err := s.conversations.NextSeq(ctx, conversationID)
	if err != nil {
		slog.Error("send message allocate seq failed", "err", err)
		return entity.Message{}, err
	}

	// 建消息并标记 delivered（FR-005：服务端成功接收即「已送达」）
	msg := entity.NewMessage(messageID, conversationID, senderID, messageType, content, seq)
	msg.MarkDelivered()
	created, err := s.messages.Save(ctx, msg)
	if err != nil {
		slog.Error("send message save failed", "err", err)
		return entity.Message{}, err
	}

	// 更新会话摘要/未读/活跃时间（FR-008；摘要失败不阻塞消息已落库）
	if err := s.conversations.TouchOnSend(ctx, conversationID, previewFor(messageType, content.String())); err != nil {
		slog.Error("send message touch conversation failed", "err", err)
	}
	// 失效发送方与接收方会话缓存及未读缓存（cache-aside 写后失效，保证未读/摘要秒级一致）
	s.convCache.Del(ctx, senderID.String())
	s.convCache.Del(ctx, recipient.String())
	s.unread.Del(ctx, conversationID)
	slog.Info("send message success", "message_id", messageID.String(), "seq", seq)
	return created, nil
}

// previewFor 生成会话列表「最新消息摘要」（FR-008；图片/文件显示占位，撤回显示占位见召回后语义）。
func previewFor(messageType entity.MessageType, content string) string {
	switch messageType {
	case entity.MessageImage:
		return "[图片]"
	case entity.MessageFile:
		return "[文件]"
	default:
		return content
	}
}

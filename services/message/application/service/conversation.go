package service

import (
	"context"
	"log/slog"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
	"isletalk/message/infrastructure/cache"
)

// ConversationService 是会话列表与管理用例（T023，FR-008/009/024）。
type ConversationService struct {
	conversations port.ConversationRepository
	unread        *cache.UnreadCache
	convCache     *cache.ConversationCache
}

func NewConversationService(conversations port.ConversationRepository, unread *cache.UnreadCache, convCache *cache.ConversationCache) *ConversationService {
	return &ConversationService{conversations: conversations, unread: unread, convCache: convCache}
}

// List 会话列表（置顶优先，最近活跃倒序，FR-008）。
// cache-aside 读：命中读缓存返回，未命中回源 MongoDB 并回填缓存。
func (s *ConversationService) List(ctx context.Context, userID valueobject.UserId, limit int) ([]entity.Conversation, error) {
	slog.Info("list conversations", "user", userID.String())
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if convs, ok := s.convCache.Get(ctx, userID.String()); ok {
		return convs, nil
	}
	convs, err := s.conversations.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	s.convCache.Set(ctx, userID.String(), convs)
	return convs, nil
}

// Pin 置顶 / 取消置顶（FR-024）。
func (s *ConversationService) Pin(ctx context.Context, conversationID string, userID valueobject.UserId, pinned bool) error {
	c, err := s.get(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if pinned {
		c.Pin()
	} else {
		c.Unpin()
	}
	if _, err := s.conversations.Save(ctx, c); err != nil {
		return err
	}
	s.convCache.Del(ctx, userID.String())
	return nil
}

// Mute 免打扰 / 取消免打扰（FR-024）。
func (s *ConversationService) Mute(ctx context.Context, conversationID string, userID valueobject.UserId, muted bool) error {
	c, err := s.get(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if muted {
		c.Mute()
	} else {
		c.Unmute()
	}
	if _, err := s.conversations.Save(ctx, c); err != nil {
		return err
	}
	s.convCache.Del(ctx, userID.String())
	return nil
}

// Delete 删除会话（FR-024；对方发新消息时重新出现，见 TouchOnSend）。
func (s *ConversationService) Delete(ctx context.Context, conversationID string, userID valueobject.UserId) error {
	c, err := s.get(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	c.MarkDeleted()
	if _, err := s.conversations.Save(ctx, c); err != nil {
		return err
	}
	s.convCache.Del(ctx, userID.String())
	return nil
}

// MarkUnread 标记未读（FR-024）。
func (s *ConversationService) MarkUnread(ctx context.Context, conversationID string, userID valueobject.UserId) error {
	c, err := s.get(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	c.IncrementUnread()
	if _, err := s.conversations.Save(ctx, c); err != nil {
		return err
	}
	s.unread.Del(ctx, conversationID)
	s.convCache.Del(ctx, userID.String())
	return nil
}

// get 加载会话并校验参与者身份（FR-011 授权边界）。
func (s *ConversationService) get(ctx context.Context, conversationID string, userID valueobject.UserId) (entity.Conversation, error) {
	c, err := s.conversations.FindByID(ctx, conversationID)
	if err != nil {
		if port.IsNotFound(err) {
			return entity.Conversation{}, ErrConversationNotFound
		}
		return entity.Conversation{}, err
	}
	if !c.HasParticipant(userID) {
		return entity.Conversation{}, ErrNotParticipant
	}
	return c, nil
}

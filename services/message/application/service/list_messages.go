package service

import (
	"context"
	"log/slog"
	"strconv"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
)

const (
	// DefaultPageSize 是默认分页大小（FR-012）。
	DefaultPageSize = 50
	// MaxPageSize 是分页上限（FR-012）。
	MaxPageSize = 1000
)

// ListMessagesService 是分页拉取消息用例（T016，FR-012）。
type ListMessagesService struct {
	messages port.MessageRepository
}

func NewListMessagesService(messages port.MessageRepository) *ListMessagesService {
	return &ListMessagesService{messages: messages}
}

// List 分页拉取会话消息（seq 倒序，page_token 游标，AIP-158，禁止 offset）。
func (s *ListMessagesService) List(ctx context.Context, conversationID string, pageSize int32, pageToken string) ([]entity.Message, string, error) {
	slog.Info("list messages", "conversation_id", conversationID, "page_size", pageSize)
	size := int(pageSize)
	if size <= 0 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	beforeSeq := uint64(0)
	if pageToken != "" {
		v, err := strconv.ParseUint(pageToken, 10, 64)
		if err != nil {
			return nil, "", ErrInvalidPageToken
		}
		beforeSeq = v
	}
	// 多取一条判断是否还有下一页
	msgs, err := s.messages.ListByConversation(ctx, conversationID, beforeSeq, size+1)
	if err != nil {
		slog.Error("list messages failed", "err", err)
		return nil, "", err
	}
	hasMore := len(msgs) > size
	if hasMore {
		msgs = msgs[:size]
	}
	next := ""
	if hasMore && len(msgs) > 0 {
		next = strconv.FormatUint(msgs[len(msgs)-1].Seq(), 10)
	}
	return msgs, next, nil
}

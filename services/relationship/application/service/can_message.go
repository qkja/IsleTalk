package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// CanMessageService 是消息授权校验用例（FR-012，供 message-svc 调用）。
type CanMessageService struct {
	friendships port.FriendshipRepository
	blocks      port.BlockRepository
}

func NewCanMessageService(friendships port.FriendshipRepository, blocks port.BlockRepository) *CanMessageService {
	return &CanMessageService{friendships: friendships, blocks: blocks}
}

// CanMessage 判定「两人是否可互发消息」：互为好友且未被任一方拉黑。
func (s *CanMessageService) CanMessage(ctx context.Context, senderID, recipientID valueobject.UserId) (bool, error) {
	// 互为好友（双向 active）
	f1, err := s.friendships.Find(ctx, senderID, recipientID)
	if err != nil || !f1.IsActive() {
		return false, nil
	}
	f2, err := s.friendships.Find(ctx, recipientID, senderID)
	if err != nil || !f2.IsActive() {
		return false, nil
	}
	// 任一方拉黑 → 拒绝（FR-009）
	if b, err := s.blocks.Find(ctx, senderID, recipientID); err == nil && b.IsActive() {
		return false, nil
	}
	if b, err := s.blocks.Find(ctx, recipientID, senderID); err == nil && b.IsActive() {
		return false, nil
	}
	slog.Info("can message", "sender", senderID.String(), "recipient", recipientID.String(), "allowed", true)
	return true, nil
}

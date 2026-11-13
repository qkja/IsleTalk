package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/valueobject"
)

// Delete 删除好友（FR-007，双向解除：双删两条 Friendship 文档）。
func (s *FriendListService) Delete(ctx context.Context, userID, friendID valueobject.UserId) error {
	slog.Info("delete friend", "user_id", userID.String(), "friend_id", friendID.String())
	if err := s.friendships.Delete(ctx, userID, friendID); err != nil {
		slog.Error("delete friend failed", "err", err, "user_id", userID.String())
		return err
	}
	if err := s.friendships.Delete(ctx, friendID, userID); err != nil {
		slog.Error("delete reverse friend failed", "err", err)
		return err
	}
	slog.Info("delete friend success", "user_id", userID.String(), "friend_id", friendID.String())
	return nil
}

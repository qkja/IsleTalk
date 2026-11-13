package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// FriendListService 是好友列表用例（FR-006，列表分页默认 50 最大 1000）。
type FriendListService struct {
	friendships port.FriendshipRepository
}

func NewFriendListService(friendships port.FriendshipRepository) *FriendListService {
	return &FriendListService{friendships: friendships}
}

// List 列出某用户的好友（分页）。
func (s *FriendListService) List(ctx context.Context, userID valueobject.UserId, page, limit int) ([]entity.Friendship, error) {
	slog.Info("list friends", "user_id", userID.String(), "page", page, "limit", limit)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	friends, err := s.friendships.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("list friends failed", "err", err, "user_id", userID.String())
		return nil, err
	}
	start := page * limit
	if start >= len(friends) {
		return []entity.Friendship{}, nil
	}
	end := start + limit
	if end > len(friends) {
		end = len(friends)
	}
	return friends[start:end], nil
}

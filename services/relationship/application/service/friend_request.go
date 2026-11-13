package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// FriendRequestService 是发好友申请用例（FR-002/003/010）。
type FriendRequestService struct {
	requests port.FriendRequestRepository
	identity port.IdentityClient
}

func NewFriendRequestService(requests port.FriendRequestRepository, identity port.IdentityClient) *FriendRequestService {
	return &FriendRequestService{requests: requests, identity: identity}
}

// Send 发起好友申请：校验不自加 → 陌生人开关（FR-010）→ 重复申请（FR-003）→ 建申请。
func (s *FriendRequestService) Send(ctx context.Context, initiatorID, targetID valueobject.UserId, message string) (entity.FriendRequest, error) {
	slog.Info("send friend request", "initiator", initiatorID.String(), "target", targetID.String())

	if initiatorID.Equal(targetID) {
		slog.Warn("send friend request to self", "user_id", initiatorID.String())
		return entity.FriendRequest{}, ErrSelfRequest
	}

	// 陌生人添加开关（FR-010）：非好友且目标关闭开关则拒绝
	target, err := s.identity.GetUser(ctx, targetID)
	if err != nil {
		slog.Error("send friend request get target failed", "err", err, "target", targetID.String())
		return entity.FriendRequest{}, err
	}
	if !target.AllowStrangerAdd {
		slog.Warn("send friend request stranger add disabled", "target", targetID.String())
		return entity.FriendRequest{}, ErrStrangerAddDisabled
	}

	// 重复申请（FR-003，唯一索引兜底）
	if _, err := s.requests.Find(ctx, initiatorID, targetID); err == nil {
		slog.Warn("send friend request already exists", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrRequestExists
	} else if !port.IsNotFound(err) {
		slog.Error("send friend request find failed", "err", err)
		return entity.FriendRequest{}, err
	}

	req := entity.NewFriendRequest(initiatorID, targetID, message)
	created, err := s.requests.Create(ctx, req)
	if err != nil {
		slog.Error("send friend request create failed", "err", err)
		return entity.FriendRequest{}, err
	}
	slog.Info("send friend request success", "initiator", initiatorID.String(), "target", targetID.String())
	return created, nil
}

package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// RespondService 是接受/拒绝好友申请用例（FR-002/004）。
type RespondService struct {
	requests    port.FriendRequestRepository
	friendships port.FriendshipRepository
}

func NewRespondService(requests port.FriendRequestRepository, friendships port.FriendshipRepository) *RespondService {
	return &RespondService{requests: requests, friendships: friendships}
}

// Respond 接受/拒绝申请。接受时本地事务双写两条 Friendship（双向一致，SC-002）。
func (s *RespondService) Respond(ctx context.Context, initiatorID, targetID valueobject.UserId, accept bool) (entity.FriendRequest, error) {
	slog.Info("respond friend request", "initiator", initiatorID.String(), "target", targetID.String(), "accept", accept)

	req, err := s.requests.Find(ctx, initiatorID, targetID)
	if err != nil {
		if port.IsNotFound(err) {
			return entity.FriendRequest{}, ErrRequestNotFound
		}
		slog.Error("respond find request failed", "err", err)
		return entity.FriendRequest{}, err
	}
	if !req.IsPending() {
		slog.Warn("respond request not pending", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrRequestNotFound
	}

	if accept {
		req.Accept()
		// 双向一致：双写两条 Friendship（A→B 与 B→A）
		if _, err := s.friendships.Create(ctx, entity.NewFriendship(initiatorID, targetID)); err != nil {
			slog.Error("respond create friendship failed", "err", err)
			return entity.FriendRequest{}, err
		}
		if _, err := s.friendships.Create(ctx, entity.NewFriendship(targetID, initiatorID)); err != nil {
			slog.Error("respond create reverse friendship failed", "err", err)
			return entity.FriendRequest{}, err
		}
	} else {
		req.Reject()
	}

	if err := s.requests.Update(ctx, req); err != nil {
		slog.Error("respond update request failed", "err", err)
		return entity.FriendRequest{}, err
	}
	slog.Info("respond friend request success", "initiator", initiatorID.String(), "target", targetID.String())
	return req, nil
}

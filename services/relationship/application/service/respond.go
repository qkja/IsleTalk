package service

import (
	"context"
	"log/slog"
	"time"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// requestExpiry 是待处理申请的过期时长（Assumptions：默认 30 天）。测试可覆盖。
var requestExpiry = 30 * 24 * time.Hour

// RespondService 是接受/拒绝好友申请用例（FR-002/004）。
type RespondService struct {
	requests    port.FriendRequestRepository
	friendships port.FriendshipRepository
}

func NewRespondService(requests port.FriendRequestRepository, friendships port.FriendshipRepository) *RespondService {
	return &RespondService{requests: requests, friendships: friendships}
}

// Respond 接受/拒绝申请。接受时本地事务双写两条 Friendship（双向一致，SC-002），
// 并使反向待处理申请自动失效（Edge Case）；过期申请不可接受（T044）。
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

	// 申请过期（T044）：超过 30 天的待处理申请不可接受
	if time.Since(req.CreatedAt()) > requestExpiry {
		req.Reject()
		_ = s.requests.Update(ctx, req)
		slog.Warn("respond request expired", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrRequestExpired
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
		// 反向申请自动失效（T042）：B→A 的待处理申请标记 rejected
		if reverse, err := s.requests.Find(ctx, targetID, initiatorID); err == nil && reverse.IsPending() {
			reverse.Reject()
			if err := s.requests.Update(ctx, reverse); err != nil {
				slog.Error("respond reject reverse request failed", "err", err)
				return entity.FriendRequest{}, err
			}
			slog.Info("respond rejected reverse request", "initiator", targetID.String(), "target", initiatorID.String())
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

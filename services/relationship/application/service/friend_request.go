package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
	"isletalk/relationship/infrastructure/cache"
)

// FriendRequestService 是发好友申请用例（FR-002/003/009/010/014）。
type FriendRequestService struct {
	requests    port.FriendRequestRepository
	friendships port.FriendshipRepository
	blocks      port.BlockRepository
	identity    port.IdentityClient
	limiter     *cache.RateLimiter
}

func NewFriendRequestService(
	requests port.FriendRequestRepository,
	friendships port.FriendshipRepository,
	blocks port.BlockRepository,
	identity port.IdentityClient,
	limiter *cache.RateLimiter,
) *FriendRequestService {
	return &FriendRequestService{
		requests:    requests,
		friendships: friendships,
		blocks:      blocks,
		identity:    identity,
		limiter:     limiter,
	}
}

// Send 发起好友申请：限流（FR-014）→ 不自加 → 拉黑拦截（FR-009）→ 陌生人开关（FR-010）
// → 已是好友（FR-003/004）→ 重复申请（FR-003）→ 建申请。
func (s *FriendRequestService) Send(ctx context.Context, initiatorID, targetID valueobject.UserId, message string) (entity.FriendRequest, error) {
	slog.Info("send friend request", "initiator", initiatorID.String(), "target", targetID.String())

	// 限流（FR-014，账号维度；IP 维度由 api 层中间件承载）
	if s.limiter != nil {
		if ok, err := s.limiter.Allow(ctx, initiatorID.String()); err == nil && !ok {
			slog.Warn("send friend request rate limited", "initiator", initiatorID.String())
			return entity.FriendRequest{}, ErrRateLimited
		}
	}

	if initiatorID.Equal(targetID) {
		slog.Warn("send friend request to self", "user_id", initiatorID.String())
		return entity.FriendRequest{}, ErrSelfRequest
	}

	// 拉黑拦截（FR-009）：目标拉黑了我 或 我拉黑了目标，均拒绝发申请
	if b, err := s.blocks.Find(ctx, targetID, initiatorID); err == nil && b.IsActive() {
		slog.Warn("send friend request blocked by target", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrBlocked
	}
	if b, err := s.blocks.Find(ctx, initiatorID, targetID); err == nil && b.IsActive() {
		slog.Warn("send friend request blocked initiator", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrBlocked
	}

	// 已是好友（FR-003/004）：好友不重复发申请
	if f, err := s.friendships.Find(ctx, initiatorID, targetID); err == nil && f.IsActive() {
		slog.Warn("send friend request already friends", "initiator", initiatorID.String(), "target", targetID.String())
		return entity.FriendRequest{}, ErrAlreadyFriends
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

// ListPending 列出某用户收到的待处理申请（「新的朋友」入口，FR-011）。
func (s *FriendRequestService) ListPending(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error) {
	slog.Info("list pending friend requests", "target", targetID.String())
	res, err := s.requests.ListPendingByTarget(ctx, targetID)
	if err != nil {
		slog.Error("list pending friend requests failed", "err", err, "target", targetID.String())
		return nil, err
	}
	return res, nil
}

// ListHistory 列出某用户的历史申请（已接受/已拒绝，FR-011）。
func (s *FriendRequestService) ListHistory(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error) {
	slog.Info("list history friend requests", "target", targetID.String())
	res, err := s.requests.ListHistoryByTarget(ctx, targetID)
	if err != nil {
		slog.Error("list history friend requests failed", "err", err, "target", targetID.String())
		return nil, err
	}
	return res, nil
}

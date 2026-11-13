package port

import (
	"context"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/valueobject"
)

// FriendRequestRepository 是好友申请的持久化接口。
type FriendRequestRepository interface {
	// Create 创建一条好友申请（(initiator_id, target_id) 唯一）。
	Create(ctx context.Context, r entity.FriendRequest) (entity.FriendRequest, error)
	// Find 查询指定一对用户之间的申请。
	Find(ctx context.Context, initiatorID, targetID valueobject.UserId) (entity.FriendRequest, error)
	// ListPendingByTarget 列出某用户收到的待处理申请（「新的朋友」入口）。
	ListPendingByTarget(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error)
	// ListHistoryByTarget 列出某用户的历史申请（已接受/已拒绝）。
	ListHistoryByTarget(ctx context.Context, targetID valueobject.UserId) ([]entity.FriendRequest, error)
	// Update 更新申请状态（接受/拒绝）。
	Update(ctx context.Context, r entity.FriendRequest) error
}

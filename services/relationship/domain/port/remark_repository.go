package port

import (
	"context"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/valueobject"
)

// RemarkRepository 是好友备注的持久化接口（单向，仅 owner 可见）。
type RemarkRepository interface {
	// Save 保存/更新备注（upsert，按 (owner_id, friend_id)）。
	Save(ctx context.Context, r entity.Remark) (entity.Remark, error)
	// Find 查询备注。
	Find(ctx context.Context, ownerID, friendID valueobject.UserId) (entity.Remark, error)
}

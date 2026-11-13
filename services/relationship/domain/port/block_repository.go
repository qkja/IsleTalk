package port

import (
	"context"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/valueobject"
)

// BlockRepository 是拉黑关系的持久化接口（单向）。
type BlockRepository interface {
	// Set 设置拉黑状态（blocked=true 拉黑，false 解除）。
	Set(ctx context.Context, ownerID, targetID valueobject.UserId, blocked bool) error
	// Find 查询 owner 是否拉黑了 target。
	Find(ctx context.Context, ownerID, targetID valueobject.UserId) (entity.Block, error)
}

package port

import (
	"context"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/valueobject"
)

// SessionRepository 是设备会话的持久化接口。
type SessionRepository interface {
	// Create 创建会话。
	Create(ctx context.Context, s entity.Session) (entity.Session, error)
	// FindByID 按会话标识查询。
	FindByID(ctx context.Context, id string) (entity.Session, error)
	// ListByUser 列出某用户的全部会话。
	ListByUser(ctx context.Context, userID valueobject.UserId) ([]entity.Session, error)
	// Revoke 撤销会话（MUST 以 MongoDB 为持久事实，撤销后立即拒绝）。
	Revoke(ctx context.Context, id string) error
	// RevokeAllByUser 撤销某用户全部会话（改密/注销全失效）。
	RevokeAllByUser(ctx context.Context, userID valueobject.UserId) error
	// Touch 更新最近活跃时间。
	Touch(ctx context.Context, id string) error
}

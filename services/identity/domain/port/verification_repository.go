package port

import (
	"context"
	"time"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/valueobject"
)

// VerificationRepository 是实名登记信息的持久化接口。
type VerificationRepository interface {
	// Save 保存实名登记（含 AEAD 加密后的身份摘要）。
	Save(ctx context.Context, v entity.Verification) (entity.Verification, error)
	// FindByUser 按用户查询实名登记。
	FindByUser(ctx context.Context, userID valueobject.UserId) (entity.Verification, error)
	// Revoke 撤回实名（删除身份信息，信任降级）。
	Revoke(ctx context.Context, userID valueobject.UserId) error
	// DeleteExpired 删除保留期限到期的实名信息，返回删除条数（FR-025）。
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}

// NameClaimRepository 是公开账号名占用记录的持久化接口。
// 承载"保留期"语义：释放的账号名在保留期内仍占位（FR-015）。
type NameClaimRepository interface {
	// Claim 占用账号名（唯一索引保证原子性）。
	Claim(ctx context.Context, n valueobject.PublicName, userID valueobject.UserId) error
	// Release 释放账号名，进入保留期（保留期内不可被他人占用）。
	Release(ctx context.Context, n valueobject.PublicName) error
	// IsAvailable 判断账号名当前是否可占用（未被占用且不在保留期）。
	IsAvailable(ctx context.Context, n valueobject.PublicName) (bool, error)
}

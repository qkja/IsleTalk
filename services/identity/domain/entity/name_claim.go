package entity

import (
	"time"

	"isletalk/identity/domain/valueobject"
)

// NameClaim 是公开账号名的占用记录（FR-013~015）。
// 承载保留期语义：释放的账号名在保留期内仍占位。
type NameClaim struct {
	name          valueobject.PublicName
	ownerUserID   valueobject.UserId
	releasedAt    *time.Time // nil 表示仍占用
	reservedUntil *time.Time // nil 表示不在保留期
}

// NewNameClaim 占用一个账号名。
func NewNameClaim(name valueobject.PublicName, owner valueobject.UserId) NameClaim {
	return NameClaim{name: name, ownerUserID: owner}
}

// Name 返回账号名。
func (n NameClaim) Name() valueobject.PublicName { return n.name }

// OwnerUserID 返回占用者。
func (n NameClaim) OwnerUserID() valueobject.UserId { return n.ownerUserID }

// Release 释放账号名，进入保留期（默认 90 天，FR-015）。
func (n *NameClaim) Release(now time.Time, retention time.Duration) {
	released := now
	reserved := now.Add(retention)
	n.releasedAt = &released
	n.reservedUntil = &reserved
}

// Available 判断账号名当前是否可被占用。
// 未占用且不在保留期内才可用。
func (n NameClaim) Available(now time.Time) bool {
	if n.releasedAt == nil {
		return false // 仍占用
	}
	if n.reservedUntil == nil {
		return true
	}
	return now.After(*n.reservedUntil)
}

package service

import (
	"context"

	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// NameUniqueness 是账号名唯一性领域服务（FR-013~015）。
// 唯一性由仓储的唯一索引 + name_claims 保留期语义保证；
// 本服务封装"可占用判定"这一领域规则，供 application 用例调用。
type NameUniqueness struct {
	claims port.NameClaimRepository
	users  port.UserRepository
}

func NewNameUniqueness(claims port.NameClaimRepository, users port.UserRepository) *NameUniqueness {
	return &NameUniqueness{claims: claims, users: users}
}

// IsAvailable 判断账号名是否可被占用：未被占用、不在保留期、且无活跃用户持有同名。
func (n *NameUniqueness) IsAvailable(ctx context.Context, name valueobject.PublicName) (bool, error) {
	// 活跃用户已持有同名则不可占用
	if _, err := n.users.FindByPublicName(ctx, name); err == nil {
		return false, nil
	} else if !port.IsNotFound(err) {
		return false, err
	}
	return n.claims.IsAvailable(ctx, name)
}

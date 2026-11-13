package port

import (
	"context"

	"isletalk/relationship/domain/valueobject"
)

// IdentityUser 是 identity-svc 返回的用户资料（relationship 所需的最小字段集）。
type IdentityUser struct {
	ID               valueobject.UserId
	PublicName       string
	AllowStrangerAdd bool
}

// IdentityClient 是 relationship → identity 的服务间调用接口（T006）。
// 用于发申请前读陌生人的 allow_stranger_add（FR-010）。
type IdentityClient interface {
	// GetUser 按 ID 查询用户（读 allow_stranger_add）。
	GetUser(ctx context.Context, userID valueobject.UserId) (IdentityUser, error)
}

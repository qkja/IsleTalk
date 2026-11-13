package port

import (
	"context"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/valueobject"
)

// UserRepository 是用户账号的持久化接口。领域层只依赖此接口，
// 由 infrastructure 层实现并注入（宪法原则 V：依赖方向单向向内）。
type UserRepository interface {
	// Create 创建账号，返回创建的实体。
	Create(ctx context.Context, u entity.User) (entity.User, error)
	// FindByID 按不可变内部 ID 查询。
	FindByID(ctx context.Context, id valueobject.UserId) (entity.User, error)
	// FindByPhoneHash 按手机号哈希查询（注册去重、登录定位）。
	FindByPhoneHash(ctx context.Context, h valueobject.PhoneHash) (entity.User, error)
	// FindByPublicName 按公开账号名精确查询。
	FindByPublicName(ctx context.Context, n valueobject.PublicName) (entity.User, error)
	// Update 保存用户变更（昵称、头像、公开账号名、状态等）。
	Update(ctx context.Context, u entity.User) error
}

// ErrNotFound 由仓储实现返回，表示实体不存在。
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "entity not found" }

// IsNotFound 判断错误是否为"实体不存在"。
func IsNotFound(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

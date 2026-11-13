package port

import (
	"context"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/valueobject"
)

// FriendshipRepository 是好友关系的持久化接口。
// 一对好友 = 两条文档（A→B 与 B→A），保证双向一致（SC-002）。
type FriendshipRepository interface {
	// Create 创建一条好友关系（user_id → friend_id）。
	Create(ctx context.Context, f entity.Friendship) (entity.Friendship, error)
	// Find 查询 user_id 与 friend_id 之间是否为好友。
	Find(ctx context.Context, userID, friendID valueobject.UserId) (entity.Friendship, error)
	// ListByUser 列出某用户的全部好友。
	ListByUser(ctx context.Context, userID valueobject.UserId) ([]entity.Friendship, error)
	// Delete 删除一条好友关系（双向解除时删除两条）。
	Delete(ctx context.Context, userID, friendID valueobject.UserId) error
}

// ErrNotFound 表示实体不存在。
var ErrNotFound = errNotFound{}

type errNotFound struct{}

func (errNotFound) Error() string { return "entity not found" }

// IsNotFound 判断错误是否为"实体不存在"。
func IsNotFound(err error) bool {
	_, ok := err.(errNotFound)
	return ok
}

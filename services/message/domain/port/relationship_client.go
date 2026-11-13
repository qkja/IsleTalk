package port

import (
	"context"

	"isletalk/message/domain/valueobject"
)

// RelationshipClient 是 message → relationship 的服务间调用接口（T005/T008）。
// 发消息前校验双方互为好友且未被拉黑（FR-011）。
type RelationshipClient interface {
	// CanMessage 判定 sender 是否可向 recipient 发消息。
	// 返回 (allowed, blocked, err)：blocked=true 表示被拉黑，allowed=false 且 blocked=false 表示非好友。
	CanMessage(ctx context.Context, senderID, recipientID valueobject.UserId) (bool, bool, error)
}

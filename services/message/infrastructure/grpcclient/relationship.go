package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	identityv1 "isletalk/message/api/identity/v1"
	relationshipv1 "isletalk/message/api/relationship/v1"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
)

// RelationshipClient 封装 message → relationship 的 CanMessage 调用（T008）。
// 生产 MUST 100% mTLS（宪法 E 组）；当前骨架用 insecure，证书体系属 002 待办，装配时替换。
type RelationshipClient struct {
	client relationshipv1.RelationshipServiceClient
}

func NewRelationshipClient(addr string) (*RelationshipClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &RelationshipClient{client: relationshipv1.NewRelationshipServiceClient(conn)}, nil
}

// CanMessage 判定 sender 是否可向 recipient 发消息（FR-011）。
// 返回 (allowed, blocked, err)：blocked=true 表示被拉黑，allowed=false 且 blocked=false 表示非好友。
func (c *RelationshipClient) CanMessage(ctx context.Context, senderID, recipientID valueobject.UserId) (bool, bool, error) {
	resp, err := c.client.CanMessage(ctx, &relationshipv1.CanMessageRequest{
		SenderId:    &identityv1.UserId{Value: senderID.String()},
		RecipientId: &identityv1.UserId{Value: recipientID.String()},
	})
	if err != nil {
		return false, false, err
	}
	return resp.GetAllowed(), resp.GetStatus() == relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_BLOCKED, nil
}

var _ port.RelationshipClient = (*RelationshipClient)(nil)

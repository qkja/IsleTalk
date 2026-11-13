package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	identityv1 "isletalk/relationship/api/identity/v1"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// IdentityClient 封装 relationship → identity 的 gRPC 调用（T009）。
// 生产 MUST 100% mTLS（宪法 E 组）；当前骨架用 insecure，证书体系属 002 待办，装配时替换。
type IdentityClient struct {
	client identityv1.IdentityServiceClient
}

func NewIdentityClient(addr string) (*IdentityClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &IdentityClient{client: identityv1.NewIdentityServiceClient(conn)}, nil
}

// GetUser 按 ID 查询用户，读取 allow_stranger_add（FR-010）。
func (c *IdentityClient) GetUser(ctx context.Context, userID valueobject.UserId) (port.IdentityUser, error) {
	resp, err := c.client.GetUser(ctx, &identityv1.GetUserRequest{Id: &identityv1.UserId{Value: userID.String()}})
	if err != nil {
		return port.IdentityUser{}, err
	}
	return port.IdentityUser{
		ID:               userID,
		PublicName:       resp.User.GetPublicName().GetValue(),
		AllowStrangerAdd: resp.User.GetAllowStrangerAdd(),
	}, nil
}

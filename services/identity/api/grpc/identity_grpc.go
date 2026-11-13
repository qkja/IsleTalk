package grpc

import (
	"context"

	"isletalk/identity/api/identity/v1"
	"isletalk/identity/application/service"
	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// IdentityGRPCServer 实现 identity.v1.IdentityServiceServer（T020/T026/T030/T040）。
// 对外服务能力：GetUser / SearchByPublicName / CheckUser / ListSessions / RevokeSession。
type IdentityGRPCServer struct {
	identityv1.UnimplementedIdentityServiceServer
	users    port.UserRepository
	sessions *service.SessionService
}

func NewIdentityGRPCServer(users port.UserRepository, sessions *service.SessionService) *IdentityGRPCServer {
	return &IdentityGRPCServer{users: users, sessions: sessions}
}

// entityToProto 将领域实体映射为 pb 消息（MUST NOT 泄露手机号/身份信息，FR-029）。
func entityToProto(u entity.User) *identityv1.User {
	return &identityv1.User{
		Id:               &identityv1.UserId{Value: u.ID().String()},
		PublicName:       &identityv1.PublicName{Value: u.PublicName().String()},
		Nickname:         u.Nickname(),
		AvatarUrl:        u.AvatarURL(),
		Signature:        u.Signature(),
		Visibility:       string(u.Visibility()),
		AllowStrangerAdd: u.AllowStrangerAdd(),
	}
}

func (s *IdentityGRPCServer) GetUser(ctx context.Context, req *identityv1.GetUserRequest) (*identityv1.GetUserResponse, error) {
	id, err := valueobject.NewUserId(req.GetId().GetValue())
	if err != nil {
		return nil, err
	}
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &identityv1.GetUserResponse{User: entityToProto(u)}, nil
}

func (s *IdentityGRPCServer) SearchByPublicName(ctx context.Context, req *identityv1.SearchByPublicNameRequest) (*identityv1.SearchByPublicNameResponse, error) {
	name, err := valueobject.NewPublicName(req.GetPublicName())
	if err != nil {
		return nil, err
	}
	u, err := s.users.FindByPublicName(ctx, name)
	if err != nil {
		if port.IsNotFound(err) {
			return &identityv1.SearchByPublicNameResponse{}, nil // 未命中返回空
		}
		return nil, err
	}
	return &identityv1.SearchByPublicNameResponse{User: entityToProto(u)}, nil
}

func (s *IdentityGRPCServer) CheckUser(ctx context.Context, req *identityv1.CheckUserRequest) (*identityv1.CheckUserResponse, error) {
	id, err := valueobject.NewUserId(req.GetId().GetValue())
	if err != nil {
		return nil, err
	}
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		if port.IsNotFound(err) {
			return &identityv1.CheckUserResponse{Active: false}, nil // 不存在/注销/封禁统一为不可用
		}
		return nil, err
	}
	return &identityv1.CheckUserResponse{Active: u.IsActive()}, nil
}

func (s *IdentityGRPCServer) ListSessions(ctx context.Context, req *identityv1.ListSessionsRequest) (*identityv1.ListSessionsResponse, error) {
	id, err := valueobject.NewUserId(req.GetUserId().GetValue())
	if err != nil {
		return nil, err
	}
	sessions, err := s.sessions.List(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]*identityv1.Session, 0, len(sessions))
	for _, sess := range sessions {
		out = append(out, &identityv1.Session{
			SessionId: sess.ID(),
			DeviceId:  sess.DeviceID(),
			Revoked:   sess.Revoked(),
		})
	}
	return &identityv1.ListSessionsResponse{Sessions: out}, nil
}

func (s *IdentityGRPCServer) RevokeSession(ctx context.Context, req *identityv1.RevokeSessionRequest) (*identityv1.RevokeSessionResponse, error) {
	if err := s.sessions.Revoke(ctx, req.GetSessionId()); err != nil {
		return nil, err
	}
	return &identityv1.RevokeSessionResponse{}, nil
}

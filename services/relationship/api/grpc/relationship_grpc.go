package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	relationshipv1 "isletalk/relationship/api/relationship/v1"
	"isletalk/relationship/application/service"
	"isletalk/relationship/domain/valueobject"
)

// RelationshipGRPCServer 实现 relationship.v1.RelationshipService（4 RPC）。
type RelationshipGRPCServer struct {
	relationshipv1.UnimplementedRelationshipServiceServer
	friendRequest *service.FriendRequestService
	respond       *service.RespondService
	block         *service.BlockService
	canMessage    *service.CanMessageService
}

func NewRelationshipGRPCServer(
	friendRequest *service.FriendRequestService,
	respond *service.RespondService,
	block *service.BlockService,
	canMessage *service.CanMessageService,
) *RelationshipGRPCServer {
	return &RelationshipGRPCServer{
		friendRequest: friendRequest,
		respond:       respond,
		block:         block,
		canMessage:    canMessage,
	}
}

func (s *RelationshipGRPCServer) SendFriendRequest(ctx context.Context, req *relationshipv1.FriendRequest) (*relationshipv1.SendFriendRequestResponse, error) {
	initiator, err := valueobject.NewUserId(req.GetInitiatorId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	target, err := valueobject.NewUserId(req.GetTargetId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err := s.friendRequest.Send(ctx, initiator, target, req.GetRequestMessage()); err != nil {
		return nil, grpcErr(err)
	}
	return &relationshipv1.SendFriendRequestResponse{Status: relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_PENDING}, nil
}

func (s *RelationshipGRPCServer) RespondToFriendRequest(ctx context.Context, req *relationshipv1.RespondFriendRequest) (*relationshipv1.RespondFriendRequestResponse, error) {
	initiator, err := valueobject.NewUserId(req.GetInitiatorId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	target, err := valueobject.NewUserId(req.GetTargetId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	got, err := s.respond.Respond(ctx, initiator, target, req.GetAccept())
	if err != nil {
		return nil, grpcErr(err)
	}
	st := relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_REJECTED
	if req.GetAccept() {
		st = relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED
	}
	_ = got
	return &relationshipv1.RespondFriendRequestResponse{Status: st}, nil
}

func (s *RelationshipGRPCServer) SetBlock(ctx context.Context, req *relationshipv1.SetBlockRequest) (*relationshipv1.SetBlockResponse, error) {
	owner, err := valueobject.NewUserId(req.GetOwnerId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	target, err := valueobject.NewUserId(req.GetTargetId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.block.SetBlock(ctx, owner, target, req.GetBlocked()); err != nil {
		return nil, grpcErr(err)
	}
	return &relationshipv1.SetBlockResponse{}, nil
}

func (s *RelationshipGRPCServer) CanMessage(ctx context.Context, req *relationshipv1.CanMessageRequest) (*relationshipv1.CanMessageResponse, error) {
	sender, err := valueobject.NewUserId(req.GetSenderId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	recipient, err := valueobject.NewUserId(req.GetRecipientId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	allowed, blocked, err := s.canMessage.CanMessage(ctx, sender, recipient)
	if err != nil {
		return nil, grpcErr(err)
	}
	st := relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_PENDING // 非好友
	if blocked {
		st = relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_BLOCKED
	}
	if allowed {
		st = relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED
	}
	return &relationshipv1.CanMessageResponse{Allowed: allowed, Status: st}, nil
}

// grpcErr 映射用例错误到 gRPC status code（错误码注册表见 002）。
func grpcErr(err error) error {
	switch {
	case errors.Is(err, service.ErrSelfRequest),
		errors.Is(err, service.ErrRequestNotFound):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrStrangerAddDisabled),
		errors.Is(err, service.ErrBlocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrRequestExists),
		errors.Is(err, service.ErrAlreadyFriends):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

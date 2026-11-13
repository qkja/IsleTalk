package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "isletalk/relationship/api/identity/v1"
	relationshipv1 "isletalk/relationship/api/relationship/v1"
	"isletalk/relationship/application/service"
)

func TestGrpcErr(t *testing.T) {
	cases := []struct {
		err  error
		want codes.Code
	}{
		{service.ErrSelfRequest, codes.InvalidArgument},
		{service.ErrRequestNotFound, codes.InvalidArgument},
		{service.ErrStrangerAddDisabled, codes.PermissionDenied},
		{service.ErrBlocked, codes.PermissionDenied},
		{service.ErrRequestExists, codes.AlreadyExists},
		{service.ErrAlreadyFriends, codes.AlreadyExists},
	}
	for _, c := range cases {
		got := status.Code(grpcErr(c.err))
		if got != c.want {
			t.Fatalf("grpcErr(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestHandlerInvalidUserID(t *testing.T) {
	srv := &RelationshipGRPCServer{}

	if _, err := srv.SendFriendRequest(context.Background(), &relationshipv1.FriendRequest{
		InitiatorId: &identityv1.UserId{Value: ""},
		TargetId:    &identityv1.UserId{Value: "b"},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty initiator should be InvalidArgument, got %v", err)
	}

	if _, err := srv.SetBlock(context.Background(), &relationshipv1.SetBlockRequest{
		OwnerId:  &identityv1.UserId{Value: ""},
		TargetId: &identityv1.UserId{Value: "b"},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty owner should be InvalidArgument, got %v", err)
	}

	if _, err := srv.CanMessage(context.Background(), &relationshipv1.CanMessageRequest{
		SenderId:    &identityv1.UserId{Value: ""},
		RecipientId: &identityv1.UserId{Value: "b"},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty sender should be InvalidArgument, got %v", err)
	}
}

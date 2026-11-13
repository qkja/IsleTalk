package contract

import (
	"context"
	"testing"

	grpcserver "isletalk/relationship/api/grpc"
	identityv1 "isletalk/relationship/api/identity/v1"
	relationshipv1 "isletalk/relationship/api/relationship/v1"
	"isletalk/relationship/application/service"
	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// 契约测试（T032）：基于 relationship.proto 的 4 RPC 全覆盖。

type memFriendships struct{ m map[string]entity.Friendship }

func (r *memFriendships) Create(_ context.Context, f entity.Friendship) (entity.Friendship, error) {
	r.m[f.UserID().String()+":"+f.FriendID().String()] = f
	return f, nil
}
func (r *memFriendships) Find(_ context.Context, a, b valueobject.UserId) (entity.Friendship, error) {
	if f, ok := r.m[a.String()+":"+b.String()]; ok {
		return f, nil
	}
	return entity.Friendship{}, port.ErrNotFound
}
func (r *memFriendships) ListByUser(context.Context, valueobject.UserId) ([]entity.Friendship, error) {
	return nil, nil
}
func (r *memFriendships) Delete(context.Context, valueobject.UserId, valueobject.UserId) error {
	return nil
}

type memRequests struct {
	m map[string]entity.FriendRequest
}

func (r *memRequests) Create(_ context.Context, req entity.FriendRequest) (entity.FriendRequest, error) {
	r.m[req.InitiatorID().String()+":"+req.TargetID().String()] = req
	return req, nil
}
func (r *memRequests) Find(_ context.Context, a, b valueobject.UserId) (entity.FriendRequest, error) {
	if req, ok := r.m[a.String()+":"+b.String()]; ok {
		return req, nil
	}
	return entity.FriendRequest{}, port.ErrNotFound
}
func (r *memRequests) ListPendingByTarget(context.Context, valueobject.UserId) ([]entity.FriendRequest, error) {
	return nil, nil
}
func (r *memRequests) ListHistoryByTarget(context.Context, valueobject.UserId) ([]entity.FriendRequest, error) {
	return nil, nil
}
func (r *memRequests) Update(_ context.Context, req entity.FriendRequest) error {
	r.m[req.InitiatorID().String()+":"+req.TargetID().String()] = req
	return nil
}

type memBlocks struct{ m map[string]entity.Block }

func (r *memBlocks) Set(_ context.Context, o, t valueobject.UserId, blocked bool) error {
	if blocked {
		r.m[o.String()+":"+t.String()] = entity.NewBlock(o, t)
	}
	return nil
}
func (r *memBlocks) Find(_ context.Context, o, t valueobject.UserId) (entity.Block, error) {
	if b, ok := r.m[o.String()+":"+t.String()]; ok {
		return b, nil
	}
	return entity.Block{}, port.ErrNotFound
}

type memIdentity struct{}

func (memIdentity) GetUser(_ context.Context, uid valueobject.UserId) (port.IdentityUser, error) {
	return port.IdentityUser{ID: uid, PublicName: "u", AllowStrangerAdd: true}, nil
}

func newServer() *grpcserver.RelationshipGRPCServer {
	f := &memFriendships{m: map[string]entity.Friendship{}}
	r := &memRequests{m: map[string]entity.FriendRequest{}}
	b := &memBlocks{m: map[string]entity.Block{}}
	return grpcserver.NewRelationshipGRPCServer(
		service.NewFriendRequestService(r, memIdentity{}),
		service.NewRespondService(r, f),
		service.NewBlockService(b),
		service.NewCanMessageService(f, b),
	)
}

func uid(s string) *identityv1.UserId { return &identityv1.UserId{Value: s} }

func TestContractSendFriendRequest(t *testing.T) {
	srv := newServer()
	_, err := srv.SendFriendRequest(context.Background(), &relationshipv1.FriendRequest{
		InitiatorId: uid("a"), TargetId: uid("b"), RequestMessage: "hi",
	})
	if err != nil {
		t.Fatalf("SendFriendRequest: %v", err)
	}
}

func TestContractRespondAccept(t *testing.T) {
	srv := newServer()
	if _, err := srv.SendFriendRequest(context.Background(), &relationshipv1.FriendRequest{
		InitiatorId: uid("a"), TargetId: uid("b"), RequestMessage: "hi",
	}); err != nil {
		t.Fatal(err)
	}
	resp, err := srv.RespondToFriendRequest(context.Background(), &relationshipv1.RespondFriendRequest{
		InitiatorId: uid("a"), TargetId: uid("b"), Accept: true,
	})
	if err != nil {
		t.Fatalf("RespondToFriendRequest: %v", err)
	}
	if resp.Status != relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED {
		t.Fatalf("status = %v, want ACCEPTED", resp.Status)
	}
}

func TestContractSetBlock(t *testing.T) {
	srv := newServer()
	if _, err := srv.SetBlock(context.Background(), &relationshipv1.SetBlockRequest{
		OwnerId: uid("a"), TargetId: uid("b"), Blocked: true,
	}); err != nil {
		t.Fatalf("SetBlock: %v", err)
	}
}

func TestContractCanMessage(t *testing.T) {
	srv := newServer()
	// 未加好友 → 拒绝
	resp, err := srv.CanMessage(context.Background(), &relationshipv1.CanMessageRequest{
		SenderId: uid("a"), RecipientId: uid("b"),
	})
	if err != nil {
		t.Fatalf("CanMessage: %v", err)
	}
	if resp.Allowed {
		t.Fatal("non-friends should not be allowed to message")
	}
}

package grpcclient

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	relationshipv1 "isletalk/message/api/relationship/v1"
	"isletalk/message/domain/valueobject"
)

type fakeRelServer struct {
	relationshipv1.UnimplementedRelationshipServiceServer
	allowed bool
	blocked bool
}

func (s *fakeRelServer) CanMessage(_ context.Context, _ *relationshipv1.CanMessageRequest) (*relationshipv1.CanMessageResponse, error) {
	st := relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_PENDING
	if s.blocked {
		st = relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_BLOCKED
	} else if s.allowed {
		st = relationshipv1.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED
	}
	return &relationshipv1.CanMessageResponse{Allowed: s.allowed, Status: st}, nil
}

func newBufConnClient(t *testing.T, allowed, blocked bool) *RelationshipClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	relationshipv1.RegisterRelationshipServiceServer(srv, &fakeRelServer{allowed: allowed, blocked: blocked})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &RelationshipClient{client: relationshipv1.NewRelationshipServiceClient(conn)}
}

func TestCanMessageAllowed(t *testing.T) {
	a, _ := valueobject.NewUserId("a")
	b, _ := valueobject.NewUserId("b")
	c := newBufConnClient(t, true, false)
	allowed, blocked, err := c.CanMessage(context.Background(), a, b)
	if err != nil || !allowed || blocked {
		t.Fatalf("allowed=%v blocked=%v err=%v, want allowed=true blocked=false", allowed, blocked, err)
	}
}

func TestCanMessageBlocked(t *testing.T) {
	a, _ := valueobject.NewUserId("a")
	b, _ := valueobject.NewUserId("b")
	c := newBufConnClient(t, false, true)
	allowed, blocked, err := c.CanMessage(context.Background(), a, b)
	if err != nil || allowed || !blocked {
		t.Fatalf("allowed=%v blocked=%v err=%v, want allowed=false blocked=true", allowed, blocked, err)
	}
}

func TestNewRelationshipClient(t *testing.T) {
	// grpc.Dial 是惰性连接，不实际拨号，故返回非 nil 客户端不报错
	c, err := NewRelationshipClient("localhost:1")
	if err != nil {
		t.Fatalf("NewRelationshipClient: %v", err)
	}
	if c == nil || c.client == nil {
		t.Fatal("client should not be nil")
	}
}

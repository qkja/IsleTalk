package contract

import (
	"context"
	"testing"

	grpcserver "isletalk/identity/api/grpc"
	identityv1 "isletalk/identity/api/identity/v1"
	"isletalk/identity/application/service"
	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// 契约测试（T041）：基于 identity.proto 的 5 个 RPC 全覆盖。
// 提供方契约测试，验证 gRPC handler 与 proto 契约一致。

type users struct{ m map[string]entity.User }

func (u *users) Create(_ context.Context, x entity.User) (entity.User, error) { return x, nil }
func (u *users) FindByID(_ context.Context, id valueobject.UserId) (entity.User, error) {
	if x, ok := u.m[id.String()]; ok {
		return x, nil
	}
	return entity.User{}, port.ErrNotFound
}
func (u *users) FindByPhoneHash(_ context.Context, _ valueobject.PhoneHash) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (u *users) FindByPublicName(_ context.Context, n valueobject.PublicName) (entity.User, error) {
	for _, x := range u.m {
		if x.PublicName() == n {
			return x, nil
		}
	}
	return entity.User{}, port.ErrNotFound
}
func (u *users) Update(_ context.Context, x entity.User) error { return nil }

type sessions struct{ m map[string]entity.Session }

func (s *sessions) Create(_ context.Context, x entity.Session) (entity.Session, error) { return x, nil }
func (s *sessions) FindByID(_ context.Context, id string) (entity.Session, error) {
	if x, ok := s.m[id]; ok {
		return x, nil
	}
	return entity.Session{}, port.ErrNotFound
}
func (s *sessions) ListByUser(_ context.Context, uid valueobject.UserId) ([]entity.Session, error) {
	out := []entity.Session{}
	for _, x := range s.m {
		if x.UserID() == uid {
			out = append(out, x)
		}
	}
	return out, nil
}
func (s *sessions) Revoke(_ context.Context, id string) error {
	if x, ok := s.m[id]; ok {
		x.Revoke()
		s.m[id] = x
	}
	return nil
}
func (s *sessions) RevokeAllByUser(_ context.Context, _ valueobject.UserId) error { return nil }
func (s *sessions) Touch(_ context.Context, _ string) error                       { return nil }

func mustID(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func mustName(t *testing.T, s string) valueobject.PublicName {
	t.Helper()
	v, err := valueobject.NewPublicName(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func testUser(t *testing.T, id, name string) entity.User {
	t.Helper()
	phone, _ := valueobject.NewPhoneHash("phone:" + id)
	u, err := entity.NewUser(mustID(t, id), mustName(t, name), phone, "zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func newServer(t *testing.T) (*grpcserver.IdentityGRPCServer, *users, *sessions) {
	u := &users{m: map[string]entity.User{}}
	s := &sessions{m: map[string]entity.Session{}}
	svc := service.NewSessionService(s)
	return grpcserver.NewIdentityGRPCServer(u, svc), u, s
}

func TestContractGetUser(t *testing.T) {
	srv, u, _ := newServer(t)
	u.m["u-1"] = testUser(t, "u-1", "alice")

	resp, err := srv.GetUser(context.Background(), &identityv1.GetUserRequest{Id: &identityv1.UserId{Value: "u-1"}})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if resp.User.GetPublicName().GetValue() != "alice" {
		t.Fatalf("public name = %q", resp.User.GetPublicName().GetValue())
	}
}

func TestContractSearchByPublicName(t *testing.T) {
	srv, u, _ := newServer(t)
	u.m["u-1"] = testUser(t, "u-1", "alice")

	resp, err := srv.SearchByPublicName(context.Background(), &identityv1.SearchByPublicNameRequest{PublicName: "alice"})
	if err != nil {
		t.Fatalf("SearchByPublicName: %v", err)
	}
	if resp.User == nil || resp.User.GetId().GetValue() != "u-1" {
		t.Fatal("search should return u-1")
	}

	// 未命中返回空
	resp2, _ := srv.SearchByPublicName(context.Background(), &identityv1.SearchByPublicNameRequest{PublicName: "nobody"})
	if resp2.User != nil {
		t.Fatal("not found should return nil user")
	}
}

func TestContractCheckUser(t *testing.T) {
	srv, u, _ := newServer(t)
	u.m["u-1"] = testUser(t, "u-1", "alice")

	resp, _ := srv.CheckUser(context.Background(), &identityv1.CheckUserRequest{Id: &identityv1.UserId{Value: "u-1"}})
	if !resp.Active {
		t.Fatal("active user should return true")
	}
	resp2, _ := srv.CheckUser(context.Background(), &identityv1.CheckUserRequest{Id: &identityv1.UserId{Value: "u-x"}})
	if resp2.Active {
		t.Fatal("nonexistent user should return false")
	}
}

func TestContractListSessions(t *testing.T) {
	srv, _, s := newServer(t)
	uid := mustID(t, "u-1")
	s.m["s-1"] = entity.NewSession("s-1", uid, "d1")
	s.m["s-2"] = entity.NewSession("s-2", uid, "d2")

	resp, err := srv.ListSessions(context.Background(), &identityv1.ListSessionsRequest{UserId: &identityv1.UserId{Value: "u-1"}})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(resp.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(resp.Sessions))
	}
}

func TestContractRevokeSession(t *testing.T) {
	srv, _, s := newServer(t)
	uid := mustID(t, "u-1")
	s.m["s-1"] = entity.NewSession("s-1", uid, "d1")

	if _, err := srv.RevokeSession(context.Background(), &identityv1.RevokeSessionRequest{SessionId: "s-1"}); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if !s.m["s-1"].Revoked() {
		t.Fatal("session should be revoked")
	}
}
